package checker

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/utils"
	"strconv"
	"strings"
)

func (c *checker) VisitParen(expr *ast.Paren) {
	expr.AcceptChildren(c)

	if expr.Expr != nil {
		*expr.Result() = *expr.Expr.Result()
		expr.Result().Flags = 0
	}
}

func (c *checker) VisitLiteral(expr *ast.Literal) {
	expr.AcceptChildren(c)

	var kind ast.PrimitiveKind
	pointer := false

	switch expr.Token().Kind {
	case scanner.Nil:
		kind = ast.Void
		pointer = true

	case scanner.True, scanner.False:
		kind = ast.Bool

	case scanner.Number:
		raw := expr.String()
		last := raw[len(raw)-1]

		if last == 'f' || last == 'F' {
			_, err := strconv.ParseFloat(raw[:len(raw)-1], 32)
			if err != nil {
				c.error(expr, "Invalid float")
				expr.Result().SetInvalid()

				return
			}

			kind = ast.F32
		} else if strings.ContainsRune(raw, '.') {
			_, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				c.error(expr, "Invalid double")
				expr.Result().SetInvalid()

				return
			}

			kind = ast.F64
		} else {
			kind = ast.I32
		}

	case scanner.Hex:
		_, err := strconv.ParseUint(expr.String()[2:], 16, 64)
		if err != nil {
			c.error(expr, "Invalid hex integer")
			expr.Result().SetInvalid()

			return
		}

		kind = ast.U32

	case scanner.Binary:
		_, err := strconv.ParseUint(expr.String()[2:], 2, 64)
		if err != nil {
			c.error(expr, "Invalid binary integer")
			expr.Result().SetInvalid()

			return
		}

		kind = ast.U32

	case scanner.Character:
		kind = ast.U8

	case scanner.String:
		kind = ast.U8
		pointer = true

	default:
		panic("checker.VisitLiteral() - Not implemented")
	}

	expr.Result().SetValue(&ast.Primitive{Kind: kind}, 0, nil)

	if pointer {
		expr.Result().SetValue(&ast.Pointer{Pointee: expr.Result().Type}, 0, nil)
	}
}

func (c *checker) VisitStructInitializer(expr *ast.StructInitializer) {
	expr.AcceptChildren(c)

	if expr.Type == nil {
		return
	}

	// Check struct
	var struct_ *ast.Struct

	if s, ok := ast.As[*ast.Struct](expr.Type); ok {
		struct_ = s
	} else {
		c.error(expr.Type, "Expected a struct")
		expr.Result().SetInvalid()

		return
	}

	if expr.New {
		expr.Result().SetValue(&ast.Pointer{Pointee: struct_}, 0, nil)
	} else {
		expr.Result().SetValue(struct_, 0, nil)
	}

	// Check fields
	assignedFields := utils.NewSet[string]()

	for _, initField := range expr.Fields {
		if initField.Name == nil {
			continue
		}

		// Check name collision
		if !assignedFields.Add(initField.Name.String()) {
			c.error(initField.Name, "Field with the name '%s' was already assigned", initField.Name)
		}

		// Check field
		_, field := struct_.GetField(initField.Name.String())
		if field == nil {
			c.error(initField.Name, "Field with the name '%s' doesn't exist on the struct '%s'", initField.Name, struct_)
			continue
		}

		// Check value result
		if initField.Value != nil {
			if initField.Value.Result().Kind == ast.InvalidResultKind {
				continue // Do not cascade errors
			}

			if initField.Value.Result().Kind != ast.ValueResultKind {
				c.error(initField.Value, "Cannot assign this value to a field with type '%s'", field.Type)
				continue
			}

			if !initField.Value.Result().Type.CanAssignTo(field.Type) {
				c.error(initField.Value, "Expected a '%s' but got '%s'", ast.PrintType(field.Type), ast.PrintType(initField.Value.Result().Type))
			}
		}
	}

	// Check malloc
	if expr.New {
		c.checkMalloc(expr)
	}
}

