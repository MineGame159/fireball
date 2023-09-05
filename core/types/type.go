package types

import (
	"fireball/core"
)

type Type interface {
	Range() core.Range
	Size() int

	WithoutRange() Type

	Equals(other Type) bool
	CanAssignTo(other Type) bool

	AcceptTypes(visitor Visitor)
	AcceptTypesPtr(visitor PtrVisitor)

	String() string
}

type Visitor interface {
	VisitType(type_ Type)
}

type PtrVisitor interface {
	VisitType(type_ *Type)
}
