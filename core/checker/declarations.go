package checker

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fireball/core/utils"
)

func (c *checker) VisitStruct(decl *ast.Struct) {
	decl.AcceptChildren(c)

	// Check fields
	fields := utils.NewSet[string]()

	for _, field := range decl.Fields {
		// Check name collision
		if !fields.Add(field.Name.Lexeme) {
			c.errorToken(field.Name, "Field with the name '%s' already exists.", field.Name)
		}

		// Check void type
		if types.IsPrimitive(field.Type, types.Void) {
			c.errorToken(field.Name, "Field cannot be of type 'void'.")
		}
	}
}

func (c *checker) VisitImpl(decl *ast.Impl) {
	if decl.Type_ != nil {
		c.pushScope()
		c.addVariable(scanner.Token{Kind: scanner.Identifier, Lexeme: "this"}, decl.Type_)
	}

	decl.AcceptChildren(c)

	if decl.Type_ != nil {
		c.popScope()
	}
}

func (c *checker) VisitEnum(decl *ast.Enum) {
	decl.AcceptChildren(c)

	// Check type
	if decl.Type != nil {
		if v, ok := decl.Type.(*types.PrimitiveType); !ok || !types.IsInteger(v.Kind) {
			c.errorRange(decl.Type.Range(), "Invalid type '%s', can only be a signed or unsigned integer.", decl.Type)
		} else {
			// Check if all cases fit inside the type
			min_, max_ := types.GetRangeTrunc(v.Kind)

			for _, case_ := range decl.Cases {
				if int64(case_.Value) < min_ || int64(case_.Value) > max_ {
					c.errorToken(case_.Name, "Value '%d' does not fit inside the range of '%s'.", case_.Value, decl.Type)
				}
			}
		}
	}
}

func (c *checker) VisitFunc(decl *ast.Func) {
	// Need to resolve return type sooner
	decl.AcceptTypesPtr(c)

	// Check flags
	_, isImpl := decl.Parent().(*ast.Impl)

	if isImpl && decl.IsExtern() && !decl.IsStatic() {
		c.errorToken(decl.Name, "Non static methods can't be extern.")
	}
	if isImpl && decl.IsIntrinsic() && !decl.IsStatic() {
		c.errorToken(decl.Name, "Non static methods can't be intrinsics.")
	}

	if decl.IsVariadic() && !decl.IsExtern() {
		c.errorToken(decl.Name, "Only extern functions can be variadic.")
	}

	// Intrinsic
	if decl.IsIntrinsic() {
		c.checkIntrinsic(decl)
		return
	}

	// Push scope
	c.function = decl
	c.pushScope()

	// Params
	for _, param := range decl.Params {
		if c.hasVariableInScope(param.Name) {
			c.errorToken(param.Name, "Parameter with the name '%s' already exists.", param.Name)
		} else {
			c.addVariable(param.Name, param.Type).param = true
		}
	}

	// Body
	for _, stmt := range decl.Body {
		c.AcceptStmt(stmt)
	}

	// Pop scope
	c.popScope()
	c.function = nil

	// Check parameter void type
	for _, param := range decl.Params {
		if types.IsPrimitive(param.Type, types.Void) {
			c.errorToken(param.Name, "Parameter cannot be of type 'void'.")
		}
	}

	// Check last return
	if decl.HasBody() && !types.IsPrimitive(decl.Returns, types.Void) {
		valid := len(decl.Body) > 0

		if valid {
			if _, ok := decl.Body[len(decl.Body)-1].(*ast.Return); !ok {
				valid = false
			}
		}

		if !valid {
			c.errorToken(decl.Name, "Function needs to return a '%s' value.", decl.Returns)
		}
	}
}

func (c *checker) checkIntrinsic(decl *ast.Func) {
	valid := false

	switch decl.Name.Lexeme {
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
		valid = isExactIntrinsic(decl, types.Void, types.Void, types.U32)

	case "memset":
		valid = isExactIntrinsic(decl, types.Void, types.U8, types.U32)
	}

	if !valid {
		c.errorToken(decl.Name, "Unknown intrinsic.")
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

func isSimpleIntrinsicType(type_ types.Type, predicate simpleIntrinsicPredicate) bool {
	valid := false

	if v, ok := type_.(*types.PrimitiveType); ok {
		if predicate&unsignedPredicate != 0 && types.IsUnsigned(v.Kind) {
			valid = true
		} else if predicate&signedPredicate != 0 && types.IsSigned(v.Kind) {
			valid = true
		} else if predicate&floatingPredicate != 0 && types.IsFloating(v.Kind) {
			valid = true
		}
	}

	return valid
}

func isExactIntrinsic(decl *ast.Func, params ...types.PrimitiveKind) bool {
	if len(decl.Params) != len(params) {
		return false
	}

	for i, param := range decl.Params {
		if params[i] == types.Void {
			if _, ok := param.Type.(*types.PointerType); !ok {
				return false
			}
		} else {
			if !types.IsPrimitive(param.Type, params[i]) {
				return false
			}
		}
	}

	if !types.IsPrimitive(decl.Returns, types.Void) {
		return false
	}

	return true
}
