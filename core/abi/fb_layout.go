package abi

import (
	"fireball/core/ast"
	"slices"
)

var FbLayout Layout = &fbLayout{}

type fbLayout struct{}

func (f *fbLayout) Size(abi Abi, decl *ast.Struct) uint32 {
	fields := f.sorted(abi, decl)

	layout := cFieldAligner{}

	for _, field := range fields {
		layout.add(abi.Size(field.Type), abi.Align(field.Type))
	}

	return layout.size()
}

func (f *fbLayout) Fields(abi Abi, decl *ast.Struct) ([]*ast.Field, []uint32) {
	fields := f.sorted(abi, decl)

	layout := cFieldAligner{}
	offsets := make([]uint32, len(fields))

	for i, field := range fields {
		offsets[i] = layout.add(abi.Size(field.Type), abi.Align(field.Type))
	}

	return fields, offsets
}

func (f *fbLayout) sorted(abi Abi, decl *ast.Struct) []*ast.Field {
	fields := make([]*ast.Field, len(decl.Fields))
	copy(fields, decl.Fields)

	slices.SortStableFunc(fields, func(f1, f2 *ast.Field) int {
		a1 := abi.Align(f1.Type)
		a2 := abi.Align(f2.Type)

		if a1 < a2 {
			return +1
		}
		if a1 > a2 {
			return -1
		}
		return 0
	})

	return fields
}
