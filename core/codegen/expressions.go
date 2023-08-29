package codegen

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fmt"
	"log"
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

	case scanner.True, scanner.False, scanner.Number:
		raw = expr.Value.Lexeme

	case scanner.Character:
		c := expr.Value.Lexeme[1 : len(expr.Value.Lexeme)-1]
		var char uint8

		switch c {
		case "'":
			char = '\''
		case "\\0":
			char = '\000'

		case "\n":
			char = '\n'
		case "\r":
			char = '\r'
		case "\t":
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
	c.exprValue = c.locals.constant(raw, expr.Type())
}

func (c *codegen) VisitUnary(expr *ast.Unary) {
	c.acceptExpr(expr.Right)

	switch expr.Op.Kind {
	case scanner.Bang:
		res := c.locals.unnamed(expr.Right.Type())
		c.writeFmt("%s = xor i1 %s, true\n", res, c.load(c.exprValue, expr.Right.Type()))
		c.exprValue = res

	case scanner.Minus:
		if v, ok := expr.Right.Type().(*types.PrimitiveType); ok {
			res := c.locals.unnamed(expr.Right.Type())
			val := c.load(c.exprValue, expr.Right.Type())

			if types.IsFloating(v.Kind) {
				// floating
				c.writeFmt("%s = fneg %s %s\n", res, c.getType(expr.Right.Type()), val)
			} else {
				// signed
				c.writeFmt("%s = sub nsw %s 0, %s\n", res, c.getType(expr.Right.Type()), val)
			}

			c.exprValue = res
		} else {
			log.Fatalln("Invalid type")
		}

	case scanner.Ampersand:
		c.exprValue = value{
			identifier: c.exprValue.identifier,
			type_:      types.PointerType{Pointee: c.exprValue.type_},
		}

	default:
		log.Fatalln("Invalid unary operator")
	}
}

func (c *codegen) VisitBinary(expr *ast.Binary) {
	left := c.acceptExpr(expr.Left)
	right := c.acceptExpr(expr.Right)

	c.exprValue = c.binary(expr.Op.Kind, left, expr.Left.Type(), right, expr.Right.Type())
}

func (c *codegen) VisitIdentifier(expr *ast.Identifier) {
	// Functions
	f := c.getFunction(expr.Identifier)

	if f != nil {
		c.exprValue = f.val
		return
	}

	// Variables
	v := c.getVariable(expr.Identifier)

	if v != nil {
		c.exprValue = v.val
		return
	}

	// Error
	log.Fatalln("Invalid identifier")
}

func (c *codegen) VisitAssignment(expr *ast.Assignment) {
	// Assignee
	assignee := c.acceptExpr(expr.Assignee)

	// Value
	val := c.load(c.acceptExpr(expr.Value), expr.Value.Type())

	if expr.Op.Kind != scanner.Equal {
		val = c.binary(expr.Op.Kind, c.load(assignee, expr.Assignee.Type()), expr.Assignee.Type(), val, expr.Value.Type())
	}

	// Store
	c.writeFmt("store %s %s, ptr %s\n", c.getType(expr.Value.Type()), val, assignee)
	c.exprValue = assignee
}

