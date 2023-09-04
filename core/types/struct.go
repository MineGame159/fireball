package types

import (
	"fireball/core"
)

type StructType struct {
	range_ core.Range

	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type Type
}

func Struct(name string, fields []Field, range_ core.Range) *StructType {
	return &StructType{
		range_: range_,
		Name:   name,
		Fields: fields,
	}
}

func (s *StructType) Range() core.Range {
	return s.range_
}

func (s *StructType) Size() int {
	size := 0

	for _, field := range s.Fields {
		size += field.Type.Size()
	}

	return size
}

func (s *StructType) Copy() Type {
	fields := make([]Field, len(s.Fields))

	for i, field := range s.Fields {
		fields[i] = Field{
			Name: field.Name,
			Type: field.Type.Copy(),
		}
	}

	return &StructType{
		Name:   s.Name,
		Fields: fields,
	}
}

func (s *StructType) Equals(other Type) bool {
	if v, ok := other.(*StructType); ok {
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

func (s *StructType) CanAssignTo(other Type) bool {
	return s.Equals(other)
}

func (s *StructType) AcceptChildren(visitor Visitor) {
	for i := range s.Fields {
		visitor.VisitType(s.Fields[i].Type)
	}
}

func (s *StructType) AcceptChildrenPtr(visitor PtrVisitor) {
	for i := range s.Fields {
		visitor.VisitType(&s.Fields[i].Type)
	}
}

func (s *StructType) String() string {
	return s.Name
}

func (s *StructType) GetField(name string) (int, *Field) {
	for i := range s.Fields {
		if s.Fields[i].Name == name {
			return i, &s.Fields[i]
		}
	}

	return 0, nil
}
