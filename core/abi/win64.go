package abi

import (
	"fireball/core/ast"
)

var WIN64 Abi = &win64{}

type win64 struct{}

func (w *win64) Size(type_ ast.Type) uint32 {
	return getX64Size(w, type_)
}

func (w *win64) Align(type_ ast.Type) uint32 {
	return getX64Align(type_)
}

func (w *win64) Classify(type_ ast.Type, args []Arg) []Arg {
	switch type_ := ast.Resolved(type_).(type) {
	case *ast.Primitive:
		var arg Arg

		switch type_.Kind {
		case ast.Bool:
			arg = i1

		case ast.U8, ast.I8:
			arg = i8
		case ast.U16, ast.I16:
			arg = i16
		case ast.U32, ast.I32:
			arg = i32
		case ast.U64, ast.I64:
			arg = i64

		case ast.F32:
			arg = f32
		case ast.F64:
			arg = f64

		default:
			return args
		}

		return append(args, arg)

	case *ast.Pointer, ast.FuncType:
		return append(args, ptr)

	case *ast.Array, ast.StructType, *ast.Interface:
		var arg Arg

		switch w.Size(type_) {
		case 1:
			arg = i8
		case 2:
			arg = i16
		case 4:
			arg = i32
		case 8:
			arg = i64
		default:
			arg = memory
		}

		return append(args, arg)

	case *ast.Enum:
		return w.Classify(type_.ActualType, args)

	default:
		return args
	}
}
