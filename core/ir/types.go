package ir

import "slices"

type Type interface {
	isType()

	Equals(other Type) bool
}

// Void

var Void = &VoidType{}

type VoidType struct{}

func (v *VoidType) isType() {}

func (v *VoidType) Equals(other Type) bool {
	if _, ok := other.(*VoidType); ok {
		return true
	}

	return false
}

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

func (i *IntType) Equals(other Type) bool {
	if other, ok := other.(*IntType); ok {
		return i.BitSize == other.BitSize
	}

	return false
}

// Float

var F32 = &FloatType{BitSize: 32}
var F64 = &FloatType{BitSize: 64}

type FloatType struct {
	BitSize uint8
}

func (f *FloatType) isType() {}

func (f *FloatType) Equals(other Type) bool {
	if other, ok := other.(*FloatType); ok {
		return f.BitSize == other.BitSize
	}

	return false
}

// Pointer

type PointerType struct {
	Pointee Type

	ByVal Type
	SRet  Type
}

func (p *PointerType) isType() {}

func (p *PointerType) Equals(other Type) bool {
	if other, ok := other.(*PointerType); ok {
		return p.Pointee.Equals(other.Pointee)
	}

	return false
}

// Array

type ArrayType struct {
	Count uint32
	Base  Type
}

func (a *ArrayType) isType() {}

func (a *ArrayType) Equals(other Type) bool {
	if other, ok := other.(*ArrayType); ok {
		return a.Count == other.Count && a.Base.Equals(other.Base)
	}

	return false
}

// Struct

type StructType struct {
	Name   string
	Fields []Type
}

func (s *StructType) isType() {}

func (s *StructType) Equals(other Type) bool {
	if other, ok := other.(*StructType); ok {
		return slices.EqualFunc(s.Fields, other.Fields, func(t Type, t2 Type) bool {
			return t.Equals(t2)
		})
	}

	return false
}

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

func (f *FuncType) Equals(other Type) bool {
	if other, ok := other.(*FuncType); ok {
		return f.Returns.Equals(other.Returns) && f.Variadic == other.Variadic && slices.EqualFunc(f.Params, other.Params, func(param *Param, param2 *Param) bool {
			return param.Typ.Equals(param2.Typ)
		})
	}

	return false
}

// Meta

var MetaT = &MetaType{}

type MetaType struct{}

func (m *MetaType) isType() {}

func (m *MetaType) Equals(other Type) bool {
	if _, ok := other.(*MetaType); ok {
		return true
	}

	return false
}
