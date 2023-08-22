package types

import "fmt"

type Type interface {
	fmt.Stringer

	Size() int

	CanAssignTo(other Type) bool
}
