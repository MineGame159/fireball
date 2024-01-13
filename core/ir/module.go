package ir

type Module struct {
	Path string

	Structs []*StructType

	Globals   []*GlobalVar
	Functions []*Func

	NamedMetadata map[string]GroupMeta
	Metadata      []Meta

	metaIndex      uint32
	expressionMeta MetaID
}

func (m *Module) Struct(s *StructType) {
	m.Structs = append(m.Structs, s)
}

func (m *Module) Constant(name string, value Value) *GlobalVar {
	m.Globals = append(m.Globals, &GlobalVar{
		name:     name,
		Typ:      value.Type(),
		ptrType:  PointerType{Pointee: value.Type()},
		Value:    value,
		Constant: true,
	})

	return m.Globals[len(m.Globals)-1]
}

func (m *Module) Global(name string, typ Type, value Value) *GlobalVar {
	m.Globals = append(m.Globals, &GlobalVar{
		name:  name,
		Typ:   typ,
		Value: value,
	})

	return m.Globals[len(m.Globals)-1]
}

func (m *Module) Define(name string, typ *FuncType, flags FuncFlags) *Func {
	m.Functions = append(m.Functions, &Func{
		name:  name,
		Typ:   typ,
		Flags: flags,
	})

	return m.Functions[len(m.Functions)-1]
}

func (m *Module) Declare(name string, typ *FuncType) *Func {
	m.Functions = append(m.Functions, &Func{
		name: name,
		Typ:  typ,
	})

	return m.Functions[len(m.Functions)-1]
}

func (m *Module) NamedMeta(name string, metadata []MetaID) {
	if m.NamedMetadata == nil {
		m.NamedMetadata = make(map[string]GroupMeta)
	}

	m.NamedMetadata[name] = GroupMeta{Metadata: metadata}
}

func (m *Module) Meta(meta Meta) MetaID {
	if _, ok := meta.(*ExpressionMeta); ok {
		if !m.expressionMeta.Valid() {
			meta.(metaImpl).setIndex(m.metaIndex)
			m.metaIndex++

			m.Metadata = append(m.Metadata, meta)
			m.expressionMeta = MetaID(len(m.Metadata))
		}

		return m.expressionMeta
	}

	if !IsMetaInline(meta) {
		meta.(metaImpl).setIndex(m.metaIndex)
		m.metaIndex++
	}

	m.Metadata = append(m.Metadata, meta)
	return MetaID(len(m.Metadata))
}
