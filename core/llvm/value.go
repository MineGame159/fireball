package llvm

import (
	"fireball/core/cst"
)

type ValueKind uint8

const (
	GlobalValue ValueKind = iota
	LocalValue
	LiteralValue
)

type Value interface {
	Kind() ValueKind
	Type() Type
	Name() string
}

type NameableValue interface {
	Value

	SetName(name string)
}

type Instruction interface {
	SetLocation(node *cst.Node)
}

type AlignedInstruction interface {
	Instruction

	SetAlign(align uint32)
}

type InstructionValue interface {
	NameableValue
	Instruction
}

type AlignedInstructionValue interface {
	InstructionValue
	AlignedInstruction
}
