package checker

import (
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
	"fireball/core/utils"
	"strconv"
)

func (c *checker) VisitStruct(decl *ast.Struct) {
	decl.AcceptChildren(c)

	// Check static fields
	fields := utils.NewSet[string]()

	for _, field := range decl.StaticFields {
		// Check name collision
		if !fields.Add(field.Name.String()) {
			c.error(field.Name, "Static field with the name '%s' already exists", field.Name)
		}

		// Check void type
		if ast.IsPrimitive(field.Type, ast.Void) {
			c.error(field.Name, "Static field cannot be of type 'void'")
		}
	}

	// Check fields
	fields = utils.NewSet[string]()

	for _, field := range decl.Fields {
		// Check name collision
		if !fields.Add(field.Name.String()) {
			c.error(field.Name, "Field with the name '%s' already exists", field.Name)
		}

		// Check void type
		if ast.IsPrimitive(field.Type, ast.Void) {
			c.error(field.Name, "Field cannot be of type 'void'")
		}
	}
}

func (c *checker) VisitImpl(decl *ast.Impl) {
	if decl.Type != nil {
		c.pushScope()
		c.addVariable(ast.NewToken(cst.Node{}, scanner.Token{Kind: scanner.Identifier, Lexeme: "this"}), decl.Type)
	}

	decl.AcceptChildren(c)

	if decl.Type != nil {
		c.popScope()
	}
}

func (c *checker) VisitEnum(decl *ast.Enum) {
	decl.AcceptChildren(c)

	// Set case values
	lastValue := int64(-1)

	for _, case_ := range decl.Cases {
		if case_.Value == nil {
			lastValue++
			case_.ActualValue = lastValue
		} else {
			value, err := strconv.ParseInt(case_.Value.String(), 10, 64)

			if err == nil {
				lastValue = value
				case_.ActualValue = value
			} else {
				c.error(case_.Value, "Failed to parse number")

				lastValue++
				case_.ActualValue = lastValue
			}
		}
	}

	// Check type
	if decl.Type != nil {
		if v, ok := ast.As[*ast.Primitive](decl.Type); !ok || !ast.IsInteger(v.Kind) {
			c.error(decl.Type, "Invalid type '%s', can only be a signed or unsigned integer", decl.Type)
		} else {
			// Check if all cases fit inside the type
			min_, max_ := ast.GetRangeTrunc(v.Kind)

			for _, case_ := range decl.Cases {
				if case_.ActualValue < min_ || case_.ActualValue > max_ {
					c.error(case_.Name, "Value '%d' does not fit inside the range of '%s'", case_.Value, decl.Type)
				}
			}
		}
	}
}

func (c *checker) VisitFunc(decl *ast.Func) {
	// Check attributes
	for i, attribute := range decl.Attributes {
		switch attribute := attribute.(type) {
		case ast.ExternAttribute:
			if attribute.Name == "" {
				decl.Attributes[i] = ast.ExternAttribute{Name: decl.Name.String()}
			}

		case ast.IntrinsicAttribute:
			if attribute.Name == "" {
				decl.Attributes[i] = ast.IntrinsicAttribute{Name: decl.Name.String()}
			}

		case ast.InlineAttribute:

		default:
			c.error(decl.Name, "Invalid attribute for a function")
		}
	}

	// Check flags
	_, isImpl := decl.Parent().(*ast.Impl)

	var extern ast.ExternAttribute
	isExtern := decl.GetAttribute(&extern)

	var intrinsic ast.IntrinsicAttribute
	isIntrinsic := decl.GetAttribute(&intrinsic)

	if isImpl && isExtern && !decl.IsStatic() {
		c.error(decl.Name, "Non static methods can't be extern")
	}
	if isImpl && isIntrinsic && !decl.IsStatic() {
		c.error(decl.Name, "Non static methods can't be intrinsics")
	}

	if decl.IsVariadic() && !isExtern {
		c.error(decl.Name, "Only extern functions can be variadic")
	}

	// Intrinsic
	if isIntrinsic {
		c.checkIntrinsic(decl, intrinsic)
		return
	}

	// Push scope
	c.function = decl
	c.pushScope()

	// Params
	for _, param := range decl.Params {
		if c.hasVariableInScope(param.Name) {
			c.error(param.Name, "Parameter with the name '%s' already exists", param.Name)
		} else {
			c.addVariable(param.Name, param.Type).param = true
		}
	}

	// Body
	for _, stmt := range decl.Body {
		c.VisitNode(stmt)
	}

	// Pop scope
	c.popScope()
	c.function = nil

	// Check parameter void type
	for _, param := range decl.Params {
		if ast.IsPrimitive(param.Type, ast.Void) {
			c.error(param.Name, "Parameter cannot be of type 'void'")
		}
	}

	// Check last return
	if decl.HasBody() && !ast.IsPrimitive(decl.Returns, ast.Void) {
		valid := len(decl.Body) > 0

		if valid {
			if _, ok := decl.Body[len(decl.Body)-1].(*ast.Return); !ok {
				valid = false
			}
		}

		if !valid {
			c.error(decl.Name, "Function needs to return a '%s' value", decl.Returns)
		}
	}
}

