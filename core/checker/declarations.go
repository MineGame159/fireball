package checker

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/utils"
)

func (c *checker) VisitNamespace(_ *ast.Namespace) {}

func (c *checker) VisitUsing(_ *ast.Using) {}

func (c *checker) VisitStruct(decl *ast.Struct) {
	decl.AcceptChildren(c)

	c.checkNameCollision(decl, decl.Name)

	// Check attributes
	for _, attribute := range decl.Attributes {
		c.visitStructAttribute(attribute)
	}

	prevResolver := c.resolver
	if len(decl.GenericParams) != 0 {
		c.resolver = ast.NewGenericResolver(c.resolver, decl.GenericParams)
	}

	// Check static fields
	fields := utils.NewSet[string]()

	for _, field := range decl.StaticFields {
		// Check name collision
		if field.Name() != nil && !fields.Add(field.Name().String()) {
			c.error(field.Name(), "Static field with the name '%s' already exists", field.Name())
		}

		// Check void type
		if ast.IsPrimitive(field.Type(), ast.Void) {
			c.error(field.Name(), "Static field cannot be of type 'void'")
		}
	}

	// Check fields
	fields = utils.NewSet[string]()

	for _, field := range decl.Fields {
		// Check name collision
		if field.Name() != nil && !fields.Add(field.Name().String()) {
			c.error(field.Name(), "Field with the name '%s' already exists", field.Name)
		}

		// Check void type
		if ast.IsPrimitive(field.Type(), ast.Void) {
			c.error(field.Name(), "Field cannot be of type 'void'")
		}
	}

	c.resolver = prevResolver
}

func (c *checker) VisitImpl(decl *ast.Impl) {
	// Implements
	if decl.Implements != nil {
		// Check interface
		if inter, ok := ast.As[*ast.Interface](decl.Implements); ok {
			// Check methods
			count := 0

			for _, method := range decl.Methods {
				if method.Name == nil {
					continue
				}

				interMethod, _ := inter.GetMethod(method.Name.String())

				if interMethod != nil && interMethod.NameAndSignatureEquals(method) {
					count++
				} else {
					c.error(method, "Interface '%s' does not contain method '%s'", ast.PrintType(inter), ast.PrintTypeOptions(method, ast.TypePrintOptions{FuncNames: true}))
				}
			}

			if count != len(inter.Methods) {
				c.error(decl.Struct, "Missing method from interface '%s'", ast.PrintType(decl.Implements))
			}
		} else {
			c.error(decl.Implements, "'%s' is not an interface", ast.PrintType(decl.Implements))
		}
	}

	// Children
	prevResolver := c.resolver

	if decl.Type != nil {
		c.pushScope()
		c.addVariable(&ast.Token{Token_: scanner.Token{Kind: scanner.Identifier, Lexeme: "this"}}, decl.Type, nil)

		s, _ := ast.As[*ast.Struct](decl.Type)
		if len(s.GenericParams) > 0 {
			c.resolver = ast.NewGenericResolver(c.resolver, s.GenericParams)
		}
	}

	decl.AcceptChildren(c)

	if decl.Type != nil {
		c.popScope()
	}

	c.resolver = prevResolver
}

