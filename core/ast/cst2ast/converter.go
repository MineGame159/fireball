package cst2ast

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/cst"
	"fireball/core/scanner"
	"fireball/core/utils"
)

type converter struct {
	reporter utils.Reporter
}

func Convert(reporter utils.Reporter, node cst.Node) []ast.Decl {
	c := converter{
		reporter: reporter,
	}

	var nodes []ast.Decl

	for _, child := range node.Children {
		nodes = append(nodes, c.convertDecl(child))
	}

	return nodes
}

func tokenEnd(node cst.Node) scanner.Token {
	return node.Children[len(node.Children)-1].Token
}

func (c *converter) error(token scanner.Token, msg string) {
	c.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   core.TokenToRange(token),
		Message: msg,
	})
}
