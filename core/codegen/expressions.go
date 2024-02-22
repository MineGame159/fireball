package codegen

import (
	"fireball/core/abi"
	"fireball/core/ast"
	"fireball/core/common"
	"fireball/core/ir"
	"fireball/core/scanner"
	"log"
	"strconv"
	"strings"
)

func (c *codegen) VisitParen(expr *ast.Paren) {
	c.acceptExpr(expr.Expr)
}

func (c *codegen) VisitLiteral(expr *ast.Literal) {
	// Convert fireball constant into a LLVM IR constant
	var value ir.Value
	type_ := c.types.get(expr.Result().Type)

	switch expr.Token().Kind {
	case scanner.Nil:
		value = ir.Null

	case scanner.True:
		value = &ir.IntConst{Typ: type_, Value: ir.Unsigned(1)}

	case scanner.False:
		value = &ir.IntConst{Typ: type_, Value: ir.Unsigned(0)}

	case scanner.Number:
		raw := expr.String()
		last := raw[len(raw)-1]

		if last == 'f' || last == 'F' {
			v, _ := strconv.ParseFloat(raw[:len(raw)-1], 32)
			value = &ir.FloatConst{Typ: type_, Value: v}
		} else if strings.ContainsRune(raw, '.') {
			v, _ := strconv.ParseFloat(raw, 64)
			value = &ir.FloatConst{Typ: type_, Value: v}
		} else {
			t, _ := ast.As[*ast.Primitive](expr.Result().Type)

			if ast.IsSigned(t.Kind) {
				v, _ := strconv.ParseInt(raw, 10, 64)
				value = &ir.IntConst{Typ: type_, Value: ir.Signed(v)}
			} else {
				v, _ := strconv.ParseUint(raw, 10, 64)
				value = &ir.IntConst{Typ: type_, Value: ir.Unsigned(v)}
			}
		}

	case scanner.Hex:
		v, _ := strconv.ParseUint(expr.String()[2:], 16, 64)
		value = &ir.IntConst{Typ: type_, Value: ir.Unsigned(v)}

	case scanner.Binary:
		v, _ := strconv.ParseUint(expr.String()[2:], 2, 64)
		value = &ir.IntConst{Typ: type_, Value: ir.Unsigned(v)}

	case scanner.Character:
		char := expr.String()[1 : len(expr.String())-1]
		var number uint8

		switch char {
		case "'":
			number = '\''
		case "\\0":
			number = '\000'

		case "\\n":
			number = '\n'
		case "\\r":
			number = '\r'
		case "\\t":
			number = '\t'

		default:
			number = char[0]
		}

		value = &ir.IntConst{Typ: type_, Value: ir.Unsigned(uint64(number))}

	case scanner.String:
		value = c.module.Constant("", convertString(expr.String()[1:len(expr.String())-1]))

	default:
		panic("codegen.VisitLiteral() - Invalid literal kind")
	}

	// Emit
	c.exprResult = exprValue{
		v: value,
	}
}

func convertString(s string) *ir.StringConst {
	b := &ir.StringConst{
		Length: 0,
		Value:  make([]byte, 0, len(s)),
	}

	for i := 0; i < len(s); i++ {
		ch := s[i]

		switch ch {
		case '\\':
			if i+1 >= len(s) {
				break
			}

			switch s[i+1] {
			case '0':
				b.Value = append(b.Value, '\\')
				b.Value = append(b.Value, '0')
				b.Value = append(b.Value, '0')

			case 'n':
				b.Value = append(b.Value, '\\')
				b.Value = append(b.Value, '0')
				b.Value = append(b.Value, 'A')

			case 'r':
				b.Value = append(b.Value, '\\')
				b.Value = append(b.Value, '0')
				b.Value = append(b.Value, 'D')

			case 't':
				b.Value = append(b.Value, '\\')
				b.Value = append(b.Value, '0')
				b.Value = append(b.Value, '9')
			}

			i++

		default:
			b.Value = append(b.Value, ch)
		}

		b.Length++
	}

	b.Length++

	b.Value = append(b.Value, '\\')
	b.Value = append(b.Value, '0')
	b.Value = append(b.Value, '0')

	return b
}

