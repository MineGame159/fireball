package cst2ast

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
	"fireball/core/types"
	"strconv"
)

func (c *converter) convertDecl(node cst.Node) ast.Decl {
	switch node.Kind {
	case cst.StructNode:
		return c.convertStructDecl(node)
	case cst.ImplNode:
		return c.convertImplDecl(node)
	case cst.EnumNode:
		return c.convertEnumDecl(node)
	case cst.FuncNode:
		return c.convertFuncDecl(node)

	default:
		panic("cst2ast.convertDecl() - Not implemented")
	}
}

// Struct

func (c *converter) convertStructDecl(node cst.Node) ast.Decl {
	s := &ast.Struct{}

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			s.Name = child.Token
		} else if child.Kind == cst.StructFieldNode {
			field, static := c.convertStructField(s, child)

			if static {
				s.StaticFields = append(s.StaticFields, field)
			} else {
				s.Fields = append(s.Fields, field)
			}
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0].Token, "Structs cannot have attributes")
		}
	}

	s.SetRangeToken(node.Token, tokenEnd(node))
	s.SetChildrenParent()

	return s
}

func (c *converter) convertStructField(s *ast.Struct, node cst.Node) (ast.Field, bool) {
	f := ast.Field{Parent: s}
	static := false

	for _, child := range node.Children {
		if child.Token.Kind == scanner.Static {
			static = true
		} else if child.Kind == cst.IdentifierNode {
			f.Name = child.Token
		} else if child.Kind.IsType() {
			f.Type = c.convertType(child)
		}
	}

	return f, static
}

// Impl

func (c *converter) convertImplDecl(node cst.Node) ast.Decl {
	i := &ast.Impl{}

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			i.Struct = child.Token
		} else if child.Kind == cst.FuncNode {
			i.Functions = append(i.Functions, c.convertFuncDecl(child))
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0].Token, "Implementations cannot have attributes")
		}
	}

	i.SetRangeToken(node.Token, tokenEnd(node))
	i.SetChildrenParent()

	return i
}

// Enum

func (c *converter) convertEnumDecl(node cst.Node) ast.Decl {
	e := &ast.Enum{InferType: true}

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			e.Name = child.Token
		} else if child.Kind.IsType() {
			e.Type = c.convertType(child)
			e.InferType = false
		} else if child.Kind == cst.EnumCaseNode {
			e.Cases = append(e.Cases, c.convertEnumCase(child))
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0].Token, "Enums cannot have attributes")
		}
	}

	e.SetRangeToken(node.Token, tokenEnd(node))
	e.SetChildrenParent()

	return e
}

func (c *converter) convertEnumCase(node cst.Node) ast.EnumCase {
	e := ast.EnumCase{InferValue: true}

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			e.Name = child.Token
		} else if child.Kind == cst.NumberExprNode {
			value, _ := strconv.ParseInt(child.Token.Lexeme, 10, 32)

			e.Value = int(value)
			e.InferValue = false
		}
	}

	return e
}

// Func

func (c *converter) convertFuncDecl(node cst.Node) ast.Decl {
	f := &ast.Func{}
	reported := false

	for i, child := range node.Children {
		if child.Token.Kind == scanner.Static {
			f.Flags |= ast.Static
		} else if child.Kind == cst.IdentifierNode {
			f.Name = child.Token
		} else if child.Kind == cst.FuncParamNode {
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
		} else if child.Kind.IsStmt() {
			f.Body = append(f.Body, c.convertStmt(child))
		} else if child.Kind == cst.AttributesNode {
			f.Attributes = c.convertAttributes(child)
		}
	}

	var extern types.ExternAttribute
	var intrinsic types.IntrinsicAttribute

	if !node.Contains(scanner.LeftBrace) && !f.GetAttribute(&extern) && !f.GetAttribute(&intrinsic) {
		c.error(f.Name, "Functions without the extern or intrinsic attribute need to have a body")
	}

	if f.Returns == nil {
		f.Returns = types.Primitive(types.Void, core.Range{})
	}

	f.SetRangeToken(node.Token, tokenEnd(node))
	f.SetChildrenParent()

	return f
}

func (c *converter) convertFuncParam(node cst.Node) (ast.Param, bool) {
	p := ast.Param{}
	varArgs := false

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			p.Name = child.Token
		} else if child.Kind.IsType() {
			p.Type = c.convertType(child)
		} else if child.Token.Kind == scanner.DotDotDot {
			varArgs = true
		}
	}

	return p, varArgs
}
