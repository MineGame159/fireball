package ir

type Value interface {
	Type() Type
	Name() string
}

type NamedValue interface {
	Value

	SetName(name string)
}

type MetaValue interface {
	Value

	Meta() MetaID
	SetMeta(id MetaID)
}
