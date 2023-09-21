package checker

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fireball/core/utils"
	"log"
	"strconv"
	"strings"
)

func (c *checker) VisitGroup(expr *ast.Group) {
	expr.AcceptChildren(c)

	*expr.Result() = *expr.Expr.Result()
	expr.Result().Flags = 0
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
		raw := expr.Value.Lexeme
		last := raw[len(raw)-1]

		if last == 'f' || last == 'F' {
			_, err := strconv.ParseFloat(raw[:len(raw)-1], 32)
			if err != nil {
				c.errorToken(expr.Value, "Invalid float.")
				expr.Result().SetInvalid()

				return
			}

			kind = types.F32
		} else if strings.ContainsRune(raw, '.') {
			_, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				c.errorToken(expr.Value, "Invalid double.")
				expr.Result().SetInvalid()

				return
			}

			kind = types.F64
		} else {
			kind = types.I32
		}

	case scanner.Hex:
		_, err := strconv.ParseUint(expr.Value.Lexeme[2:], 16, 64)
		if err != nil {
			c.errorToken(expr.Value, "Invalid hex integer.")
			expr.Result().SetInvalid()

			return
		}

		kind = types.U32

	case scanner.Binary:
		_, err := strconv.ParseUint(expr.Value.Lexeme[2:], 2, 64)
		if err != nil {
			c.errorToken(expr.Value, "Invalid binary integer.")
			expr.Result().SetInvalid()

			return
		}

		kind = types.U32

	case scanner.Character:
		kind = types.U8

	case scanner.String:
		kind = types.U8
		pointer = true
	}

	expr.Result().SetValue(types.Primitive(kind, core.Range{}), 0)

	if pointer {
		expr.Result().SetValue(types.Pointer(expr.Result().Type, core.Range{}), 0)
	}
}

func (c *checker) VisitStructInitializer(expr *ast.StructInitializer) {
	expr.AcceptChildren(c)

	// Check struct
	var type_ *ast.Struct

	if t, _ := c.resolver.GetType(expr.Name.Lexeme); t != nil {
		if s, ok := t.(*ast.Struct); ok {
			type_ = s
		}
	}

	if type_ == nil {
		c.errorToken(expr.Name, "Unknown type '%s'.", expr.Name)
		expr.Result().SetInvalid()

		return
	}

	expr.Result().SetValue(type_, 0)

	// Check fields
	assignedFields := utils.NewSet[string]()

	for _, initField := range expr.Fields {
		// Check name collision
		if !assignedFields.Add(initField.Name.Lexeme) {
			c.errorToken(initField.Name, "Field with the name '%s' was already assigned.", initField.Name)
		}

		// Check field
		_, field := type_.GetField(initField.Name.Lexeme)
		if field == nil {
			c.errorToken(initField.Name, "Field with the name '%s' doesn't exist on the struct '%s'.", initField.Name, expr.Name)
			continue
		}

		// Check value result
		if initField.Value.Result().Kind == ast.InvalidResultKind {
			continue // Do not cascade errors
		}

		if initField.Value.Result().Kind != ast.ValueResultKind {
			c.errorRange(initField.Value.Range(), "Cannot assign this value to a field with type '%s'.", field.Type)
			continue
		}

		if !initField.Value.Result().Type.CanAssignTo(field.Type) {
			c.errorRange(initField.Value.Range(), "Expected a '%s' but got '%s'.", field.Type, initField.Value.Result().Type)
		}
	}
}

func (c *checker) VisitArrayInitializer(expr *ast.ArrayInitializer) {
	expr.AcceptChildren(c)

	// Check empty
	if len(expr.Values) == 0 {
		c.errorRange(expr.Range(), "Array initializers need to have at least one value.")
		expr.Result().SetInvalid()

		return
	}

	// Check values
	ok := true
	var type_ types.Type

	for _, value := range expr.Values {
		if value.Result().Kind == ast.InvalidResultKind {
			ok = false
			continue
		}

		if value.Result().Kind != ast.ValueResultKind {
			c.errorRange(value.Range(), "Invalid value.")
			ok = false

			continue
		}

		if type_ == nil {
			type_ = value.Result().Type
		} else {
			if !value.Result().Type.CanAssignTo(type_) {
				c.errorRange(value.Range(), "Expected a '%s' but got '%s'.", type_, value.Result().Type)
				ok = false
			}
		}
	}

	if ok {
		expr.Result().SetValue(types.Array(uint32(len(expr.Values)), type_, core.Range{}), 0)
	} else {
		expr.Result().SetInvalid()
	}
}

