package llvm

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
	SetLocation(location Location)
}

type AlignedInstruction interface {
	Instruction

	SetAlign(align int)
}

type InstructionValue interface {
	NameableValue
	Instruction
}

type AlignedInstructionValue interface {
	InstructionValue
	AlignedInstruction
}