func (c *codegen) VisitStructInitializer(expr *ast.StructInitializer) {
	// Value
	struct_, _ := ast.As[ast.StructType](expr.Type)
	fields, _ := abi.GetStructLayout(struct_.Underlying()).Fields(abi.GetTargetAbi(), struct_)

	type_ := c.types.get(struct_)

	var result ir.Value = &ir.ZeroInitConst{Typ: type_}

	for _, initField := range expr.Fields {
		field, i := getField(fields, initField.Name)
		element := c.implicitCastLoadExpr(field.Type(), initField.Value)

		r := c.block.Add(&ir.InsertValueInst{
			Value:   result,
			Element: element.v,
			Indices: []uint32{uint32(i)},
		})

		c.setLocationMeta(r, initField)
		result = r
	}

	c.exprResult = exprValue{v: result}

	// Malloc
	if expr.New {
		mallocFunc := c.resolver.GetFunction("malloc")
		malloc := c.getFunction(mallocFunc)

		pointer := c.block.Add(&ir.CallInst{
			Callee: malloc.v,
			Args: []ir.Value{&ir.IntConst{
				Typ:   c.types.get(mallocFunc.ParameterIndex(0).Type),
				Value: ir.Unsigned(uint64(abi.GetTargetAbi().Size(struct_))),
			}},
		})

		c.block.Add(&ir.StoreInst{
			Pointer: pointer,
			Value:   result,
			Align:   abi.GetTargetAbi().Align(struct_),
		})

		c.exprResult = exprValue{v: pointer}
	}
}

func (c *codegen) VisitArrayInitializer(expr *ast.ArrayInitializer) {
	baseType := expr.Result().Type.(*ast.Array).Base
	type_ := c.types.get(expr.Result().Type)

	var result ir.Value = &ir.ZeroInitConst{Typ: type_}

	for i, value := range expr.Values {
		element := c.implicitCastLoadExpr(baseType, value)

		r := c.block.Add(&ir.InsertValueInst{
			Value:   result,
			Element: element.v,
			Indices: []uint32{uint32(i)},
		})

		c.setLocationMeta(r, value)

		result = r
	}

	c.exprResult = exprValue{v: result}
}

func (c *codegen) VisitAllocateArray(expr *ast.AllocateArray) {
	mallocFunc := c.resolver.GetFunction("malloc")
	malloc := c.getFunction(mallocFunc)

	count := c.loadExpr(expr.Count)
	count = c.cast(count, expr.Count.Result().Type, mallocFunc.ParameterIndex(0).Type, expr)

	pointer := c.block.Add(&ir.CallInst{
		Callee: malloc.v,
		Args: []ir.Value{&ir.IntConst{
			Typ:   c.types.get(mallocFunc.ParameterIndex(0).Type),
			Value: ir.Unsigned(uint64(abi.GetTargetAbi().Size(expr.Type))),
		}},
	})

	c.exprResult = exprValue{v: pointer}
}

