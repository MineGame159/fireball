package types

import (
	"fireball/core"
)

type PointerType struct {
	range_  core.Range
	Pointee Type
}

func Pointer(pointee Type, range_ core.Range) *PointerType {
	return &PointerType{
		range_:  range_,
		Pointee: pointee,
	}
}

func (p *PointerType) Range() core.Range {
	return p.range_
}

func (p *PointerType) Size() int {
	return 4
}

func (p *PointerType) WithRange(range_ core.Range) Type {
	return &PointerType{
		range_:  range_,
		Pointee: p.Pointee.WithRange(core.Range{}),
	}
}

func (p *PointerType) Equals(other Type) bool {
	if v, ok := other.(*PointerType); ok {
		return p.Pointee.Equals(v.Pointee)
	}

	return false
}

func (p *PointerType) CanAssignTo(other Type) bool {
	if v, ok := other.(*PointerType); ok {
		return IsPrimitive(v.Pointee, Void) || p.Pointee.CanAssignTo(v.Pointee)
	}

	return false
}

func (p *PointerType) AcceptTypes(visitor Visitor) {
	visitor.VisitType(p.Pointee)
}

func (p *PointerType) AcceptTypesPtr(visitor PtrVisitor) {
	visitor.VisitType(&p.Pointee)
}

func (p *PointerType) String() string {
	return "*" + p.Pointee.String()
}
