package ast

import "strings"

type FieldLike interface {
	Node

	Underlying() *Field

	Struct() StructType

	Name() *Token
	Type() Type
}

type StructType interface {
	Type

	Underlying() *Struct

	StaticFieldCount() int
	StaticFieldIndex(index int) FieldLike
	StaticFieldName(name string) FieldLike

	FieldCount() int
	FieldIndex(index int) FieldLike
	FieldName(name string) FieldLike

	MangledName(name *strings.Builder)
}

type SpecializableFunc interface {
	FuncType

	Generics() []*Generic

	Specialize(types []Type) FuncType
}

type SpecializedParam struct {
	Param *Param
	Type  Type
}

type FuncType interface {
	Type

	Underlying() *Func
	Receiver() Type

	ParameterCount() int
	ParameterIndex(index int) SpecializedParam

	Returns() Type

	MangledName(name *strings.Builder)
}
