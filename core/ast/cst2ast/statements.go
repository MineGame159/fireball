package cst2ast

import (
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
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
	var expr ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			expr = c.convertExpr(child)
		}
	}

	if e := ast.NewExpression(node, expr); e != nil {
		return e
	}

	return nil
}

func (c *converter) convertBlockStmt(node cst.Node) ast.Stmt {
	var stmts []ast.Stmt

	for _, child := range node.Children {
		if child.Kind.IsStmt() {
			stmt := c.convertStmt(child)

			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		}
	}

	if b := ast.NewBlock(node, stmts); b != nil {
		return b
	}

	return nil
}

func (c *converter) convertVarStmt(node cst.Node) ast.Stmt {
	var name *ast.Token
	var type_ ast.Type
	var value ast.Expr

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			if name == nil {
				name = c.convertToken(child)
			} else {
				value = c.convertExpr(child)
			}
		} else if child.Kind.IsType() {
			type_ = c.convertType(child)
		} else if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		}
	}

	if v := ast.NewVar(node, name, type_, value); v != nil {
		return v
	}

	return nil
}

func (c *converter) convertIfStmt(node cst.Node) ast.Stmt {
	var condition ast.Expr
	var then ast.Stmt
	var else_ ast.Stmt

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			condition = c.convertExpr(child)
		} else if child.Kind.IsStmt() {
			if then == nil {
				then = c.convertStmt(child)
			} else {
				else_ = c.convertStmt(child)
			}
		}
	}

	if i := ast.NewIf(node, condition, then, else_); i != nil {
		return i
	}

	return nil
}

func (c *converter) convertForStmt(node cst.Node) ast.Stmt {
	var initializer ast.Stmt
	var condition ast.Expr
	var increment ast.Expr
	var body ast.Stmt

	semicolons := 0

	for _, child := range node.Children {
		if child.Kind.IsStmt() {
			if semicolons == 0 {
				initializer = c.convertStmt(child)
				semicolons++
			} else {
				body = c.convertStmt(child)
			}
		} else if child.Kind.IsExpr() {
			if semicolons < 2 {
				condition = c.convertExpr(child)
			} else {
				increment = c.convertExpr(child)
			}
		} else if child.Token.Kind == scanner.Semicolon {
			semicolons++
		}
	}

	if f := ast.NewFor(node, initializer, condition, increment, body); f != nil {
		return f
	}

	return nil
}

func (c *converter) convertReturnStmt(node cst.Node) ast.Stmt {
	var value ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		}
	}

	if r := ast.NewReturn(node, value); r != nil {
		return r
	}

	return nil
}

func (c *converter) convertBreakStmt(node cst.Node) ast.Stmt {
	if b := ast.NewBreak(node); b != nil {
		return b
	}

	return nil
}

func (c *converter) convertContinueStmt(node cst.Node) ast.Stmt {
	if c := ast.NewContinue(node); c != nil {
		return c
	}

	return nil
}