func (c *checker) VisitUnary(expr *ast.Unary) {
	expr.AcceptChildren(c)

	// Check value
	result := expr.Value.Result()

	if result.Kind == ast.InvalidResultKind {
		return // Do not cascade errors
	}

	if expr.Prefix {
		// Prefix
		switch expr.Op.Kind {
		case scanner.Bang:
			if result.Kind != ast.ValueResultKind {
				c.errorRange(expr.Value.Range(), "Cannot negate this value.")
				expr.Result().SetInvalid()

				return
			}

			if !types.IsPrimitive(result.Type, types.Bool) {
				c.errorRange(expr.Value.Range(), "Expected a 'bool' but got a '%s'.", result.Type)
			}

			expr.Result().SetValue(types.Primitive(types.Bool, core.Range{}), 0)

		case scanner.Minus:
			if result.Kind != ast.ValueResultKind {
				c.errorRange(expr.Value.Range(), "Cannot negate this value.")
				expr.Result().SetInvalid()

				return
			}

			if v, ok := result.Type.(*types.PrimitiveType); ok {
				if types.IsFloating(v.Kind) || types.IsSigned(v.Kind) {
					expr.Result().SetValue(result.Type, 0)
					return
				}
			}

			c.errorRange(expr.Value.Range(), "Expected either a floating pointer number or signed integer but got a '%s'.", result.Type)
			expr.Result().SetInvalid()

		case scanner.Ampersand:
			if result.IsAddressable() {
				expr.Result().SetValue(types.Pointer(result.Type, core.Range{}), 0)
			} else if result.Kind == ast.FunctionResultKind {
				if result.Function.Method() != nil {
					c.errorRange(expr.Value.Range(), "Cannot take address of a non-static method.")
					expr.Result().SetInvalid()

					return
				}

				expr.Result().SetValue(expr.Value.Result().Function, 0)
			} else {
				c.errorRange(expr.Value.Range(), "Cannot take address of this expression.")
				expr.Result().SetInvalid()
			}

		case scanner.Star:
			if result.Kind != ast.ValueResultKind {
				c.errorRange(expr.Value.Range(), "Cannot dereference this value.")
				expr.Result().SetInvalid()

				return
			}

			if p, ok := result.Type.(*types.PointerType); ok {
				expr.Result().SetValue(p.Pointee, ast.AssignableFlag)
			} else {
				c.errorRange(expr.Value.Range(), "Can only dereference pointer types, not '%s'.", result.Type)
				expr.Result().SetInvalid()
			}

		case scanner.PlusPlus, scanner.MinusMinus:
			if !result.IsAddressable() {
				c.errorRange(expr.Value.Range(), "Invalid value.")
				expr.Result().SetInvalid()

				return
			}

			if type_, ok := result.Type.(*types.PrimitiveType); !ok || (!types.IsInteger(type_.Kind) && !types.IsFloating(type_.Kind)) {
				c.errorRange(expr.Value.Range(), "Cannot increment or decrement '%s'.", result.Type)
				expr.Result().SetInvalid()

				return
			}

			expr.Result().SetValue(result.Type, 0)

		default:
			panic("checker.VisitUnary() - Invalid unary prefix operator")
		}
	} else {
		// Postfix
		switch expr.Op.Kind {
		case scanner.PlusPlus, scanner.MinusMinus:
			if !result.IsAddressable() {
				c.errorRange(expr.Value.Range(), "Invalid value.")
				expr.Result().SetInvalid()

				return
			}

			if type_, ok := result.Type.(*types.PrimitiveType); !ok || (!types.IsInteger(type_.Kind) && !types.IsFloating(type_.Kind)) {
				c.errorRange(expr.Value.Range(), "Cannot increment or decrement '%s'.", result.Type)
				expr.Result().SetInvalid()

				return
			}

			expr.Result().SetValue(result.Type, 0)

		default:
			panic("checker.VisitUnary() - Invalid unary postfix operator")
		}
	}
}

