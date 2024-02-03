package abi

import "fireball/core/ast"

func getX64Size(abi Abi, type_ ast.Type) uint32 {
	switch type_ := type_.Resolved().(type) {
	case *ast.Primitive:
		return getX64PrimitiveSize(type_.Kind)

	case *ast.Pointer:
		return 8

	case *ast.Array:
		return abi.Size(type_.Base) * type_.Count

	case *ast.Enum:
		return abi.Size(type_.ActualType)

	case *ast.Interface:
		return 8 * 2

	case *ast.Func:
		return 8

	default:
		panic("abi.getX64Size() - Not implemented")
	}
}

func getX64Align(type_ ast.Type) uint32 {
	switch type_ := type_.Resolved().(type) {
	case *ast.Primitive:
		return getX64PrimitiveSize(type_.Kind)

	case *ast.Pointer:
		return 8

	case *ast.Array:
		return getX64Align(type_.Base)

	case *ast.Struct:
		maxAlign := uint32(0)

		for _, field := range type_.Fields {
			maxAlign = max(maxAlign, getX64Align(field.Type))
		}

		return maxAlign

	case *ast.Enum:
		return getX64Align(type_.ActualType)

	case *ast.Interface:
		return 8

	case *ast.Func:
		return 8

	default:
		panic("abi.getX64Align() - Not implemented")
	}
}

func getX64PrimitiveSize(kind ast.PrimitiveKind) uint32 {
	switch kind {
	case ast.Void:
		return 0
	case ast.Bool, ast.U8, ast.I8:
		return 1
	case ast.U16, ast.I16:
		return 2
	case ast.U32, ast.I32, ast.F32:
		return 4
	case ast.U64, ast.I64, ast.F64:
		return 8

	default:
		panic("abi.getX64Size.Primitive() - Not implemented")
	}
}