func (c *checker) VisitArrayInitializer(expr *ast.ArrayInitializer) {
	expr.AcceptChildren(c)

	// Check empty
	if len(expr.Values) == 0 {
		c.error(expr, "Array initializers need to have at least one value")
		expr.Result().SetInvalid()

		return
	}

	// Check values
	ok := true
	var type_ ast.Type

	for _, value := range expr.Values {
		if value.Result().Kind == ast.InvalidResultKind {
			ok = false
			continue
		}

		if value.Result().Kind != ast.ValueResultKind {
			c.error(value, "Invalid value")
			ok = false

			continue
		}

		if type_ == nil {
			type_ = value.Result().Type
		} else {
			if !value.Result().Type.CanAssignTo(type_) {
				c.error(value, "Expected a '%s' but got '%s'", ast.PrintType(type_), ast.PrintType(value.Result().Type))
				ok = false
			}
		}
	}

	if ok {
		expr.Result().SetValue(&ast.Array{Base: type_, Count: uint32(len(expr.Values))}, 0, nil)
	} else {
		expr.Result().SetInvalid()
	}
}

func (c *checker) VisitAllocateArray(expr *ast.AllocateArray) {
	expr.AcceptChildren(c)

	c.checkMalloc(expr)

	// Check count
	if expr.Count != nil {
		if expr.Count.Result().Kind != ast.ValueResultKind {
			c.error(expr.Count, "Invalid value")
			expr.Result().SetInvalid()

			return
		}

		if !ast.IsPrimitive(expr.Count.Result().Type, ast.I32) {
			c.error(expr.Count, "Expected an 'i32' but got '%s'", ast.PrintType(expr.Count.Result().Type))
			expr.Result().SetInvalid()

			return
		}
	}

	// Set result
	type_ := expr.Type

	if type_ == nil {
		type_ = &ast.Primitive{Kind: ast.Void}
	}

	expr.Result().SetValue(&ast.Pointer{Pointee: type_}, 0, nil)
}

