package ir

type Block struct {
	name string

	Instructions []Inst
}

func (b *Block) Add(inst Inst) Inst {
	b.Instructions = append(b.Instructions, inst)
	return inst
}

// Value

func (b *Block) Type() Type {
	return nil
}

func (b *Block) Name() string {
	return b.name
}

func (b *Block) SetName(name string) {
	b.name = name
}
