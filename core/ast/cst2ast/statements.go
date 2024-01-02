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
	var expr ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			expr = c.convertExpr(child)
		}
	}

	return ast.NewExpression(node, expr)
}

func (c *converter) convertBlockStmt(node cst.Node) ast.Stmt {
	var stmts []ast.Stmt

	for _, child := range node.Children {
		if child.Kind.IsStmt() {
			stmts = append(stmts, c.convertStmt(child))
		}
	}

	return ast.NewBlock(node, stmts)
}

func (c *converter) convertVarStmt(node cst.Node) ast.Stmt {
	var name *ast.Token
	var type_ ast.Type
	var value ast.Expr

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind.IsType() {
			type_ = c.convertType(child)
		} else if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		}
	}

	return ast.NewVar(node, name, type_, value)
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

	return ast.NewIf(node, condition, then, else_)
}

func (c *converter) convertForStmt(node cst.Node) ast.Stmt {
	var initializer ast.Stmt
	var condition ast.Expr
	var increment ast.Expr
	var body ast.Stmt

	for _, child := range node.Children {
		if child.Kind.IsStmt() {
			if initializer == nil {
				initializer = c.convertStmt(child)
			} else {
				body = c.convertStmt(child)
			}
		} else if child.Kind.IsExpr() {
			if condition == nil {
				condition = c.convertExpr(child)
			} else {
				increment = c.convertExpr(child)
			}
		}
	}

	return ast.NewFor(node, initializer, condition, increment, body)
}

func (c *converter) convertReturnStmt(node cst.Node) ast.Stmt {
	var value ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		}
	}

	return ast.NewReturn(node, value)
}

func (c *converter) convertBreakStmt(node cst.Node) ast.Stmt {
	return ast.NewBreak(node)
}

func (c *converter) convertContinueStmt(node cst.Node) ast.Stmt {
	return ast.NewContinue(node)
}
