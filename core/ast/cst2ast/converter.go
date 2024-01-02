package cst2ast

import (
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
	"fireball/core/utils"
)

type converter struct {
	reporter utils.Reporter
}

func Convert(reporter utils.Reporter, path string, node cst.Node) *ast.File {
	c := converter{
		reporter: reporter,
	}

	var decls []ast.Decl

	for _, child := range node.Children {
		decls = append(decls, c.convertDecl(child))
	}

	return ast.NewFile(node, path, decls)
}

func (c *converter) convertToken(node cst.Node) *ast.Token {
	return ast.NewToken(node, node.Token)
}

func tokenEnd(node cst.Node) scanner.Token {
	return node.Children[len(node.Children)-1].Token
}

func (c *converter) error(node cst.Node, msg string) {
	c.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   node.Range,
		Message: msg,
	})
}
