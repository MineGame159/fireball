package checker

import "fireball/core/ast"

func (c *checker) visitAttribute(decl *ast.Func, attribute *ast.Attribute) {
	if attribute.Name == nil {
		return
	}

	switch attribute.Name.String() {
	case "Extern":
		if len(attribute.Args) > 1 {
			c.error(attribute.Name, "Extern attribute can only have one argument")
		}

	case "Intrinsic":
		if len(attribute.Args) > 1 {
			c.error(attribute.Name, "Intrinsic attribute can only have one argument")
		}

		c.visitIntrinsic(decl, attribute)

	case "Inline":
		if len(attribute.Args) > 0 {
			c.error(attribute.Name, "Inline doesn't have any arguments")
		}

	default:
		c.error(attribute.Name, "Attribute with this name doesn't exist")
	}
}

func (c *checker) visitIntrinsic(decl *ast.Func, attribute *ast.Attribute) {
	valid := false

	var token *ast.Token
	name := ""

	if len(attribute.Args) > 0 {
		arg := attribute.Args[0]
		if arg == nil {
			return
		}

		token = arg
		name = arg.String()[1 : len(arg.String())-1]
	} else {
		token = decl.Name
		name = decl.Name.String()
	}

	switch name {
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
		c.error(token, "Unknown intrinsic")
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
			if param.Type == nil || !param.Type.Equals(decl.Params[0].Type) {
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
