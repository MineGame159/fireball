package abi

import (
	"fireball/core/ast"
)

var AMD64 Abi = &amd64{}

type amd64 struct{}

func (a *amd64) Size(type_ ast.Type) uint32 {
	return getX64Size(a, type_)
}

func (a *amd64) Align(type_ ast.Type) uint32 {
	return getX64Align(type_)
}

func (a *amd64) Classify(type_ ast.Type, args []Arg) []Arg {
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

	case *ast.Pointer:
		return append(args, ptr)

	case ast.StructType, *ast.Array:
		if a.Size(type_) > 64 {
			return append(args, memory)
		}

		args = a.flatten(type_, getSize(args), args)
		size := getSize(args)

		if size > 16 {
			args = args[0:0]
			return append(args, memory)
		}

		for _, arg := range args {
			if arg.Class == Memory {
				args = args[0:0]
				return append(args, memory)
			}
		}

		return args

	case *ast.Enum:
		return a.Classify(type_.ActualType, args)

	case *ast.Interface:
		args = append(args, ptr)
		return append(args, ptr)

	case ast.FuncType:
		return append(args, ptr)

	default:
		return args
	}
}

func (a *amd64) flatten(type_ ast.Type, baseOffset uint32, args []Arg) []Arg {
	switch type_ := ast.Resolved(type_).(type) {
	case *ast.Array:
		baseSize := a.Size(type_.Base)
		baseAlign := a.Align(type_.Base)

		for i := uint32(0); i < type_.Count; i++ {
			offset := alignBytes(baseOffset+baseSize*i, baseAlign)
			args = a.flatten(type_.Base, offset, args)
		}

		return args

	case ast.StructType:
		fields, offsets := GetStructLayout(type_.Underlying()).Fields(a, type_)

		for i, field := range fields {
			offset := baseOffset + offsets[i]
			args = a.flatten(field.Type(), offset, args)
		}

		return args

	case *ast.Interface:
		void := ast.Primitive{Kind: ast.Void}
		ptr := ast.Pointer{Pointee: &void}

		args = a.flatten(&ptr, baseOffset, args)
		return a.flatten(&ptr, baseOffset+8, args)

	default:
		typeArgs := a.Classify(type_, nil)
		if len(typeArgs) != 1 {
			panic("abi.amd64.flatten() - Failed to flatten type")
		}

		arg := typeArgs[0]
		offset := alignBytes(baseOffset, a.Align(type_))

		var finalArg *Arg
		args = getArg(args, offset, &finalArg)

		if finalArg.Class == None {
			*finalArg = arg
		} else if finalArg.Class == arg.Class {
			finalArg.Bits += arg.Bits
		} else if finalArg.Class == Memory || arg.Class == Memory {
			finalArg.Class = Memory
			finalArg.Bits += arg.Bits
		} else if finalArg.Class == Integer || arg.Class == Integer {
			finalArg.Class = Integer
			finalArg.Bits += arg.Bits
		} else {
			finalArg.Class = SSE
			finalArg.Bits += arg.Bits
		}

		return args
	}
}
