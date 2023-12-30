package cst2ast

import (
	"fireball/core/cst"
	"fireball/core/types"
)

func (c *converter) convertAttributes(node cst.Node) []any {
	var attributes []any

	for _, child := range node.Children {
		if child.Kind == cst.AttributeNode {
			attributes = append(attributes, c.convertAttribute(child))
		}
	}

	return attributes
}

func (c *converter) convertAttribute(node cst.Node) any {
	switch node.Children[0].Token.Lexeme {
	case "Extern":
		return c.convertExternAttribute(node)
	case "Intrinsic":
		return c.convertIntrinsicAttribute(node)
	case "Inline":
		return c.convertInlineAttribute(node)

	default:
		panic("cst2ast.convertAttribute() - Not implemented")
	}
}

func (c *converter) convertExternAttribute(node cst.Node) any {
	e := types.ExternAttribute{}

	for _, child := range node.Children {
		if child.Kind == cst.StringExprNode {
			if e.Name == "" {
				e.Name = child.Token.Lexeme[1 : len(child.Token.Lexeme)-1]
			} else {
				c.error(node.Token, "Extern attribute only have 1 optional parameter")
			}
		}
	}

	return e
}

func (c *converter) convertIntrinsicAttribute(node cst.Node) any {
	e := types.IntrinsicAttribute{}

	for _, child := range node.Children {
		if child.Kind == cst.StringExprNode {
			if e.Name == "" {
				e.Name = child.Token.Lexeme[1 : len(child.Token.Lexeme)-1]
			} else {
				c.error(node.Token, "Intrinsic attribute only have 1 optional parameter")
			}
		}
	}

	return e
}

func (c *converter) convertInlineAttribute(node cst.Node) any {
	e := types.InlineAttribute{}

	for _, child := range node.Children {
		if child.Kind == cst.StringExprNode {
			c.error(node.Token, "Inline attribute doesn't have any parameters")
		}
	}

	return e
}
