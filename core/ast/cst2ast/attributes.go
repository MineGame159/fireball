package cst2ast

import (
	"fireball/core/ast"
	"fireball/core/cst"
)

func (c *converter) convertAttributes(node cst.Node) []any {
	var attributes []any

	for _, child := range node.Children {
		if child.Kind == cst.AttributeNode {
			attribute := c.convertAttribute(child)

			if attribute != nil {
				attributes = append(attributes, attribute)
			}
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
		return nil
	}
}

func (c *converter) convertExternAttribute(node cst.Node) any {
	e := ast.ExternAttribute{}

	for _, child := range node.Children {
		if child.Kind == cst.StringExprNode {
			if e.Name == "" {
				e.Name = child.Token.Lexeme[1 : len(child.Token.Lexeme)-1]
			} else {
				c.error(node, "Extern attribute only have 1 optional parameter")
			}
		}
	}

	return e
}

func (c *converter) convertIntrinsicAttribute(node cst.Node) any {
	e := ast.IntrinsicAttribute{}

	for _, child := range node.Children {
		if child.Kind == cst.StringExprNode {
			if e.Name == "" {
				e.Name = child.Token.Lexeme[1 : len(child.Token.Lexeme)-1]
			} else {
				c.error(node, "Intrinsic attribute only have 1 optional parameter")
			}
		}
	}

	return e
}

func (c *converter) convertInlineAttribute(node cst.Node) any {
	e := ast.InlineAttribute{}

	for _, child := range node.Children {
		if child.Kind == cst.StringExprNode {
			c.error(node, "Inline attribute doesn't have any parameters")
		}
	}

	return e
}
