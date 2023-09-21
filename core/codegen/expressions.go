package codegen

import (
	"fireball/core/ast"
	"fireball/core/llvm"
	"fireball/core/scanner"
	"fireball/core/types"
	"log"
	"strconv"
	"strings"
)

func (c *codegen) VisitGroup(expr *ast.Group) {
	c.acceptExpr(expr.Expr)
}

func (c *codegen) VisitLiteral(expr *ast.Literal) {
	// Convert fireball constant into a LLVM IR constant
	var value llvm.Value
	type_ := c.getType(expr.Result().Type)

	switch expr.Value.Kind {
	case scanner.Nil:
		value = c.function.LiteralRaw(type_, "null")

	case scanner.True, scanner.False:
		value = c.function.LiteralRaw(type_, expr.Value.Lexeme)

	case scanner.Number:
		raw := expr.Value.Lexeme
		last := raw[len(raw)-1]

		if last == 'f' || last == 'F' {
			v, _ := strconv.ParseFloat(raw[:len(raw)-1], 32)
			value = c.function.Literal(type_, llvm.Literal{Floating: v})
		} else if strings.ContainsRune(raw, '.') {
			v, _ := strconv.ParseFloat(raw, 64)
			value = c.function.Literal(type_, llvm.Literal{Floating: v})
		} else {
			t := expr.Result().Type.(*types.PrimitiveType)

			if types.IsSigned(t.Kind) {
				v, _ := strconv.ParseInt(raw, 10, 64)
				value = c.function.Literal(type_, llvm.Literal{Signed: v})
			} else {
				v, _ := strconv.ParseUint(raw, 10, 64)
				value = c.function.Literal(type_, llvm.Literal{Unsigned: v})
			}
		}

	case scanner.Hex:
		v, _ := strconv.ParseUint(expr.Value.Lexeme[2:], 16, 64)
		value = c.function.Literal(type_, llvm.Literal{Unsigned: v})

	case scanner.Binary:
		v, _ := strconv.ParseUint(expr.Value.Lexeme[2:], 2, 64)
		value = c.function.Literal(type_, llvm.Literal{Unsigned: v})

	case scanner.Character:
		char := expr.Value.Lexeme[1 : len(expr.Value.Lexeme)-1]
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
		value = c.module.Constant(expr.Value.Lexeme[1 : len(expr.Value.Lexeme)-1])

	default:
		panic("codegen.VisitLiteral() - Invalid literal kind")
	}

	// Emit
	c.exprResult = exprValue{
		v: value,
	}
}

func (c *codegen) VisitStructInitializer(expr *ast.StructInitializer) {
	struct_, _ := expr.Result().Type.(*ast.Struct)
	type_ := c.getType(expr.Result().Type)

	result := c.function.LiteralRaw(type_, "zeroinitializer")

	for _, field := range expr.Fields {
		element := c.loadExpr(field.Value)
		i, _ := struct_.GetField(field.Name.Lexeme)

		r := c.block.InsertValue(result, element.v, i)
		r.SetLocation(field.Name)

		result = r
	}

	c.exprResult = exprValue{v: result}
}

func (c *codegen) VisitArrayInitializer(expr *ast.ArrayInitializer) {
	type_ := c.getType(expr.Result().Type)

	result := c.function.LiteralRaw(type_, "zeroinitializer")

	for i, valueExpr := range expr.Values {
		element := c.loadExpr(valueExpr)

		r := c.block.InsertValue(result, element.v, i)
		r.SetLocation(valueExpr.Token())

		result = r
	}

	c.exprResult = exprValue{v: result}
}

