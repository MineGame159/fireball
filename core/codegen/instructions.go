package codegen

import (
	"fireball/core/ast"
	"fireball/core/ir"
)

func (c *codegen) alloca(type_ ast.Type, name string, node ast.Node) ir.Value {
	pointer := c.block.Add(&ir.AllocaInst{
		Typ:   c.types.get(type_),
		Align: type_.Align(),
	})

	if name != "" {
		pointer.SetName(name)
	}

	if node != nil {
		c.setLocationMeta(pointer, node)
	}

	return pointer
}

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
