package types

import "fmt"

type ArrayType struct {
	Count uint32
	Base  Type
}

func (a ArrayType) Size() int {
	return int(a.Count) * a.Base.Size()
}

func (a ArrayType) CanAssignTo(other Type) bool {
	if v, ok := other.(ArrayType); ok {
		return a.Count == v.Count && a.Base.CanAssignTo(v.Base)
	}

	return false
}

func (a ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", a.Count, a.Base)
}