func (c *checker) VisitUnary(expr *ast.Unary) {
	expr.AcceptChildren(c)

	if expr.Operator == nil || expr.Value == nil {
		if expr.Value != nil {
			*expr.Result() = *expr.Value.Result()
			expr.Result().Flags = 0
		}

		return
	}

	// Check value
	result := expr.Value.Result()

	if result.Kind == ast.InvalidResultKind {
		return // Do not cascade errors
	}

	if expr.Prefix {
		// Prefix
		switch expr.Operator.Token().Kind {
		case scanner.Bang:
			c.expectPrimitiveValue(expr.Value, ast.Bool)
			expr.Result().SetValue(&ast.Primitive{Kind: ast.Bool}, 0, nil)

		case scanner.Minus:
			if result.Kind != ast.ValueResultKind {
				c.error(expr.Value, "Cannot negate this value")
				expr.Result().SetInvalid()

				return
			}

			if v, ok := ast.As[*ast.Primitive](result.Type); ok {
				if ast.IsFloating(v.Kind) || ast.IsSigned(v.Kind) {
					expr.Result().SetValue(result.Type, 0, nil)
					return
				}
			}

			c.error(expr.Value, "Expected either a floating pointer number or signed integer but got a '%s'", ast.PrintType(result.Type))
			expr.Result().SetInvalid()

		case scanner.Ampersand:
			if result.IsAddressable() {
				expr.Result().SetValue(&ast.Pointer{Pointee: result.Type}, 0, nil)
			} else {
				c.error(expr.Value, "Cannot take address of this expression")
				expr.Result().SetInvalid()
			}

		case scanner.Star:
			if result.Kind != ast.ValueResultKind {
				c.error(expr.Value, "Cannot dereference this value")
				expr.Result().SetInvalid()

				return
			}

			if p, ok := ast.As[*ast.Pointer](result.Type); ok {
				expr.Result().SetValue(p.Pointee, ast.AssignableFlag, nil)
			} else {
				c.error(expr.Value, "Can only dereference pointer types, not '%s'", ast.PrintType(result.Type))
				expr.Result().SetInvalid()
			}

		case scanner.PlusPlus, scanner.MinusMinus:
			if !result.IsAddressable() {
				c.error(expr.Value, "Invalid value")
				expr.Result().SetInvalid()

				return
			}

			if type_, ok := ast.As[*ast.Primitive](result.Type); !ok || (!ast.IsInteger(type_.Kind) && !ast.IsFloating(type_.Kind)) {
				c.error(expr.Value, "Cannot increment or decrement '%s'", ast.PrintType(result.Type))
				expr.Result().SetInvalid()

				return
			}

			expr.Result().SetValue(result.Type, 0, nil)

		case scanner.FuncPtr:
			if f, ok := result.Callable().(*ast.Func); ok && f.Method() != nil {
				c.error(expr.Value, "Cannot take address of a non-static method")
			}

			expr.Result().SetValue(result.Type, 0, nil)

		default:
			panic("checker.VisitUnary() - Invalid unary prefix operator")
		}
	} else {
		// Postfix
		switch expr.Operator.Token().Kind {
		case scanner.PlusPlus, scanner.MinusMinus:
			if !result.IsAddressable() {
				c.error(expr.Value, "Invalid value")
				expr.Result().SetInvalid()

				return
			}

			if type_, ok := ast.As[*ast.Primitive](result.Type); !ok || (!ast.IsInteger(type_.Kind) && !ast.IsFloating(type_.Kind)) {
				c.error(expr.Value, "Cannot increment or decrement '%s'", ast.PrintType(result.Type))
				expr.Result().SetInvalid()

				return
			}

			expr.Result().SetValue(result.Type, 0, nil)

		default:
			panic("checker.VisitUnary() - Invalid unary postfix operator")
		}
	}
}

