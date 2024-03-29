package codegen

import (
	"fireball/core/ast"
	"fireball/core/ir"
	"fireball/core/scanner"
)

func (c *codegen) setLocationMeta(value ir.MetaValue, node ast.Node) {
	if node == nil {
		return
	}

	meta := &ir.LocationMeta{
		Scope: c.scopes.getMeta(),
	}

	if node.Cst() != nil {
		pos := node.Cst().Range.Start

		meta.Line = uint32(pos.Line)
		meta.Column = uint32(pos.Column)
	}

	value.SetMeta(c.module.Meta(meta))
}

func (c *codegen) setLocationMetaCst(value ir.MetaValue, node ast.Node, kind scanner.TokenKind) {
	if node == nil {
		return
	}

	meta := &ir.LocationMeta{
		Scope: c.scopes.getMeta(),
	}

	if node.Cst() != nil {
		token := node.Cst().Get(kind)

		meta.Line = uint32(token.Range.Start.Line)
		meta.Column = uint32(token.Range.Start.Column)
	}

	value.SetMeta(c.module.Meta(meta))
}