func (c *checker) VisitBinary(expr *ast.Binary) {
	expr.AcceptChildren(c)

	if expr.Left.Result().Kind == ast.InvalidResultKind || expr.Right.Result().Kind == ast.InvalidResultKind {
		return // // Do not cascade errors
	}

	// Check results
	ok := true

	if expr.Left.Result().Kind != ast.ValueResultKind {
		c.errorRange(expr.Left.Range(), "Invalid value.")
		ok = false
	}

	if expr.Right.Result().Kind != ast.ValueResultKind {
		c.errorRange(expr.Right.Range(), "Invalid value")
		ok = false
	}

	if !ok {
		expr.Result().SetInvalid()
		return
	}

	leftType := expr.Left.Result().Type
	rightType := expr.Right.Result().Type

	// Check based on the operator
	if scanner.IsArithmetic(expr.Op.Kind) {
		// Arithmetic
		if left, ok := leftType.(*types.PrimitiveType); ok {
			if right, ok := rightType.(*types.PrimitiveType); ok {
				if types.IsNumber(left.Kind) && types.IsNumber(right.Kind) && left.Equals(right) {
					expr.Result().SetValue(leftType, 0)
					return
				}
			}
		}

		c.errorRange(expr.Range(), "Expected two number types.")
		expr.Result().SetInvalid()
	} else if scanner.IsEquality(expr.Op.Kind) {
		// Equality
		valid := false

		if leftType.Equals(rightType) {
			// left type == right type
			valid = true
		} else if left, ok := leftType.(*types.PrimitiveType); ok {
			// integer == integer || floating == floating
			if right, ok := rightType.(*types.PrimitiveType); ok {
				if (types.IsInteger(left.Kind) && types.IsInteger(right.Kind)) || (types.IsFloating(left.Kind) && types.IsFloating(right.Kind)) {
					valid = true
				}
			}
		} else if left, ok := leftType.(*types.PointerType); ok {
			if right, ok := rightType.(*types.PointerType); ok {
				// *void == *? || *? == *void
				if types.IsPrimitive(left.Pointee, types.Void) || types.IsPrimitive(right.Pointee, types.Void) {
					valid = true
				}
			}
		}

		if !valid {
			c.errorRange(expr.Range(), "Cannot check equality for '%s' and '%s'.", leftType, rightType)
			expr.Result().SetInvalid()
		} else {
			expr.Result().SetValue(types.Primitive(types.Bool, core.Range{}), 0)
		}
	} else if scanner.IsComparison(expr.Op.Kind) {
		// Comparison
		if left, ok := leftType.(*types.PrimitiveType); ok {
			if right, ok := rightType.(*types.PrimitiveType); ok {
				if !types.IsNumber(left.Kind) || !types.IsNumber(right.Kind) || !left.Equals(right) {
					c.errorRange(expr.Range(), "Expected two number types.")
					expr.Result().SetInvalid()

					return
				}
			}
		}

		expr.Result().SetValue(types.Primitive(types.Bool, core.Range{}), 0)
	} else if scanner.IsBitwise(expr.Op.Kind) {
		// Bitwise
		if left, ok := leftType.(*types.PrimitiveType); ok {
			if right, ok := rightType.(*types.PrimitiveType); ok {
				if left.Equals(right) && types.IsInteger(left.Kind) {
					expr.Result().SetValue(leftType, 0)
					return
				}
			}
		}

		c.errorRange(expr.Range(), "Expected two identical integer types.")
		expr.Result().SetInvalid()
	} else {
		// Error
		log.Fatalln("checker.VisitBinary() - Invalid operator kind")
	}
}