func (c *codegen) VisitCast(expr *ast.Cast) {
	c.acceptExpr(expr.Expr)
	val := c.load(c.exprValue, expr.Expr.Type())

	res := c.locals.unnamed(expr.Type())
	c.exprValue = res

	if from, ok := expr.Expr.Type().(*types.PrimitiveType); ok {
		if to, ok := expr.Type().(*types.PrimitiveType); ok {
			if (types.IsInteger(from.Kind) || from.Kind == types.Bool) && types.IsInteger(to.Kind) {
				// integer / bool to integer
				if from.Size() > to.Size() {
					c.writeFmt("%s = trunc %s %s to %s\n", res, c.getType(from), val, c.getType(to))
				} else {
					c.writeFmt("%s = zext %s %s to %s\n", res, c.getType(from), val, c.getType(to))
				}

				return
			} else if types.IsFloating(from.Kind) && types.IsFloating(to.Kind) {
				// floating to floating
				if from.Size() > to.Size() {
					c.writeFmt("%s = fptrunc %s %s to %s\n", res, c.getType(from), val, c.getType(to))
				} else {
					c.writeFmt("%s = fpext %s %s to %s\n", res, c.getType(from), val, c.getType(to))
				}

				return
			} else if (types.IsInteger(from.Kind) || from.Kind == types.Bool) && types.IsFloating(to.Kind) {
				// integer / bool to floating
				if types.IsSigned(from.Kind) {
					c.writeFmt("%s = sitofp %s %s to %s\n", res, c.getType(from), val, c.getType(to))
				} else {
					c.writeFmt("%s = uitofp %s %s to %s\n", res, c.getType(from), val, c.getType(to))
				}

				return
			} else if types.IsFloating(from.Kind) && types.IsInteger(to.Kind) {
				// floating to integer
				if types.IsSigned(to.Kind) {
					c.writeFmt("%s = fptosi %s %s to %s\n", res, c.getType(from), val, c.getType(to))
				} else {
					c.writeFmt("%s = fptoui %s %s to %s\n", res, c.getType(from), val, c.getType(to))
				}

				return
			} else if types.IsInteger(from.Kind) && to.Kind == types.Bool {
				// integer to bool
				c.writeFmt("%s = icmp ne %s %s, 0\n", res, c.getType(from), val)
				return
			} else if types.IsFloating(from.Kind) && to.Kind == types.Bool {
				// floating to bool
				c.writeFmt("%s = fcmp une %s %s, 0\n", res, c.getType(from), val)
				return
			}
		}
	}

	if _, ok := expr.Expr.Type().(types.PointerType); ok {
		if _, ok := expr.Type().(types.PointerType); ok {
			// pointer to pointer
			return
		}
	}

	// Error
	log.Fatalln("Invalid cast")
}

func (c *codegen) VisitCall(expr *ast.Call) {
	var f types.FunctionType

	if v, ok := expr.Callee.Type().(types.FunctionType); ok {
		f = v
	}

	args := make([]value, len(expr.Args))

	for i, arg := range expr.Args {
		args[i] = c.load(c.acceptExpr(arg), arg.Type())
	}

	builder := strings.Builder{}

	type_ := c.getType(expr.Type())
	callee := c.acceptExpr(expr.Callee)

	if types.IsPrimitive(f.Returns, types.Void) {
		builder.WriteString(fmt.Sprintf("call %s %s(", type_, callee))

		c.exprValue = value{
			identifier: "",
			type_:      types.Primitive(types.Void),
		}
	} else {
		val := c.locals.unnamed(expr.Type())
		builder.WriteString(fmt.Sprintf("%s = call %s %s(", val, type_, callee))
		c.exprValue = val
	}

	for i, arg := range args {
		if i > 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(fmt.Sprintf("%s %s", c.getType(expr.Args[i].Type()), arg))
	}

	builder.WriteString(")\n")
	c.writeStr(builder.String())
}

func (c *codegen) VisitIndex(expr *ast.Index) {
	val := c.toPtrOrLoad(c.acceptExpr(expr.Value), expr.Value.Type())
	index := c.load(c.acceptExpr(expr.Index), expr.Index.Type())

	type_ := c.getType(expr.Type())

	res := c.locals.unnamed(expr.Type())
	res.needsLoading = true

	c.writeFmt("%s = getelementptr inbounds %s, %s %s, %s %s\n", res, type_, c.getType(val.type_), val, c.getType(expr.Index.Type()), index)

	c.exprValue = res
}

// Utils

func (c *codegen) binary(op scanner.TokenKind, left value, leftType types.Type, right value, rightType types.Type) value {
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

	switch op {
	case scanner.Plus, scanner.PlusEqual:
		inst = ternary(floating, "fadd", "add")
	case scanner.Minus, scanner.MinusEqual:
		inst = ternary(floating, "fsub", "sub")
	case scanner.Star, scanner.StarEqual:
		inst = ternary(floating, "fmul", "mul")
	case scanner.Slash, scanner.SlashEqual:
		inst = ternary(floating, "fdiv", "div")
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

	default:
		log.Fatalln("Invalid operator kind")
	}

	// Emit
	val := c.locals.unnamed(left.type_)
	c.writeFmt("%s = %s %s %s, %s\n", val, inst, c.getType(left.type_), left, right)

	return val
}
