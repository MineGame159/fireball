package llvm

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

func (i *instruction) SetLocation(location Location) {
	i.location = i.module.addMetadata(Metadata{
		Type: "DILocation",
		Fields: []MetadataField{
			{
				Name:  "scope",
				Value: refMetadataValue(i.module.getScope()),
			},
			{
				Name:  "line",
				Value: numberMetadataValue(location.Line()),
			},
			{
				Name:  "column",
				Value: numberMetadataValue(location.Column()),
			},
		},
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
	align int
}

func (a *alloca) SetAlign(align int) {
	a.align = align
}

type load struct {
	instruction
	pointer Value
	align   int
}

func (l *load) SetAlign(align int) {
	l.align = align
}

type store struct {
	instruction
	pointer Value
	value   Value
	align   int
}

func (s *store) SetAlign(align int) {
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
