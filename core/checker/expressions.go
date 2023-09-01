package checker

import (
	"fireball/core"
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

	var kind types.PrimitiveKind
	pointer := false

	switch expr.Value.Kind {
	case scanner.Nil:
		kind = types.Void
		pointer = true

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

	case scanner.Character:
		kind = types.U8

	case scanner.String:
		kind = types.U8
		pointer = true
	}

	expr.SetType(types.Primitive(kind, core.Range{}))

	if pointer {
		expr.SetType(types.Pointer(expr.Type(), core.Range{}))
	}
}

func (c *checker) VisitUnary(expr *ast.Unary) {
	expr.AcceptChildren(c)

	switch expr.Op.Kind {
	case scanner.Bang:
		if !types.IsPrimitive(expr.Right.Type(), types.Bool) {
			c.errorNode(expr, "Expected a 'bool' but got a '%s'.", expr.Right.Type())
		}

		expr.SetType(types.Primitive(types.Bool, core.Range{}))

	case scanner.Minus:
		if v, ok := expr.Right.Type().(*types.PrimitiveType); ok {
			if types.IsFloating(v.Kind) || types.IsSigned(v.Kind) {
				expr.SetType(expr.Right.Type())
				break
			}
		}

		expr.SetType(types.Primitive(types.I32, core.Range{}))
		c.errorNode(expr, "Expected either a floating pointer number or signed integer but got a '%s'.", expr.Right.Type())

	case scanner.Ampersand:
		expr.SetType(types.Pointer(expr.Right.Type().Copy(), core.Range{}))

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
				if types.IsNumber(left.Kind) && types.IsNumber(right.Kind) && left.Equals(right) {
					expr.SetType(expr.Left.Type().Copy())
					return
				}
			}
		}

		expr.SetType(types.Primitive(types.I32, core.Range{}))
		c.errorNode(expr, "Expected two number types.")
	} else if scanner.IsEquality(expr.Op.Kind) {
		// Equality
		expr.SetType(types.Primitive(types.Bool, core.Range{}))
	} else if scanner.IsComparison(expr.Op.Kind) {
		// Comparison
		expr.SetType(types.Primitive(types.Bool, core.Range{}))

		if left, ok := expr.Left.Type().(*types.PrimitiveType); ok {
			if right, ok := expr.Right.Type().(*types.PrimitiveType); ok {
				if !types.IsNumber(left.Kind) || !types.IsNumber(right.Kind) || !left.Equals(right) {
					c.errorNode(expr, "Expected two number types.")
				}
			}
		}
	} else {
		// Error
		expr.SetType(types.Primitive(types.Void, core.Range{}))
		c.errorNode(expr, "Invalid operator.")
	}
}

func (c *checker) VisitLogical(expr *ast.Logical) {
	expr.AcceptChildren(c)

	// Check bool types
	if !types.IsPrimitive(expr.Left.Type(), types.Bool) {
		c.errorNode(expr, "Left - Expected a 'bool' but got a '%s'.", expr.Left.Type())
	}

	if !types.IsPrimitive(expr.Right.Type(), types.Bool) {
		c.errorNode(expr, "Right - Expected a 'bool' but got a '%s'.", expr.Right.Type())
	}

	// Set type
	expr.SetType(types.Primitive(types.Bool, core.Range{}))
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
		variable.used = true
		expr.SetType(variable.type_)
		return
	}

	// Error
	expr.SetType(types.Primitive(types.Void, core.Range{}))
	c.errorNode(expr, "Unknown identifier.")
}

func (c *checker) VisitAssignment(expr *ast.Assignment) {
	expr.AcceptChildren(c)

	expr.SetType(expr.Value.Type())
}

func (c *checker) VisitCast(expr *ast.Cast) {
	expr.AcceptChildren(c)

	if types.IsPrimitive(expr.Expr.Type(), types.Void) || types.IsPrimitive(expr.Type(), types.Void) {
		c.errorNode(expr, "Cannot cast to or from type 'void'.")
	}
}

func (c *checker) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(c)

	if v, ok := expr.Callee.Type().(*types.FunctionType); ok {
		toCheck := min(len(v.Params), len(expr.Args))

		if v.Variadic {
			if len(expr.Args) < len(v.Params) {
				c.errorNode(expr, "Got '%d' arguments but function takes at least '%d'.", len(expr.Args), len(v.Params))
			}
		} else {
			if len(v.Params) != len(expr.Args) {
				c.errorNode(expr, "Got '%d' arguments but function only takes '%d'.", len(expr.Args), len(v.Params))
			}
		}

		for i := 0; i < toCheck; i++ {
			if !expr.Args[i].Type().CanAssignTo(v.Params[i]) {
				c.errorNode(expr, "Argument with type '%s' cannot be assigned to a parameter wth type '%s'.", expr.Args[i].Type(), v.Params[i])
			}
		}

		expr.SetType(v.Returns.Copy())
	} else {
		expr.SetType(types.Primitive(types.Void, core.Range{}))
		c.errorNode(expr, "Can't call type '%s'.", expr.Callee.Type())
	}
}

func (c *checker) VisitIndex(expr *ast.Index) {
	expr.AcceptChildren(c)

	if v, ok := expr.Value.Type().(*types.ArrayType); ok {
		expr.SetType(v.Base.Copy())
	} else if v, ok := expr.Value.Type().(*types.PointerType); ok {
		expr.SetType(v.Pointee.Copy())
	} else {
		c.errorNode(expr, "Can only index into array and pointer types, not '%s'.", expr.Value.Type())
		expr.SetType(types.Pointer(types.Primitive(types.Void, core.Range{}), core.Range{}))
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
			expr.SetType(field.Type.Copy())
		} else {
			c.errorNode(expr, "Struct '%s' does not contain member '%s'.", s, expr.Name)
			expr.SetType(types.Primitive(types.Void, core.Range{}))
		}
	} else {
		c.errorNode(expr, "Only structs and pointers to structs can have members, not '%s'.", expr.Value.Type())
		expr.SetType(types.Primitive(types.Void, core.Range{}))
	}
}
