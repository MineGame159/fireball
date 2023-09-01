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
			c.errorRange(expr.Right.Range(), "Expected a 'bool' but got a '%s'.", expr.Right.Type())
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
		c.errorRange(expr.Right.Range(), "Expected either a floating pointer number or signed integer but got a '%s'.", expr.Right.Type())

	case scanner.Ampersand:
		expr.SetType(types.Pointer(expr.Right.Type().Copy(), core.Range{}))

	default:
		log.Fatalln("checker.VisitUnary() - Invalid unary operator")
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
		c.errorRange(expr.Range(), "Expected two number types.")
	} else if scanner.IsEquality(expr.Op.Kind) {
		// Equality
		expr.SetType(types.Primitive(types.Bool, core.Range{}))
	} else if scanner.IsComparison(expr.Op.Kind) {
		// Comparison
		expr.SetType(types.Primitive(types.Bool, core.Range{}))

		if left, ok := expr.Left.Type().(*types.PrimitiveType); ok {
			if right, ok := expr.Right.Type().(*types.PrimitiveType); ok {
				if !types.IsNumber(left.Kind) || !types.IsNumber(right.Kind) || !left.Equals(right) {
					c.errorRange(expr.Range(), "Expected two number types.")
				}
			}
		}
	} else {
		// Error
		log.Fatalln("checker.VisitBinary() - Invalid operator kind")
	}
}

func (c *checker) VisitLogical(expr *ast.Logical) {
	expr.AcceptChildren(c)

	// Check bool types
	if !types.IsPrimitive(expr.Left.Type(), types.Bool) {
		c.errorRange(expr.Left.Range(), "Expected a 'bool' but got a '%s'.", expr.Left.Type())
	}

	if !types.IsPrimitive(expr.Right.Type(), types.Bool) {
		c.errorRange(expr.Right.Range(), "Expected a 'bool' but got a '%s'.", expr.Right.Type())
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
		expr.SetType(variable.type_.Copy())
		return
	}

	// Error
	expr.SetType(types.Primitive(types.Void, core.Range{}))
	c.errorToken(expr.Identifier, "Unknown identifier.")
}

func (c *checker) VisitAssignment(expr *ast.Assignment) {
	expr.AcceptChildren(c)

	expr.SetType(expr.Value.Type())

	// Check assignee
	validAssignee := false

	if v, ok := expr.Assignee.(*ast.Identifier); ok {
		if _, ok := v.Type().(*types.FunctionType); !ok {
			validAssignee = true
		}
	} else if _, ok := expr.Assignee.(*ast.Member); ok {
		validAssignee = true
	} else if _, ok := expr.Assignee.(*ast.Index); ok {
		validAssignee = true
	} else if _, ok := expr.Assignee.Type().(*types.PointerType); ok {
		validAssignee = true
	}

	if !validAssignee {
		c.errorRange(expr.Assignee.Range(), "Can only assign to variables, fields, array indexes and pointers.")
	}

	// Check type
	if expr.Op.Kind == scanner.Equal {
		// Equal
		if !expr.Value.Type().CanAssignTo(expr.Assignee.Type()) {
			c.errorRange(expr.Value.Range(), "Expected a '%s' but got '%s'.", expr.Assignee.Type(), expr.Value.Type())
		}
	} else {
		// Arithmetic
		valid := false

		if assignee, ok := expr.Assignee.Type().(*types.PrimitiveType); ok {
			if value, ok := expr.Value.Type().(*types.PrimitiveType); ok {
				if types.IsNumber(assignee.Kind) && types.IsNumber(value.Kind) && assignee.Equals(value) {
					valid = true
				}
			}
		}

		if !valid {
			c.errorRange(expr.Value.Range(), "Expected two number types.")
		}
	}
}

func (c *checker) VisitCast(expr *ast.Cast) {
	expr.AcceptChildren(c)

	if types.IsPrimitive(expr.Expr.Type(), types.Void) || types.IsPrimitive(expr.Type(), types.Void) {
		c.errorRange(expr.Range(), "Cannot cast to or from type 'void'.")
	}
}

func (c *checker) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(c)

	if v, ok := expr.Callee.Type().(*types.FunctionType); ok {
		toCheck := min(len(v.Params), len(expr.Args))

		if v.Variadic {
			if len(expr.Args) < len(v.Params) {
				c.errorRange(expr.Range(), "Got '%d' arguments but function takes at least '%d'.", len(expr.Args), len(v.Params))
			}
		} else {
			if len(v.Params) != len(expr.Args) {
				c.errorRange(expr.Range(), "Got '%d' arguments but function only takes '%d'.", len(expr.Args), len(v.Params))
			}
		}

		for i := 0; i < toCheck; i++ {
			arg := expr.Args[i]
			param := v.Params[i]

			if !arg.Type().CanAssignTo(param) {
				c.errorRange(arg.Range(), "Argument with type '%s' cannot be assigned to a parameter wth type '%s'.", arg.Type(), param)
			}
		}

		expr.SetType(v.Returns.Copy())
	} else {
		expr.SetType(types.Primitive(types.Void, core.Range{}))
		c.errorRange(expr.Callee.Range(), "Can't call type '%s'.", expr.Callee.Type())
	}
}

func (c *checker) VisitIndex(expr *ast.Index) {
	expr.AcceptChildren(c)

	// Check value type
	if v, ok := expr.Value.Type().(*types.ArrayType); ok {
		expr.SetType(v.Base.Copy())
	} else if v, ok := expr.Value.Type().(*types.PointerType); ok {
		expr.SetType(v.Pointee.Copy())
	} else {
		c.errorRange(expr.Value.Range(), "Can only index into array and pointer types, not '%s'.", expr.Value.Type())
		expr.SetType(types.Pointer(types.Primitive(types.Void, core.Range{}), core.Range{}))
	}

	// Check index type
	if v, ok := expr.Index.Type().(*types.PrimitiveType); ok {
		if types.IsInteger(v.Kind) {
			return
		}
	}

	c.errorRange(expr.Index.Range(), "Can only index using integer types, not '%s'.", expr.Index.Type())
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
			c.errorToken(expr.Name, "Struct '%s' does not contain member '%s'.", s, expr.Name)
			expr.SetType(types.Primitive(types.Void, core.Range{}))
		}
	} else {
		c.errorRange(expr.Value.Range(), "Only structs and pointers to structs can have members, not '%s'.", expr.Value.Type())
		expr.SetType(types.Primitive(types.Void, core.Range{}))
	}
}
