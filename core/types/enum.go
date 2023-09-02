package types

import "fireball/core"

type EnumType struct {
	range_ core.Range

	Name  string
	Type  Type
	Cases []EnumCase
}

type EnumCase struct {
	Name  string
	Value int
}

func Enum(name string, type_ Type, Cases []EnumCase, range_ core.Range) *EnumType {
	return &EnumType{
		range_: range_,
		Name:   name,
		Type:   type_,
		Cases:  Cases,
	}
}

func (e *EnumType) Range() core.Range {
	return e.range_
}

func (e *EnumType) Size() int {
	return e.Type.Size()
}

func (e *EnumType) Copy() Type {
	return &EnumType{
		Name:  e.Name,
		Type:  e.Type.Copy(),
		Cases: e.Cases,
	}
}

func (e *EnumType) Equals(other Type) bool {
	if v, ok := other.(*EnumType); ok {
		if !e.Type.Equals(v.Type) || len(e.Cases) != len(v.Cases) {
			return false
		}

		for i, eCase := range e.Cases {
			vCase := v.Cases[i]

			if eCase.Name != vCase.Name || eCase.Value != vCase.Value {
				return false
			}
		}

		return true
	}

	return false
}

func (e *EnumType) CanAssignTo(other Type) bool {
	return e.Equals(other)
}

func (e *EnumType) AcceptChildren(visitor Visitor) {
	visitor.VisitType(e.Type)
}

func (e *EnumType) AcceptChildrenPtr(visitor PtrVisitor) {
	visitor.VisitType(&e.Type)
}

func (e *EnumType) String() string {
	return e.Name
}

func (e *EnumType) GetCase(name string) *EnumCase {
	for i, case_ := range e.Cases {
		if case_.Name == name {
			return &e.Cases[i]
		}
	}

	return nil
}