func (c *codegen) VisitUnary(expr *ast.Unary) {
	value := c.acceptExpr(expr.Value)
	var result ir.Value

	if expr.Prefix {
		// Prefix
		switch expr.Operator.Token().Kind {
		case scanner.Bang:
			result = c.block.Add(&ir.XorInst{
				Left:  ir.True,
				Right: c.load(value, expr.Value.Result().Type).v,
			})

		case scanner.Minus:
			if v, ok := ast.As[*ast.Primitive](expr.Value.Result().Type); ok {
				value := c.load(value, expr.Value.Result().Type)

				if ast.IsFloating(v.Kind) {
					// floating
					result = c.block.Add(&ir.FNegInst{Value: value.v})
				} else {
					// signed
					result = c.block.Add(&ir.SubInst{
						Left: &ir.IntConst{
							Typ:   c.types.get(expr.Value.Result().Type),
							Value: ir.Unsigned(0),
						},
						Right: value.v,
					})
				}
			}

		case scanner.Ampersand, scanner.FuncPtr:
			c.exprResult = exprValue{
				v:           value.v,
				addressable: false,
			}

			return

		case scanner.Star:
			result = c.load(value, expr.Value.Result().Type).v

			if _, ok := expr.Parent().(*ast.Assignment); !ok {
				result = c.block.Add(&ir.LoadInst{
					Typ:     result.Type().(*ir.PointerType).Pointee,
					Pointer: result,
					Align:   abi.GetTargetAbi().Align(expr.Result().Type),
				})
			}

		case scanner.PlusPlus, scanner.MinusMinus:
			var one ir.Value

			if isFloating(expr.Value.Result().Type) {
				one = &ir.FloatConst{
					Typ:   c.types.get(expr.Value.Result().Type),
					Value: 1,
				}
			} else {
				one = &ir.IntConst{
					Typ:   c.types.get(expr.Value.Result().Type),
					Value: ir.Unsigned(1),
				}
			}

			newValue := c.binary(
				expr.Operator,
				value,
				exprValue{v: one},
				expr.Value.Result().Type,
			)

			c.block.Add(&ir.StoreInst{
				Pointer: value.v,
				Value:   newValue.v,
				Align:   abi.GetTargetAbi().Align(expr.Value.Result().Type),
			})

			result = newValue.v

		default:
			panic("codegen.VisitUnary() - Invalid unary prefix operator")
		}

		if metaValue, ok := result.(ir.MetaValue); ok {
			c.setLocationMeta(metaValue, expr)
		}
	} else {
		// Postfix
		switch expr.Operator.Token().Kind {
		case scanner.PlusPlus, scanner.MinusMinus:
			prevValue := c.load(value, expr.Value.Result().Type)

			var one ir.Value

			if isFloating(expr.Value.Result().Type) {
				one = &ir.FloatConst{
					Typ:   c.types.get(expr.Value.Result().Type),
					Value: 1,
				}
			} else {
				one = &ir.IntConst{
					Typ:   c.types.get(expr.Value.Result().Type),
					Value: ir.Unsigned(1),
				}
			}

			newValue := c.binary(
				expr.Operator,
				prevValue,
				exprValue{v: one},
				expr.Value.Result().Type,
			)

			c.block.Add(&ir.StoreInst{
				Pointer: value.v,
				Value:   newValue.v,
				Align:   abi.GetTargetAbi().Align(expr.Value.Result().Type),
			})

			result = prevValue.v

		default:
			panic("codegen.VisitUnary() - Invalid unary prefix operator")
		}
	}

	c.exprResult = exprValue{v: result}
}

func (c *codegen) VisitBinary(expr *ast.Binary) {
	c.exprResult = c.binaryLoad(expr.Left, expr.Right, expr.Operator)
}

func (c *codegen) VisitLogical(expr *ast.Logical) {
	type_ := ast.Primitive{Kind: ast.Bool}

	left := c.implicitCastLoadExpr(&type_, expr.Left)
	right := c.implicitCastLoadExpr(&type_, expr.Right)

	switch expr.Operator.Token().Kind {
	case scanner.Or:
		false_ := c.function.Block("or.false")
		end := c.function.Block("or.end")

		// Start
		startBlock := c.block

		c.setLocationMeta(
			c.block.Add(&ir.BrInst{Condition: left.v, True: end, False: false_}),
			expr,
		)

		// False
		c.beginBlock(false_)

		c.setLocationMeta(
			c.block.Add(&ir.BrInst{True: end}),
			expr,
		)

		// End
		c.beginBlock(end)

		result := c.block.Add(&ir.PhiInst{Incs: []ir.Incoming{
			{
				Value: ir.True,
				Label: startBlock,
			},
			{
				Value: right.v,
				Label: false_,
			},
		}})

		c.setLocationMeta(result, expr)
		c.exprResult = exprValue{v: result}

	case scanner.And:
		true_ := c.function.Block("and.true")
		end := c.function.Block("and.end")

		// Start
		startBlock := c.block

		c.setLocationMeta(
			c.block.Add(&ir.BrInst{Condition: left.v, True: true_, False: end}),
			expr,
		)

		// True
		c.beginBlock(true_)

		c.setLocationMeta(
			c.block.Add(&ir.BrInst{True: end}),
			expr,
		)

		// End
		c.beginBlock(end)

		result := c.block.Add(&ir.PhiInst{Incs: []ir.Incoming{
			{
				Value: ir.False,
				Label: startBlock,
			},
			{
				Value: right.v,
				Label: true_,
			},
		}})

		c.setLocationMeta(result, expr)
		c.exprResult = exprValue{v: result}

	default:
		log.Fatalln("Invalid logical operator")
	}
}

