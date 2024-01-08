package codegen

import (
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/llvm"
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
	var value llvm.Value
	type_ := c.getType(expr.Result().Type)

	switch expr.Token().Kind {
	case scanner.Nil:
		value = c.function.LiteralRaw(type_, "null")

	case scanner.True, scanner.False:
		value = c.function.LiteralRaw(type_, expr.String())

	case scanner.Number:
		raw := expr.String()
		last := raw[len(raw)-1]

		if last == 'f' || last == 'F' {
			v, _ := strconv.ParseFloat(raw[:len(raw)-1], 32)
			value = c.function.Literal(type_, llvm.Literal{Floating: v})
		} else if strings.ContainsRune(raw, '.') {
			v, _ := strconv.ParseFloat(raw, 64)
			value = c.function.Literal(type_, llvm.Literal{Floating: v})
		} else {
			t, _ := ast.As[*ast.Primitive](expr.Result().Type)

			if ast.IsSigned(t.Kind) {
				v, _ := strconv.ParseInt(raw, 10, 64)
				value = c.function.Literal(type_, llvm.Literal{Signed: v})
			} else {
				v, _ := strconv.ParseUint(raw, 10, 64)
				value = c.function.Literal(type_, llvm.Literal{Unsigned: v})
			}
		}

	case scanner.Hex:
		v, _ := strconv.ParseUint(expr.String()[2:], 16, 64)
		value = c.function.Literal(type_, llvm.Literal{Unsigned: v})

	case scanner.Binary:
		v, _ := strconv.ParseUint(expr.String()[2:], 2, 64)
		value = c.function.Literal(type_, llvm.Literal{Unsigned: v})

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

		value = c.function.Literal(type_, llvm.Literal{Unsigned: uint64(number)})

	case scanner.String:
		value = c.module.Constant(expr.String()[1 : len(expr.String())-1])

	default:
		panic("codegen.VisitLiteral() - Invalid literal kind")
	}

	// Emit
	c.exprResult = exprValue{
		v: value,
	}
}

func (c *codegen) VisitStructInitializer(expr *ast.StructInitializer) {
	// Value
	struct_, _ := ast.As[*ast.Struct](expr.Type)
	type_ := c.getType(struct_)

	result := c.function.LiteralRaw(type_, "zeroinitializer")

	for _, field := range expr.Fields {
		element := c.loadExpr(field.Value)
		i, _ := struct_.GetField(field.Name.String())

		r := c.block.InsertValue(result, element.v, i)
		r.SetLocation(field.Name.Cst())

		result = r
	}

	c.exprResult = exprValue{v: result}

	// Malloc
	if expr.New {
		mallocFunc := c.resolver.GetFunction("malloc")
		malloc := c.getFunction(mallocFunc)

		pointer := c.block.Call(
			malloc.v,
			[]llvm.Value{c.function.Literal(
				c.getType(mallocFunc.Params[0].Type),
				llvm.Literal{Unsigned: uint64(struct_.Size())},
			)},
			c.getType(mallocFunc.Returns),
		)

		store := c.block.Store(pointer, result)
		store.SetAlign(struct_.Align())

		c.exprResult = exprValue{v: pointer}
	}
}

func (c *codegen) VisitArrayInitializer(expr *ast.ArrayInitializer) {
	type_ := c.getType(expr.Result().Type)

	result := c.function.LiteralRaw(type_, "zeroinitializer")

	for i, valueExpr := range expr.Values {
		element := c.loadExpr(valueExpr)

		r := c.block.InsertValue(result, element.v, i)
		r.SetLocation(valueExpr.Cst())

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
		expr.Cst(),
	)
	count = c.exprResult

	pointer := c.block.Call(
		malloc.v,
		[]llvm.Value{c.block.Binary(
			llvm.Mul,
			c.function.Literal(
				c.getType(mallocFunc.Params[0].Type),
				llvm.Literal{Unsigned: uint64(expr.Type.Size())},
			),
			count.v,
		)},
		c.getType(mallocFunc.Returns),
	)

	c.exprResult = exprValue{v: pointer}
}