func (c *checker) VisitBinary(expr *ast.Binary) {
	expr.AcceptChildren(c)

	if expr.Left == nil || expr.Operator == nil || expr.Right == nil {
		if expr.Left != nil {
			*expr.Result() = *expr.Left.Result()
			expr.Result().Flags = 0
		}

		return
	}

	if expr.Left.Result().Kind == ast.InvalidResultKind || expr.Right.Result().Kind == ast.InvalidResultKind {
		return // // Do not cascade errors
	}

	// Check results
	ok := true

	if expr.Left.Result().Kind != ast.ValueResultKind {
		c.error(expr.Left, "Invalid value")
		ok = false
	}

	if expr.Right.Result().Kind != ast.ValueResultKind {
		c.error(expr.Right, "Invalid value")
		ok = false
	}

	if !ok {
		expr.Result().SetInvalid()
		return
	}

	leftType := expr.Left.Result().Type
	rightType := expr.Right.Result().Type

	// Check based on the operator
	if scanner.IsArithmetic(expr.Operator.Token().Kind) {
		// Arithmetic
		if left, ok := ast.As[*ast.Primitive](leftType); ok {
			if right, ok := ast.As[*ast.Primitive](rightType); ok {
				if ast.IsNumber(left.Kind) && ast.IsNumber(right.Kind) && left.Equals(right) {
					expr.Result().SetValue(leftType, 0, nil)
					return
				}
			}
		}

		c.error(expr, "Expected two equal number types")
		expr.Result().SetInvalid()
	} else if scanner.IsEquality(expr.Operator.Token().Kind) {
		// Equality
		valid := false

		if leftType.Equals(rightType) {
			// left type == right type
			valid = true
		} else if left, ok := ast.As[*ast.Primitive](leftType); ok {
			// integer == integer || floating == floating
			if right, ok := ast.As[*ast.Primitive](rightType); ok {
				if (ast.IsInteger(left.Kind) && ast.IsInteger(right.Kind)) || (ast.IsFloating(left.Kind) && ast.IsFloating(right.Kind)) {
					valid = true
				}
			}
		} else if left, ok := ast.As[*ast.Pointer](leftType); ok {
			if right, ok := ast.As[*ast.Pointer](rightType); ok {
				// *void == *? || *? == *void
				if ast.IsPrimitive(left.Pointee, ast.Void) || ast.IsPrimitive(right.Pointee, ast.Void) {
					valid = true
				}
			}
		}

		if !valid {
			c.error(expr, "Cannot check equality for '%s' and '%s'", ast.PrintType(leftType), ast.PrintType(rightType))
			expr.Result().SetInvalid()
		} else {
			expr.Result().SetValue(&ast.Primitive{Kind: ast.Bool}, 0, nil)
		}
	} else if scanner.IsComparison(expr.Operator.Token().Kind) {
		// Comparison
		if left, ok := ast.As[*ast.Primitive](leftType); ok {
			if right, ok := ast.As[*ast.Primitive](rightType); ok {
				if !ast.IsNumber(left.Kind) || !ast.IsNumber(right.Kind) || !left.Equals(right) {
					c.error(expr, "Expected two equal number types")
					expr.Result().SetInvalid()

					return
				}
			}
		}

		expr.Result().SetValue(&ast.Primitive{Kind: ast.Bool}, 0, nil)
	} else if scanner.IsBitwise(expr.Operator.Token().Kind) {
		// Bitwise
		if left, ok := ast.As[*ast.Primitive](leftType); ok {
			if right, ok := ast.As[*ast.Primitive](rightType); ok {
				if left.Equals(right) && ast.IsInteger(left.Kind) {
					expr.Result().SetValue(leftType, 0, nil)
					return
				}
			}
		}

		c.error(expr, "Expected two equal integer types")
		expr.Result().SetInvalid()
	} else {
		// Error
		panic("checker.VisitBinary() - Invalid operator kind")
	}
}

func (c *checker) VisitLogical(expr *ast.Logical) {
	expr.AcceptChildren(c)

	// Check expressions
	c.expectPrimitiveValue(expr.Left, ast.Bool)
	c.expectPrimitiveValue(expr.Right, ast.Bool)

	// Set type
	expr.Result().SetValue(&ast.Primitive{Kind: ast.Bool}, 0, nil)
}

func (c *checker) VisitIdentifier(expr *ast.Identifier) {
	expr.AcceptChildren(c)

	// Function / function pointer
	if parentWantsFunction(expr) {
		// Function
		if f := c.resolver.GetFunction(expr.String()); f != nil {
			expr.Result().SetCallable(f, f)
			return
		}

		// Variable
		if variable := c.getVariable(expr.String()); variable != nil {
			if f, ok := ast.As[*ast.Func](variable.type_); ok {
				variable.used = true

				expr.Result().SetCallable(f, variable.node)
				return
			}
		}

		// Error
		c.error(expr, "Unknown function")
		expr.Result().SetInvalid()

		return
	}

	// Resolver
	if r := c.resolver.GetChild(expr.String()); r != nil {
		expr.Result().SetResolver(r)
		return
	}

	// Type
	if t := c.resolver.GetType(expr.String()); t != nil {
		expr.Result().SetType(t)
		return
	}

	// Variable
	if variable := c.getVariable(expr.String()); variable != nil {
		variable.used = true

		expr.Result().SetValue(variable.type_, ast.AssignableFlag|ast.AddressableFlag, variable.node)
		return
	}

	// Error
	c.error(expr, "Unknown identifier")
	expr.Result().SetInvalid()
}

