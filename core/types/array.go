package types

import (
	"fireball/core"
	"fmt"
)

type ArrayType struct {
	range_ core.Range

	Count uint32
	Base  Type
}

func Array(count uint32, base Type, range_ core.Range) *ArrayType {
	return &ArrayType{
		range_: range_,
		Count:  count,
		Base:   base,
	}
}

func (a *ArrayType) Range() core.Range {
	return a.range_
}

func (a *ArrayType) Size() int {
	return int(a.Count) * a.Base.Size()
}

func (a *ArrayType) WithoutRange() Type {
	return &ArrayType{
		Count: a.Count,
		Base:  a.Base.WithoutRange(),
	}
}

func (a *ArrayType) Equals(other Type) bool {
	if v, ok := other.(*ArrayType); ok {
		return a.Count == v.Count && a.Base.Equals(v.Base)
	}

	return false
}

func (a *ArrayType) CanAssignTo(other Type) bool {
	if v, ok := other.(*ArrayType); ok {
		return a.Count == v.Count && a.Base.CanAssignTo(v.Base)
	}

	return false
}

func (a *ArrayType) AcceptTypes(visitor Visitor) {
	visitor.VisitType(a.Base)
}

func (a *ArrayType) AcceptTypesPtr(visitor PtrVisitor) {
	visitor.VisitType(&a.Base)
}

func (a *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", a.Count, a.Base)
}
