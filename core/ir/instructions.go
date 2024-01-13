package ir

type Inst interface {
	NamedValue
	MetaValue
}

type baseInst struct {
	name string
	meta MetaID
}

func (b *baseInst) Name() string {
	return b.name
}

func (b *baseInst) SetName(name string) {
	b.name = name
}

func (b *baseInst) Meta() MetaID {
	return b.meta
}

func (b *baseInst) SetMeta(id MetaID) {
	b.meta = id
}

// Ret

type RetInst struct {
	baseInst

	Value Value
}

func (r *RetInst) Type() Type {
	return nil
}

// Br

type BrInst struct {
	baseInst

	Condition Value
	True      Value
	False     Value
}

func (b *BrInst) Type() Type {
	return nil
}

// FNeg

type FNegInst struct {
	baseInst

	Value Value
}

func (f *FNegInst) Type() Type {
	return f.Value.Type()
}

// Add

type AddInst struct {
	baseInst

	Left  Value
	Right Value
}

func (a *AddInst) Type() Type {
	return a.Left.Type()
}

// Sub

type SubInst struct {
	baseInst

	Left  Value
	Right Value
}

func (s *SubInst) Type() Type {
	return s.Left.Type()
}

// Mul

type MulInst struct {
	baseInst

	Left  Value
	Right Value
}

func (m *MulInst) Type() Type {
	return m.Left.Type()
}

// IDiv

type IDivInst struct {
	baseInst

	Signed bool

	Left  Value
	Right Value
}

func (i *IDivInst) Type() Type {
	return i.Left.Type()
}

// FDiv

type FDivInst struct {
	baseInst

	Left  Value
	Right Value
}

func (f *FDivInst) Type() Type {
	return f.Left.Type()
}

// IRem

type IRemInst struct {
	baseInst

	Signed bool

	Left  Value
	Right Value
}

func (i *IRemInst) Type() Type {
	return i.Left.Type()
}

// FRem

type FRemInst struct {
	baseInst

	Left  Value
	Right Value
}

func (f *FRemInst) Type() Type {
	return f.Left.Type()
}

// Shl

type ShlInst struct {
	baseInst

	Left  Value
	Right Value
}

func (s *ShlInst) Type() Type {
	return s.Left.Type()
}

// Shr

type ShrInst struct {
	baseInst

	SignExtend bool

	Left  Value
	Right Value
}

func (s *ShrInst) Type() Type {
	return s.Left.Type()
}

// And

type AndInst struct {
	baseInst

	Left  Value
	Right Value
}

func (s *AndInst) Type() Type {
	return s.Left.Type()
}

// Or

type OrInst struct {
	baseInst

	Left  Value
	Right Value
}

func (s *OrInst) Type() Type {
	return s.Left.Type()
}

// Xor

type XorInst struct {
	baseInst

	Left  Value
	Right Value
}

func (s *XorInst) Type() Type {
	return s.Left.Type()
}

// Extract Value

type ExtractValueInst struct {
	baseInst

	Value   Value
	Indices []uint32
}

func (e *ExtractValueInst) Type() Type {
	return indexType(e.Value.Type(), e.Indices)
}

// Insert Value

type InsertValueInst struct {
	baseInst

	Value   Value
	Element Value
	Indices []uint32
}

func (i *InsertValueInst) Type() Type {
	return i.Value.Type()
}

// Alloca

type AllocaInst struct {
	baseInst

	Typ    Type
	TypPtr PointerType
	Align  uint32
}

func (a *AllocaInst) Type() Type {
	if a.TypPtr.Pointee == nil {
		a.TypPtr.Pointee = a.Typ
	}

	return &a.TypPtr
}

// Load

type LoadInst struct {
	baseInst

	Typ     Type
	Pointer Value
	Align   uint32
}

func (l *LoadInst) Type() Type {
	return l.Typ
}

// Store

type StoreInst struct {
	baseInst

	Pointer Value
	Value   Value
	Align   uint32
}

func (s *StoreInst) Type() Type {
	return nil
}

type GetElementPtrInst struct {
	baseInst

	PointerTyp Type
	Typ        Type

	Pointer Value
	Indices []Value

	Inbounds bool
}

func (g *GetElementPtrInst) Type() Type {
	return g.PointerTyp
}

// Trunc

type TruncInst struct {
	baseInst

	Value Value
	Typ   Type
}

func (t *TruncInst) Type() Type {
	return t.Typ
}

// Ext

type ExtInst struct {
	baseInst

	SignExtend bool

	Value Value
	Typ   Type
}

func (e *ExtInst) Type() Type {
	return e.Typ
}

// FExt

type FExtInst struct {
	baseInst

	Value Value
	Typ   Type
}

func (f *FExtInst) Type() Type {
	return f.Typ
}

// F2I

type F2IInst struct {
	baseInst

	Signed bool

	Value Value
	Typ   Type
}

func (f *F2IInst) Type() Type {
	return f.Typ
}

// I2F

type I2FInst struct {
	baseInst

	Signed bool

	Value Value
	Typ   Type
}

func (i *I2FInst) Type() Type {
	return i.Typ
}

// ICmp

type CmpKind uint8

const (
	Eq CmpKind = iota
	Ne
	Gt
	Ge
	Lt
	Le
)

type ICmpInst struct {
	baseInst

	Kind   CmpKind
	Signed bool

	Left  Value
	Right Value
}

func (i *ICmpInst) Type() Type {
	return I1
}

// FCmp

type FCmpInst struct {
	baseInst

	Kind    CmpKind
	Ordered bool

	Left  Value
	Right Value
}

func (f *FCmpInst) Type() Type {
	return I1
}

// Phi

type Incoming struct {
	Value Value
	Label Value
}

type PhiInst struct {
	baseInst

	Incs []Incoming
}

func (p *PhiInst) Type() Type {
	return p.Incs[0].Value.Type()
}

// Select

type SelectInst struct {
	baseInst

	Condition Value
	True      Value
	False     Value
}

func (s *SelectInst) Type() Type {
	return s.True.Type()
}

// Call

type CallInst struct {
	baseInst

	Callee Value
	Args   []Value
}

func (c *CallInst) Type() Type {
	return c.Callee.Type().(*FuncType).Returns
}
