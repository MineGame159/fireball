package ir

type GlobalVar struct {
	name    string
	Typ     Type
	ptrType PointerType

	Value    Value
	Constant bool

	meta MetaID
}

// Value

func (g *GlobalVar) Type() Type {
	if g.ptrType.Pointee == nil {
		g.ptrType.Pointee = g.Typ
	}

	return &g.ptrType
}

func (g *GlobalVar) Name() string {
	return g.name
}

func (g *GlobalVar) SetName(name string) {
	g.name = name
}

func (g *GlobalVar) Meta() MetaID {
	return g.meta
}

func (g *GlobalVar) SetMeta(id MetaID) {
	g.meta = id
}
