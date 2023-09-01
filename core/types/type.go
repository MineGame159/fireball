package types

import (
	"fireball/core"
	"fmt"
)

type Type interface {
	Range() core.Range
	Size() int
	Copy() Type

	Equals(other Type) bool
	CanAssignTo(other Type) bool

	AcceptChildren(visitor Visitor)
	AcceptChildrenPtr(visitor PtrVisitor)

	fmt.Stringer
}

type Visitor interface {
	VisitType(type_ Type)
}

type PtrVisitor interface {
	VisitType(type_ *Type)
}