func (c *codegen) VisitUnary(expr *ast.Unary) {
	value := c.acceptExpr(expr.Value)
	var result llvm.Value

	if expr.Prefix {
		// Prefix
		switch expr.Operator.Token().Kind {
		case scanner.Bang:
			t := ast.Primitive{Kind: ast.Bool}

			r := c.block.Binary(
				llvm.Xor,
				c.function.Literal(
					c.getType(&t),
					llvm.Literal{Signed: 1},
				),
				c.load(value, expr.Value.Result().Type).v,
			)

			r.SetLocation(expr.Cst())
			result = r

		case scanner.Minus:
			if v, ok := ast.As[*ast.Primitive](expr.Value.Result().Type); ok {
				value := c.load(value, expr.Value.Result().Type)

				if ast.IsFloating(v.Kind) {
					// floating
					result = c.block.FNeg(value.v)
				} else {
					// signed
					r := c.block.Binary(
						llvm.Sub,
						c.function.Literal(
							c.getType(expr.Value.Result().Type),
							llvm.Literal{},
						),
						value.v,
					)

					r.SetLocation(expr.Cst())
					result = r
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
				load := c.block.Load(result)
				load.SetAlign(expr.Result().Type.Align())

				result = load
			}

		case scanner.PlusPlus, scanner.MinusMinus:
			newValue := c.binary(
				expr.Operator,
				value,
				exprValue{v: c.function.Literal(
					c.getType(expr.Value.Result().Type),
					llvm.Literal{
						Signed:   1,
						Unsigned: 1,
						Floating: 1,
					},
				)},
				expr.Value.Result().Type,
			)

			store := c.block.Store(value.v, newValue.v)
			store.SetAlign(expr.Value.Result().Type.Align())
			store.SetLocation(expr.Cst())

			result = newValue.v

		default:
			panic("codegen.VisitUnary() - Invalid unary prefix operator")
		}
	} else {
		// Postfix
		switch expr.Operator.Token().Kind {
		case scanner.PlusPlus, scanner.MinusMinus:
			prevValue := c.load(value, expr.Value.Result().Type)

			newValue := c.binary(
				expr.Operator,
				prevValue,
				exprValue{v: c.function.Literal(
					c.getType(expr.Value.Result().Type),
					llvm.Literal{
						Signed:   1,
						Unsigned: 1,
						Floating: 1,
					},
				)},
				expr.Value.Result().Type,
			)

			store := c.block.Store(value.v, newValue.v)
			store.SetAlign(expr.Value.Result().Type.Align())
			store.SetLocation(expr.Cst())

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
		c.block.Br(left.v, end, false_).SetLocation(expr.Cst())

		// False
		c.beginBlock(false_)
		c.block.Br(nil, end, nil).SetLocation(expr.Cst())

		// End
		c.beginBlock(end)

		result := c.block.Phi(c.function.LiteralRaw(c.getType(expr.Result().Type), "true"), startBlock, right.v, false_)
		result.SetLocation(expr.Cst())

		c.exprResult = exprValue{v: result}

	case scanner.And:
		true_ := c.function.Block("and.true")
		end := c.function.Block("and.end")

		// Start
		startBlock := c.block
		c.block.Br(left.v, true_, end).SetLocation(expr.Cst())

		// True
		c.beginBlock(true_)
		c.block.Br(nil, end, nil).SetLocation(expr.Cst())

		// End
		c.beginBlock(end)

		result := c.block.Phi(c.function.LiteralRaw(c.getType(expr.Result().Type), "false"), startBlock, right.v, true_)
		result.SetLocation(expr.Cst())

		c.exprResult = exprValue{v: result}

	default:
		log.Fatalln("Invalid logical operator")
	}
}

func (c *codegen) VisitIdentifier(expr *ast.Identifier) {
	switch expr.Result().Kind {
	case ast.TypeResultKind:
		// Nothing

	case ast.ValueResultKind:
		if v := c.getVariable(expr.Name); v != nil {
			c.exprResult = v.value
		}

	case ast.CallableResultKind:
		switch node := expr.Result().Callable().(type) {
		case *ast.Func:
			c.exprResult = c.getFunction(node)

		case *ast.Var, *ast.Param:
			if v := c.getVariable(expr.Name); v != nil {
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
	store := c.block.Store(assignee.v, value.v)
	store.SetAlign(expr.Result().Type.Align())
	store.SetLocation(expr.Cst())

	c.exprResult = assignee
}

func (c *codegen) VisitCast(expr *ast.Cast) {
	value := c.acceptExpr(expr.Value)

	if from, ok := ast.As[*ast.Primitive](expr.Value.Result().Type); ok {
		if to, ok := ast.As[*ast.Primitive](expr.Result().Type); ok {
			// primitive to primitive
			c.castPrimitiveToPrimitive(value, from, to, from.Kind, to.Kind, expr.Cst())
			return
		}
	}

	if from, ok := ast.As[*ast.Enum](expr.Value.Result().Type); ok {
		if to, ok := ast.As[*ast.Primitive](expr.Result().Type); ok {
			// enum to integer
			fromT, _ := ast.As[*ast.Primitive](from.Type)

			c.castPrimitiveToPrimitive(value, from, to, fromT.Kind, to.Kind, expr.Cst())
			return
		}
	}

	if from, ok := ast.As[*ast.Primitive](expr.Value.Result().Type); ok {
		if to, ok := expr.Result().Type.(*ast.Enum); ok {
			// integer to enum
			toT, _ := ast.As[*ast.Primitive](to.Type)

			c.castPrimitiveToPrimitive(value, from, to, from.Kind, toT.Kind, expr.Cst())
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

func (c *codegen) castPrimitiveToPrimitive(value exprValue, from, to ast.Type, fromKind, toKind ast.PrimitiveKind, location *cst.Node) {
	if fromKind == toKind || (ast.EqualsPrimitiveCategory(fromKind, toKind) && ast.GetBitSize(fromKind) == ast.GetBitSize(toKind)) {
		c.exprResult = value
		return
	}

	value = c.load(value, from)

	if (ast.IsInteger(fromKind) || ast.IsFloating(fromKind)) && toKind == ast.Bool {
		// integer / floating to bool
		result := c.block.Binary(
			llvm.Ne,
			value.v,
			c.function.Literal(
				value.v.Type(),
				llvm.Literal{},
			),
		)

		result.SetLocation(location)
		c.exprResult = exprValue{
			v: result,
		}
	} else {
		var kind llvm.CastKind

		if (ast.IsInteger(fromKind) || fromKind == ast.Bool) && ast.IsInteger(toKind) {
			// integer / bool to integer
			if from.Size() > to.Size() {
				kind = llvm.Trunc
			} else {
				kind = llvm.ZExt
			}
		} else if ast.IsFloating(fromKind) && ast.IsFloating(toKind) {
			// floating to floating
			if from.Size() > to.Size() {
				kind = llvm.FpTrunc
			} else {
				kind = llvm.FpExt
			}
		} else if (ast.IsInteger(fromKind) || fromKind == ast.Bool) && ast.IsFloating(toKind) {
			// integer / bool to floating
			if ast.IsSigned(fromKind) {
				kind = llvm.SiToFp
			} else {
				kind = llvm.UiToFp
			}
		} else if ast.IsFloating(fromKind) && ast.IsInteger(toKind) {
			// floating to integer
			if ast.IsSigned(toKind) {
				kind = llvm.FpToSi
			} else {
				kind = llvm.FpToUi
			}
		}

		result := c.block.Cast(kind, value.v, c.getType(to))
		result.SetLocation(location)

		c.exprResult = exprValue{
			v: result,
		}
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
		panic("codegen.VisitTypeCall() - Invalid name")
	}

	c.exprResult = exprValue{
		v: c.function.Literal(
			c.getType(expr.Result().Type),
			llvm.Literal{Signed: int64(value)},
		),
	}
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

	args := make([]llvm.Value, argCount)

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
	result := c.block.Call(callee.v, args, c.getType(function.Returns))
	result.SetLocation(expr.Cst())

	c.exprResult = exprValue{
		v: result,
	}

	// If the function returns a constant-sized array and the array is immediately indexed then store it in an alloca first
	if callNeedsTempVariable(expr) {
		pointer := c.allocas[expr]

		store := c.block.Store(pointer.v, c.exprResult.v)
		store.SetAlign(function.Returns.Align())

		c.exprResult = pointer
	}
}

func (c *codegen) VisitIndex(expr *ast.Index) {
	value := c.acceptExpr(expr.Value)
	index := c.loadExpr(expr.Index)

	if pointer, ok := ast.As[*ast.Pointer](expr.Value.Result().Type); ok {
		load := c.block.Load(value.v)
		load.SetAlign(pointer.Pointee.Align())

		value = exprValue{v: load}
	}

	t := ast.Pointer{Pointee: expr.Result().Type}
	type_ := c.getType(&t)

	result := c.block.GetElementPtr(
		value.v,
		[]llvm.Value{index.v},
		type_,
		c.getType(expr.Result().Type),
	)

	result.SetLocation(expr.Cst())

	c.exprResult = exprValue{
		v:           result,
		addressable: true,
	}
}

func (c *codegen) VisitMember(expr *ast.Member) {
	value := c.acceptExpr(expr.Value)

	switch expr.Result().Kind {
	case ast.ValueResultKind:
		switch node := expr.Result().Value().(type) {
		case *ast.EnumCase:
			c.exprResult = exprValue{
				v: c.function.Literal(
					c.getType(node.Parent().(*ast.Enum).ActualType),
					llvm.Literal{Signed: node.ActualValue, Unsigned: uint64(node.ActualValue)},
				),
			}

		case *ast.Field:
			if node.IsStatic() {
				c.exprResult = c.getStaticVariable(node)
			} else {
				value, s := c.memberLoad(expr.Value.Result().Type, value)

				if value.addressable {
					i32Type_ := ast.Primitive{Kind: ast.I32}
					i32Type := c.getType(&i32Type_)

					t := ast.Pointer{Pointee: node.Type}

					result := c.block.GetElementPtr(
						value.v,
						[]llvm.Value{
							c.function.Literal(i32Type, llvm.Literal{Signed: 0}),
							c.function.Literal(i32Type, llvm.Literal{Signed: int64(node.Index())}),
						},
						c.getType(&t),
						c.getType(s),
					)

					result.SetLocation(expr.Cst())

					c.exprResult = exprValue{
						v:           result,
						addressable: true,
					}
				} else {
					result := c.block.ExtractValue(value.v, node.Index())
					result.SetLocation(expr.Cst())

					c.exprResult = exprValue{
						v: result,
					}
				}
			}
		}

	case ast.CallableResultKind:
		switch node := expr.Result().Callable().(type) {
		case *ast.Func:
			value, _ := c.memberLoad(expr.Value.Result().Type, value)

			c.exprResult = c.getFunction(node)
			c.this = value

		case *ast.Field:
			if node.IsStatic() {
				c.exprResult = c.getStaticVariable(node)
			} else {
				value, _ := c.memberLoad(expr.Value.Result().Type, value)

				c.exprResult = c.getFunction(expr.Result().Type.(*ast.Func))
				c.this = value
			}

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
			load := c.block.Load(value.v)
			load.SetAlign(v.Align())

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

	var kind llvm.BinaryKind

	switch op.Token().Kind {
	case scanner.Plus, scanner.PlusEqual, scanner.PlusPlus:
		kind = llvm.Add
	case scanner.Minus, scanner.MinusEqual, scanner.MinusMinus:
		kind = llvm.Sub
	case scanner.Star, scanner.StarEqual:
		kind = llvm.Mul
	case scanner.Slash, scanner.SlashEqual:
		kind = llvm.Div
	case scanner.Percentage, scanner.PercentageEqual:
		kind = llvm.Rem

	case scanner.EqualEqual:
		kind = llvm.Eq
	case scanner.BangEqual:
		kind = llvm.Ne

	case scanner.Less:
		kind = llvm.Lt
	case scanner.LessEqual:
		kind = llvm.Le
	case scanner.Greater:
		kind = llvm.Gt
	case scanner.GreaterEqual:
		kind = llvm.Ge

	case scanner.Pipe, scanner.PipeEqual:
		kind = llvm.Or
	case scanner.Xor, scanner.XorEqual:
		kind = llvm.Xor
	case scanner.Ampersand, scanner.AmpersandEqual:
		kind = llvm.And
	case scanner.LessLess, scanner.LessLessEqual:
		kind = llvm.Shl
	case scanner.GreaterGreater, scanner.GreaterGreaterEqual:
		kind = llvm.Shr

	default:
		panic("codegen.binary() - Invalid operator kind")
	}

	result := c.block.Binary(kind, left.v, right.v)
	result.SetLocation(op.Cst())

	return exprValue{v: result}
}
