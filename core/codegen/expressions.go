package codegen

import (
	"fireball/core/ast"
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
	struct_, _ := ast.As[*ast.Struct](expr.Type)
	type_ := c.types.get(struct_)

	var result ir.Value = &ir.ZeroInitConst{Typ: type_}

	for _, field := range expr.Fields {
		element := c.loadExpr(field.Value)
		i, _ := struct_.GetField(field.Name.String())

		r := c.block.Add(&ir.InsertValueInst{
			Value:   result,
			Element: element.v,
			Indices: []uint32{uint32(i)},
		})

		c.setLocationMeta(r, field)
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
				Typ:   c.types.get(mallocFunc.Params[0].Type),
				Value: ir.Unsigned(uint64(struct_.Size())),
			}},
		})

		c.block.Add(&ir.StoreInst{
			Pointer: pointer,
			Value:   result,
			Align:   struct_.Align() * 8,
		})

		c.exprResult = exprValue{v: pointer}
	}
}

func (c *codegen) VisitArrayInitializer(expr *ast.ArrayInitializer) {
	type_ := c.types.get(expr.Result().Type)

	var result ir.Value = &ir.ZeroInitConst{Typ: type_}

	for i, value := range expr.Values {
		element := c.loadExpr(value)

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
	count := c.loadExpr(expr.Count)

	mallocFunc := c.resolver.GetFunction("malloc")
	malloc := c.getFunction(mallocFunc)

	a, _ := ast.As[*ast.Primitive](expr.Count.Result().Type)
	b, _ := ast.As[*ast.Primitive](mallocFunc.Params[0].Type)

	c.castPrimitiveToPrimitive(
		count,
		expr.Count.Result().Type,
		mallocFunc.Params[0].Type,
		a.Kind,
		b.Kind,
		expr,
	)
	count = c.exprResult

	pointer := c.block.Add(&ir.CallInst{
		Callee: malloc.v,
		Args: []ir.Value{&ir.IntConst{
			Typ:   c.types.get(mallocFunc.Params[0].Type),
			Value: ir.Unsigned(uint64(expr.Type.Size())),
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
					Align:   expr.Result().Type.Align() * 8,
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
				Align:   expr.Value.Result().Type.Align() * 8,
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
				Align:   expr.Value.Result().Type.Align() * 8,
			})

			result = prevValue.v

		default:
			panic("codegen.VisitUnary() - Invalid unary prefix operator")
		}
	}

	c.exprResult = exprValue{v: result}
}

func (c *codegen) VisitBinary(expr *ast.Binary) {
	left := c.acceptExpr(expr.Left)
	right := c.acceptExpr(expr.Right)

	c.exprResult = c.binary(expr.Operator, left, right, expr.Left.Result().Type)
}

func (c *codegen) VisitLogical(expr *ast.Logical) {
	left := c.loadExpr(expr.Left)
	right := c.loadExpr(expr.Right)

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
		case *ast.Func:
			c.exprResult = c.getFunction(node)

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
	value := c.loadExpr(expr.Value)

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
		Align:   expr.Result().Type.Align() * 8,
	})

	c.setLocationMeta(store, expr)
	c.exprResult = assignee
}

func (c *codegen) VisitCast(expr *ast.Cast) {
	value := c.acceptExpr(expr.Value)

	if from, ok := ast.As[*ast.Primitive](expr.Value.Result().Type); ok {
		if to, ok := ast.As[*ast.Primitive](expr.Result().Type); ok {
			// primitive to primitive
			c.castPrimitiveToPrimitive(value, from, to, from.Kind, to.Kind, expr)
			return
		}
	}

	if from, ok := ast.As[*ast.Enum](expr.Value.Result().Type); ok {
		if to, ok := ast.As[*ast.Primitive](expr.Result().Type); ok {
			// enum to integer
			fromT, _ := ast.As[*ast.Primitive](from.Type)

			c.castPrimitiveToPrimitive(value, from, to, fromT.Kind, to.Kind, expr)
			return
		}
	}

	if from, ok := ast.As[*ast.Primitive](expr.Value.Result().Type); ok {
		if to, ok := expr.Result().Type.(*ast.Enum); ok {
			// integer to enum
			toT, _ := ast.As[*ast.Primitive](to.Type)

			c.castPrimitiveToPrimitive(value, from, to, from.Kind, toT.Kind, expr)
			return
		}
	}

	if _, ok := ast.As[*ast.Pointer](expr.Value.Result().Type); ok {
		if _, ok := ast.As[*ast.Pointer](expr.Result().Type); ok {
			// pointer to pointer
			c.exprResult = value
			return
		}

		if _, ok := ast.As[*ast.Func](expr.Result().Type); ok {
			// pointer to function pointer
			c.exprResult = value
			return
		}
	}

	// Error
	panic("codegen.VisitCast() - Invalid cast")
}

