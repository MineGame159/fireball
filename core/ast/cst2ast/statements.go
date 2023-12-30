package cst2ast

import (
	"fireball/core/ast"
	"fireball/core/cst"
)

func (c *converter) convertStmt(node cst.Node) ast.Stmt {
	switch node.Kind {
	case cst.ExprStmtNode:
		return c.convertExprStmt(node)
	case cst.BlockStmtNode:
		return c.convertBlockStmt(node)
	case cst.VarStmtNode:
		return c.convertVarStmt(node)
	case cst.IfStmtNode:
		return c.convertIfStmt(node)
	case cst.ForStmtNode:
		return c.convertForStmt(node)
	case cst.ReturnStmtNode:
		return c.convertReturnStmt(node)
	case cst.BreakStmtNode:
		return c.convertBreakStmt(node)
	case cst.ContinueStmtNode:
		return c.convertContinueStmt(node)

	default:
		panic("cst2ast.convertStmt() - Not implemented")
	}
}

func (c *converter) convertExprStmt(node cst.Node) ast.Stmt {
	e := &ast.Expression{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			e.Expr = c.convertExpr(child)
		}
	}

	e.SetRangeToken(node.Token, tokenEnd(node))
	e.SetChildrenParent()

	return e
}

func (c *converter) convertBlockStmt(node cst.Node) ast.Stmt {
	b := &ast.Block{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsStmt() {
			b.Stmts = append(b.Stmts, c.convertStmt(child))
		}
	}

	b.SetRangeToken(node.Token, tokenEnd(node))
	b.SetChildrenParent()

	return b
}

func (c *converter) convertVarStmt(node cst.Node) ast.Stmt {
	v := &ast.Variable{InferType: true}

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			v.Name = child.Token
		} else if child.Kind.IsType() {
			v.Type = c.convertType(child)
			v.InferType = false
		} else if child.Kind.IsExpr() {
			v.Initializer = c.convertExpr(child)
		}
	}

	v.SetRangeToken(node.Token, tokenEnd(node))
	v.SetChildrenParent()

	return v
}

func (c *converter) convertIfStmt(node cst.Node) ast.Stmt {
	i := &ast.If{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			i.Condition = c.convertExpr(child)
		} else if child.Kind.IsStmt() {
			if i.Then == nil {
				i.Then = c.convertStmt(child)
			} else {
				i.Else = c.convertStmt(child)
			}
		}
	}

	i.SetRangeToken(node.Token, tokenEnd(node))
	i.SetChildrenParent()

	return i
}

func (c *converter) convertForStmt(node cst.Node) ast.Stmt {
	f := &ast.For{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsStmt() {
			if f.Initializer == nil {
				f.Initializer = c.convertStmt(child)
			} else {
				f.Body = c.convertStmt(child)
			}
		} else if child.Kind.IsExpr() {
			if f.Condition == nil {
				f.Condition = c.convertExpr(child)
			} else {
				f.Increment = c.convertExpr(child)
			}
		}
	}

	f.SetRangeToken(node.Token, tokenEnd(node))
	f.SetChildrenParent()

	return f
}

func (c *converter) convertReturnStmt(node cst.Node) ast.Stmt {
	r := &ast.Return{Token_: node.Token}

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			r.Expr = c.convertExpr(child)
		}
	}

	r.SetRangeToken(node.Token, tokenEnd(node))
	r.SetChildrenParent()

	return r
}

func (c *converter) convertBreakStmt(node cst.Node) ast.Stmt {
	b := &ast.Break{Token_: node.Token}

	b.SetRangeToken(node.Token, tokenEnd(node))
	b.SetChildrenParent()

	return b
}

func (c *converter) convertContinueStmt(node cst.Node) ast.Stmt {
	co := &ast.Continue{Token_: node.Token}

	co.SetRangeToken(node.Token, tokenEnd(node))
	co.SetChildrenParent()

	return co
}