func (c *checker) VisitAssignment(expr *ast.Assignment) {
	expr.AcceptChildren(c)

	// Check assignee
	if expr.Assignee == nil || expr.Assignee.Result().Kind == ast.InvalidResultKind {
		expr.Result().SetInvalid()
		return
	}

	if !expr.Assignee.Result().IsAssignable() {
		c.error(expr.Assignee, "Cannot assign to this value")
	}

	expr.Result().SetValue(expr.Assignee.Result().Type, 0, nil)

	// Check operator and value
	if expr.Operator == nil || expr.Value == nil || expr.Value.Result().Kind == ast.InvalidResultKind {
		return
	}

	if expr.Value.Result().Kind != ast.ValueResultKind {
		c.error(expr.Value, "Invalid value")
	}

	// Check type
	if expr.Operator.Token().Kind == scanner.Equal {
		// Equal
		if !expr.Value.Result().Type.CanAssignTo(expr.Assignee.Result().Type) {
			c.error(expr.Value, "Expected a '%s' but got '%s'", ast.PrintType(expr.Assignee.Result().Type), ast.PrintType(expr.Value.Result().Type))
		}
	} else {
		if scanner.IsArithmetic(expr.Operator.Token().Kind) {
			// Arithmetic
			valid := false

			if assignee, ok := ast.As[*ast.Primitive](expr.Assignee.Result().Type); ok {
				if value, ok := ast.As[*ast.Primitive](expr.Value.Result().Type); ok {
					if ast.IsNumber(assignee.Kind) && ast.IsNumber(value.Kind) && assignee.Equals(value) {
						valid = true
					}
				}
			}

			if !valid {
				c.error(expr.Value, "Expected two equal number types")
			}
		} else if scanner.IsBitwise(expr.Operator.Token().Kind) {
			// Bitwise
			valid := false

			if left, ok := ast.As[*ast.Primitive](expr.Assignee.Result().Type); ok {
				if right, ok := ast.As[*ast.Primitive](expr.Value.Result().Type); ok {
					if left.Equals(right) && ast.IsInteger(left.Kind) {
						valid = true
					}
				}
			}

			if !valid {
				c.error(expr.Value, "Expected two equal integer types")
			}
		} else {
			panic("checker.VisitAssignment() - Invalid operator")
		}
	}
}

func (c *checker) VisitCast(expr *ast.Cast) {
	expr.AcceptChildren(c)

	// Check nil and results
	expr.Result().SetInvalid()

	if expr.Value == nil || expr.Target == nil {
		return
	}

	if expr.Value.Result().Kind == ast.InvalidResultKind {
		return
	}

	if expr.Value.Result().Kind != ast.ValueResultKind {
		c.error(expr.Value, "Cannot cast this value")
		return
	}

	expr.Result().SetValue(expr.Target, 0, nil)

	// Check type
	if ast.IsPrimitive(expr.Value.Result().Type, ast.Void) || ast.IsPrimitive(expr.Target, ast.Void) {
		// void
		c.error(expr, "Cannot cast to or from type 'void'")
	} else if _, ok := ast.As[*ast.Enum](expr.Value.Result().Type); ok {
		// enum to non integer
		if to, ok := ast.As[*ast.Primitive](expr.Target); !ok || !ast.IsInteger(to.Kind) {
			c.error(expr, "Can only cast enums to integers, not '%s'", ast.PrintType(to))
		}
	} else if _, ok := ast.As[*ast.Enum](expr.Target); ok {
		// non integer to enum
		if from, ok := ast.As[*ast.Primitive](expr.Value.Result().Type); !ok || !ast.IsInteger(from.Kind) {
			c.error(expr, "Can only cast to enums from integers, not '%s'", ast.PrintType(from))
		}
	}
}

func (c *checker) VisitTypeCall(expr *ast.TypeCall) {
	expr.AcceptChildren(c)

	expr.Result().SetValue(&ast.Primitive{Kind: ast.I32}, 0, nil)
}

