package types

import (
	"fireball/core"
)

type Type interface {
	Range() core.Range
	Size() int

	WithRange(range_ core.Range) Type

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
