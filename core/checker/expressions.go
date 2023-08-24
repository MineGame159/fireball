package checker

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"log"
	"strings"
)

func (c *checker) VisitGroup(expr *ast.Group) {
	c.acceptExpr(expr.Expr)
	expr.SetType(expr.Expr.Type())
}

func (c *checker) VisitLiteral(expr *ast.Literal) {
	kind := types.Void

	switch expr.Value.Kind {
	case scanner.True, scanner.False:
		kind = types.Bool

	case scanner.Number:
		if strings.HasSuffix(expr.Value.Lexeme, "f") {
			kind = types.F32
		} else if strings.ContainsRune(expr.Value.Lexeme, '.') {
			kind = types.F64
		} else {
			kind = types.I32
		}
	}

	expr.SetType(types.Primitive(kind))
}

func (c *checker) VisitUnary(expr *ast.Unary) {
	c.acceptExpr(expr.Right)

	switch expr.Op.Kind {
	case scanner.Bang:
		if !types.IsPrimitive(expr.Right.Type(), types.Bool) {
			c.error(expr, "Expected a 'bool' but got a '%s'.", expr.Right.Type())
		}

		expr.SetType(types.Primitive(types.Bool))

	case scanner.Minus:
		if v, ok := expr.Right.Type().(*types.PrimitiveType); ok {
			if types.IsFloating(v.Kind) || types.IsSigned(v.Kind) {
				expr.SetType(expr.Right.Type())
				break
			}
		}

		expr.SetType(types.Primitive(types.I32))
		c.error(expr, "Expected either a floating pointer number or signed integer but got a '%s'.", expr.Right.Type())

	default:
		log.Fatalln("Invalid unary operator")
	}
}

func (c *checker) VisitBinary(expr *ast.Binary) {
	// Accept
	c.acceptExpr(expr.Left)
	c.acceptExpr(expr.Right)

	if scanner.IsArithmetic(expr.Op.Kind) {
		// Arithmetic
		if left, ok := expr.Left.Type().(*types.PrimitiveType); ok {
			if right, ok := expr.Right.Type().(*types.PrimitiveType); ok {
				if types.IsNumber(left.Kind) && types.IsNumber(right.Kind) && left == right {
					expr.SetType(expr.Left.Type())
					return
				}
			}
		}

		expr.SetType(types.Primitive(types.I32))
		c.error(expr, "Expected two number types.")
	} else if scanner.IsEquality(expr.Op.Kind) {
		// Equality
		expr.SetType(types.Primitive(types.Bool))
	} else if scanner.IsComparison(expr.Op.Kind) {
		// Comparison
		expr.SetType(types.Primitive(types.Bool))

		if left, ok := expr.Left.Type().(*types.PrimitiveType); ok {
			if right, ok := expr.Right.Type().(*types.PrimitiveType); ok {
				if !types.IsNumber(left.Kind) || !types.IsNumber(right.Kind) || left != right {
					c.error(expr, "Expected two number types.")
				}
			}
		}
	} else {
		// Error
		expr.SetType(types.Primitive(types.Void))
		c.error(expr, "Invalid operator.")
	}
}

func (c *checker) VisitIdentifier(expr *ast.Identifier) {
	// Function
	function := c.getFunction(expr.Identifier)

	if function != nil {
		params := make([]types.Type, len(function.Params))

		for i, param := range function.Params {
			params[i] = param.Type
		}

		expr.SetType(types.FunctionType{
			Params:  params,
			Returns: function.Returns,
		})
		return
	}

	// Variable
	variable := c.getVariable(expr.Identifier)

	if variable != nil {
		expr.SetType(variable.type_)
		return
	}

	// Error
	expr.SetType(types.Primitive(types.Void))
	c.error(expr, "Unknown identifier.")
}

func (c *checker) VisitAssignment(expr *ast.Assignment) {
	c.acceptExpr(expr.Assignee)
	c.acceptExpr(expr.Value)
}

func (c *checker) VisitCast(expr *ast.Cast) {
	c.acceptExpr(expr.Expr)

	if types.IsPrimitive(expr.Expr.Type(), types.Void) || types.IsPrimitive(expr.Type(), types.Void) {
		c.error(expr, "Cannot cast to or from type 'void'.")
	}
}

func (c *checker) VisitCall(expr *ast.Call) {
	c.acceptExpr(expr.Callee)

	for _, arg := range expr.Args {
		c.acceptExpr(arg)
	}

	if v, ok := expr.Callee.Type().(types.FunctionType); ok {
		if len(v.Params) == len(expr.Args) {
			for i, arg := range expr.Args {
				if !arg.Type().CanAssignTo(v.Params[i]) {
					c.error(expr, "Argument with type '%s' cannot be assigned to a parameter wth type '%s'.", arg.Type(), v.Params[i])
				}
			}
		} else {
			c.error(expr, "Got '%d' arguments but function only takes '%d'.", len(expr.Args), len(v.Params))
		}

		expr.SetType(v.Returns)
	} else {
		expr.SetType(types.Primitive(types.Void))
		c.error(expr, "Can't call type '%s'.", expr.Callee.Type())
	}
}