func (c *codegen) castPrimitiveToPrimitive(value exprValue, from, to ast.Type, fromKind, toKind ast.PrimitiveKind, location ast.Node) {
	if fromKind == toKind || (ast.EqualsPrimitiveCategory(fromKind, toKind) && ast.GetBitSize(fromKind) == ast.GetBitSize(toKind)) {
		c.exprResult = value
		return
	}

	value = c.load(value, from)

	if (ast.IsInteger(fromKind) || ast.IsFloating(fromKind)) && toKind == ast.Bool {
		// integer / floating to bool
		var result ir.MetaValue

		if ast.IsFloating(fromKind) {
			result = c.block.Add(&ir.FCmpInst{
				Kind:    ir.Ne,
				Ordered: false,
				Left:    value.v,
				Right:   ir.False,
			})
		} else {
			result = c.block.Add(&ir.ICmpInst{
				Kind:   ir.Ne,
				Signed: ast.IsSigned(fromKind),
				Left:   value.v,
				Right:  ir.False,
			})
		}

		c.setLocationMeta(result, location)
		c.exprResult = exprValue{v: result}
	} else {
		type_ := c.types.get(to)
		var result ir.MetaValue

		if (ast.IsInteger(fromKind) || fromKind == ast.Bool) && ast.IsInteger(toKind) {
			// integer / bool to integer
			if from.Size() > to.Size() {
				result = c.block.Add(&ir.TruncInst{
					Value: value.v,
					Typ:   type_,
				})
			} else {
				result = c.block.Add(&ir.ExtInst{
					SignExtend: false,
					Value:      value.v,
					Typ:        type_,
				})
			}
		} else if ast.IsFloating(fromKind) && ast.IsFloating(toKind) {
			// floating to floating
			if from.Size() > to.Size() {
				result = c.block.Add(&ir.TruncInst{
					Value: value.v,
					Typ:   type_,
				})
			} else {
				result = c.block.Add(&ir.FExtInst{
					Value: value.v,
					Typ:   type_,
				})
			}
		} else if (ast.IsInteger(fromKind) || fromKind == ast.Bool) && ast.IsFloating(toKind) {
			// integer / bool to floating
			result = c.block.Add(&ir.I2FInst{
				Signed: ast.IsSigned(fromKind),
				Value:  value.v,
				Typ:    type_,
			})
		} else if ast.IsFloating(fromKind) && ast.IsInteger(toKind) {
			// floating to integer
			result = c.block.Add(&ir.F2IInst{
				Signed: ast.IsSigned(toKind),
				Value:  value.v,
				Typ:    type_,
			})
		}

		c.setLocationMeta(result, location)
		c.exprResult = exprValue{v: result}
	}
}

func (c *codegen) VisitTypeCall(expr *ast.TypeCall) {
	value := uint32(0)

	switch expr.Callee.String() {
	case "sizeof":
		value = expr.Arg.Size()
	case "alignof":
		value = expr.Arg.Align()

	default:
		panic("codegen.VisitTypeCall() - Not implemented")
	}

	c.exprResult = exprValue{v: &ir.IntConst{
		Typ:   c.types.get(expr.Result().Type),
		Value: ir.Unsigned(uint64(value)),
	}}
}

