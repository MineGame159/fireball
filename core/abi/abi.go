package abi

import (
	"fireball/core/ast"
	"runtime"
)

type ClassKind uint8

const (
	None ClassKind = iota
	Integer
	SSE
	Memory
)

type Arg struct {
	Class ClassKind
	Bits  uint32
}

func (a Arg) Bytes() uint32 {
	return max(a.Bits, 8) / 8
}

type Abi interface {
	Size(type_ ast.Type) uint32
	Align(type_ ast.Type) uint32

	Fields(decl *ast.Struct) ([]*ast.Field, []uint32)

	Classify(type_ ast.Type, args []Arg) []Arg
}

func GetTargetAbi() Abi {
	switch runtime.GOOS {
	case "windows":
		return WIN64
	case "linux", "darwin":
		return AMD64

	default:
		panic("abi.GetTargetAbi() - Not implemented")
	}
}

func GetStructAbi(decl *ast.Struct) Abi {
	return GetTargetAbi()
}

func GetFuncAbi(decl *ast.Func) Abi {
	return GetTargetAbi()

	/*for _, attribute := range decl.Attributes {
		if attribute.Name.String() == "Extern" {
			return GetTargetAbi()
		}
	}

	return FB*/
}