func (c *checker) checkIntrinsic(decl *ast.Func, intrinsic ast.IntrinsicAttribute) {
	valid := false

	switch intrinsic.Name {
	case "abs":
		valid = isSimpleIntrinsic(decl, 1, signedPredicate|floatingPredicate)

	case "min", "max":
		valid = isSimpleIntrinsic(decl, 2, unsignedPredicate|signedPredicate|floatingPredicate)

	case "pow", "copysign":
		valid = isSimpleIntrinsic(decl, 2, floatingPredicate)

	case "sqrt", "sin", "cos", "exp", "exp2", "exp10", "log", "log2", "log10", "floor", "ceil", "round":
		valid = isSimpleIntrinsic(decl, 1, floatingPredicate)

	case "fma":
		valid = isSimpleIntrinsic(decl, 3, floatingPredicate)

	case "memcpy", "memmove":
		valid = isExactIntrinsic(decl, ast.Void, ast.Void, ast.U32)

	case "memset":
		valid = isExactIntrinsic(decl, ast.Void, ast.U8, ast.U32)
	}

	if !valid {
		c.error(decl.Name, "Unknown intrinsic")
	}
}

type simpleIntrinsicPredicate uint8

const (
	unsignedPredicate simpleIntrinsicPredicate = 1 << iota
	signedPredicate
	floatingPredicate
)

func isSimpleIntrinsic(decl *ast.Func, paramCount int, predicate simpleIntrinsicPredicate) bool {
	if len(decl.Params) != paramCount {
		return false
	}

	for _, param := range decl.Params {
		if !isSimpleIntrinsicType(param.Type, predicate) {
			return false
		}
	}

	if !isSimpleIntrinsicType(decl.Returns, predicate) {
		return false
	}

	for i, param := range decl.Params {
		if i > 0 {
			if !param.Type.Equals(decl.Params[0].Type) {
				return false
			}
		}
	}

	if !decl.Returns.Equals(decl.Params[0].Type) {
		return false
	}

	return true
}

func isSimpleIntrinsicType(type_ ast.Type, predicate simpleIntrinsicPredicate) bool {
	valid := false

	if v, ok := ast.As[*ast.Primitive](type_); ok {
		if predicate&unsignedPredicate != 0 && ast.IsUnsigned(v.Kind) {
			valid = true
		} else if predicate&signedPredicate != 0 && ast.IsSigned(v.Kind) {
			valid = true
		} else if predicate&floatingPredicate != 0 && ast.IsFloating(v.Kind) {
			valid = true
		}
	}

	return valid
}

func isExactIntrinsic(decl *ast.Func, params ...ast.PrimitiveKind) bool {
	if len(decl.Params) != len(params) {
		return false
	}

	for i, param := range decl.Params {
		if params[i] == ast.Void {
			if _, ok := ast.As[*ast.Pointer](param.Type); !ok {
				return false
			}
		} else {
			if !ast.IsPrimitive(param.Type, params[i]) {
				return false
			}
		}
	}

	if !ast.IsPrimitive(decl.Returns, ast.Void) {
		return false
	}

	return true
}
