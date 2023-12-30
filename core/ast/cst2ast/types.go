package cst2ast

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/types"
	"strconv"
)

func (c *converter) convertType(node cst.Node) types.Type {
	switch node.Kind {
	case cst.IdentifierTypeNode:
		return c.convertIdentifierType(node)
	case cst.PointerTypeNode:
		return c.convertPointerType(node)
	case cst.ArrayTypeNode:
		return c.convertArrayType(node)
	case cst.FuncTypeNode:
		return c.convertFuncType(node)

	default:
		panic("cst2ast.convertType() - Not implemented")
	}
}

func (c *converter) convertIdentifierType(node cst.Node) types.Type {
	token := node.Children[0].Token
	range_ := core.TokenToRange(token)

	var kind types.PrimitiveKind

	switch token.Lexeme {
	case "void":
		kind = types.Void
	case "bool":
		kind = types.Bool

	case "u8":
		kind = types.U8
	case "u16":
		kind = types.U16
	case "u32":
		kind = types.U32
	case "u64":
		kind = types.U64

	case "i8":
		kind = types.I8
	case "i16":
		kind = types.I16
	case "i32":
		kind = types.I32
	case "i64":
		kind = types.I64

	case "f32":
		kind = types.F32
	case "f64":
		kind = types.F64

	default:
		return types.Unresolved(token, range_)
	}

	return types.Primitive(kind, range_)
}

func (c *converter) convertPointerType(node cst.Node) types.Type {
	range_ := core.TokenToRange(node.Token)
	return types.Pointer(c.convertType(node.Children[1]), range_)
}

func (c *converter) convertArrayType(node cst.Node) types.Type {
	range_ := core.TokenToRange(node.Token)
	value, _ := strconv.ParseUint(node.Children[1].Token.Lexeme, 10, 32)

	return types.Array(uint32(value), c.convertType(node.Children[3]), range_)
}

func (c *converter) convertFuncType(node cst.Node) types.Type {
	f := &ast.Func{}
	reported := false

	for i, child := range node.Children {
		if child.Kind == cst.FuncTypeParamNode {
			param, varArgs := c.convertFuncParam(child)

			if varArgs {
				f.Flags |= ast.Variadic
			} else {
				if f.IsVariadic() && !reported {
					c.error(node.Children[i-2].Children[0].Token, "Variadic arguments can only appear at the end of the parameter list")
					reported = true
				}

				f.Params = append(f.Params, param)
			}
		} else if child.Kind.IsType() {
			f.Returns = c.convertType(child)
		}
	}

	return f
}
