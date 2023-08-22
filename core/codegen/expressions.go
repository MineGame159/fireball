package codegen

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fmt"
	"log"
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

	default:
		log.Fatalln("Invalid literal kind")
	}

	// Emit
	c.exprValue = c.locals.constant(raw, expr.Type())
}

func (c *codegen) VisitUnary(expr *ast.Unary) {
}

func (c *codegen) VisitBinary(expr *ast.Binary) {
	left := c.acceptExpr(expr.Left)
	right := c.acceptExpr(expr.Right)

	c.exprValue = c.binary(expr.Op.Kind, left, right)
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
	val := c.load(c.acceptExpr(expr.Value))
	val = c.binary(expr.Op.Kind, assignee, val)

	// Store
	c.writeFmt("store %s %s, ptr %s\n", c.getType(val.type_), val, assignee)
	c.exprValue = assignee
}

func (c *codegen) VisitCall(expr *ast.Call) {
	args := make([]value, len(expr.Args))

	for i, arg := range expr.Args {
		args[i] = c.acceptExpr(arg)
	}

	val := c.locals.unnamed(expr.Type())
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s = call %s %s(", val, c.getType(expr.Type()), c.acceptExpr(expr.Callee)))

	for i, arg := range args {
		if i > 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(fmt.Sprintf("%s %s", c.getType(arg.type_), arg))
	}

	builder.WriteString(")\n")
	c.writeStr(builder.String())

	c.exprValue = val
}

// Utils

func (c *codegen) binary(op scanner.TokenKind, left value, right value) value {
	// Load arguments in case they are pointers
	left = c.load(left)
	right = c.load(right)

	// Check for floating point numbers
	floating := false

	if v, ok := left.type_.(*types.PrimitiveType); ok {
		floating = types.IsFloating(v.Kind)
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

	default:
		log.Fatalln("Invalid operator kind")
	}

	// Emit
	val := c.locals.unnamed(left.type_)
	c.writeFmt("%s = %s %s %s, %s\n", val, inst, c.getType(left.type_), left, right)

	return val
}
