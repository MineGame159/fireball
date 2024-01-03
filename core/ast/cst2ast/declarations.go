package cst2ast

import (
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
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
	var name *ast.Token
	var fields []*ast.Field
	var staticFields []*ast.Field

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind == cst.StructFieldNode {
			field, static := c.convertStructField(child)

			if static {
				staticFields = append(staticFields, field)
			} else {
				fields = append(fields, field)
			}
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0], "Structs cannot have attributes")
		}
	}

	return ast.NewStruct(node, name, fields, staticFields)
}

func (c *converter) convertStructField(node cst.Node) (*ast.Field, bool) {
	var name *ast.Token
	var type_ ast.Type

	static := false

	for _, child := range node.Children {
		if child.Token.Kind == scanner.Static {
			static = true
		} else if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind.IsType() {
			type_ = c.convertType(child)
		}
	}

	return ast.NewField(node, name, type_), static
}

// Impl

func (c *converter) convertImplDecl(node cst.Node) ast.Decl {
	var struct_ *ast.Token
	var methods []*ast.Func

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			struct_ = c.convertToken(child)
		} else if child.Kind == cst.FuncNode {
			methods = append(methods, c.convertFuncDecl(child))
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0], "Implementations cannot have attributes")
		}
	}

	return ast.NewImpl(node, struct_, methods)
}

// Enum

func (c *converter) convertEnumDecl(node cst.Node) ast.Decl {
	var name *ast.Token
	var type_ ast.Type
	var cases []*ast.EnumCase

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind.IsType() {
			type_ = c.convertType(child)
		} else if child.Kind == cst.EnumCaseNode {
			cases = append(cases, c.convertEnumCase(child))
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0], "Enums cannot have attributes")
		}
	}

	return ast.NewEnum(node, name, type_, cases)
}

func (c *converter) convertEnumCase(node cst.Node) *ast.EnumCase {
	var name *ast.Token
	var value *ast.Token

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind == cst.NumberExprNode {
			value = c.convertToken(child)
		}
	}

	return ast.NewEnumCase(node, name, value)
}

// Func

func (c *converter) convertFuncDecl(node cst.Node) *ast.Func {
	var attributes []*ast.Attribute
	var flags ast.FuncFlags
	var name *ast.Token
	var params []*ast.Param
	var returns ast.Type
	var body []ast.Stmt

	reported := false

	for _, child := range node.Children {
		if child.Token.Kind == scanner.Static {
			flags |= ast.Static
		} else if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind == cst.FuncParamNode {
			param, varArgs := c.convertFuncParam(child)

			if varArgs {
				flags |= ast.Variadic
			} else {
				if flags&ast.Variadic != 0 && !reported {
					c.error(child, "Variadic arguments can only appear at the end of the parameter list")
					reported = true
				}

				params = append(params, param)
			}
		} else if child.Kind.IsType() {
			returns = c.convertType(child)
		} else if child.Kind.IsStmt() {
			body = append(body, c.convertStmt(child))
		} else if child.Kind == cst.AttributesNode {
			attributes = c.convertAttributes(child)
		}
	}

	needsBody := true

	for _, attribute := range attributes {
		if attribute.Name.String() == "Extern" || attribute.Name.String() == "Intrinsic" {
			needsBody = false
			break
		}
	}

	if needsBody {
		if !node.Contains(scanner.LeftBrace) {
			c.error(node, "Functions without the extern or intrinsic attribute need to have a body")
		}
	} else {
		if node.Contains(scanner.LeftBrace) {
			c.error(node, "Functions with the extern or intrinsic attribute can't have a body")
		}
	}

	if returns == nil {
		returns = ast.NewPrimitive(cst.Node{}, ast.Void, scanner.Token{})
	}

	return ast.NewFunc(node, attributes, flags, name, params, returns, body)
}

func (c *converter) convertFuncParam(node cst.Node) (*ast.Param, bool) {
	var name *ast.Token
	var type_ ast.Type

	varArgs := false

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind.IsType() {
			type_ = c.convertType(child)
		} else if child.Token.Kind == scanner.DotDotDot {
			varArgs = true
		}
	}

	return ast.NewParam(node, name, type_), varArgs
}

// Attributes

func (c *converter) convertAttributes(node cst.Node) []*ast.Attribute {
	var attributes []*ast.Attribute

	for _, child := range node.Children {
		if child.Kind == cst.AttributeNode {
			attributes = append(attributes, c.convertAttribute(child))
		}
	}

	return attributes
}

func (c *converter) convertAttribute(node cst.Node) *ast.Attribute {
	var name *ast.Token
	var args []*ast.Token

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind == cst.StringExprNode {
			args = append(args, c.convertToken(child))
		}
	}

	return ast.NewAttribute(node, name, args)
}
