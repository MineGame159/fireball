package types

type PointerType struct {
	Pointee Type
}

func (p *PointerType) Size() int {
	return 4
}

func (p *PointerType) CanAssignTo(other Type) bool {
	if v, ok := other.(*PointerType); ok {
		return p.Pointee.CanAssignTo(v.Pointee)
	}

	return false
}

func (p *PointerType) AcceptTypes(visitor Visitor) {
	visitor.VisitType(&p.Pointee)
}

func (p *PointerType) String() string {
	return "*" + p.Pointee.String()
}
