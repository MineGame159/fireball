package abi

import (
	"fireball/core/ast"
)

var CLayout Layout = &cLayout{}

type cLayout struct{}

func (c *cLayout) Size(abi Abi, decl ast.StructType) uint32 {
	layout := cFieldAligner{}

	for i := 0; i < decl.FieldCount(); i++ {
		type_ := decl.FieldIndex(i).Type()
		layout.add(abi.Size(type_), abi.Align(type_))
	}

	return layout.size()
}

func (c *cLayout) Fields(abi Abi, decl ast.StructType) ([]ast.FieldLike, []uint32) {
	layout := cFieldAligner{}

	fields := make([]ast.FieldLike, decl.FieldCount())
	offsets := make([]uint32, len(fields))

	for i := 0; i < decl.FieldCount(); i++ {
		field := decl.FieldIndex(i)

		fields[i] = field
		offsets[i] = layout.add(abi.Size(field.Type()), abi.Align(field.Type()))
	}

	return fields, offsets
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
