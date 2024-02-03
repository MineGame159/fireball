package abi

import "fireball/core/ast"

var FB Abi = &fb{}

type fb struct{}

func (f *fb) Size(type_ ast.Type) uint32 {
	if type_, ok := ast.As[*ast.Struct](type_); ok {
		layout := cLayout{}

		for _, field := range type_.Fields {
			layout.add(f.Size(field.Type), f.Align(field.Type))
		}

		return layout.size()
	}

	return getX64Size(f, type_)
}

func (f *fb) Align(type_ ast.Type) uint32 {
	return getX64Align(type_)
}

func (f *fb) Fields(decl *ast.Struct) ([]*ast.Field, []uint32) {
	layout := cLayout{}
	offsets := make([]uint32, len(decl.Fields))

	for i, field := range decl.Fields {
		offsets[i] = layout.add(f.Size(field.Type), f.Align(field.Type))
	}

	return decl.Fields, offsets
}

func (f *fb) Classify(type_ ast.Type, args []Arg) []Arg {
	switch type_.Resolved().(type) {
	default:
		panic("abi.amd64.Classify() - Not implemented")
	}
}
