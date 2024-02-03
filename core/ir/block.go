package ir

import "slices"

type Block struct {
	name string

	Instructions []Inst
}

func (b *Block) Add(inst Inst) Inst {
	b.Instructions = append(b.Instructions, inst)
	return inst
}

func (b *Block) Insert(index int, inst Inst) Inst {
	if len(b.Instructions) == 0 {
		b.Instructions = append(b.Instructions, inst)
		return inst
	}

	b.Instructions = slices.Insert(b.Instructions, index, inst)
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
