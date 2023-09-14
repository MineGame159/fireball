package codegen

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

func (c *codegen) VisitGroup(expr *ast.Group) {
	c.acceptExpr(expr.Expr)
}

func (c *codegen) VisitLiteral(expr *ast.Literal) {
	// Convert fireball constant into a LLVM IR constant
	raw := ""

	switch expr.Value.Kind {
	case scanner.Nil:
		raw = "null"

	case scanner.True, scanner.False:
		raw = expr.Value.Lexeme

	case scanner.Number:
		raw = expr.Value.Lexeme
		last := raw[len(raw)-1]

		if last == 'f' || last == 'F' {
			v, _ := strconv.ParseFloat(raw[:len(raw)-1], 32)
			raw = fmt.Sprintf("0x%X", math.Float64bits(v))
		} else if strings.ContainsRune(raw, '.') {
			v, _ := strconv.ParseFloat(raw, 64)
			raw = fmt.Sprintf("0x%X", math.Float64bits(v))
		}

	case scanner.Hex:
		v, _ := strconv.ParseUint(expr.Value.Lexeme[2:], 16, 64)
		raw = strconv.FormatUint(v, 10)

	case scanner.Binary:
		v, _ := strconv.ParseUint(expr.Value.Lexeme[2:], 2, 64)
		raw = strconv.FormatUint(v, 10)

	case scanner.Character:
		c := expr.Value.Lexeme[1 : len(expr.Value.Lexeme)-1]
		var char uint8

		switch c {
		case "'":
			char = '\''
		case "\\0":
			char = '\000'

		case "\\n":
			char = '\n'
		case "\\r":
			char = '\r'
		case "\\t":
			char = '\t'

		default:
			char = c[0]
		}

		raw = strconv.Itoa(int(char))

	case scanner.String:
		raw = c.getConstant(expr.Value.Lexeme[1 : len(expr.Value.Lexeme)-1])

	default:
		log.Fatalln("Invalid literal kind")
	}

	// Emit
	c.exprResult = c.locals.constant(raw)
}

func (c *codegen) VisitInitializer(expr *ast.Initializer) {
	struct_, _ := expr.Result().Type.(*ast.Struct)
	type_ := c.getType(expr.Result().Type)

	result := c.locals.constant("zeroinitializer")

	for _, field := range expr.Fields {
		loc := c.debug.location(field.Name)
		newResult := c.locals.unnamed()

		value, valueType := c.loadExpr(field.Value)
		i, _ := struct_.GetField(field.Name.Lexeme)

		c.writeFmt("%s = insertvalue %s %s, %s %s, %d, !dbg %s\n", newResult, type_, result, valueType, value, i, loc)

		result = newResult
	}

	c.exprResult = result
}

func (c *codegen) VisitUnary(expr *ast.Unary) {
	loc := c.debug.location(expr.Token())
	value := c.acceptExpr(expr.Value)

	if expr.Prefix {
		// Prefix
		switch expr.Op.Kind {
		case scanner.Bang:
			result := c.locals.unnamed()
			c.writeFmt("%s = xor i1 %s, true, !dbg %s\n", result, c.load(value, expr.Value.Result().Type), loc)
			c.exprResult = result

		case scanner.Minus:
			if v, ok := expr.Value.Result().Type.(*types.PrimitiveType); ok {
				result := c.locals.unnamed()
				value := c.load(value, expr.Value.Result().Type)

				if types.IsFloating(v.Kind) {
					// floating
					c.writeFmt("%s = fneg %s %s, !dbg %s\n", result, c.getType(expr.Value.Result().Type), value, loc)
				} else {
					// signed
					c.writeFmt("%s = sub nsw %s 0, %s, !dbg %s\n", result, c.getType(expr.Value.Result().Type), value, loc)
				}

				c.exprResult = result
			} else {
				log.Fatalln("codegen.VisitUnary() - Invalid type")
			}

		case scanner.Ampersand:
			c.exprResult = exprValue{
				identifier:  value.identifier,
				addressable: true,
			}

		case scanner.Star:
			result := c.locals.unnamed()
			c.writeFmt("%s = load %s, ptr %s, !dbg %s\n", result, c.getType(expr.Result().Type), c.load(value, expr.Value.Result().Type), loc)
			c.exprResult = result

		case scanner.PlusPlus, scanner.MinusMinus:
			newValue := c.binary(expr.Op, value, expr.Value.Result().Type, c.globals.constant("1"), expr.Value.Result().Type)
			c.writeFmt("store %s %s, ptr %s, !dbg %s\n", c.getType(expr.Value.Result().Type), newValue, value, loc)

			c.exprResult = newValue

		default:
			panic("codegen.VisitUnary() - Invalid unary prefix operator")
		}
	} else {
		// Postfix
		switch expr.Op.Kind {
		case scanner.PlusPlus, scanner.MinusMinus:
			prevValue := c.load(value, expr.Value.Result().Type)

			newValue := c.binary(expr.Op, prevValue, expr.Value.Result().Type, c.globals.constant("1"), expr.Value.Result().Type)
			c.writeFmt("store %s %s, ptr %s, !dbg %s\n", c.getType(expr.Value.Result().Type), newValue, value, loc)

			c.exprResult = prevValue

		default:
			panic("codegen.VisitUnary() - Invalid unary prefix operator")
		}
	}
}

