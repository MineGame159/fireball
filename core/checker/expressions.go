package checker

import (
	"fireball/core/ast"
	"fireball/core/common"
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
	var struct_ ast.StructType

	if s, ok := ast.As[ast.StructType](expr.Type); ok {
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
		field := struct_.FieldName(initField.Name.String())
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

			c.checkRequired(field.Type(), initField.Value)
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
			c.checkRequired(type_, value)
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
			if result.Kind != ast.CallableResultKind {
				c.error(expr.Value, "Invalid value")
				expr.Result().SetInvalid()

				return
			}

			if f, ok := result.Callable().(*ast.Func); ok && f.Receiver() != nil {
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

	expr.Result().SetInvalid()

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
		return
	}

	// Check
	c.checkBinary(expr, expr.Left, expr.Right, expr.Operator, false)
}

func (c *checker) VisitLogical(expr *ast.Logical) {
	expr.AcceptChildren(c)

	// Check expressions
	type_ := ast.Primitive{Kind: ast.Bool}

	c.checkRequired(&type_, expr.Left)
	c.checkRequired(&type_, expr.Right)

	// Set type
	expr.Result().SetValue(&ast.Primitive{Kind: ast.Bool}, 0, nil)
}

func (c *checker) VisitIdentifier(expr *ast.Identifier) {
	expr.AcceptChildren(c)

	if expr.Name == nil {
		expr.Result().SetInvalid()
		return
	}

	// Function / function pointer
	if parentWantsFunction(expr) {
		var function ast.FuncType
		var node ast.Node

		// Function
		if f := c.resolver.GetFunction(expr.Name.String()); f != nil {
			function = f
			node = f
		}

		// Global variable
		if variable := c.resolver.GetVariable(expr.Name.String()); variable != nil && variable.Type != nil {
			if f, ok := ast.As[ast.FuncType](variable.Type); ok {
				function = f
				node = variable
			}
		}

		// Variable
		if variable := c.getVariable(expr.Name.String()); variable != nil {
			if f, ok := ast.As[ast.FuncType](variable.type_); ok {
				variable.used = true

				function = f
				node = variable.node
			}
		}

		// Ok
		if function != nil {
			if f, ok := node.(*ast.Func); ok {
				c.specializeFuncIfNeeded(expr, f, expr.GenericArgs)
			} else {
				expr.Result().SetCallable(function, node)
			}

			return
		}

		// Error
		c.error(expr, "Unknown function")
		expr.Result().SetInvalid()

		return
	}

	if len(expr.GenericArgs) != 0 {
		errorSlice(c, expr.GenericArgs, "This identifier doesn't have any generic parameters")
	}

	// Resolver
	if r := c.resolver.GetChild(expr.Name.String()); r != nil {
		expr.Result().SetResolver(r)
		return
	}

	// Type
	if t := c.resolver.GetType(expr.Name.String()); t != nil {
		expr.Result().SetType(t)
		return
	}

	// Primitive type
	for kind := ast.Void; kind <= ast.F64; kind++ {
		if expr.Name.String() == kind.String() {
			expr.Result().SetType(&ast.Primitive{Kind: kind})
			return
		}
	}

	// Global variable
	if variable := c.resolver.GetVariable(expr.Name.String()); variable != nil && variable.Type != nil {
		expr.Result().SetValue(variable.Type, ast.AssignableFlag|ast.AddressableFlag, variable)
		return
	}

	// Variable
	if variable := c.getVariable(expr.Name.String()); variable != nil {
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
		c.checkRequired(expr.Assignee.Result().Type, expr.Value)
	} else {
		// Binary
		c.checkBinary(expr, expr.Assignee, expr.Value, expr.Operator, true)

		if expr.Result().Kind == ast.InvalidResultKind {
			panic("checker.VisitAssignment() - Not implemented")
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

	// Check based on the operator
	switch expr.Operator.Token().Kind {
	case scanner.As:
		if _, ok := common.GetCast(expr.Value.Result().Type, expr.Target); !ok {
			c.error(expr, "Cannot cast type '%s' to type '%s'", ast.PrintType(expr.Value.Result().Type), ast.PrintType(expr.Target))
		}

		expr.Result().SetValue(expr.Target, 0, nil)

	case scanner.Is:
		if _, ok := ast.As[*ast.Interface](expr.Value.Result().Type); !ok {
			c.error(expr.Value, "Runtime type checking is only supported for interfaces")
		}

		expr.Result().SetValue(&ast.Primitive{Kind: ast.Bool}, 0, nil)

	default:
		panic("checker.VisitCast() - Not implemented")
	}
}

func (c *checker) VisitTypeCall(expr *ast.TypeCall) {
	expr.AcceptChildren(c)

	expr.Result().SetValue(&ast.Primitive{Kind: ast.I32}, 0, nil)
}

func (c *checker) VisitTypeof(expr *ast.Typeof) {
	expr.AcceptChildren(c)

	expr.Result().SetValue(&ast.Primitive{Kind: ast.U32}, 0, nil)

	// Check arg
	if expr.Arg != nil && expr.Arg.Result().Kind == ast.InvalidResultKind {
		c.error(expr.Arg, "Invalid expression")
	}
}

func (c *checker) VisitCall(expr *ast.Call) {
	expr.AcceptChildren(c)

	// Check callee
	if expr.Callee == nil || expr.Callee.Result().Kind == ast.InvalidResultKind {
		expr.Result().SetInvalid()
		return
	}

	var function ast.FuncType

	if f, ok := ast.As[ast.FuncType](expr.Callee.Result().Type); ok {
		function = f
	} else {
		c.error(expr.Callee, "Cannot call this value")
		expr.Result().SetInvalid()
		return
	}

	expr.Result().SetValue(function.Returns(), 0, nil)

	// Check argument count
	if function.Underlying().IsVariadic() {
		if len(expr.Args) < function.ParameterCount() {
			c.error(expr, "Got '%d' arguments but function takes at least '%d'", len(expr.Args), function.ParameterCount())
		}
	} else {
		if function.ParameterCount() != len(expr.Args) {
			c.error(expr, "Got '%d' arguments but function takes '%d'", len(expr.Args), function.ParameterCount())
		}
	}

	// Check argument types
	toCheck := min(function.ParameterCount(), len(expr.Args))

	for i := 0; i < toCheck; i++ {
		arg := expr.Args[i]
		param := function.ParameterIndex(i)

		if arg.Result().Kind == ast.InvalidResultKind {
			continue
		}

		if arg.Result().Kind != ast.ValueResultKind {
			c.error(arg, "Invalid value")
			continue
		}

		c.checkRequired(param.Type, arg)
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
		case ast.StructType:
			// Callable
			if parentWantsFunction(expr) {
				function := c.resolver.GetMethod(t, expr.Name.String(), true)

				if function != nil {
					c.specializeFuncIfNeeded(expr, function, expr.GenericArgs)
					return
				}
			}

			if len(expr.GenericArgs) != 0 {
				c.error(expr.Name, "")
			}

			if parentWantsFunction(expr) {
				field := t.StaticFieldName(expr.Name.String())

				if field != nil {
					if f, ok := ast.As[ast.FuncType](field.Type()); ok {
						expr.Result().SetCallable(f, field)
						return
					}
				}

				c.error(expr.Name, "Struct '%s' does not contain static method with the name '%s'", ast.PrintType(t), expr.Name)
				return
			}

			// Field
			field := t.StaticFieldName(expr.Name.String())

			if field == nil {
				c.error(expr.Name, "Struct '%s' does not contain static field '%s'", ast.PrintType(t), expr.Name)
				return
			}

			expr.Result().SetValue(field.Type(), ast.AssignableFlag|ast.AddressableFlag, field)
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

		// Callable
		if parentWantsFunction(expr) {
			// Function
			if f := resolver.GetFunction(expr.Name.String()); f != nil {
				c.specializeFuncIfNeeded(expr, f, expr.GenericArgs)
				return
			}
		}

		if len(expr.GenericArgs) != 0 {
			c.error(expr.Name, "")
		}

		if parentWantsFunction(expr) {
			// Global variable
			if variable := resolver.GetVariable(expr.Name.String()); variable != nil && variable.Type != nil {
				if f, ok := ast.As[ast.FuncType](variable.Type); ok {
					expr.Result().SetCallable(f, variable)
					return
				}
			}

			// Error
			c.error(expr.Name, "Unknown identifier")
			expr.Result().SetInvalid()

			return
		}

		// Resolver
		if r := resolver.GetChild(expr.Name.String()); r != nil {
			expr.Result().SetResolver(r)
			return
		}

		// Type
		if t := resolver.GetType(expr.Name.String()); t != nil {
			expr.Result().SetType(t)
			return
		}

		// Global variable
		if variable := resolver.GetVariable(expr.Name.String()); variable != nil && variable.Type != nil {
			expr.Result().SetValue(variable.Type, ast.AssignableFlag|ast.AddressableFlag, variable)
			return
		}

		// Error
		c.error(expr.Name, "Unknown identifier")
		expr.Result().SetInvalid()

	case ast.ValueResultKind:
		// Get struct
		var s ast.StructType

		if v, ok := ast.As[ast.StructType](expr.Value.Result().Type); ok {
			s = v
		} else if v, ok := ast.As[*ast.Pointer](expr.Value.Result().Type); ok {
			if v, ok := ast.As[ast.StructType](v.Pointee); ok {
				s = v
			}
		}

		if s == nil {
			if inter, ok := ast.As[*ast.Interface](expr.Value.Result().Type); ok {
				// Interface
				method, _ := inter.GetMethod(expr.Name.String())

				if method != nil {
					expr.Result().SetCallable(method, method)
					return
				}

				c.error(expr.Name, "Interface '%s' does not contain method with the name '%s'", ast.PrintType(inter), expr.Name)
				return
			}

			c.error(expr.Value, "Only structs, pointers to structs and interfaces can have members, not '%s'", ast.PrintType(expr.Value.Result().Type))
			return
		}

		// Callable
		if parentWantsFunction(expr) {
			function := c.resolver.GetMethod(s, expr.Name.String(), false)

			if function != nil {
				c.specializeFuncIfNeeded(expr, function, expr.GenericArgs)
				return
			}
		}

		if len(expr.GenericArgs) != 0 {
			c.error(expr.Name, "")
		}

		if parentWantsFunction(expr) {
			field := s.FieldName(expr.Name.String())

			if field != nil {
				if f, ok := ast.As[*ast.Func](field.Type()); ok {
					expr.Result().SetCallable(f, field)
					return
				}
			}

			c.error(expr.Name, "Struct '%s' does not contain method with the name '%s'", ast.PrintType(s), expr.Name)
			return
		}

		// Field
		field := s.FieldName(expr.Name.String())

		if field == nil {
			c.error(expr.Name, "Struct '%s' does not contain field '%s'", ast.PrintType(s), expr.Name)
			return
		}

		expr.Result().SetValue(field.Type(), ast.AssignableFlag|ast.AddressableFlag, field)
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

func (c *checker) specializeFuncIfNeeded(expr ast.Expr, f ast.FuncType, genericArgs []ast.Type) {
	if sf, ok := f.(ast.SpecializableFunc); ok {
		if len(genericArgs) != len(sf.Generics()) {
			if genericArgs != nil {
				errorSlice(c, genericArgs, "Got '%d' generic arguments but function takes '%d'", len(genericArgs), len(sf.Generics()))
			} else {
				c.error(expr, "Got '%d' generic arguments but function takes '%d'", len(genericArgs), len(sf.Generics()))
			}
		}

		specialized := sf.Specialize(genericArgs)
		expr.Result().SetCallable(specialized, specialized)
	} else {
		expr.Result().SetCallable(f, f)
	}
}

func (c *checker) checkBinary(expr, left, right ast.Expr, operator *ast.Token, assignment bool) {
	// Implicitly cast between left and right types
	leftType := left.Result().Type
	rightType := right.Result().Type

	castType := leftType
	_, castOk := common.GetImplicitCast(rightType, leftType)

	if !castOk {
		if assignment {
			c.error(expr, "Expected a '%s' but got a '%s'", ast.PrintType(leftType), ast.PrintType(rightType))
			return
		}

		castType = rightType
		_, castOk = common.GetImplicitCast(leftType, rightType)
	}

	// Arithmetic
	if scanner.IsArithmetic(operator.Token().Kind) {
		if left, ok := ast.As[*ast.Primitive](leftType); ok {
			if right, ok := ast.As[*ast.Primitive](rightType); ok {
				if ast.IsNumber(left.Kind) && ast.IsNumber(right.Kind) && castOk {
					expr.Result().SetValue(castType, 0, nil)
					return
				}
			}
		}

		c.error(expr, "Operator '%s' cannot be applied to '%s' and '%s'", operator.String(), ast.PrintType(leftType), ast.PrintType(rightType))
		return
	}

	// Equality
	if !assignment && scanner.IsEquality(operator.Token().Kind) {
		if castOk {
			expr.Result().SetValue(&ast.Primitive{Kind: ast.Bool}, 0, nil)
			return
		}

		if _, ok := ast.As[*ast.Interface](leftType); ok {
			if _, ok := ast.As[*ast.Pointer](rightType); ok {
				expr.Result().SetValue(&ast.Primitive{Kind: ast.Bool}, 0, nil)
				return
			}
		}

		c.error(expr, "Operator '%s' cannot be applied to '%s' and '%s'", operator.String(), ast.PrintType(leftType), ast.PrintType(rightType))
		return
	}

	// Comparison
	if scanner.IsComparison(operator.Token().Kind) {
		if left, ok := ast.As[*ast.Primitive](leftType); ok {
			if right, ok := ast.As[*ast.Primitive](rightType); ok {
				if ast.IsNumber(left.Kind) && ast.IsNumber(right.Kind) && castOk {
					expr.Result().SetValue(&ast.Primitive{Kind: ast.Bool}, 0, nil)
					return
				}
			}
		}

		c.error(expr, "Operator '%s' cannot be applied to '%s' and '%s'", operator.String(), ast.PrintType(leftType), ast.PrintType(rightType))
		return
	}

	// Bitwise
	if !assignment && scanner.IsBitwise(operator.Token().Kind) {
		if left, ok := ast.As[*ast.Primitive](leftType); ok {
			if right, ok := ast.As[*ast.Primitive](rightType); ok {
				if ast.IsInteger(left.Kind) && ast.IsInteger(right.Kind) && castOk {
					expr.Result().SetValue(castType, 0, nil)
					return
				}
			}
		}

		c.error(expr, "Operator '%s' cannot be applied to '%s' and '%s'", operator.String(), ast.PrintType(leftType), ast.PrintType(rightType))
		return
	}
}

func (c *checker) checkMalloc(expr ast.Expr) {
	function := c.resolver.GetFunction("malloc")

	if function == nil {
		c.error(expr, "Malloc function not found")
		return
	}

	if function.ParameterCount() != 1 || !ast.IsPrimitive(function.ParameterIndex(0).Type, ast.U64) {
		c.error(expr, "Malloc parameter needs to be a u64")
	}

	if _, ok := ast.As[*ast.Pointer](function.Returns()); !ok {
		c.error(expr, "Malloc needs to return a pointer")
	}
}