func (c *codegen) VisitIdentifier(expr *ast.Identifier) {
	switch expr.Result().Kind {
	case ast.TypeResultKind, ast.ResolverResultKind:
		// Nothing

	case ast.ValueResultKind:
		switch node := expr.Result().Value().(type) {
		case *ast.GlobalVar:
			c.exprResult = c.getGlobalVariable(node)

		default:
			if v := c.scopes.getVariable(expr.Name); v != nil {
				c.exprResult = v.value
			}
		}

	case ast.CallableResultKind:
		switch node := expr.Result().Callable().(type) {
		case ast.FuncType:
			c.exprResult = c.getFunction(expr.Result().Type.(ast.FuncType))

		case *ast.GlobalVar:
			c.exprResult = c.getGlobalVariable(node)

		case *ast.Var, *ast.Param:
			if v := c.scopes.getVariable(expr.Name); v != nil {
				c.exprResult = v.value
			}
		}

	default:
		panic("codegen.VisitIdentifier() - Not implemented")
	}
}

func (c *codegen) VisitAssignment(expr *ast.Assignment) {
	// Assignee
	assignee := c.acceptExpr(expr.Assignee)

	// Value
	value := c.implicitCastLoadExpr(expr.Assignee.Result().Type, expr.Value)

	if expr.Operator.Token().Kind != scanner.Equal {
		value = c.binary(
			expr.Operator,
			c.load(assignee, expr.Assignee.Result().Type),
			value,
			expr.Assignee.Result().Type,
		)
	}

	// Store
	store := c.block.Add(&ir.StoreInst{
		Pointer: assignee.v,
		Value:   value.v,
		Align:   abi.GetTargetAbi().Align(expr.Result().Type),
	})

	c.setLocationMeta(store, expr)
	c.exprResult = assignee
}

func (c *codegen) VisitCast(expr *ast.Cast) {
	switch expr.Operator.Token().Kind {
	case scanner.As:
		value := c.acceptExpr(expr.Value)
		c.exprResult = c.cast(value, expr.Value.Result().Type, expr.Target, expr)

	case scanner.Is:
		value := c.loadExpr(expr.Value)

		// Get vtable pointer
		vtablePtr := c.block.Add(&ir.ExtractValueInst{
			Value:   value.v,
			Indices: []uint32{uint32(0)},
		})

		// Get type id
		typ := c.vtables.getType(expr.Value.Result().Type.Resolved().(*ast.Interface))
		typPtr := &ir.PointerType{Pointee: typ}

		typeIdPtr := c.block.Add(&ir.GetElementPtrInst{
			PointerTyp: typPtr,
			Typ:        typ,
			Pointer:    vtablePtr,
			Indices: []ir.Value{
				&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)},
				&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)},
			},
			Inbounds: true,
		})

		typeId := c.block.Add(&ir.LoadInst{
			Typ:     ir.I32,
			Pointer: typeIdPtr,
		})

		// Compare
		result := c.block.Add(&ir.ICmpInst{
			Kind:   ir.Eq,
			Signed: false,
			Left:   typeId,
			Right: &ir.IntConst{
				Typ:   ir.I32,
				Value: ir.Unsigned(uint64(c.ctx.GetTypeID(expr.Target))),
			},
		})

		c.setLocationMeta(result, expr.Operator)
		c.exprResult = exprValue{v: result}

	default:
		panic("codegen.VisitCast() - Not implemented")
	}
}

