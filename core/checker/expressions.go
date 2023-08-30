package checker

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"log"
	"strings"
)

func (c *checker) VisitGroup(expr *ast.Group) {
	expr.AcceptChildren(c)

	expr.SetType(expr.Expr.Type())
}

func (c *checker) VisitLiteral(expr *ast.Literal) {
	expr.AcceptChildren(c)

	switch expr.Value.Kind {
	case scanner.Nil:
		expr.SetType(&types.PointerType{Pointee: types.Primitive(types.Void)})

	case scanner.True, scanner.False:
		expr.SetType(types.Primitive(types.Bool))

	case scanner.Number:
		if strings.HasSuffix(expr.Value.Lexeme, "f") {
			expr.SetType(types.Primitive(types.F32))
		} else if strings.ContainsRune(expr.Value.Lexeme, '.') {
			expr.SetType(types.Primitive(types.F64))
		} else {
			expr.SetType(types.Primitive(types.I32))
		}

	case scanner.Character:
		expr.SetType(types.Primitive(types.U8))

	case scanner.String:
		expr.SetType(&types.PointerType{Pointee: types.Primitive(types.U8)})
	}
}

func (c *checker) VisitUnary(expr *ast.Unary) {
	expr.AcceptChildren(c)

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

	case scanner.Ampersand:
		expr.SetType(&types.PointerType{Pointee: expr.Right.Type()})

	default:
		log.Fatalln("Invalid unary operator")
	}
}

func (c *checker) VisitBinary(expr *ast.Binary) {
	expr.AcceptChildren(c)

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
	expr.AcceptChildren(c)

	// Function
	function := c.getFunction(expr.Identifier)

	if function != nil {
		params := make([]types.Type, len(function.Params))

		for i, param := range function.Params {
			params[i] = param.Type
		}

		expr.SetType(&types.FunctionType{
			Params:   params,
			Variadic: function.Variadic,
			Returns:  function.Returns,
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
	expr.AcceptChildren(c)
}

func (c *checker) VisitCast(expr *ast.Cast) {
	expr.AcceptChildren(c)

	if types.IsPrimitive(expr.Expr.Type(), types.Void) || types.IsPrimitive(expr.Type(), types.Void) {
		c.error(expr, "Cannot cast to or from type 'void'.")
	}
}

func (c *checker) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(c)

	if v, ok := expr.Callee.Type().(*types.FunctionType); ok {
		toCheck := min(len(v.Params), len(expr.Args))

		if v.Variadic {
			if len(expr.Args) < len(v.Params) {
				c.error(expr, "Got '%d' arguments but function takes at least '%d'.", len(expr.Args), len(v.Params))
			}
		} else {
			if len(v.Params) != len(expr.Args) {
				c.error(expr, "Got '%d' arguments but function only takes '%d'.", len(expr.Args), len(v.Params))
			}
		}

		for i := 0; i < toCheck; i++ {
			if !expr.Args[i].Type().CanAssignTo(v.Params[i]) {
				c.error(expr, "Argument with type '%s' cannot be assigned to a parameter wth type '%s'.", expr.Args[i].Type(), v.Params[i])
			}
		}

		expr.SetType(v.Returns)
	} else {
		expr.SetType(types.Primitive(types.Void))
		c.error(expr, "Can't call type '%s'.", expr.Callee.Type())
	}
}

func (c *checker) VisitIndex(expr *ast.Index) {
	expr.AcceptChildren(c)

	if v, ok := expr.Value.Type().(*types.ArrayType); ok {
		expr.SetType(v.Base)
	} else if v, ok := expr.Value.Type().(*types.PointerType); ok {
		expr.SetType(v.Pointee)
	} else {
		c.error(expr, "Can only index into array and pointer types, not '%s'.", expr.Value.Type())
		expr.SetType(&types.PointerType{Pointee: types.Primitive(types.Void)})
	}
}

func (c *checker) VisitMember(expr *ast.Member) {
	expr.AcceptChildren(c)

	var s *types.StructType

	if v, ok := expr.Value.Type().(*types.StructType); ok {
		s = v
	} else if v, ok := expr.Value.Type().(*types.PointerType); ok {
		if v, ok := v.Pointee.(*types.StructType); ok {
			s = v
		}
	}

	if s != nil {
		_, field := s.GetField(expr.Name.Lexeme)

		if field != nil {
			expr.SetType(field.Type)
		} else {
			c.error(expr, "Struct '%s' does not contain member '%s'.", s, expr.Name)
			expr.SetType(types.Primitive(types.Void))
		}
	} else {
		c.error(expr, "Only structs and pointers to structs can have members, not '%s'.", expr.Value.Type())
		expr.SetType(types.Primitive(types.Void))
	}
}
