package ir

type FuncFlags uint8

const (
	InlineFlag FuncFlags = 1 << iota
	HotFlag
)

type Func struct {
	name string
	Typ  *FuncType

	Flags FuncFlags
	meta  MetaID

	Blocks []*Block
}

func (f *Func) Block(name string) *Block {
	f.Blocks = append(f.Blocks, &Block{name: name})
	return f.Blocks[len(f.Blocks)-1]
}

// Value

func (f *Func) Type() Type {
	return f.Typ
}

func (f *Func) Name() string {
	return f.name
}

func (f *Func) SetName(name string) {
	f.name = name
}

func (f *Func) Meta() MetaID {
	return f.meta
}

func (f *Func) SetMeta(id MetaID) {
	f.meta = id
}