func (c *codegen) VisitTypeCall(expr *ast.TypeCall) {
	value := uint32(0)

	switch expr.Callee.String() {
	case "sizeof":
		value = abi.GetTargetAbi().Size(expr.Arg)
	case "alignof":
		value = abi.GetTargetAbi().Align(expr.Arg)

	default:
		panic("codegen.VisitTypeCall() - Not implemented")
	}

	c.exprResult = exprValue{v: &ir.IntConst{
		Typ:   c.types.get(expr.Result().Type),
		Value: ir.Unsigned(uint64(value)),
	}}
}

func (c *codegen) VisitTypeof(expr *ast.Typeof) {
	c.exprResult = exprValue{v: &ir.IntConst{
		Typ:   c.types.get(expr.Result().Type),
		Value: ir.Unsigned(uint64(c.ctx.GetTypeID(expr.Arg.Result().Type))),
	}}
}

func (c *codegen) VisitCall(expr *ast.Call) {
	// Get type
	callee := c.acceptExpr(expr.Callee)

	var function ast.FuncType

	if _, ok := expr.Callee.Result().Callable().(*ast.Func); ok {
		function = expr.Callee.Result().Type.(ast.FuncType)
	}

	if f, ok := ast.As[ast.FuncType](expr.Callee.Result().Type); ok && function == nil {
		function = f
		callee = c.load(callee, expr.Callee.Result().Type)
	}

	// Load arguments
	hasReceiver := function.Receiver() != nil
	if _, ok := function.Parent().(*ast.Interface); ok {
		hasReceiver = true
	}

	argCount := len(expr.Args)
	if hasReceiver {
		argCount++
	}

	funcAbi := abi.GetFuncAbi(function.Underlying())
	returnArgs := funcAbi.Classify(function.Returns(), nil)

	args := make([]ir.Value, 0, argCount)
	hasReturnPtr := false

	if len(returnArgs) == 1 && returnArgs[0].Class == abi.Memory {
		pointer := c.allocas.get(function.Returns(), "")
		pointer.TypPtr.SRet = c.types.get(function.Returns())

		args = append(args, pointer)
		hasReturnPtr = true
	}

	if hasReceiver {
		args = append(args, c.this.v)
	}

	for i, arg := range expr.Args {
		if i >= function.ParameterCount() {
			args = append(args, c.loadExpr(arg).v)
		} else {
			param := function.ParameterIndex(i)
			var value exprValue

			if needsImplicitCast(arg.Result().Type, param.Type) {
				value = c.implicitCastLoadExpr(param.Type, arg)
			} else {
				value = c.acceptExpr(arg)
			}

			args = c.valueToParams(funcAbi, value, param.Type, args)
		}
	}

	// Intrinsic
	intrinsicName := function.Underlying().IntrinsicName()

	if intrinsicName != "" {
		args = c.modifyIntrinsicArgs(function.Underlying(), intrinsicName, args)
	}

	// Call
	call := c.block.Add(&ir.CallInst{
		Typ:    c.types.get(function).(*ir.FuncType),
		Callee: callee.v,
		Args:   args,
	})

	c.setLocationMetaCst(call, expr, scanner.LeftParen)

	if ast.IsPrimitive(function.Returns(), ast.Void) {
		c.exprResult = exprValue{v: call}
	} else {
		result := exprValue{v: call}

		if hasReturnPtr {
			result = exprValue{
				v:           args[0],
				addressable: true,
			}
		}

		c.exprResult = c.returnValueToValue(funcAbi, result, function.Returns())
	}

	// If the function returns a constant-sized array and the array is immediately indexed then store it in an alloca first
	if callNeedsTempVariable(expr) {
		pointer := c.allocas.get(function.Returns(), "")

		c.block.Add(&ir.StoreInst{
			Pointer: pointer,
			Value:   c.exprResult.v,
			Align:   abi.GetTargetAbi().Align(function.Returns()),
		})

		c.exprResult = exprValue{
			v:           pointer,
			addressable: true,
		}
	}
}