func (c *checker) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(c)

	// Check callee
	if expr.Callee == nil || expr.Callee.Result().Kind == ast.InvalidResultKind {
		expr.Result().SetInvalid()
		return
	}

	var function *ast.Func

	if f, ok := ast.As[*ast.Func](expr.Callee.Result().Type); ok {
		function = f
	} else {
		c.error(expr.Callee, "Cannot call this value")
		expr.Result().SetInvalid()
		return
	}

	expr.Result().SetValue(function.Returns, 0, nil)

	// Check argument count
	if function.IsVariadic() {
		if len(expr.Args) < len(function.Params) {
			c.error(expr, "Got '%d' arguments but function takes at least '%d'", len(expr.Args), len(function.Params))
		}
	} else {
		if len(function.Params) != len(expr.Args) {
			c.error(expr, "Got '%d' arguments but function only takes '%d'", len(expr.Args), len(function.Params))
		}
	}

	// Check argument types
	toCheck := min(len(function.Params), len(expr.Args))

	for i := 0; i < toCheck; i++ {
		arg := expr.Args[i]
		param := function.Params[i]

		if arg.Result().Kind == ast.InvalidResultKind {
			continue
		}

		if arg.Result().Kind != ast.ValueResultKind {
			c.error(arg, "Invalid value")
			continue
		}

		if !arg.Result().Type.CanAssignTo(param.Type) {
			c.error(arg, "Argument with type '%s' cannot be assigned to a parameter with type '%s'", ast.PrintType(arg.Result().Type), ast.PrintType(param.Type))
		}
	}
}

func (c *checker) VisitIndex(expr *ast.Index) {
	expr.AcceptChildren(c)

	if expr.Value == nil || expr.Value.Result().Kind == ast.InvalidResultKind {
		return // Do not cascade errors
	}

	// Check value
	var base ast.Type

	if expr.Value.Result().Kind == ast.ValueResultKind {
		if v, ok := ast.As[*ast.Array](expr.Value.Result().Type); ok {
			base = v.Base
		} else if v, ok := ast.As[*ast.Pointer](expr.Value.Result().Type); ok {
			base = v.Pointee
		}

		if base == nil {
			c.error(expr.Value, "Can only index into array and pointer types, not '%s'", ast.PrintType(expr.Value.Result().Type))
		}
	} else {
		c.error(expr.Value, "Invalid value")
	}

	if base != nil {
		expr.Result().SetValue(base, ast.AssignableFlag|ast.AddressableFlag, nil)
	} else {
		expr.Result().SetInvalid()
	}

	// Check index
	if expr.Index != nil && expr.Index.Result().Kind != ast.InvalidResultKind {
		if expr.Index.Result().Kind == ast.ValueResultKind {
			ok2 := false

			if v, ok := ast.As[*ast.Primitive](expr.Index.Result().Type); ok {
				if ast.IsInteger(v.Kind) {
					ok2 = true
				}
			}

			if !ok2 {
				c.error(expr.Index, "Can only index using integer types, not '%s'", ast.PrintType(expr.Index.Result().Type))
			}
		} else {
			c.error(expr.Index, "Invalid value")
		}
	}

	// Check if value is not an array initializer
	if _, ok := expr.Value.(*ast.ArrayInitializer); ok {
		c.error(expr.Value, "Cannot index into a temporary array created directly from an array initializer")
	}
}