func (c *codegen) VisitCall(expr *ast.Call) {
	// Get type
	callee := c.acceptExpr(expr.Callee)

	var function *ast.Func

	if f, ok := expr.Callee.Result().Callable().(*ast.Func); ok {
		function = f
	}

	if f, ok := ast.As[*ast.Func](expr.Callee.Result().Type); ok && function == nil {
		function = f
		callee = c.load(callee, expr.Callee.Result().Type)
	}

	// Load arguments
	argCount := len(expr.Args)
	if function.Method() != nil {
		argCount++
	}

	args := make([]ir.Value, argCount)

	if function.Method() != nil {
		args[0] = c.this.v
	}

	for i, arg := range expr.Args {
		index := i
		if function.Method() != nil {
			index++
		}

		args[index] = c.loadExpr(arg).v
	}

	// Intrinsic
	intrinsicName := function.IntrinsicName()

	if intrinsicName != "" {
		args = c.modifyIntrinsicArgs(function, intrinsicName, args)
	}

	// Call
	result := c.block.Add(&ir.CallInst{
		Callee: callee.v,
		Args:   args,
	})

	c.setLocationMeta(result, expr)
	c.exprResult = exprValue{v: result}

	// If the function returns a constant-sized array and the array is immediately indexed then store it in an alloca first
	if callNeedsTempVariable(expr) {
		pointer := c.allocas[expr]

		c.block.Add(&ir.StoreInst{
			Pointer: pointer.v,
			Value:   c.exprResult.v,
			Align:   function.Returns.Align() * 8,
		})

		c.exprResult = pointer
	}
}

func (c *codegen) VisitIndex(expr *ast.Index) {
	value := c.acceptExpr(expr.Value)
	index := c.loadExpr(expr.Index)

	if pointer, ok := ast.As[*ast.Pointer](expr.Value.Result().Type); ok {
		value = exprValue{v: c.block.Add(&ir.LoadInst{
			Typ:     value.v.Type().(*ir.PointerType).Pointee,
			Pointer: value.v,
			Align:   pointer.Pointee.Align() * 8,
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

	c.setLocationMeta(result, expr)

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

		case *ast.Field:
			if node.IsStatic() {
				c.exprResult = c.getStaticVariable(node)
			} else {
				value, s := c.memberLoad(expr.Value.Result().Type, value)

				if value.addressable {
					ptrType := ast.Pointer{Pointee: node.Type}

					result := c.block.Add(&ir.GetElementPtrInst{
						PointerTyp: c.types.get(&ptrType),
						Typ:        c.types.get(s),
						Pointer:    value.v,
						Indices: []ir.Value{
							&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)},
							&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(uint64(node.Index()))},
						},
						Inbounds: true,
					})

					c.setLocationMeta(result, expr)

					c.exprResult = exprValue{
						v:           result,
						addressable: true,
					}
				} else {
					result := c.block.Add(&ir.ExtractValueInst{
						Value:   value.v,
						Indices: []uint32{uint32(node.Index())},
					})

					c.setLocationMeta(result, expr)
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

				c.exprResult = c.getFunction(expr.Result().Type.(*ast.Func))
				c.this = value
			}

		case *ast.Func:
			if expr.Value.Result().Type != nil {
				value, _ = c.memberLoad(expr.Value.Result().Type, value)
			}

			if node.Method() != nil && !value.addressable {
				pointer := c.allocas[expr]

				c.block.Add(&ir.StoreInst{
					Pointer: pointer.v,
					Value:   value.v,
					Align:   node.Align() * 8,
				})

				value = pointer
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

func (c *codegen) memberLoad(type_ ast.Type, value exprValue) (exprValue, *ast.Struct) {
	if s, ok := ast.As[*ast.Struct](type_); ok {
		return value, s
	}

	if v, ok := ast.As[*ast.Pointer](type_); ok {
		if v, ok := ast.As[*ast.Struct](v.Pointee); ok {
			load := c.block.Add(&ir.LoadInst{
				Typ:     value.v.Type().(*ir.PointerType).Pointee,
				Pointer: value.v,
				Align:   v.Align() * 8,
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