func (c *codegen) VisitIndex(expr *ast.Index) {
	value := c.acceptExpr(expr.Value)
	index := c.loadExpr(expr.Index)

	if pointer, ok := ast.As[*ast.Pointer](expr.Value.Result().Type); ok {
		value = exprValue{v: c.block.Add(&ir.LoadInst{
			Typ:     value.v.Type().(*ir.PointerType).Pointee,
			Pointer: value.v,
			Align:   abi.GetTargetAbi().Align(pointer.Pointee),
		})}
	}

	ptrType := ast.Pointer{Pointee: expr.Result().Type}

	result := c.block.Add(&ir.GetElementPtrInst{
		PointerTyp: c.types.get(&ptrType),
		Typ:        c.types.get(expr.Result().Type),
		Pointer:    value.v,
		Indices:    []ir.Value{index.v},
		Inbounds:   true,
	})

	c.setLocationMetaCst(result, expr, scanner.LeftBracket)

	c.exprResult = exprValue{
		v:           result,
		addressable: true,
	}
}

func (c *codegen) VisitMember(expr *ast.Member) {
	value := c.acceptExpr(expr.Value)

	switch expr.Result().Kind {
	case ast.TypeResultKind, ast.ResolverResultKind:
		// Nothing

	case ast.ValueResultKind:
		switch node := expr.Result().Value().(type) {
		case *ast.EnumCase:
			c.exprResult = exprValue{v: &ir.IntConst{
				Typ:   c.types.get(node.Parent().(*ast.Enum).ActualType),
				Value: ir.Signed(node.ActualValue),
			}}

		case ast.FieldLike:
			if node.Underlying().IsStatic() {
				c.exprResult = c.getStaticVariable(node)
			} else {
				struct_ := node.Struct()
				fields, _ := abi.GetStructLayout(struct_.Underlying()).Fields(abi.GetTargetAbi(), struct_)
				_, i := getField(fields, node.Name())

				value, s := c.memberLoad(expr.Value.Result().Type, value)

				if value.addressable {
					ptrType := ast.Pointer{Pointee: node.Type()}

					result := c.block.Add(&ir.GetElementPtrInst{
						PointerTyp: c.types.get(&ptrType),
						Typ:        c.types.get(s),
						Pointer:    value.v,
						Indices: []ir.Value{
							&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)},
							&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(uint64(i))},
						},
						Inbounds: true,
					})

					c.setLocationMeta(result, expr.Name)

					c.exprResult = exprValue{
						v:           result,
						addressable: true,
					}
				} else {
					result := c.block.Add(&ir.ExtractValueInst{
						Value:   value.v,
						Indices: []uint32{uint32(i)},
					})

					c.setLocationMeta(result, expr.Name)
					c.exprResult = exprValue{v: result}
				}
			}

		case *ast.GlobalVar:
			c.exprResult = c.getGlobalVariable(node)
		}

	case ast.CallableResultKind:
		switch node := expr.Result().Callable().(type) {
		case *ast.Field:
			if node.IsStatic() {
				c.exprResult = c.getStaticVariable(node)
			} else {
				value, _ := c.memberLoad(expr.Value.Result().Type, value)

				c.exprResult = c.getFunction(expr.Result().Type.(ast.FuncType))
				c.this = value
			}

		case ast.FuncType:
			// Interface
			if inter, ok := ast.As[*ast.Interface](expr.Value.Result().Type); ok {
				if value.addressable {
					value = c.load(value, expr.Value.Result().Type)
				}

				// Get vtable pointer
				vtablePtr := c.block.Add(&ir.ExtractValueInst{
					Value:   value.v,
					Indices: []uint32{uint32(0)},
				})

				// Get function
				_, index := inter.GetMethod(expr.Name.String())

				typ := c.vtables.getType(inter)
				typPtr := &ir.PointerType{Pointee: typ}

				functionPtr := c.block.Add(&ir.GetElementPtrInst{
					PointerTyp: typPtr,
					Typ:        typ,
					Pointer:    vtablePtr,
					Indices: []ir.Value{
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)},
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(1)},
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(uint64(index))},
					},
					Inbounds: true,
				})

				fnPtr := typ.Fields[1].(*ir.ArrayType).Base

				function := c.block.Add(&ir.LoadInst{
					Typ:     fnPtr,
					Pointer: functionPtr,
				})

				// Get data pointer
				dataPtr := c.block.Add(&ir.ExtractValueInst{
					Value:   value.v,
					Indices: []uint32{uint32(1)},
				})

				// Return
				c.exprResult = exprValue{v: function}
				c.this = exprValue{v: dataPtr, addressable: true}

				return
			}

			// Struct
			if expr.Value.Result().Type != nil {
				value, _ = c.memberLoad(expr.Value.Result().Type, value)
			}

			if node.Receiver() != nil && !value.addressable {
				pointer := c.allocas.get(expr.Value.Result().Type, "")

				c.block.Add(&ir.StoreInst{
					Pointer: pointer,
					Value:   value.v,
					Align:   abi.GetTargetAbi().Align(node),
				})

				value = exprValue{
					v:           pointer,
					addressable: true,
				}
			}

			c.exprResult = c.getFunction(node)
			c.this = value

		case *ast.GlobalVar:
			c.exprResult = c.getGlobalVariable(node)

		default:
			panic("codegen.VisitMember() - Callable not implemented")
		}

	default:
		panic("codegen.VisitMember() - Result kind not implemented")
	}
}

