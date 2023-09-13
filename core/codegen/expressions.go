package codegen

import (
	"fireball/core"
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
	c.exprValue = c.locals.constant(raw, expr.Result().Type)
}

func (c *codegen) VisitInitializer(expr *ast.Initializer) {
	// TODO: Use LLVM's structure literal syntax if possible, aka if all fields are assigned
	struct_, _ := expr.Result().Type.(*ast.Struct)
	type_ := c.getType(expr.Result().Type)

	// Allocate struct
	res := c.locals.unnamed(expr.Result().Type)
	c.writeFmt("%s = alloca %s\n", res, type_)

	// Assign fields
	for _, field := range expr.Fields {
		i, _ := struct_.GetField(field.Name.Lexeme)

		value := c.load(c.acceptExpr(field.Value), field.Value.Result().Type)

		ptr := c.locals.unnamed(expr.Result().Type)
		loc := c.debug.location(field.Name)

		c.writeFmt("%s = getelementptr inbounds %s, ptr %s, i32 0, i32 %d\n", ptr, type_, res, i)
		c.writeFmt("store %s %s, ptr %s, !dbg %s\n", c.getType(field.Value.Result().Type), value, ptr, loc)
	}

	// Return
	res.needsLoading = true
	c.exprValue = res
}

func (c *codegen) VisitUnary(expr *ast.Unary) {
	loc := c.debug.location(expr.Token())
	val := c.acceptExpr(expr.Right)

	switch expr.Op.Kind {
	case scanner.Bang:
		res := c.locals.unnamed(expr.Right.Result().Type)
		c.writeFmt("%s = xor i1 %s, true, !dbg %s\n", res, c.load(val, expr.Right.Result().Type), loc)
		c.exprValue = res

	case scanner.Minus:
		if v, ok := expr.Right.Result().Type.(*types.PrimitiveType); ok {
			res := c.locals.unnamed(expr.Right.Result().Type)
			val := c.load(val, expr.Right.Result().Type)

			if types.IsFloating(v.Kind) {
				// floating
				c.writeFmt("%s = fneg %s %s, !dbg %s\n", res, c.getType(expr.Right.Result().Type), val, loc)
			} else {
				// signed
				c.writeFmt("%s = sub nsw %s 0, %s, !dbg %s\n", res, c.getType(expr.Right.Result().Type), val, loc)
			}

			c.exprValue = res
		} else {
			log.Fatalln("codegen.VisitUnary() - Invalid type")
		}

	case scanner.Ampersand:
		c.exprValue = value{
			identifier: val.identifier,
			type_:      &types.PointerType{Pointee: val.type_},
		}

	case scanner.Star:
		res := c.locals.unnamed(expr.Result().Type)
		c.writeFmt("%s = load %s, ptr %s, !dbg %s\n", res, c.getType(expr.Result().Type), c.load(val, expr.Right.Result().Type), loc)
		c.exprValue = res

	default:
		log.Fatalln("codegen.VisitUnary() - Invalid unary operator")
	}
}

func (c *codegen) VisitBinary(expr *ast.Binary) {
	left := c.acceptExpr(expr.Left)
	right := c.acceptExpr(expr.Right)

	c.exprValue = c.binary(expr.Op, left, expr.Left.Result().Type, right, expr.Right.Result().Type)
}

func (c *codegen) VisitLogical(expr *ast.Logical) {
	loc := c.debug.location(expr.Token())

	left := c.load(c.acceptExpr(expr.Left), expr.Left.Result().Type)
	right := c.load(c.acceptExpr(expr.Right), expr.Right.Result().Type)

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

		res := c.locals.unnamed(expr.Result().Type)
		c.writeFmt("%s = phi i1 [ true, %%%s ], [ %s, %%%s ], !dbg %s\n", res, startBlock, right, false_, loc)

		c.exprValue = res

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

		res := c.locals.unnamed(expr.Result().Type)
		c.writeFmt("%s = phi i1 [ false, %%%s ], [ %s, %%%s ], !dbg %s\n", res, startBlock, right, true_, loc)

		c.exprValue = res

	default:
		log.Fatalln("Invalid logical operator")
	}
}

func (c *codegen) VisitIdentifier(expr *ast.Identifier) {
	switch expr.Kind {
	case ast.FunctionKind:
		if v := c.getFunction(expr.Identifier); v.identifier != "" {
			c.exprValue = v
			return
		}

	case ast.EnumKind:
		if expr.Kind == ast.EnumKind {
			c.exprValue = value{identifier: "$enum$"}
			return
		}

	case ast.VariableKind, ast.ParameterKind:
		if v := c.getVariable(expr.Identifier); v != nil {
			c.exprValue = v.val
			return
		}
	}

	log.Fatalln("Invalid identifier")
}

