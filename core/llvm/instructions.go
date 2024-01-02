package llvm

import (
	"fireball/core/cst"
)

type instruction struct {
	module *Module

	type_ Type
	name  string

	location int
}

func (i *instruction) Kind() ValueKind {
	return LocalValue
}

func (i *instruction) Type() Type {
	return i.type_
}

func (i *instruction) Name() string {
	return i.name
}

func (i *instruction) SetName(name string) {
	i.name = name
}

func (i *instruction) SetLocation(node *cst.Node) {
	fields := []MetadataField{
		{
			Name:  "scope",
			Value: refMetadataValue(i.module.getScope()),
		},
	}

	if node != nil {
		fields = append(fields, MetadataField{
			Name:  "line",
			Value: numberMetadataValue(int(node.Range.Start.Line)),
		})
		fields = append(fields, MetadataField{
			Name:  "column",
			Value: numberMetadataValue(int(node.Range.Start.Column)),
		})
	}

	i.location = i.module.addMetadata(Metadata{
		Type:   "DILocation",
		Fields: fields,
	})
}

type variableMetadata struct {
	instruction
	pointer  Value
	metadata int
}

type lifetimeMetadata struct {
	instruction
	pointer Value
	start   bool
}

type BinaryKind uint8

type fNeg struct {
	instruction
	value Value
}

const (
	Add BinaryKind = iota
	Sub
	Mul
	Div
	Rem

	Eq
	Ne
	Lt
	Le
	Gt
	Ge

	Or
	Xor
	And
	Shl
	Shr
)

type binary struct {
	instruction
	op    BinaryKind
	left  Value
	right Value
}

type CastKind uint8

const (
	Trunc CastKind = iota
	ZExt
	SExt
	FpTrunc
	FpExt
	FpToUi
	FpToSi
	UiToFp
	SiToFp
	PtrToInt
	IntToPtr
	Bitcast
)

type cast struct {
	instruction
	kind  CastKind
	value Value
}

type extractValue struct {
	instruction
	value Value
	index int
}

type insertValue struct {
	instruction
	value   Value
	element Value
	index   int
}

type alloca struct {
	instruction
	type_ Type
	align uint32
}

func (a *alloca) SetAlign(align uint32) {
	a.align = align
}

type load struct {
	instruction
	pointer Value
	align   uint32
}

func (l *load) SetAlign(align uint32) {
	l.align = align
}

type store struct {
	instruction
	pointer Value
	value   Value
	align   uint32
}

func (s *store) SetAlign(align uint32) {
	s.align = align
}

type getElementPtr struct {
	instruction
	type_   Type
	pointer Value
	indices []Value
}

type br struct {
	instruction
	condition Value
	true      *Block
	false     *Block
}

type phi struct {
	instruction
	firstValue  Value
	firstBlock  *Block
	secondValue Value
	secondBlock *Block
}

type call struct {
	instruction
	value     Value
	arguments []Value
}

type ret struct {
	instruction
	value Value
}