func (c *codegen) memberLoad(type_ ast.Type, value exprValue) (exprValue, ast.StructType) {
	if s, ok := ast.As[ast.StructType](type_); ok {
		return value, s
	}

	if v, ok := ast.As[*ast.Pointer](type_); ok {
		if v, ok := ast.As[ast.StructType](v.Pointee); ok {
			load := c.block.Add(&ir.LoadInst{
				Typ:     value.v.Type().(*ir.PointerType).Pointee,
				Pointer: value.v,
				Align:   abi.GetTargetAbi().Align(v),
			})

			return exprValue{
				v:           load,
				addressable: true,
			}, v
		}
	}

	panic("codegen.memberLoad() - Not implemented")
}

// Utils

func getField(fields []ast.FieldLike, name ast.Node) (ast.FieldLike, int) {
	for i, field := range fields {
		if field.Name().String() == name.String() {
			return field, i
		}
	}

	panic("codegen.getField() - Slice doesn't contain a field with the name")
}

func (c *codegen) binaryLoad(left, right ast.Expr, operator *ast.Token) exprValue {
	// Interface == Pointer
	if operator.Token().Kind == scanner.EqualEqual || operator.Token().Kind == scanner.BangEqual {
		if _, ok := ast.As[*ast.Interface](left.Result().Type); ok {
			if _, ok := ast.As[*ast.Pointer](right.Result().Type); ok {
				left := c.loadExpr(left)
				right := c.loadExpr(right)

				kind := ir.Eq
				if operator.Token().Kind == scanner.BangEqual {
					kind = ir.Ne
				}

				result := c.block.Add(&ir.ICmpInst{
					Kind:   kind,
					Signed: false,
					Left: c.block.Add(&ir.ExtractValueInst{
						Value:   left.v,
						Indices: []uint32{1},
					}),
					Right: right.v,
				})

				c.setLocationMeta(result, operator)
				return exprValue{v: result}
			}
		}
	}

	// Left -> Right
	cast, castOk := common.GetImplicitCast(left.Result().Type, right.Result().Type)

	if castOk {
		to := right.Result().Type

		left := c.convertCast(c.loadExpr(left), cast, left.Result().Type, to, operator)
		right := c.loadExpr(right)

		return c.binary(operator, left, right, to)
	}

	// Right -> Left
	cast, castOk = common.GetImplicitCast(right.Result().Type, left.Result().Type)

	if castOk {
		to := left.Result().Type

		left := c.loadExpr(left)
		right := c.convertCast(c.loadExpr(right), cast, right.Result().Type, to, operator)

		return c.binary(operator, left, right, to)
	}

	// Invalid
	panic("codegen.binaryLoad() - Not implemented")
}

