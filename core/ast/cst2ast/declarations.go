package cst2ast

import (
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
)

func (c *converter) convertDecl(node cst.Node) ast.Decl {
	switch node.Kind {
	case cst.NamespaceDeclNode:
		if namespace := c.convertNamespaceDecl(node); namespace != nil {
			return namespace
		}

		return nil
	case cst.UsingDeclNode:
		return c.convertUsingDecl(node)

	case cst.StructDeclNode:
		return c.convertStructDecl(node)
	case cst.ImplDeclNode:
		return c.convertImplDecl(node)
	case cst.EnumDeclNode:
		return c.convertEnumDecl(node)
	case cst.InterfaceDeclNode:
		return c.convertInterfaceDecl(node)
	case cst.FuncDeclNode:
		if f := c.convertFuncDecl(node); f != nil {
			return f
		}

		return nil
	case cst.VarDeclNode:
		return c.convertVarDecl(node)

	default:
		panic("cst2ast.convertDecl() - Not implemented")
	}
}

// Namespace

func (c *converter) convertNamespaceDecl(node cst.Node) *ast.Namespace {
	var name *ast.NamespaceName

	for _, child := range node.Children {
		if child.Kind == cst.NamespaceNameNode {
			name = c.convertNamespaceName(child)
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0], "Namespaces cannot have attributes")
		}
	}

	return ast.NewNamespace(node, name)
}

func (c *converter) convertUsingDecl(node cst.Node) ast.Decl {
	var name *ast.NamespaceName

	for _, child := range node.Children {
		if child.Kind == cst.NamespaceNameNode {
			name = c.convertNamespaceName(child)
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0], "Using cannot have attributes")
		}
	}

	if u := ast.NewUsing(node, name); u != nil {
		return u
	}

	return nil
}

func (c *converter) convertNamespaceName(node cst.Node) *ast.NamespaceName {
	var parts []*ast.Token

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			part := c.convertToken(child)

			if part != nil {
				parts = append(parts, part)
			}
		}
	}

	return ast.NewNamespaceName(node, parts)
}

// Struct

func (c *converter) convertStructDecl(node cst.Node) ast.Decl {
	var attributes []*ast.Attribute
	var name *ast.Token
	var fields []*ast.Field
	var staticFields []*ast.Field

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind == cst.StructFieldNode {
			field, static := c.convertStructField(child)

			if field != nil {
				if static {
					staticFields = append(staticFields, field)
				} else {
					fields = append(fields, field)
				}
			}
		} else if child.Kind == cst.AttributesNode {
			attributes = c.convertAttributes(child)
		}
	}

	if s := ast.NewStruct(node, attributes, name, fields, staticFields); s != nil {
		return s
	}

	return nil
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
	var implements ast.Type
	var methods []*ast.Func

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			struct_ = c.convertToken(child)
		} else if child.Kind.IsType() {
			implements = c.convertType(child)
		} else if child.Kind == cst.FuncDeclNode {
			method := c.convertFuncDecl(child)

			if method != nil {
				methods = append(methods, method)
			}
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0], "Implementations cannot have attributes")
		}
	}

	if i := ast.NewImpl(node, struct_, implements, methods); i != nil {
		return i
	}

	return nil
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
			case_ := c.convertEnumCase(child)

			if case_ != nil {
				cases = append(cases, case_)
			}
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0], "Enums cannot have attributes")
		}
	}

	if e := ast.NewEnum(node, name, type_, cases); e != nil {
		return e
	}

	return nil
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

// Interface

func (c *converter) convertInterfaceDecl(node cst.Node) ast.Decl {
	var name *ast.Token
	var methods []*ast.Func

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind == cst.FuncDeclNode {
			method := c.convertFuncDecl(child)

			if method != nil {
				methods = append(methods, method)
			}
		}
	}

	if i := ast.NewInterface(node, name, methods); i != nil {
		return i
	}

	return nil
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
			} else if param != nil {
				if flags&ast.Variadic != 0 && !reported {
					c.error(child, "Variadic arguments can only appear at the end of the parameter list")
					reported = true
				}

				params = append(params, param)
			}
		} else if child.Kind.IsType() {
			returns = c.convertType(child)
		} else if child.Kind.IsStmt() {
			stmt := c.convertStmt(child)

			if stmt != nil {
				body = append(body, stmt)
			}
		} else if child.Kind == cst.AttributesNode {
			attributes = c.convertAttributes(child)
		}
	}

	if returns == nil && (attributes != nil || flags != 0 || name != nil || params != nil || body != nil) {
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

	if p := ast.NewParam(node, name, type_); p != nil {
		return p, varArgs
	}

	return nil, varArgs
}

// Var

func (c *converter) convertVarDecl(node cst.Node) ast.Decl {
	var name *ast.Token
	var type_ ast.Type

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind.IsType() {
			type_ = c.convertType(child)
		} else if child.Kind == cst.AttributesNode {
			c.error(child.Children[0], "Global variables cannot have attributes")
		}
	}

	if v := ast.NewGlobalVar(node, name, type_); v != nil {
		return v
	}

	return nil
}

// Attributes

func (c *converter) convertAttributes(node cst.Node) []*ast.Attribute {
	var attributes []*ast.Attribute

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

func (c *converter) convertAttribute(node cst.Node) *ast.Attribute {
	var name *ast.Token
	var args []*ast.Token

	for _, child := range node.Children {
		if child.Kind == cst.IdentifierNode {
			name = c.convertToken(child)
		} else if child.Kind == cst.StringExprNode {
			arg := c.convertToken(child)

			if arg != nil {
				args = append(args, arg)
			}
		}
	}

	return ast.NewAttribute(node, name, args)
}