func (c *codegen) VisitBinary(expr *ast.Binary) {
	left := c.acceptExpr(expr.Left)
	right := c.acceptExpr(expr.Right)

	c.exprResult = c.binary(expr.Op, left, expr.Left.Result().Type, right, expr.Right.Result().Type)
}

func (c *codegen) VisitLogical(expr *ast.Logical) {
	loc := c.debug.location(expr.Token())

	left, _ := c.loadExpr(expr.Left)
	right, _ := c.loadExpr(expr.Right)

	switch expr.Op.Kind {
	case scanner.Or:
		false_ := c.blocks.unnamedRaw()
		end := c.blocks.unnamedRaw()

		// Start
		startBlock := c.block
		c.writeFmt("br i1 %s, label %%%s, label %%%s\n", left, end, false_)

		// False
		c.writeBlock(false_)
		c.writeFmt("br label %%%s\n", end)

		// End
		c.writeBlock(end)

		result := c.locals.unnamed()
		c.writeFmt("%s = phi i1 [ true, %%%s ], [ %s, %%%s ], !dbg %s\n", result, startBlock, right, false_, loc)

		c.exprResult = result

	case scanner.And:
		true_ := c.blocks.unnamedRaw()
		end := c.blocks.unnamedRaw()

		// Start
		startBlock := c.block
		c.writeFmt("br i1 %s, label %%%s, label %%%s\n", left, true_, end)

		// True
		c.writeBlock(true_)
		c.writeFmt("br label %%%s\n", end)

		// End
		c.writeBlock(end)

		result := c.locals.unnamed()
		c.writeFmt("%s = phi i1 [ false, %%%s ], [ %s, %%%s ], !dbg %s\n", result, startBlock, right, true_, loc)

		c.exprResult = result

	default:
		log.Fatalln("Invalid logical operator")
	}
}

func (c *codegen) VisitIdentifier(expr *ast.Identifier) {
	switch expr.Kind {
	case ast.FunctionKind:
		if v := c.getFunction(expr.Identifier); v.identifier != "" {
			c.exprResult = v
			return
		}

	case ast.EnumKind:
		if expr.Kind == ast.EnumKind {
			c.exprResult = exprValue{identifier: "$enum$"}
			return
		}

	case ast.VariableKind, ast.ParameterKind:
		if v := c.getVariable(expr.Identifier); v != nil {
			c.exprResult = v.value
			return
		}
	}

	log.Fatalln("Invalid identifier")
}

func (c *codegen) VisitAssignment(expr *ast.Assignment) {
	// Assignee
	assignee := c.acceptExpr(expr.Assignee)

	// Value
	value, valueType := c.loadExpr(expr.Value)

	if expr.Op.Kind != scanner.Equal {
		value = c.binary(expr.Op, c.load(assignee, expr.Assignee.Result().Type), expr.Assignee.Result().Type, value, expr.Value.Result().Type)
	}

	// Store
	loc := c.debug.location(expr.Token())
	c.writeFmt("store %s %s, ptr %s, !dbg %s\n", valueType, value, assignee, loc)

	c.exprResult = assignee
}