func (c *codegen) VisitAssignment(expr *ast.Assignment) {
	// Assignee
	assignee := c.acceptExpr(expr.Assignee)

	// Value
	val := c.load(c.acceptExpr(expr.Value), expr.Value.Result().Type)

	if expr.Op.Kind != scanner.Equal {
		val = c.binary(expr.Op, c.load(assignee, expr.Assignee.Result().Type), expr.Assignee.Result().Type, val, expr.Value.Result().Type)
	}

	// Store
	loc := c.debug.location(expr.Token())
	c.writeFmt("store %s %s, ptr %s, !dbg %s\n", c.getType(expr.Value.Result().Type), val, assignee, loc)

	c.exprValue = assignee
}

func (c *codegen) VisitCast(expr *ast.Cast) {
	loc := c.debug.location(expr.Token())
	val := c.acceptExpr(expr.Expr)

	if from, ok := expr.Expr.Result().Type.(*types.PrimitiveType); ok {
		if to, ok := expr.Result().Type.(*types.PrimitiveType); ok {
			// primitive to primitive
			c.castPrimitiveToPrimitive(val, loc, from, to, from.Kind, to.Kind)
			return
		}
	}

	if from, ok := expr.Expr.Result().Type.(*ast.Enum); ok {
		if to, ok := expr.Result().Type.(*types.PrimitiveType); ok {
			// enum to integer
			c.castPrimitiveToPrimitive(val, loc, from, to, from.Type.(*types.PrimitiveType).Kind, to.Kind)
			return
		}
	}

	if from, ok := expr.Expr.Result().Type.(*types.PrimitiveType); ok {
		if to, ok := expr.Result().Type.(*ast.Enum); ok {
			// integer to enum
			c.castPrimitiveToPrimitive(val, loc, from, to, from.Kind, to.Type.(*types.PrimitiveType).Kind)
			return
		}
	}

	if _, ok := expr.Expr.Result().Type.(*types.PointerType); ok {
		if _, ok := expr.Result().Type.(*types.PointerType); ok {
			// pointer to pointer
			c.exprValue = val
			return
		}
	}

	// Error
	log.Fatalln("Invalid cast")
}

func (c *codegen) castPrimitiveToPrimitive(val value, loc string, from, to types.Type, fromKind, toKind types.PrimitiveKind) {
	if fromKind == toKind {
		c.exprValue = val
		return
	}

	val = c.load(val, from)

	res := c.locals.unnamed(to)
	c.exprValue = res

	if (types.IsInteger(fromKind) || fromKind == types.Bool) && types.IsInteger(toKind) {
		// integer / bool to integer
		if from.Size() > to.Size() {
			c.writeFmt("%s = trunc %s %s to %s, !dbg %s\n", res, c.getType(from), val, c.getType(to), loc)
		} else {
			c.writeFmt("%s = zext %s %s to %s, !dbg %s\n", res, c.getType(from), val, c.getType(to), loc)
		}
	} else if types.IsFloating(fromKind) && types.IsFloating(toKind) {
		// floating to floating
		if from.Size() > to.Size() {
			c.writeFmt("%s = fptrunc %s %s to %s, !dbg %s\n", res, c.getType(from), val, c.getType(to), loc)
		} else {
			c.writeFmt("%s = fpext %s %s to %s, !dbg %s\n", res, c.getType(from), val, c.getType(to), loc)
		}
	} else if (types.IsInteger(fromKind) || fromKind == types.Bool) && types.IsFloating(toKind) {
		// integer / bool to floating
		if types.IsSigned(fromKind) {
			c.writeFmt("%s = sitofp %s %s to %s, !dbg %s\n", res, c.getType(from), val, c.getType(to), loc)
		} else {
			c.writeFmt("%s = uitofp %s %s to %s, !dbg %s\n", res, c.getType(from), val, c.getType(to), loc)
		}
	} else if types.IsFloating(fromKind) && types.IsInteger(toKind) {
		// floating to integer
		if types.IsSigned(toKind) {
			c.writeFmt("%s = fptosi %s %s to %s, !dbg %s\n", res, c.getType(from), val, c.getType(to), loc)
		} else {
			c.writeFmt("%s = fptoui %s %s to %s, !dbg %s\n", res, c.getType(from), val, c.getType(to), loc)
		}
	} else if types.IsInteger(fromKind) && toKind == types.Bool {
		// integer to bool
		c.writeFmt("%s = icmp ne %s %s, 0, !dbg %s\n", res, c.getType(from), val, loc)
	} else if types.IsFloating(fromKind) && toKind == types.Bool {
		// floating to bool
		c.writeFmt("%s = fcmp une %s %s, 0, !dbg %s\n", res, c.getType(from), val, loc)
	}
}

