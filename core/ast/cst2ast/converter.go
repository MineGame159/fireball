package cst2ast

import (
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/utils"
)

type converter struct {
	reporter utils.Reporter
}

func Convert(reporter utils.Reporter, path string, node cst.Node) *ast.File {
	c := converter{
		reporter: reporter,
	}

	var namespace *ast.Namespace
	var decls []ast.Decl

	reportedMissingNamespace := false

	for _, child := range node.Children {
		if child.Kind == cst.NamespaceDeclNode {
			if namespace == nil {
				namespace = c.convertNamespaceDecl(child)
			} else {
				c.error(child, "There can only be one top-level file namespace")
			}
		} else if child.Kind.IsDecl() {
			if namespace == nil && !reportedMissingNamespace {
				c.error(child, "First declaration of a file needs to be a namespace")
				reportedMissingNamespace = true
			}

			decl := c.convertDecl(child)

			if decl != nil {
				decls = append(decls, decl)
			}
		}
	}

	return ast.NewFile(node, path, namespace, decls)
}

func (c *converter) convertToken(node cst.Node) *ast.Token {
	return ast.NewToken(node, node.Token)
}

func (c *converter) error(node cst.Node, msg string) {
	c.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   node.Range,
		Message: msg,
	})
}
