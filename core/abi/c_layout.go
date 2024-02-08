package abi

import (
	"fireball/core/ast"
)

var CLayout Layout = &cLayout{}

type cLayout struct{}

func (c *cLayout) Size(abi Abi, decl *ast.Struct) uint32 {
	layout := cFieldAligner{}

	for _, field := range decl.Fields {
		layout.add(abi.Size(field.Type), abi.Align(field.Type))
	}

	return layout.size()
}

func (c *cLayout) Fields(abi Abi, decl *ast.Struct) ([]*ast.Field, []uint32) {
	layout := cFieldAligner{}
	offsets := make([]uint32, len(decl.Fields))

	for i, field := range decl.Fields {
		offsets[i] = layout.add(abi.Size(field.Type), abi.Align(field.Type))
	}

	return decl.Fields, offsets
}

type cFieldAligner struct {
	biggestAlign uint32
	offset       uint32
}

func (l *cFieldAligner) add(size, align uint32) uint32 {
	l.biggestAlign = max(l.biggestAlign, align)

	offset := alignBytes(l.offset, align)
	l.offset = offset + size

	return offset
}

func (l *cFieldAligner) size() uint32 {
	if l.offset == 0 {
		return 0
	}

	return alignBytes(l.offset, l.biggestAlign)
}
