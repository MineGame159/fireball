package checker

import (
	"fireball/core/ast"
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
		if field.Name != nil && !fields.Add(field.Name.String()) {
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
		if field.Name != nil && !fields.Add(field.Name.String()) {
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
		c.addVariable(&ast.Token{Token_: scanner.Token{Kind: scanner.Identifier, Lexeme: "this"}}, decl.Type, nil)
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
			c.error(decl.Type, "Invalid type '%s', can only be a signed or unsigned integer", ast.PrintType(decl.Type))
		} else {
			// Check if all cases fit inside the type
			min_, max_ := ast.GetRangeTrunc(v.Kind)

			for _, case_ := range decl.Cases {
				if case_.ActualValue < min_ || case_.ActualValue > max_ {
					c.error(case_.Name, "Value '%d' does not fit inside the range of '%s'", case_.Value, ast.PrintType(decl.Type))
				}
			}
		}
	}
}

func (c *checker) VisitFunc(decl *ast.Func) {
	// Check attributes
	for _, attribute := range decl.Attributes {
		c.visitAttribute(decl, attribute)
	}

	if decl.Name == nil {
		return
	}

	// Check flags
	_, isImpl := decl.Parent().(*ast.Impl)

	isExtern := false
	isIntrinsic := false

	for _, attribute := range decl.Attributes {
		if attribute.Name.String() == "Extern" {
			isExtern = true
		} else if attribute.Name.String() == "Intrinsic" {
			isIntrinsic = true
		}
	}

	if isImpl && isExtern && !decl.IsStatic() {
		c.error(decl.Name, "Non static methods can't be extern")
	}
	if isImpl && isIntrinsic && !decl.IsStatic() {
		c.error(decl.Name, "Non static methods can't be intrinsics")
	}

	if decl.IsVariadic() && !isExtern {
		c.error(decl.Name, "Only extern functions can be variadic")
	}

	// Push scope
	c.function = decl
	c.pushScope()

	// Params
	for _, param := range decl.Params {
		if c.hasVariableInScope(param.Name) {
			c.error(param.Name, "Parameter with the name '%s' already exists", param.Name)
		} else {
			if v := c.addVariable(param.Name, param.Type, param); v != nil {
				v.param = true
			}
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
			c.error(decl.Name, "Function needs to return a '%s' value", ast.PrintType(decl.Returns))
		}
	}
}