func (c *codegen) VisitCall(expr *ast.Call) {
	var f *ast.Func

	if v, ok := expr.Callee.Result().Type.(*ast.Func); ok {
		f = v
	}

	args := make([]value, len(expr.Args))

	for i, arg := range expr.Args {
		args[i] = c.load(c.acceptExpr(arg), arg.Result().Type)
	}

	builder := strings.Builder{}

	type_ := c.getType(expr.Result().Type)
	callee := c.acceptExpr(expr.Callee)

	if types.IsPrimitive(f.Returns, types.Void) {
		builder.WriteString(fmt.Sprintf("call %s %s(", type_, callee))

		c.exprValue = value{
			identifier: "",
			type_:      types.Primitive(types.Void, core.Range{}),
		}
	} else {
		val := c.locals.unnamed(expr.Result().Type)
		builder.WriteString(fmt.Sprintf("%s = call %s %s(", val, type_, callee))
		c.exprValue = val
	}

	for i, arg := range args {
		if i > 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(fmt.Sprintf("%s %s", c.getType(expr.Args[i].Result().Type), arg))
	}

	builder.WriteString("), !dbg ")
	builder.WriteString(c.debug.location(expr.Token()))
	builder.WriteRune('\n')
	c.writeStr(builder.String())
}

func (c *codegen) VisitIndex(expr *ast.Index) {
	val := c.toPtrOrLoad(c.acceptExpr(expr.Value), expr.Value.Result().Type)
	index := c.load(c.acceptExpr(expr.Index), expr.Index.Result().Type)

	if _, ok := expr.Value.Result().Type.(*types.PointerType); ok {
		res := c.locals.unnamed(val.type_)
		c.writeFmt("%s = load ptr, ptr %s\n", res, val)

		val = res
	}

	type_ := c.getType(expr.Result().Type)

	res := c.locals.unnamed(expr.Result().Type)
	res.needsLoading = true

	loc := c.debug.location(expr.Token())
	c.writeFmt("%s = getelementptr inbounds %s, %s %s, %s %s, !debg %s\n", res, type_, c.getType(val.type_), val, c.getType(expr.Index.Result().Type), index, loc)

	c.exprValue = res
}

func (c *codegen) VisitMember(expr *ast.Member) {
	value := c.acceptExpr(expr.Value)

	if value.identifier == "$enum$" {
		// Enum
		case_ := expr.Value.Result().Type.(*ast.Enum).GetCase(expr.Name.Lexeme)
		c.exprValue = c.locals.constant(strconv.Itoa(case_.Value), expr.Result().Type)
	} else {
		// Member
		val := c.toPtrOrLoad(value, expr.Value.Result().Type)

		var s *ast.Struct

		if v, ok := expr.Value.Result().Type.(*ast.Struct); ok {
			s = v
		} else if v, ok := expr.Value.Result().Type.(*types.PointerType); ok {
			if v, ok := v.Pointee.(*ast.Struct); ok {
				s = v

				res := c.locals.unnamed(val.type_)
				c.writeFmt("%s = load ptr, ptr %s\n", res, val)

				val = res
			}
		}

		if s == nil {
			log.Fatalln("Invalid member value")
		}

		i, _ := s.GetField(expr.Name.Lexeme)

		res := c.locals.unnamed(expr.Result().Type)
		res.needsLoading = true

		loc := c.debug.location(expr.Token())
		c.writeFmt("%s = getelementptr inbounds %s, %s %s, i32 0, i32 %d, !dbg %s\n", res, c.getType(s), c.getType(val.type_), val, i, loc)

		c.exprValue = res
	}
}

// Utils

func (c *codegen) binary(op scanner.Token, left value, leftType types.Type, right value, rightType types.Type) value {
	// Load arguments in case they are pointers
	left = c.load(left, leftType)
	right = c.load(right, rightType)

	// Check for floating point numbers and sign
	floating := false
	signed := false

	if v, ok := left.type_.(*types.PrimitiveType); ok {
		floating = types.IsFloating(v.Kind)
		signed = types.IsSigned(v.Kind)
	}

	// Select correct instruction
	inst := ""

	switch op.Kind {
	case scanner.Plus, scanner.PlusEqual:
		inst = ternary(floating, "fadd", "add")
	case scanner.Minus, scanner.MinusEqual:
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
	val := c.locals.unnamed(left.type_)

	loc := c.debug.location(op)
	c.writeFmt("%s = %s %s %s, %s, !dbg %s\n", val, inst, c.getType(left.type_), left, right, loc)

	return val
}
