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
	case cst.AllocateArrayExprNode:
		return c.convertAllocateArrayExpr(node)
	case cst.IdentifierNode:
		return c.convertIdentifierExpr(node)
	case cst.BoolExprNode, cst.NumberExprNode, cst.StringExprNode:
		return c.convertLiteral(node)

	default:
		panic("cst2ast.convertExpr() - Not implemented")
	}
}

func (c *converter) convertParenExpr(node cst.Node) ast.Expr {
	var expr ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			expr = c.convertExpr(child)
		}
	}

	return ast.NewParen(node, expr)
}

func (c *converter) convertUnaryExpr(node cst.Node) ast.Expr {
	prefix := false
	var operator *ast.Token
	var value ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		} else {
			operator = c.convertToken(child)

			if value == nil {
				prefix = true
			}
		}
	}

	return ast.NewUnary(node, prefix, operator, value)
}

func (c *converter) convertBinaryExpr(node cst.Node) ast.Expr {
	if node.Contains(scanner.Dot) {
		return c.convertMemberExpr(node)
	}
	if node.Contains(scanner.As) {
		return c.convertCastExpr(node)
	}

	var left ast.Expr
	var operator *ast.Token
	var right ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			if left == nil {
				left = c.convertExpr(child)
			} else {
				right = c.convertExpr(child)
			}
		} else {
			operator = c.convertToken(child)
		}
	}

	if node.ContainsAny(scanner.LogicalOperators) {
		return ast.NewLogical(node, left, operator, right)
	}
	if node.ContainsAny(scanner.AssignmentOperators) {
		return ast.NewAssignment(node, left, operator, right)
	}

	return ast.NewBinary(node, left, operator, right)
}

func (c *converter) convertMemberExpr(node cst.Node) ast.Expr {
	var value ast.Expr
	var name *ast.Token

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			if value == nil {
				value = c.convertExpr(child)
			} else {
				name = c.convertToken(child)
			}
		} else if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		}
	}

	return ast.NewMember(node, value, name)
}

func (c *converter) convertIndexExpr(node cst.Node) ast.Expr {
	var value ast.Expr
	var index ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			if value == nil {
				value = c.convertExpr(child)
			} else {
				index = c.convertExpr(child)
			}
		}
	}

	return ast.NewIndex(node, value, index)
}

func (c *converter) convertCastExpr(node cst.Node) ast.Expr {
	var value ast.Expr
	var target ast.Type

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		} else if child.Kind.IsType() {
			target = c.convertType(child)
		}
	}

	return ast.NewCast(node, value, target)
}

func (c *converter) convertCallExpr(node cst.Node) ast.Expr {
	var callee ast.Expr
	var args []ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			if callee == nil {
				callee = c.convertExpr(child)
			} else {
				args = append(args, c.convertExpr(child))
			}
		}
	}

	return ast.NewCall(node, callee, args)
}

func (c *converter) convertTypeCallExpr(node cst.Node) ast.Expr {
	var name *ast.Token
	var arg ast.Type

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind.IsType() {
			arg = c.convertType(child)
		}
	}

	return ast.NewTypeCall(node, name, arg)
}

func (c *converter) convertStructExpr(node cst.Node) ast.Expr {
	new_ := false
	var type_ ast.Type
	var fields []*ast.InitField

	for _, child := range node.Children {
		if child.Token.Lexeme == "new" {
			new_ = true
		} else if child.Kind.IsType() {
			type_ = c.convertType(child)
		} else if child.Kind == cst.StructFieldExprNode {
			fields = append(fields, c.convertStructFieldExpr(child))
		}
	}

	return ast.NewStructInitializer(node, new_, type_, fields)
}

func (c *converter) convertStructFieldExpr(node cst.Node) *ast.InitField {
	var name *ast.Token
	var value ast.Expr

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			if name == nil {
				name = c.convertToken(child)
			} else {
				value = c.convertExpr(child)
			}
		} else if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		}
	}

	return ast.NewInitField(node, name, value)
}

func (c *converter) convertArrayExpr(node cst.Node) ast.Expr {
	var values []ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			values = append(values, c.convertExpr(child))
		}
	}

	return ast.NewArrayInitializer(node, values)
}

func (c *converter) convertAllocateArrayExpr(node cst.Node) ast.Expr {
	var type_ ast.Type
	var count ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsType() {
			type_ = c.convertType(child)
		} else if child.Kind.IsExpr() {
			count = c.convertExpr(child)
		}
	}

	return ast.NewAllocateArray(node, type_, count)
}

func (c *converter) convertIdentifierExpr(node cst.Node) ast.Expr {
	return ast.NewIdentifier(node, node.Token)
}

func (c *converter) convertLiteral(node cst.Node) ast.Expr {
	return ast.NewLiteral(node, node.Token)
}