func (c *checker) VisitLogical(expr *ast.Logical) {
	expr.AcceptChildren(c)

	if expr.Left.Result().Kind == ast.InvalidResultKind || expr.Right.Result().Kind == ast.InvalidResultKind {
		return // Do not cascade errors
	}

	// Check results
	ok := true

	if expr.Left.Result().Kind != ast.ValueResultKind {
		c.errorRange(expr.Left.Range(), "Invalid value.")
		ok = false
	}

	if expr.Right.Result().Kind != ast.ValueResultKind {
		c.errorRange(expr.Right.Range(), "Invalid value")
		ok = false
	}

	if !ok {
		expr.Result().SetInvalid()
		return
	}

	// Check bool types
	ok = true

	if !types.IsPrimitive(expr.Left.Result().Type, types.Bool) {
		c.errorRange(expr.Left.Range(), "Expected a 'bool' but got a '%s'.", expr.Left.Result().Type)
		ok = false
	}

	if !types.IsPrimitive(expr.Right.Result().Type, types.Bool) {
		c.errorRange(expr.Right.Range(), "Expected a 'bool' but got a '%s'.", expr.Right.Result().Type)
		ok = false
	}

	if !ok {
		expr.Result().SetInvalid()
		return
	}

	// Set type
	expr.Result().SetValue(types.Primitive(types.Bool, core.Range{}), 0)
}

func (c *checker) VisitIdentifier(expr *ast.Identifier) {
	expr.AcceptChildren(c)

	// Function
	if parent, ok := expr.Parent().(*ast.Call); ok && parent.Callee == expr {
		if f, _ := c.resolver.GetFunction(expr.Identifier.Lexeme); f != nil {
			expr.Result().SetFunction(f)
			expr.Kind = ast.FunctionKind

			return
		}
	}

	// Type
	if t, _ := c.resolver.GetType(expr.Identifier.Lexeme); t != nil {
		expr.Result().SetType(t.WithRange(core.Range{}))

		if _, ok := t.(*ast.Enum); ok {
			expr.Kind = ast.EnumKind
		} else if _, ok := t.(*ast.Struct); ok {
			expr.Kind = ast.StructKind
		} else {
			panic("checker.VisitIdentifier() - Invalid type")
		}

		return
	}

	// Variable
	if variable := c.getVariable(expr.Identifier); variable != nil {
		variable.used = true

		expr.Result().SetValue(variable.type_, ast.AssignableFlag|ast.AddressableFlag)

		if variable.param {
			expr.Kind = ast.ParameterKind
		} else {
			expr.Kind = ast.VariableKind
		}

		return
	}

	// Function pointer
	if parent, ok := expr.Parent().(*ast.Unary); ok && parent.Op.Kind == scanner.Ampersand {
		if f, _ := c.resolver.GetFunction(expr.Identifier.Lexeme); f != nil {
			expr.Result().SetFunction(f)
			expr.Kind = ast.FunctionKind

			return
		}
	}

	// Error
	c.errorToken(expr.Identifier, "Unknown identifier.")
	expr.Result().SetInvalid()
}

func (c *checker) VisitAssignment(expr *ast.Assignment) {
	expr.AcceptChildren(c)

	if expr.Assignee.Result().Kind == ast.InvalidResultKind || expr.Value.Result().Kind == ast.InvalidResultKind {
		return // Do not cascade errors
	}

	// Check results
	ok := true

	if !expr.Assignee.Result().IsAssignable() {
		c.errorRange(expr.Assignee.Range(), "Cannot assign to this value.")
		ok = false
	}

	if expr.Value.Result().Kind != ast.ValueResultKind {
		c.errorRange(expr.Value.Range(), "Invalid value.")
		ok = false
	}

	if !ok {
		expr.Result().SetInvalid()
		return
	}

	// Check type
	if expr.Op.Kind == scanner.Equal {
		// Equal
		if !expr.Value.Result().Type.CanAssignTo(expr.Assignee.Result().Type) {
			c.errorRange(expr.Value.Range(), "Expected a '%s' but got '%s'.", expr.Assignee.Result().Type, expr.Value.Result().Type)
			expr.Result().SetInvalid()

			return
		}
	} else {
		// Arithmetic
		valid := false

		if assignee, ok := expr.Assignee.Result().Type.(*types.PrimitiveType); ok {
			if value, ok := expr.Value.Result().Type.(*types.PrimitiveType); ok {
				if types.IsNumber(assignee.Kind) && types.IsNumber(value.Kind) && assignee.Equals(value) {
					valid = true
				}
			}
		}

		if !valid {
			c.errorRange(expr.Value.Range(), "Expected two number types.")
			expr.Result().SetInvalid()

			return
		}
	}

	// Set result
	expr.Result().SetValue(expr.Value.Result().Type, 0)
}