func (c *checker) VisitEnum(decl *ast.Enum) {
	decl.AcceptChildren(c)

	c.checkNameCollision(decl, decl.Name)

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

func (c *checker) VisitInterface(decl *ast.Interface) {
	decl.AcceptChildren(c)

	c.checkNameCollision(decl, decl.Name)

	// Check method bodies
	for _, method := range decl.Methods {
		if method.Cst().Contains(scanner.LeftBrace) {
			errorSlice(c, method.Body, "Interface methods can't have bodies")
		}
	}
}

func (c *checker) VisitFunc(decl *ast.Func) {
	// Check name collision
	if decl.Name != nil {
		if s := decl.Struct(); s != nil {
			if method := c.resolver.GetMethod(s, decl.Name.String(), decl.IsStatic()); method != nil && method != decl {
				c.error(decl.Name, "Method with this name already exists")
			}
		} else {
			c.checkNameCollision(decl, decl.Name)
		}
	}

	// Check attributes
	for _, attribute := range decl.Attributes {
		c.visitFuncAttribute(decl, attribute)
	}

	if decl.Name == nil {
		return
	}

	// Check flags
	_, isImpl := decl.Parent().(*ast.Impl)

	isExtern := false
	isIntrinsic := false
	isTest := false

	for _, attribute := range decl.Attributes {
		switch attribute.Name.String() {
		case "Extern":
			isExtern = true
		case "Intrinsic":
			isIntrinsic = true
		case "Test":
			isTest = true
		}
	}

	if isImpl && isExtern && !decl.IsStatic() {
		c.error(decl.Name, "Non static methods can't be extern")
	}
	if isImpl && isIntrinsic && !decl.IsStatic() {
		c.error(decl.Name, "Non static methods can't be intrinsics")
	}
	if isImpl && isTest && !decl.IsStatic() {
		c.error(decl.Name, "Non static methods can't be a test")
	}

	if decl.IsVariadic() && !isExtern {
		c.error(decl.Name, "Only extern functions can be variadic")
	}

	if (isExtern && isIntrinsic) || (isExtern && isTest) || (isIntrinsic && isTest) {
		c.error(decl.Name, "Invalid combination of attributes")
	}

	// Check generics
	if len(decl.GenericParams) != 0 && isExtern {
		c.error(decl.Name, "Extern functions can't be generic")
	}

	if len(decl.GenericParams) != 0 && isIntrinsic {
		c.error(decl.Name, "Intrinsic functions can't be generic")
	}

	// Check body
	if decl.HasBody() {
		if !decl.Cst().Contains(scanner.LeftBrace) {
			c.error(decl, "Function need to have a body")
		}
	} else {
		if decl.Cst().Contains(scanner.LeftBrace) {
			c.error(decl, "Function can't have a body")
		}
	}

	// Push scope
	c.function = decl
	c.pushScope()

	prevResolver := c.resolver
	if len(decl.GenericParams) != 0 {
		c.resolver = ast.NewGenericResolver(c.resolver, decl.GenericParams)
	}

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
	c.resolver = prevResolver

	c.popScope()
	c.function = nil

	// Check last return
	if decl.HasBody() && !ast.IsPrimitive(decl.Returns(), ast.Void) {
		valid := len(decl.Body) > 0

		if valid {
			if _, ok := decl.Body[len(decl.Body)-1].(*ast.Return); !ok {
				valid = false
			}
		}

		if !valid {
			c.error(decl.Name, "Function needs to return a '%s' value", ast.PrintType(decl.Returns()))
		}
	}
}

func (c *checker) VisitGlobalVar(decl *ast.GlobalVar) {
	decl.AcceptChildren(c)

	c.checkNameCollision(decl, decl.Name)

	// Check void type
	if ast.IsPrimitive(decl.Type, ast.Void) {
		c.error(decl.Name, "Variable cannot be of type 'void'")
	}
}

// Utils

func (c *checker) checkNameCollision(decl ast.Decl, name *ast.Token) {
	if name == nil {
		return
	}

	// Children
	for _, child := range c.resolver.GetChildren() {
		if child == name.String() {
			c.error(name, "Declaration with this name already exists")
			return
		}
	}

	// Symbols
	checker := nameChecker{
		c:      c,
		except: decl,
		name:   name,
	}

	c.resolver.GetSymbols(&checker)
}

type nameChecker struct {
	c *checker

	except ast.Node
	name   *ast.Token

	found bool
}

func (s *nameChecker) VisitSymbol(node ast.Node) {
	if node != s.except && !s.found {
		var name2 *ast.Token

		switch node := node.(type) {
		case *ast.Struct:
			name2 = node.Name
		case *ast.Enum:
			name2 = node.Name
		case *ast.Func:
			name2 = node.Name
		case *ast.GlobalVar:
			name2 = node.Name
		}

		if name2 != nil && s.name.String() == name2.String() {
			s.c.error(s.name, "Declaration with this name already exists")
			s.found = true
		}
	}
}
