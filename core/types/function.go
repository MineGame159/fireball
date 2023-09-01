package types

import (
	"fireball/core"
	"strings"
)

type FunctionType struct {
	range_ core.Range

	Params   []Type
	Variadic bool
	Returns  Type
}

func Function(params []Type, variadic bool, returns Type, range_ core.Range) *FunctionType {
	return &FunctionType{
		range_:   range_,
		Params:   params,
		Variadic: variadic,
		Returns:  returns,
	}
}

func (f *FunctionType) Range() core.Range {
	return f.range_
}

func (f *FunctionType) Size() int {
	return 4
}

func (f *FunctionType) Copy() Type {
	params := make([]Type, len(f.Params))

	for i, param := range f.Params {
		params[i] = param.Copy()
	}

	return &FunctionType{
		Params:   params,
		Variadic: f.Variadic,
		Returns:  f.Returns.Copy(),
	}
}

func (f *FunctionType) Equals(other Type) bool {
	if v, ok := other.(*FunctionType); ok {
		if !f.Returns.Equals(v.Returns) {
			return false
		}

		if len(f.Params) != len(v.Params) {
			return false
		}

		for i, param := range f.Params {
			if !param.Equals(v.Params[i]) {
				return false
			}
		}

		return true
	}

	return false
}

func (f *FunctionType) CanAssignTo(other Type) bool {
	if v, ok := other.(*FunctionType); ok {
		if !f.Returns.CanAssignTo(v.Returns) {
			return false
		}

		if len(f.Params) != len(v.Params) {
			return false
		}

		for i, param := range f.Params {
			if !param.CanAssignTo(v.Params[i]) {
				return false
			}
		}

		return true
	}

	return false
}

func (f *FunctionType) AcceptChildren(visitor Visitor) {
	for i := range f.Params {
		visitor.VisitType(f.Params[i])
	}

	visitor.VisitType(f.Returns)
}

func (f *FunctionType) AcceptChildrenPtr(visitor PtrVisitor) {
	for i := range f.Params {
		visitor.VisitType(&f.Params[i])
	}

	visitor.VisitType(&f.Returns)
}

func (f *FunctionType) String() string {
	str := strings.Builder{}

	str.WriteString("func(")

	for i, param := range f.Params {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(param.String())
	}

	str.WriteString(") ")
	str.WriteString(f.Returns.String())

	return str.String()
}