func (c *checker) VisitCast(expr *ast.Cast) {
	expr.AcceptChildren(c)

	if expr.Expr.Result().Kind == ast.InvalidResultKind {
		return // // Do not cascade errors
	}

	if expr.Expr.Result().Kind != ast.ValueResultKind {
		c.errorRange(expr.Expr.Range(), "Cannot cast this value.")
		expr.Result().SetInvalid()

		return
	}

	if types.IsPrimitive(expr.Expr.Result().Type, types.Void) || types.IsPrimitive(expr.Result().Type, types.Void) {
		// void
		c.errorRange(expr.Range(), "Cannot cast to or from type 'void'.")
		expr.Result().SetInvalid()
	} else if _, ok := expr.Expr.Result().Type.(*ast.Enum); ok {
		// enum to non integer
		if to, ok := expr.Result().Type.(*types.PrimitiveType); !ok || !types.IsInteger(to.Kind) {
			c.errorRange(expr.Range(), "Can only cast enums to integers, not '%s'.", to)
			expr.Result().SetInvalid()
		}
	} else if _, ok := expr.Result().Type.(*ast.Enum); ok {
		// non integer to enum
		if from, ok := expr.Expr.Result().Type.(*types.PrimitiveType); !ok || !types.IsInteger(from.Kind) {
			c.errorRange(expr.Range(), "Can only cast to enums from integers, not '%s'.", from)
			expr.Result().SetInvalid()
		}
	}
}

func (c *checker) VisitSizeof(expr *ast.Sizeof) {
	expr.AcceptChildren(c)

	expr.Result().SetValue(types.Primitive(types.I32, core.Range{}), 0)
}

func (c *checker) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(c)

	if expr.Callee.Result().Kind == ast.InvalidResultKind {
		return // Do not cascade errors
	}

	// Check results
	ok := true
	var function *ast.Func

	if v, ok_ := expr.Callee.Result().Type.(*ast.Func); !ok_ {
		c.errorRange(expr.Callee.Range(), "Cannot call this value.")
		ok = false
	} else {
		function = v
	}

	for _, arg := range expr.Args {
		if arg.Result().Kind != ast.InvalidResultKind && arg.Result().Kind != ast.ValueResultKind {
			c.errorRange(arg.Range(), "Invalid value.")
			ok = false
		}
	}

	if !ok {
		expr.Result().SetInvalid()
		return
	}

	// Check arguments
	ok = true
	toCheck := min(len(function.Params), len(expr.Args))

	//     Check argument count
	if function.IsVariadic() {
		if len(expr.Args) < len(function.Params) {
			c.errorRange(expr.Range(), "Got '%d' arguments but function takes at least '%d'.", len(expr.Args), len(function.Params))
			ok = false
		}
	} else {
		if len(function.Params) != len(expr.Args) {
			c.errorRange(expr.Range(), "Got '%d' arguments but function only takes '%d'.", len(expr.Args), len(function.Params))
			ok = false
		}
	}

	//     Check argument types
	for i := 0; i < toCheck; i++ {
		arg := expr.Args[i]
		param := function.Params[i]

		if arg.Result().Kind == ast.InvalidResultKind {
			continue // Do not cascade errors
		}

		if !arg.Result().Type.CanAssignTo(param.Type) {
			c.errorRange(arg.Range(), "Argument with type '%s' cannot be assigned to a parameter with type '%s'.", arg.Result().Type, param.Type)
			ok = false
		}
	}

	if !ok {
		expr.Result().SetInvalid()
		return
	}

	// Set result
	expr.Result().SetValue(function.Returns, 0)
}