func (c *checker) VisitMember(expr *ast.Member) {
	expr.AcceptChildren(c)

	expr.Result().SetInvalid()

	if expr.Value == nil || expr.Name == nil || expr.Value.Result().Kind == ast.InvalidResultKind {
		return
	}

	switch expr.Value.Result().Kind {
	case ast.TypeResultKind:
		switch t := expr.Value.Result().Type.(type) {
		case *ast.Struct:
			// Callable
			if parentWantsFunction(expr) {
				function := c.resolver.GetMethod(t, expr.Name.String(), true)

				if function == nil {
					_, field := t.GetStaticField(expr.Name.String())

					if field != nil {
						if f, ok := ast.As[*ast.Func](field.Type); ok {
							expr.Result().SetCallable(f, field)
							return
						}
					}

					c.error(expr.Name, "Struct '%s' does not contain static method with the name '%s'", ast.PrintType(t), expr.Name)
					return
				}

				expr.Result().SetCallable(function, function)
				return
			}

			// Field
			_, field := t.GetStaticField(expr.Name.String())

			if field == nil {
				c.error(expr.Name, "Struct '%s' does not contain static field '%s'", ast.PrintType(t), expr.Name)
				return
			}

			expr.Result().SetValue(field.Type, ast.AssignableFlag|ast.AddressableFlag, field)
			return

		case *ast.Enum:
			case_ := t.GetCase(expr.Name.String())

			if case_ == nil {
				c.error(expr.Name, "Enum '%s' does not contain case '%s'", ast.PrintType(t), expr.Name)
				return
			}

			expr.Result().SetValue(t, 0, case_)
			return

		default:
			c.error(expr.Value, "Invalid type")
			return
		}

	case ast.ResolverResultKind:
		resolver := expr.Value.Result().Resolver()

		// Function
		if parentWantsFunction(expr) {
			if f := resolver.GetFunction(expr.Name.String()); f != nil {
				expr.Result().SetCallable(f, f)
				return
			}
		}

		// Resolver
		if r := resolver.GetChild(expr.String()); r != nil {
			expr.Result().SetResolver(r)
			return
		}

		// Type
		if t := resolver.GetType(expr.String()); t != nil {
			expr.Result().SetType(t)
			return
		}

		// Error
		c.error(expr.Name, "Unknown identifier")
		expr.Result().SetInvalid()

	case ast.ValueResultKind:
		// Get struct
		var s *ast.Struct

		if v, ok := expr.Value.Result().Type.(*ast.Struct); ok {
			s = v
		} else if v, ok := ast.As[*ast.Pointer](expr.Value.Result().Type); ok {
			if v, ok := ast.As[*ast.Struct](v.Pointee); ok {
				s = v
			}
		}

		if s == nil {
			c.error(expr.Value, "Only structs and pointers to structs can have members, not '%s'", ast.PrintType(expr.Value.Result().Type))
			return
		}

		// Callable
		if parentWantsFunction(expr) {
			function := c.resolver.GetMethod(s, expr.Name.String(), false)

			if function == nil {
				_, field := s.GetField(expr.Name.String())

				if field != nil {
					if f, ok := ast.As[*ast.Func](field.Type); ok {
						expr.Result().SetCallable(f, field)
						return
					}
				}

				c.error(expr.Name, "Struct '%s' does not contain method with the name '%s'", ast.PrintType(s), expr.Name)
				return
			}

			expr.Result().SetCallable(function, function)
			return
		}

		// Field
		_, field := s.GetField(expr.Name.String())

		if field == nil {
			c.error(expr.Name, "Struct '%s' does not contain field '%s'", ast.PrintType(s), expr.Name)
			return
		}

		expr.Result().SetValue(field.Type, ast.AssignableFlag|ast.AddressableFlag, field)
		return

	default:
		c.error(expr.Value, "Invalid value")
		return
	}
}

// Utils

func parentWantsFunction(expr ast.Expr) bool {
	switch parent := expr.Parent().(type) {
	case *ast.Call:
		return parent.Callee == expr

	case *ast.Unary:
		return parent.Operator.Token().Kind == scanner.FuncPtr

	default:
		return false
	}
}

func (c *checker) checkMalloc(expr ast.Expr) {
	function := c.resolver.GetFunction("malloc")

	if function == nil {
		c.error(expr, "Malloc function not found")
		return
	}

	if len(function.Params) != 1 || !ast.IsPrimitive(function.Params[0].Type, ast.U64) {
		c.error(expr, "Malloc parameter needs to be a u64")
	}

	if _, ok := ast.As[*ast.Pointer](function.Returns); !ok {
		c.error(expr, "Malloc needs to return a pointer")
	}
}