func (c *codegen) VisitCast(expr *ast.Cast) {
	loc := c.debug.location(expr.Token())
	value := c.acceptExpr(expr.Expr)

	if from, ok := expr.Expr.Result().Type.(*types.PrimitiveType); ok {
		if to, ok := expr.Result().Type.(*types.PrimitiveType); ok {
			// primitive to primitive
			c.castPrimitiveToPrimitive(value, loc, from, to, from.Kind, to.Kind)
			return
		}
	}

	if from, ok := expr.Expr.Result().Type.(*ast.Enum); ok {
		if to, ok := expr.Result().Type.(*types.PrimitiveType); ok {
			// enum to integer
			c.castPrimitiveToPrimitive(value, loc, from, to, from.Type.(*types.PrimitiveType).Kind, to.Kind)
			return
		}
	}

	if from, ok := expr.Expr.Result().Type.(*types.PrimitiveType); ok {
		if to, ok := expr.Result().Type.(*ast.Enum); ok {
			// integer to enum
			c.castPrimitiveToPrimitive(value, loc, from, to, from.Kind, to.Type.(*types.PrimitiveType).Kind)
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
	log.Fatalln("Invalid cast")
}

func (c *codegen) castPrimitiveToPrimitive(value exprValue, loc string, from, to types.Type, fromKind, toKind types.PrimitiveKind) {
	if fromKind == toKind {
		c.exprResult = value
		return
	}

	value = c.load(value, from)

	result := c.locals.unnamed()
	c.exprResult = result

	if (types.IsInteger(fromKind) || fromKind == types.Bool) && types.IsInteger(toKind) {
		// integer / bool to integer
		if from.Size() > to.Size() {
			c.writeFmt("%s = trunc %s %s to %s, !dbg %s\n", result, c.getType(from), value, c.getType(to), loc)
		} else {
			c.writeFmt("%s = zext %s %s to %s, !dbg %s\n", result, c.getType(from), value, c.getType(to), loc)
		}
	} else if types.IsFloating(fromKind) && types.IsFloating(toKind) {
		// floating to floating
		if from.Size() > to.Size() {
			c.writeFmt("%s = fptrunc %s %s to %s, !dbg %s\n", result, c.getType(from), value, c.getType(to), loc)
		} else {
			c.writeFmt("%s = fpext %s %s to %s, !dbg %s\n", result, c.getType(from), value, c.getType(to), loc)
		}
	} else if (types.IsInteger(fromKind) || fromKind == types.Bool) && types.IsFloating(toKind) {
		// integer / bool to floating
		if types.IsSigned(fromKind) {
			c.writeFmt("%s = sitofp %s %s to %s, !dbg %s\n", result, c.getType(from), value, c.getType(to), loc)
		} else {
			c.writeFmt("%s = uitofp %s %s to %s, !dbg %s\n", result, c.getType(from), value, c.getType(to), loc)
		}
	} else if types.IsFloating(fromKind) && types.IsInteger(toKind) {
		// floating to integer
		if types.IsSigned(toKind) {
			c.writeFmt("%s = fptosi %s %s to %s, !dbg %s\n", result, c.getType(from), value, c.getType(to), loc)
		} else {
			c.writeFmt("%s = fptoui %s %s to %s, !dbg %s\n", result, c.getType(from), value, c.getType(to), loc)
		}
	} else if types.IsInteger(fromKind) && toKind == types.Bool {
		// integer to bool
		c.writeFmt("%s = icmp ne %s %s, 0, !dbg %s\n", result, c.getType(from), value, loc)
	} else if types.IsFloating(fromKind) && toKind == types.Bool {
		// floating to bool
		c.writeFmt("%s = fcmp une %s %s, 0, !dbg %s\n", result, c.getType(from), value, loc)
	}
}

func (c *codegen) VisitCall(expr *ast.Call) {
	var f *ast.Func

	if v, ok := expr.Callee.Result().Type.(*ast.Func); ok {
		f = v
	}

	args := make([]struct {
		value exprValue
		type_ string
	}, len(expr.Args))

	for i, arg := range expr.Args {
		value, type_ := c.loadExpr(arg)

		args[i].value = value
		args[i].type_ = type_
	}

	builder := strings.Builder{}

	type_ := c.getType(expr.Result().Type)
	callee := c.acceptExpr(expr.Callee)
	loc := c.debug.location(expr.Token())
	hasReturnValue := false

	if _, ok := expr.Parent().(*ast.Expression); !ok && !types.IsPrimitive(f.Returns, types.Void) {
		result := c.locals.unnamed()

		builder.WriteString(fmt.Sprintf("%s = call %s %s(", result, type_, callee))

		c.exprResult = result
		hasReturnValue = true
	} else {
		builder.WriteString(fmt.Sprintf("call %s %s(", type_, callee))
		c.exprResult = exprValue{identifier: ""}
	}

	for i, arg := range args {
		if i > 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(fmt.Sprintf("%s %s", arg.type_, arg.value))
	}

	builder.WriteString("), !dbg ")
	builder.WriteString(loc)
	builder.WriteRune('\n')
	c.writeStr(builder.String())

	// If the function returns a constant-sized array and the array is immediately indexed then store it in an alloca first
	if hasReturnValue {
		if _, ok := expr.Result().Type.(*types.ArrayType); ok {
			if _, ok := expr.Parent().(*ast.Index); ok {
				result := c.locals.unnamed()
				result.addressable = true

				c.writeFmt("%s = alloca %s, !dbg %s\n", result, type_, loc)
				c.writeFmt("store %s %s, ptr %s, !dbg %s\n", type_, c.exprResult, result, loc)

				c.exprResult = result
			}
		}
	}
}

func (c *codegen) VisitIndex(expr *ast.Index) {
	// TODO: Does not support non-addressable values
	value := c.acceptExpr(expr.Value)
	index, indexType := c.loadExpr(expr.Index)

	if _, ok := expr.Value.Result().Type.(*types.PointerType); ok {
		res := c.locals.unnamed()
		c.writeFmt("%s = load ptr, ptr %s\n", res, value)

		value = res
	}

	type_ := c.getType(expr.Result().Type)

	result := c.locals.unnamed()
	result.addressable = true

	loc := c.debug.location(expr.Token())
	c.writeFmt("%s = getelementptr inbounds %s, ptr %s, %s %s, !debg %s\n", result, type_, value, indexType, index, loc)

	c.exprResult = result
}

func (c *codegen) VisitMember(expr *ast.Member) {
	value := c.acceptExpr(expr.Value)

	if value.identifier == "$enum$" {
		// Enum
		case_ := expr.Value.Result().Type.(*ast.Enum).GetCase(expr.Name.Lexeme)
		c.exprResult = c.locals.constant(strconv.Itoa(case_.Value))
	} else {
		// Member

		// Get struct and load the value if it is a pointer
		var s *ast.Struct

		if v, ok := expr.Value.Result().Type.(*ast.Struct); ok {
			s = v
		} else if v, ok := expr.Value.Result().Type.(*types.PointerType); ok {
			if v, ok := v.Pointee.(*ast.Struct); ok {
				s = v

				result := c.locals.unnamed()
				result.addressable = true

				c.writeFmt("%s = load ptr, ptr %s\n", result, value)

				value = result
			}
		}

		if s == nil {
			log.Fatalln("Invalid member value")
		}

		i, _ := s.GetField(expr.Name.Lexeme)

		result := c.locals.unnamed()
		loc := c.debug.location(expr.Token())

		if value.addressable {
			c.writeFmt("%s = getelementptr inbounds %s, ptr %s, i32 0, i32 %d, !dbg %s\n", result, c.getType(s), value, i, loc)
			result.addressable = true
		} else {
			c.writeFmt("%s = extractvalue %s %s, %d, !dbg %s\n", result, c.getType(s), value, i, loc)
		}

		c.exprResult = result
	}
}

// Utils

func (c *codegen) binary(op scanner.Token, left exprValue, leftType types.Type, right exprValue, rightType types.Type) exprValue {
	// Load arguments in case they are pointers
	left = c.load(left, leftType)
	right = c.load(right, rightType)

	// Check for floating point numbers and sign
	floating := false
	signed := false

	if v, ok := leftType.(*types.PrimitiveType); ok {
		floating = types.IsFloating(v.Kind)
		signed = types.IsSigned(v.Kind)
	}

	// Select correct instruction
	inst := ""

	switch op.Kind {
	case scanner.Plus, scanner.PlusEqual, scanner.PlusPlus:
		inst = ternary(floating, "fadd", "add")
	case scanner.Minus, scanner.MinusEqual, scanner.MinusMinus:
		inst = ternary(floating, "fsub", "sub")
	case scanner.Star, scanner.StarEqual:
		inst = ternary(floating, "fmul", "mul")
	case scanner.Slash, scanner.SlashEqual:
		inst = ternary(floating, "fdiv", ternary(signed, "sdiv", "udiv"))
	case scanner.Percentage, scanner.PercentageEqual:
		inst = ternary(floating, "frem", ternary(signed, "srem", "urem"))

	case scanner.EqualEqual:
		inst = ternary(floating, "fcmp oeq", "icmp eq")
	case scanner.BangEqual:
		inst = ternary(floating, "fcmp one", "icmp ne")

	case scanner.Less:
		inst = ternary(floating, "fcmp olt", ternary(signed, "icmp slt", "icmp ult"))
	case scanner.LessEqual:
		inst = ternary(floating, "fcmp ole", ternary(signed, "icmp sle", "icmp ule"))
	case scanner.Greater:
		inst = ternary(floating, "fcmp ogt", ternary(signed, "icmp sgt", "icmp ugt"))
	case scanner.GreaterEqual:
		inst = ternary(floating, "fcmp oge", ternary(signed, "icmp sge", "icmp uge"))

	case scanner.Pipe:
		inst = "or"
	case scanner.Ampersand:
		inst = "and"
	case scanner.LessLess:
		inst = "shl"
	case scanner.GreaterGreater:
		inst = ternary(signed, "ashr", "lshr")

	default:
		log.Fatalln("Invalid operator kind")
	}

	// Emit
	result := c.locals.unnamed()

	loc := c.debug.location(op)
	c.writeFmt("%s = %s %s %s, %s, !dbg %s\n", result, inst, c.getType(leftType), left, right, loc)

	return result
}
