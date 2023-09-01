package types

import "strings"

type FunctionType struct {
	Params   []Type
	Variadic bool
	Returns  Type
}

func (f *FunctionType) Size() int {
	return 4
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

func (f *FunctionType) AcceptTypes(visitor Visitor) {
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