func (c *codegen) binary(op ast.Node, left exprValue, right exprValue, type_ ast.Type) exprValue {
	left = c.load(left, type_)
	right = c.load(right, type_)

	var result ir.MetaValue

	switch op.Token().Kind {
	case scanner.Plus, scanner.PlusEqual, scanner.PlusPlus:
		result = c.block.Add(&ir.AddInst{
			Left:  left.v,
			Right: right.v,
		})

	case scanner.Minus, scanner.MinusEqual, scanner.MinusMinus:
		result = c.block.Add(&ir.SubInst{
			Left:  left.v,
			Right: right.v,
		})

	case scanner.Star, scanner.StarEqual:
		result = c.block.Add(&ir.MulInst{
			Left:  left.v,
			Right: right.v,
		})

	case scanner.Slash, scanner.SlashEqual:
		if isFloating(type_) {
			result = c.block.Add(&ir.FDivInst{
				Left:  left.v,
				Right: right.v,
			})
		} else {
			result = c.block.Add(&ir.IDivInst{
				Signed: isSigned(type_),
				Left:   left.v,
				Right:  right.v,
			})
		}

	case scanner.Percentage, scanner.PercentageEqual:
		if isFloating(type_) {
			result = c.block.Add(&ir.FRemInst{
				Left:  left.v,
				Right: right.v,
			})
		} else {
			result = c.block.Add(&ir.IRemInst{
				Signed: isSigned(type_),
				Left:   left.v,
				Right:  right.v,
			})
		}

	case scanner.Pipe, scanner.PipeEqual:
		result = c.block.Add(&ir.OrInst{
			Left:  left.v,
			Right: right.v,
		})

	case scanner.Xor, scanner.XorEqual:
		result = c.block.Add(&ir.XorInst{
			Left:  left.v,
			Right: right.v,
		})

	case scanner.Ampersand, scanner.AmpersandEqual:
		result = c.block.Add(&ir.AndInst{
			Left:  left.v,
			Right: right.v,
		})

	case scanner.LessLess, scanner.LessLessEqual:
		result = c.block.Add(&ir.ShlInst{
			Left:  left.v,
			Right: right.v,
		})

	case scanner.GreaterGreater, scanner.GreaterGreaterEqual:
		result = c.block.Add(&ir.ShrInst{
			SignExtend: false,
			Left:       left.v,
			Right:      right.v,
		})

	default:
		var kind ir.CmpKind

		switch op.Token().Kind {
		case scanner.EqualEqual:
			kind = ir.Eq
		case scanner.BangEqual:
			kind = ir.Ne

		case scanner.Less:
			kind = ir.Lt
		case scanner.LessEqual:
			kind = ir.Le
		case scanner.Greater:
			kind = ir.Gt
		case scanner.GreaterEqual:
			kind = ir.Ge

		default:
			panic("codegen.binary() - Not implemented")
		}

		if isFloating(type_) {
			result = c.block.Add(&ir.FCmpInst{
				Kind:    kind,
				Ordered: false,
				Left:    left.v,
				Right:   right.v,
			})
		} else {
			result = c.block.Add(&ir.ICmpInst{
				Kind:   kind,
				Signed: isSigned(type_),
				Left:   left.v,
				Right:  right.v,
			})
		}
	}

	c.setLocationMeta(result, op)
	return exprValue{v: result}
}
