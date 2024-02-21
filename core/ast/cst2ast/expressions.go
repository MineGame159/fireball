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
	case cst.TypeofExprNode:
		return c.convertTypeofExpr(node)
	case cst.StructExprNode:
		return c.convertStructExpr(node)
	case cst.ArrayExprNode:
		return c.convertArrayExpr(node)
	case cst.AllocateArrayExprNode:
		return c.convertAllocateArrayExpr(node)
	case cst.IdentifierExprNode:
		return c.convertIdentifierExpr(node)
	case cst.NilExprNode, cst.BoolExprNode, cst.NumberExprNode, cst.CharacterExprNode, cst.StringExprNode:
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

	if p := ast.NewParen(node, expr); p != nil {
		return p
	}

	return nil
}

func (c *converter) convertUnaryExpr(node cst.Node) ast.Expr {
	prefix := false
	var operator *ast.Token
	var value ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		} else if child.Kind == cst.MiscNode {
			operator = c.convertToken(child)

			if value == nil {
				prefix = true
			}
		}
	}

	if u := ast.NewUnary(node, prefix, operator, value); u != nil {
		return u
	}

	return nil
}

func (c *converter) convertBinaryExpr(node cst.Node) ast.Expr {
	if node.Contains(scanner.Dot) {
		return c.convertMemberExpr(node)
	}
	if node.ContainsAny(scanner.CastOperators) {
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
		} else if child.Kind == cst.MiscNode {
			operator = c.convertToken(child)
		}
	}

	if node.ContainsAny(scanner.LogicalOperators) {
		return ast.NewLogical(node, left, operator, right)
	}
	if node.ContainsAny(scanner.AssignmentOperators) {
		return ast.NewAssignment(node, left, operator, right)
	}

	if b := ast.NewBinary(node, left, operator, right); b != nil {
		return b
	}

	return nil
}

func (c *converter) convertMemberExpr(node cst.Node) ast.Expr {
	var value ast.Expr
	var name *ast.Token
	var genericArgs []ast.Type

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		} else if child.Kind == cst.TokenNode {
			name = c.convertToken(child)
		} else if child.Kind.IsType() {
			arg := c.convertType(child)

			if arg != nil {
				genericArgs = append(genericArgs, arg)
			}
		}
	}

	if m := ast.NewMember(node, value, name, genericArgs); m != nil {
		return m
	}

	return nil
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

	if i := ast.NewIndex(node, value, index); i != nil {
		return i
	}

	return nil
}

func (c *converter) convertCastExpr(node cst.Node) ast.Expr {
	var value ast.Expr
	var operator *ast.Token
	var target ast.Type

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		} else if child.Token.Kind.IsAny(scanner.CastOperators) {
			operator = c.convertToken(child)
		} else if child.Kind.IsType() {
			target = c.convertType(child)
		}
	}

	if c := ast.NewCast(node, value, operator, target); c != nil {
		return c
	}

	return nil
}

func (c *converter) convertCallExpr(node cst.Node) ast.Expr {
	var callee ast.Expr
	var args []ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			if callee == nil {
				callee = c.convertExpr(child)
			} else {
				arg := c.convertExpr(child)

				if arg != nil {
					args = append(args, arg)
				}
			}
		}
	}

	if c := ast.NewCall(node, callee, args); c != nil {
		return c
	}

	return nil
}

func (c *converter) convertTypeCallExpr(node cst.Node) ast.Expr {
	var callee *ast.Token
	var arg ast.Type

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierExprNode {
			callee = c.convertToken(child.Children[0])
		} else if child.Kind.IsType() {
			arg = c.convertType(child)
		}
	}

	if t := ast.NewTypeCall(node, callee, arg); t != nil {
		return t
	}

	return nil
}

func (c *converter) convertTypeofExpr(node cst.Node) ast.Expr {
	var callee *ast.Token
	var arg ast.Expr

	for _, child := range node.Children {
		if child.Kind == cst.TokenNode {
			callee = c.convertToken(child)
		} else if child.Kind.IsExpr() {
			arg = c.convertExpr(child)
		}
	}

	if t := ast.NewTypeof(node, callee, arg); t != nil {
		return t
	}

	return nil
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

	if s := ast.NewStructInitializer(node, new_, type_, fields); s != nil {
		return s
	}

	return nil
}

func (c *converter) convertStructFieldExpr(node cst.Node) *ast.InitField {
	var name *ast.Token
	var value ast.Expr

	for _, child := range node.Children {
		if child.Kind == cst.TokenNode {
			name = c.convertToken(child)
		} else if child.Kind.IsExpr() {
			value = c.convertExpr(child)
		}
	}

	if i := ast.NewInitField(node, name, value); i != nil {
		return i
	}

	return nil
}

func (c *converter) convertArrayExpr(node cst.Node) ast.Expr {
	var values []ast.Expr

	for _, child := range node.Children {
		if child.Kind.IsExpr() {
			values = append(values, c.convertExpr(child))
		}
	}

	if a := ast.NewArrayInitializer(node, values); a != nil {
		return a
	}

	return nil
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

	if a := ast.NewAllocateArray(node, type_, count); a != nil {
		return a
	}

	return nil
}

func (c *converter) convertIdentifierExpr(node cst.Node) ast.Expr {
	var name *ast.Token
	var genericArgs []ast.Type

	for _, child := range node.Children {
		if child.Kind == cst.TokenNode {
			name = c.convertToken(child)
		} else if child.Kind.IsType() {
			genericArg := c.convertType(child)

			if genericArg != nil {
				genericArgs = append(genericArgs, genericArg)
			}
		}
	}

	if i := ast.NewIdentifier(node, name, genericArgs); i != nil {
		return i
	}

	return nil
}

func (c *converter) convertLiteral(node cst.Node) ast.Expr {
	if l := ast.NewLiteral(node, node.Token); l != nil {
		return l
	}

	return nil
}
