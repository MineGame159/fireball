package cst2ast

import (
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
)

func (c *converter) convertExpr(node cst.Node) ast.Expr {
	switch node.Kind {
	case cst.ParenExprNode:
		return c.convertParenExpr(node)
	case cst.UnaryExprNode:
		return c.convertUnaryExpr(node)
	case cst.BinaryExprNode:
		return c.convertBinaryExpr(node)
	case cst.IndexExprNode:
		return c.convertIndexExpr(node)
	case cst.CallExprNode:
		return c.convertCallExpr(node)
	case cst.TypeCallExprNode:
		return c.convertTypeCallExpr(node)
	case cst.StructExprNode:
		return c.convertStructExpr(node)
	case cst.ArrayExprNode:
		return c.convertArrayExpr(node)
	case cst.NewArrayExprNode:
		return c.convertNewArrayExpr(node)
	case cst.IdentifierNode:
		return c.convertIdentifierExpr(node)
	case cst.BoolExprNode, cst.NumberExprNode, cst.StringExprNode:
		return c.convertLiteral(node)

	default:
		panic("cst2ast.convertExpr() - Not implemented")
	}
}

func (c *converter) convertParenExpr(node cst.Node) ast.Expr {
	p := &ast.Group{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			p.Expr = c.convertExpr(child)
		}
	}

	p.SetRangeToken(node.Token, tokenEnd(node))
	p.SetChildrenParent()

	return p
}

func (c *converter) convertUnaryExpr(node cst.Node) ast.Expr {
	u := &ast.Unary{}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			u.Value = c.convertExpr(child)
		} else {
			u.Op = child.Token

			if u.Value == nil {
				u.Prefix = true
			}
		}
	}

	u.SetRangeToken(node.Token, tokenEnd(node))
	u.SetChildrenParent()

	return u
}

func (c *converter) convertBinaryExpr(node cst.Node) ast.Expr {
	if node.ContainsAny(scanner.AssignmentOperators) {
		return c.convertAssignmentExpr(node)
	}
	if node.Contains(scanner.Dot) {
		return c.convertMemberExpr(node)
	}
	if node.Contains(scanner.As) {
		return c.convertCastExpr(node)
	}

	b := &ast.Binary{}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			if b.Left == nil {
				b.Left = c.convertExpr(child)
			} else {
				b.Right = c.convertExpr(child)
			}
		} else {
			b.Op = child.Token
		}
	}

	b.SetRangeToken(node.Token, tokenEnd(node))
	b.SetChildrenParent()

	return b
}

func (c *converter) convertAssignmentExpr(node cst.Node) ast.Expr {
	a := &ast.Assignment{}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			if a.Assignee == nil {
				a.Assignee = c.convertExpr(child)
			} else {
				a.Value = c.convertExpr(child)
			}
		} else {
			a.Op = child.Token
		}
	}

	a.SetRangeToken(node.Token, tokenEnd(node))
	a.SetChildrenParent()

	return a
}

func (c *converter) convertMemberExpr(node cst.Node) ast.Expr {
	m := &ast.Member{}

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			if m.Value == nil {
				m.Value = c.convertExpr(child)
			} else {
				m.Name = child.Token
			}
		} else if child.Kind.IsExpr() {
			m.Value = c.convertExpr(child)
		}
	}

	m.SetRangeToken(node.Token, tokenEnd(node))
	m.SetChildrenParent()

	return m
}

func (c *converter) convertCastExpr(node cst.Node) ast.Expr {
	b := &ast.Cast{}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			b.Expr = c.convertExpr(child)
		} else if child.Kind.IsType() {
			b.Target = c.convertType(child)
		}
	}

	b.SetRangeToken(node.Token, tokenEnd(node))
	b.SetChildrenParent()

	return b
}

func (c *converter) convertIndexExpr(node cst.Node) ast.Expr {
	i := &ast.Index{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			if i.Value == nil {
				i.Value = c.convertExpr(child)
			} else {
				i.Index = c.convertExpr(child)
			}
		}
	}

	i.SetRangeToken(node.Token, tokenEnd(node))
	i.SetChildrenParent()

	return i
}

func (c *converter) convertCallExpr(node cst.Node) ast.Expr {
	ca := &ast.Call{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			if ca.Callee == nil {
				ca.Callee = c.convertExpr(child)
			} else {
				ca.Args = append(ca.Args, c.convertExpr(child))
			}
		}
	}

	ca.SetRangeToken(node.Token, tokenEnd(node))
	ca.SetChildrenParent()

	return ca
}

func (c *converter) convertTypeCallExpr(node cst.Node) ast.Expr {
	t := &ast.TypeCall{}

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			t.Name = child.Token
		} else if child.Kind.IsType() {
			t.Target = c.convertType(child)
		}
	}

	t.SetRangeToken(node.Token, tokenEnd(node))
	t.SetChildrenParent()

	return t
}

func (c *converter) convertStructExpr(node cst.Node) ast.Expr {
	s := &ast.StructInitializer{Token_: node.Token}

	for _, child := range node.Children {
		if child.Token.Lexeme == "new" {
			s.New = true
		} else if child.Kind.IsType() {
			s.Target = c.convertType(child)
		} else if child.Kind == cst.StructFieldExprNode {
			s.Fields = append(s.Fields, c.convertStructFieldExpr(child))
		}
	}

	s.SetRangeToken(node.Token, tokenEnd(node))
	s.SetChildrenParent()

	return s
}

func (c *converter) convertArrayExpr(node cst.Node) ast.Expr {
	a := &ast.ArrayInitializer{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			a.Values = append(a.Values, c.convertExpr(child))
		}
	}

	a.SetRangeToken(node.Token, tokenEnd(node))
	a.SetChildrenParent()

	return a
}

func (c *converter) convertNewArrayExpr(node cst.Node) ast.Expr {
	n := &ast.NewArray{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsType() {
			n.Type_ = c.convertType(child)
		} else if child.Kind.IsExpr() {
			n.Count = c.convertExpr(child)
		}
	}

	n.SetRangeToken(node.Token, tokenEnd(node))
	n.SetChildrenParent()

	return n
}

func (c *converter) convertStructFieldExpr(node cst.Node) ast.InitField {
	i := ast.InitField{}

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			if i.Name.Lexeme == "" {
				i.Name = child.Token
			} else {
				i.Value = c.convertExpr(child)
			}
		} else if child.Kind.IsExpr() {
			i.Value = c.convertExpr(child)
		}
	}

	return i
}

func (c *converter) convertIdentifierExpr(node cst.Node) ast.Expr {
	i := &ast.Identifier{Identifier: node.Token}

	i.SetRangeToken(node.Token, node.Token)
	i.SetChildrenParent()

	return i
}

func (c *converter) convertLiteral(node cst.Node) ast.Expr {
	l := &ast.Literal{Value: node.Token}

	l.SetRangeToken(node.Token, node.Token)
	l.SetChildrenParent()

	return l
}