func (c *checker) VisitIndex(expr *ast.Index) {
	expr.AcceptChildren(c)

	if expr.Value.Result().Kind == ast.InvalidResultKind || expr.Index.Result().Kind == ast.InvalidResultKind {
		return // Do not cascade errors
	}

	// Check value
	ok := true
	var base types.Type

	if expr.Value.Result().Kind == ast.ValueResultKind {
		if v, ok := expr.Value.Result().Type.(*types.ArrayType); ok {
			base = v.Base
		} else if v, ok := expr.Value.Result().Type.(*types.PointerType); ok {
			base = v.Pointee
		}

		if base == nil {
			c.errorRange(expr.Value.Range(), "Can only index into array and pointer types, not '%s'.", expr.Value.Result().Type)
			ok = false
		}
	} else {
		c.errorRange(expr.Value.Range(), "Invalid value.")
		ok = false
	}

	// Check index
	if expr.Index.Result().Kind == ast.ValueResultKind {
		ok2 := false

		if v, ok := expr.Index.Result().Type.(*types.PrimitiveType); ok {
			if types.IsInteger(v.Kind) {
				ok2 = true
			}
		}

		if !ok2 {
			c.errorRange(expr.Index.Range(), "Can only index using integer types, not '%s'.", expr.Index.Result().Type)
			ok = false
		}
	} else {
		c.errorRange(expr.Index.Range(), "Invalid value.")
		ok = false
	}

	// Check if value is not an array initializer
	if _, ok := expr.Value.(*ast.ArrayInitializer); ok {
		c.errorRange(expr.Value.Range(), "Cannot index into a temporary array created directly from an array initializer.")
		ok = false
	}

	// Set result
	if ok {
		expr.Result().SetValue(base, ast.AssignableFlag|ast.AddressableFlag)
	} else {
		expr.Result().SetInvalid()
	}
}

func (c *checker) VisitMember(expr *ast.Member) {
	expr.AcceptChildren(c)

	if expr.Value.Result().Kind == ast.InvalidResultKind {
		return // Do not cascade errors
	}

	// Type result
	if expr.Value.Result().Kind == ast.TypeResultKind {
		// Struct
		if i, ok := expr.Value.(*ast.Identifier); ok && i.Kind == ast.StructKind {
			if v, ok := expr.Value.Result().Type.(*ast.Struct); ok {
				function, _ := c.resolver.GetMethod(v, expr.Name.Lexeme, true)

				if function == nil {
					c.errorToken(expr.Name, "Struct '%s' does not contain static method with the name '%s'.", v, expr.Name)
					expr.Result().SetInvalid()

					return
				}

				expr.Result().SetFunction(function)
				return
			}
		}

		// Enum
		if i, ok := expr.Value.(*ast.Identifier); ok && i.Kind == ast.EnumKind {
			if v, ok := expr.Value.Result().Type.(*ast.Enum); ok {
				if case_ := v.GetCase(expr.Name.Lexeme); case_ == nil {
					c.errorToken(expr.Name, "Enum '%s' does not contain case '%s'.", v, expr.Name)
				}

				expr.Result().SetValue(v, 0)
				return
			}
		}

		c.errorToken(expr.Name, "Invalid member.")
		expr.Result().SetInvalid()

		return
	}

	// Value result
	if expr.Value.Result().Kind == ast.ValueResultKind {
		// Get struct
		var s *ast.Struct

		if v, ok := expr.Value.Result().Type.(*ast.Struct); ok {
			s = v
		} else if v, ok := expr.Value.Result().Type.(*types.PointerType); ok {
			if v, ok := v.Pointee.(*ast.Struct); ok {
				s = v
			}
		}

		if s == nil {
			c.errorRange(expr.Value.Range(), "Only structs and pointers to structs can have members, not '%s'.", expr.Value.Result().Type)
			expr.Result().SetInvalid()

			return
		}

		// Check if parent expression is a call expression
		if call, ok := expr.Parent().(*ast.Call); ok && call.Callee == expr {
			function, _ := c.resolver.GetMethod(s, expr.Name.Lexeme, false)

			if function != nil {
				expr.Result().SetFunction(function)
			} else {
				c.errorToken(expr.Name, "Struct '%s' does not contain method '%s'.", s, expr.Name)
				expr.Result().SetInvalid()
			}

			return
		}

		// Check field
		_, field := s.GetField(expr.Name.Lexeme)

		if field == nil {
			c.errorToken(expr.Name, "Struct '%s' does not contain field '%s'.", s, expr.Name)
			expr.Result().SetInvalid()

			return
		}

		expr.Result().SetValue(field.Type, ast.AssignableFlag|ast.AddressableFlag)
		return
	}

	// Invalid result
	c.errorRange(expr.Value.Range(), "Invalid value.")
	expr.Result().SetInvalid()
}
