package types

import "fmt"

type Type interface {
	fmt.Stringer

	Size() int

	CanAssignTo(other Type) bool

	AcceptTypes(visitor Visitor)
}

type Visitor interface {
	VisitType(type_ *Type)
}
