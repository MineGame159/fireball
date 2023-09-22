package ast

import (
	"fireball/core"
	"fireball/core/architecture"
	"fireball/core/types"
)

// Struct

func (s *Struct) Size() int {
	layout := architecture.CLayout{}

	for _, field := range s.Fields {
		layout.Add(field.Type)
	}

	return layout.Offset
}

func (s *Struct) Align() int {
	biggest := 0

	for _, field := range s.Fields {
		biggest = max(biggest, field.Type.Align())
	}

	return biggest
}

func (s *Struct) WithRange(range_ core.Range) types.Type {
	return &Struct{
		range_: range_,
		parent: s.parent,
		Name:   s.Name,
		Fields: s.Fields,
		Type:   s.Type,
	}
}

func (s *Struct) Equals(other types.Type) bool {
	if v, ok := other.(*Struct); ok {
		if len(s.Fields) != len(v.Fields) {
			return false
		}

		for i, field := range s.Fields {
			if !field.Type.Equals(v.Fields[i].Type) {
				return false
			}
		}

		return true
	}

	return false
}

func (s *Struct) CanAssignTo(other types.Type) bool {
	return s.Equals(other)
}

// Enum

func (e *Enum) Size() int {
	return e.Type.Size()
}

func (e *Enum) Align() int {
	return e.Type.Align()
}

func (e *Enum) WithRange(range_ core.Range) types.Type {
	return &Enum{
		range_:    range_,
		parent:    e.parent,
		Name:      e.Name,
		Type:      e.Type,
		InferType: e.InferType,
		Cases:     e.Cases,
	}
}

func (e *Enum) Equals(other types.Type) bool {
	if v, ok := other.(*Enum); ok {
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

func (e *Enum) CanAssignTo(other types.Type) bool {
	return e.Equals(other)
}

// Function

func (f *Func) Size() int {
	return 8
}

func (f *Func) Align() int {
	return 8
}

func (f *Func) WithRange(range_ core.Range) types.Type {
	return &Func{
		range_:  range_,
		parent:  f.parent,
		Flags:   f.Flags,
		Name:    f.Name,
		Params:  f.Params,
		Returns: f.Returns,
		Body:    f.Body,
	}
}

func (f *Func) Equals(other types.Type) bool {
	if v, ok := other.(*Func); ok {
		if f.Name.Lexeme != v.Name.Lexeme {
			return false
		}
		if f.Parent() != v.Parent() {
			return false
		}
		if !f.Returns.Equals(v.Returns) {
			return false
		}

		if len(f.Params) != len(v.Params) {
			return false
		}

		for i, param := range f.Params {
			if !param.Type.Equals(v.Params[i].Type) {
				return false
			}
		}

		return true
	}

	return false
}

func (f *Func) CanAssignTo(other types.Type) bool {
	if v, ok := other.(*Func); ok {
		if !f.Returns.CanAssignTo(v.Returns) {
			return false
		}

		if len(f.Params) != len(v.Params) {
			return false
		}

		for i, param := range f.Params {
			if !param.Type.CanAssignTo(v.Params[i].Type) {
				return false
			}
		}

		return true
	}

	return false
}

func (f *Func) String() string {
	return f.Signature(false)
}