func (c *codegen) VisitUnary(expr *ast.Unary) {
	value := c.acceptExpr(expr.Value)
	var result llvm.Value

	if expr.Prefix {
		// Prefix
		switch expr.Op.Kind {
		case scanner.Bang:
			t := types.PrimitiveType{Kind: types.Bool}

			r := c.block.Binary(
				llvm.Xor,
				c.function.Literal(
					c.getType(&t),
					llvm.Literal{Signed: 1},
				),
				c.load(value).v,
			)

			r.SetLocation(expr.Token())
			result = r

		case scanner.Minus:
			if v, ok := expr.Value.Result().Type.(*types.PrimitiveType); ok {
				value := c.load(value)

				if types.IsFloating(v.Kind) {
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

					r.SetLocation(expr.Token())
					result = r
				}
			}

		case scanner.Ampersand:
			c.exprResult = exprValue{
				v:           value.v,
				addressable: false,
			}

			return

		case scanner.Star:
			result = c.load(value).v

			if _, ok := expr.Parent().(*ast.Assignment); !ok {
				result = c.block.Load(result)
			}

		case scanner.PlusPlus, scanner.MinusMinus:
			newValue := c.binary(
				expr.Op,
				value,
				exprValue{v: c.function.Literal(
					c.getType(expr.Value.Result().Type),
					llvm.Literal{
						Signed:   1,
						Unsigned: 1,
						Floating: 1,
					},
				)},
			)

			c.block.Store(value.v, newValue.v).SetLocation(expr.Token())
			result = newValue.v

		default:
			panic("codegen.VisitUnary() - Invalid unary prefix operator")
		}
	} else {
		// Postfix
		switch expr.Op.Kind {
		case scanner.PlusPlus, scanner.MinusMinus:
			prevValue := c.load(value)

			newValue := c.binary(
				expr.Op,
				prevValue,
				exprValue{v: c.function.Literal(
					c.getType(expr.Value.Result().Type),
					llvm.Literal{
						Signed:   1,
						Unsigned: 1,
						Floating: 1,
					},
				)},
			)

			c.block.Store(value.v, newValue.v).SetLocation(expr.Token())
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

	c.exprResult = c.binary(expr.Op, left, right)
}

func (c *codegen) VisitLogical(expr *ast.Logical) {
	left := c.loadExpr(expr.Left)
	right := c.loadExpr(expr.Right)

	switch expr.Op.Kind {
	case scanner.Or:
		false_ := c.function.Block("or.false")
		end := c.function.Block("or.end")

		// Start
		startBlock := c.block
		c.block.Br(left.v, end, false_).SetLocation(expr.Token())

		// False
		c.beginBlock(false_)
		c.block.Br(nil, end, nil).SetLocation(expr.Token())

		// End
		c.beginBlock(end)

		result := c.block.Phi(c.function.LiteralRaw(c.getType(expr.Result().Type), "true"), startBlock, right.v, false_)
		result.SetLocation(expr.Token())

		c.exprResult = exprValue{v: result}

	case scanner.And:
		true_ := c.function.Block("and.true")
		end := c.function.Block("and.end")

		// Start
		startBlock := c.block
		c.block.Br(left.v, true_, end).SetLocation(expr.Token())

		// True
		c.beginBlock(true_)
		c.block.Br(nil, end, nil).SetLocation(expr.Token())

		// End
		c.beginBlock(end)

		result := c.block.Phi(c.function.LiteralRaw(c.getType(expr.Result().Type), "false"), startBlock, right.v, true_)
		result.SetLocation(expr.Token())

		c.exprResult = exprValue{v: result}

	default:
		log.Fatalln("Invalid logical operator")
	}
}

func (c *codegen) VisitIdentifier(expr *ast.Identifier) {
	switch expr.Kind {
	case ast.FunctionKind:
		c.exprResult = c.getFunction(expr.Result().Type.(*ast.Func))
		return

	case ast.StructKind, ast.EnumKind:
		return

	case ast.VariableKind, ast.ParameterKind:
		if v := c.getVariable(expr.Identifier); v != nil {
			c.exprResult = v.value
			return
		}
	}

	panic("codegen.VisitIdentifier() - Invalid identifier")
}

func (c *codegen) VisitAssignment(expr *ast.Assignment) {
	// Assignee
	assignee := c.acceptExpr(expr.Assignee)

	// Value
	value := c.loadExpr(expr.Value)

	if expr.Op.Kind != scanner.Equal {
		value = c.binary(
			expr.Op,
			c.load(assignee),
			value,
		)
	}

	// Store
	c.block.Store(assignee.v, value.v).SetLocation(expr.Token())

	c.exprResult = assignee
}

func (c *codegen) VisitCast(expr *ast.Cast) {
	value := c.acceptExpr(expr.Expr)

	if from, ok := expr.Expr.Result().Type.(*types.PrimitiveType); ok {
		if to, ok := expr.Result().Type.(*types.PrimitiveType); ok {
			// primitive to primitive
			c.castPrimitiveToPrimitive(value, from, to, from.Kind, to.Kind, expr.Token())
			return
		}
	}

	if from, ok := expr.Expr.Result().Type.(*ast.Enum); ok {
		if to, ok := expr.Result().Type.(*types.PrimitiveType); ok {
			// enum to integer
			c.castPrimitiveToPrimitive(value, from, to, from.Type.(*types.PrimitiveType).Kind, to.Kind, expr.Token())
			return
		}
	}

	if from, ok := expr.Expr.Result().Type.(*types.PrimitiveType); ok {
		if to, ok := expr.Result().Type.(*ast.Enum); ok {
			// integer to enum
			c.castPrimitiveToPrimitive(value, from, to, from.Kind, to.Type.(*types.PrimitiveType).Kind, expr.Token())
			return
		}
	}

	if _, ok := expr.Expr.Result().Type.(*types.PointerType); ok {
		if _, ok := expr.Result().Type.(*types.PointerType); ok {
			// pointer to pointer
			c.exprResult = value
			return
		}
	}

	// Error
	panic("codegen.VisitCast() - Invalid cast")
}

func (c *codegen) castPrimitiveToPrimitive(value exprValue, from, to types.Type, fromKind, toKind types.PrimitiveKind, location llvm.Location) {
	if fromKind == toKind || (types.EqualsPrimitiveCategory(fromKind, toKind) && types.GetBitSize(fromKind) == types.GetBitSize(toKind)) {
		c.exprResult = value
		return
	}

	value = c.load(value)

	if (types.IsInteger(fromKind) || types.IsFloating(fromKind)) && toKind == types.Bool {
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

		if (types.IsInteger(fromKind) || fromKind == types.Bool) && types.IsInteger(toKind) {
			// integer / bool to integer
			if from.Size() > to.Size() {
				kind = llvm.Trunc
			} else {
				kind = llvm.ZExt
			}
		} else if types.IsFloating(fromKind) && types.IsFloating(toKind) {
			// floating to floating
			if from.Size() > to.Size() {
				kind = llvm.FpTrunc
			} else {
				kind = llvm.FpExt
			}
		} else if (types.IsInteger(fromKind) || fromKind == types.Bool) && types.IsFloating(toKind) {
			// integer / bool to floating
			if types.IsSigned(fromKind) {
				kind = llvm.SiToFp
			} else {
				kind = llvm.UiToFp
			}
		} else if types.IsFloating(fromKind) && types.IsInteger(toKind) {
			// floating to integer
			if types.IsSigned(toKind) {
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

func (c *codegen) VisitSizeof(expr *ast.Sizeof) {
	c.exprResult = exprValue{
		v: c.function.Literal(
			c.getType(expr.Result().Type),
			llvm.Literal{Signed: int64(expr.Target.Size())},
		),
	}
}

func (c *codegen) VisitCall(expr *ast.Call) {
	// Get type
	callee := c.acceptExpr(expr.Callee)

	function := expr.Callee.Result().Function

	if f, ok := expr.Callee.Result().Type.(*ast.Func); ok && function == nil {
		function = f
		callee = c.load(callee)
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
	if function.IsIntrinsic() {
		args = c.modifyIntrinsicArgs(function, args)
	}

	// Call
	result := c.block.Call(callee.v, args, c.getType(function.Returns))
	result.SetLocation(expr.Token())

	c.exprResult = exprValue{
		v: result,
	}

	// If the function returns a constant-sized array and the array is immediately indexed then store it in an alloca first
	if callNeedsTempVariable(expr) {
		pointer := c.allocas[expr]
		c.block.Store(pointer.v, c.exprResult.v)

		c.exprResult = pointer
	}
}

func (c *codegen) VisitIndex(expr *ast.Index) {
	value := c.acceptExpr(expr.Value)
	index := c.loadExpr(expr.Index)

	if _, ok := expr.Value.Result().Type.(*types.PointerType); ok {
		value = exprValue{
			v: c.block.Load(value.v),
		}
	}

	t := types.PointerType{Pointee: expr.Result().Type}
	type_ := c.getType(&t)

	result := c.block.GetElementPtr(
		value.v,
		[]llvm.Value{index.v},
		type_,
		c.getType(expr.Result().Type),
	)

	result.SetLocation(expr.Token())

	c.exprResult = exprValue{
		v:           result,
		addressable: true,
	}
}

func (c *codegen) VisitMember(expr *ast.Member) {
	value := c.acceptExpr(expr.Value)

	if expr.Value.Result().Kind == ast.TypeResultKind {
		// Type

		if _, ok := expr.Value.Result().Type.(*ast.Struct); ok {
			// Struct
			c.exprResult = c.getFunction(expr.Result().Function)
		} else if v, ok := expr.Value.Result().Type.(*ast.Enum); ok {
			// Enum
			case_ := v.GetCase(expr.Name.Lexeme)

			c.exprResult = exprValue{
				v: c.function.Literal(
					c.getType(v.Type),
					llvm.Literal{Signed: int64(case_.Value), Unsigned: uint64(case_.Value)},
				),
			}
		} else {
			panic("codegen.VisitMember() - Invalid type")
		}
	} else {
		// Member

		// Get struct and load the value if it is a pointer
		var s *ast.Struct

		if v, ok := expr.Value.Result().Type.(*ast.Struct); ok {
			s = v
		} else if v, ok := expr.Value.Result().Type.(*types.PointerType); ok {
			if v, ok := v.Pointee.(*ast.Struct); ok {
				s = v

				value = exprValue{
					v:           c.block.Load(value.v),
					addressable: true,
				}
			}
		}

		if s == nil {
			log.Fatalln("Invalid member value")
		}

		// Method
		if expr.Result().Kind == ast.FunctionResultKind {
			if !value.addressable {
				pointer := c.allocas[expr]
				c.block.Store(pointer.v, value.v)

				value = pointer
			}

			c.exprResult = c.getFunction(expr.Result().Function)
			c.this = value

			return
		}

		// Field
		i, field := s.GetField(expr.Name.Lexeme)

		if value.addressable {
			i32Type_ := types.PrimitiveType{Kind: types.I32}
			i32Type := c.getType(&i32Type_)

			t := types.PointerType{Pointee: field.Type}

			result := c.block.GetElementPtr(
				value.v,
				[]llvm.Value{
					c.function.Literal(i32Type, llvm.Literal{Signed: 0}),
					c.function.Literal(i32Type, llvm.Literal{Signed: int64(i)}),
				},
				c.getType(&t),
				c.getType(s),
			)

			result.SetLocation(expr.Token())

			c.exprResult = exprValue{
				v:           result,
				addressable: true,
			}
		} else {
			result := c.block.ExtractValue(value.v, i)
			result.SetLocation(expr.Token())

			c.exprResult = exprValue{
				v: result,
			}
		}
	}
}

// Utils

func (c *codegen) binary(op scanner.Token, left exprValue, right exprValue) exprValue {
	left = c.load(left)
	right = c.load(right)

	var kind llvm.BinaryKind

	switch op.Kind {
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

	case scanner.Pipe:
		kind = llvm.Or
	case scanner.Ampersand:
		kind = llvm.And
	case scanner.LessLess:
		kind = llvm.Shl
	case scanner.GreaterGreater:
		kind = llvm.Shr

	default:
		panic("codegen.binary() - Invalid operator kind")
	}

	result := c.block.Binary(kind, left.v, right.v)
	result.SetLocation(op)

	return exprValue{v: result}
}
