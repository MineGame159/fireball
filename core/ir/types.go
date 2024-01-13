package ir

type Type interface {
	isType()
}

// Void

var Void = &VoidType{}

type VoidType struct{}

func (v *VoidType) isType() {}

// Int

var I1 = &IntType{BitSize: 1}
var I8 = &IntType{BitSize: 8}
var I16 = &IntType{BitSize: 16}
var I32 = &IntType{BitSize: 32}
var I64 = &IntType{BitSize: 64}

type IntType struct {
	BitSize uint8
}

func (i *IntType) isType() {}

// Float

var F32 = &FloatType{BitSize: 32}
var F64 = &FloatType{BitSize: 64}

type FloatType struct {
	BitSize uint8
}

func (f *FloatType) isType() {}

// Pointer

type PointerType struct {
	Pointee Type
}

func (p *PointerType) isType() {}

// Array

type ArrayType struct {
	Count uint32
	Base  Type
}

func (a *ArrayType) isType() {}

// Struct

type StructType struct {
	Name   string
	Fields []Type
}

func (s *StructType) isType() {}

// Function

type Param struct {
	Typ   Type
	Name_ string
}

func (p *Param) Type() Type {
	return p.Typ
}

func (p *Param) Name() string {
	return p.Name_
}

func (p *Param) SetName(name string) {
	p.Name_ = name
}

type FuncType struct {
	Returns  Type
	Params   []*Param
	Variadic bool
}

func (f *FuncType) isType() {}

// Meta

var MetaT = &MetaType{}

type MetaType struct{}

func (m *MetaType) isType() {}
