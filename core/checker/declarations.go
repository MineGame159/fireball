package checker

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/utils"
	"math"
	"strconv"
)

func (c *checker) VisitNamespace(_ *ast.Namespace) {}

func (c *checker) VisitUsing(_ *ast.Using) {}

func (c *checker) VisitStruct(decl *ast.Struct) {
	decl.AcceptChildren(c)

	c.checkNameCollision(decl, decl.Name)

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

	c.checkNameCollision(decl, decl.Name)

	// Set case values
	lastValue := int64(-1)

	for _, case_ := range decl.Cases {
		if case_.Value == nil {
			lastValue++
			case_.ActualValue = lastValue
		} else {
			var value int64
			var err error

			switch case_.Value.Token().Kind {
			case scanner.Number:
				value, err = strconv.ParseInt(case_.Value.String(), 10, 64)
			case scanner.Hex:
				value, err = strconv.ParseInt(case_.Value.String()[2:], 16, 64)
			case scanner.Binary:
				value, err = strconv.ParseInt(case_.Value.String()[2:], 2, 64)

			default:
				panic("checker.VisitEnum() - Not implemented")
			}

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

	// Find type
	if decl.Type == nil {
		minValue := int64(math.MaxInt64)
		maxValue := int64(math.MinInt64)

		for _, case_ := range decl.Cases {
			minValue = min(minValue, case_.ActualValue)
			maxValue = max(maxValue, case_.ActualValue)
		}

		var kind ast.PrimitiveKind

		if minValue >= 0 {
			// Unsigned
			if maxValue <= math.MaxUint8 {
				kind = ast.U8
			} else if maxValue <= math.MaxUint16 {
				kind = ast.U16
			} else if maxValue <= math.MaxUint32 {
				kind = ast.U32
			} else {
				kind = ast.U64
			}
		} else {
			// Signed
			if minValue >= math.MinInt8 && maxValue <= math.MaxInt8 {
				kind = ast.I8
			} else if minValue >= math.MinInt16 && maxValue <= math.MaxInt16 {
				kind = ast.I16
			} else if minValue >= math.MinInt32 && maxValue <= math.MaxInt32 {
				kind = ast.I32
			} else {
				kind = ast.I64
			}
		}

		decl.ActualType = &ast.Primitive{Kind: kind}
	} else {
		decl.ActualType = decl.Type
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
	// Check name collision
	if decl.Name != nil {
		if s := decl.Receiver(); s != nil {
			if method := c.resolver.GetMethod(s, decl.Name.String(), decl.IsStatic()); method != nil && method != decl {
				c.error(decl.Name, "Method with this name already exists")
			}
		} else {
			c.checkNameCollision(decl, decl.Name)
		}
	}

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
