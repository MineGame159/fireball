package cst2ast

import (
	"fireball/core/ast"
	"fireball/core/cst"
	"strconv"
)

func (c *converter) convertType(node cst.Node) ast.Type {
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

func (c *converter) convertIdentifierType(node cst.Node) ast.Type {
	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			var kind ast.PrimitiveKind

			switch child.Token.Lexeme {
			case "void":
				kind = ast.Void
			case "bool":
				kind = ast.Bool

			case "u8":
				kind = ast.U8
			case "u16":
				kind = ast.U16
			case "u32":
				kind = ast.U32
			case "u64":
				kind = ast.U64

			case "i8":
				kind = ast.I8
			case "i16":
				kind = ast.I16
			case "i32":
				kind = ast.I32
			case "i64":
				kind = ast.I64

			case "f32":
				kind = ast.F32
			case "f64":
				kind = ast.F64

			default:
				if r := ast.NewResolvable(node, child.Token); r != nil {
					return r
				}

				return nil
			}

			if p := ast.NewPrimitive(node, kind, child.Token); p != nil {
				return p
			}

			return nil
		}
	}

	panic("cst2ast.convertIdentifierType() - Not implemented")
}

func (c *converter) convertPointerType(node cst.Node) ast.Type {
	var pointee ast.Type

	for _, child := range node.Children {
		if child.Kind.IsType() {
			pointee = c.convertType(child)
		}
	}

	if p := ast.NewPointer(node, pointee); p != nil {
		return p
	}

	return nil
}

func (c *converter) convertArrayType(node cst.Node) ast.Type {
	var base ast.Type
	count := uint32(0)

	for _, child := range node.Children {
		if child.Kind.IsType() {
			base = c.convertType(child)
		} else if child.Kind == cst.NumberExprNode {
			c, _ := strconv.ParseUint(child.Token.Lexeme, 10, 32)
			count = uint32(c)
		}
	}

	if a := ast.NewArray(node, base, count); a != nil {
		return a
	}

	return nil
}

func (c *converter) convertFuncType(node cst.Node) ast.Type {
	var flags ast.FuncFlags
	var params []*ast.Param
	var returns ast.Type

	reported := false

	for i, child := range node.Children {
		if child.Kind == cst.FuncTypeParamNode {
			param, varArgs := c.convertFuncParam(child)

			if varArgs {
				flags |= ast.Variadic
			} else if param != nil {
				if flags&ast.Variadic != 0 && !reported {
					c.error(node.Children[i-2].Children[0], "Variadic arguments can only appear at the end of the parameter list")
					reported = true
				}

				params = append(params, param)
			}
		} else if child.Kind.IsType() {
			returns = c.convertType(child)
		}
	}

	if f := ast.NewFunc(node, nil, flags, nil, params, returns, nil); f != nil {
		return f
	}

	return nil
}
