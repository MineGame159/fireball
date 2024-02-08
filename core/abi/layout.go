package abi

import "fireball/core/ast"

type Layout interface {
	Size(abi Abi, decl *ast.Struct) uint32

	Fields(abi Abi, decl *ast.Struct) ([]*ast.Field, []uint32)
}

func GetStructLayout(s *ast.Struct) Layout {
	return CLayout
}
