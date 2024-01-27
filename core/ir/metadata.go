package ir

// ID

type MetaID uint16

func (m MetaID) Valid() bool {
	return m > 0
}

func (m MetaID) Index() uint32 {
	return uint32(m) - 1
}

func (m MetaID) Name() string {
	return ""
}

func (m MetaID) Type() Type {
	return MetaT
}

// Meta

type Meta interface {
	Index() uint32

	isMeta()
}

type metaImpl interface {
	Meta

	setIndex(i uint32)
}

type baseMeta struct {
	index uint32
}

func (b *baseMeta) Index() uint32 {
	return b.index
}

func (b *baseMeta) setIndex(i uint32) {
	b.index = i
}

func (b *baseMeta) isMeta() {}

// Inline

type IntMeta struct {
	baseMeta

	Value int32
}

type StringMeta struct {
	baseMeta

	Value string
}

func IsMetaInline(meta Meta) bool {
	switch meta.(type) {
	case *IntMeta, *StringMeta:
		return true
	default:
		return false
	}
}

// Group

type GroupMeta struct {
	baseMeta

	Metadata []MetaID
}

// Compile Unit

type EmissionKind uint8

const (
	NoDebug EmissionKind = iota
	FullDebug
	LineTablesOnly
	DebugDirectivesOnly
)

func (e EmissionKind) String() string {
	switch e {
	case NoDebug:
		return "NoDebug"
	case FullDebug:
		return "FullDebug"
	case LineTablesOnly:
		return "LineTablesOnly"
	case DebugDirectivesOnly:
		return "DebugDirectivesOnly"

	default:
		panic("llvm.EmissionKind.String() - Not implemented")
	}
}

type NameTableKind uint8

const (
	Default NameTableKind = iota
	GNU
	None
	Apple
)

func (n NameTableKind) String() string {
	switch n {
	case Default:
		return "Default"
	case GNU:
		return "GNU"
	case None:
		return "None"
	case Apple:
		return "Apple"

	default:
		panic("llvm.NameTableKind.String() - Not implemented")
	}
}

type CompileUnitMeta struct {
	baseMeta

	File     MetaID
	Producer string

	Emission  EmissionKind
	NameTable NameTableKind

	Globals []MetaID
}

// File

type FileMeta struct {
	baseMeta

	Path string
}

// Basic Type

type EncodingKind uint8

const (
	AddressEncoding      EncodingKind = 1
	BooleanEncoding                   = 2
	FloatEncoding                     = 4
	SignedEncoding                    = 5
	SignedCharEncoding                = 6
	UnsignedEncoding                  = 7
	UnsignedCharEncoding              = 8
)

func (e EncodingKind) String() string {
	switch e {
	case AddressEncoding:
		return "DW_ATE_address"
	case BooleanEncoding:
		return "DW_ATE_boolean"
	case FloatEncoding:
		return "DW_ATE_float"
	case SignedEncoding:
		return "DW_ATE_signed"
	case SignedCharEncoding:
		return "DW_ATE_signed_char"
	case UnsignedEncoding:
		return "DW_ATE_unsigned"
	case UnsignedCharEncoding:
		return "DW_ATE_unsigned_char"

	default:
		panic("llvm.EncodingKind.String() - Not implemented")
	}
}

type BasicTypeMeta struct {
	baseMeta

	Name     string
	Encoding EncodingKind

	Size  uint32
	Align uint32
}

// Subroutine Type

type SubroutineTypeMeta struct {
	baseMeta

	Returns MetaID
	Params  []MetaID
}

// Derived Type

type DerivedTagKind uint8

const (
	MemberTag          DerivedTagKind = 13
	PointerTypeTag                    = 15
	ReferenceTypeTag                  = 16
	TypedefTag                        = 22
	InheritanceTag                    = 28
	PtrToMemberTypeTag                = 31
	ConstTypeTag                      = 38
	FriendTag                         = 42
	VolatileTypeTag                   = 53
	RestrictTypeTag                   = 55
	AtomicTypeTag                     = 71
	ImmutableTypeTag                  = 75
)

func (d DerivedTagKind) String() string {
	switch d {
	case MemberTag:
		return "DW_TAG_member"
	case PointerTypeTag:
		return "DW_TAG_pointer_type"
	case ReferenceTypeTag:
		return "DW_TAG_reference_type"
	case TypedefTag:
		return "DW_TAG_typedef"
	case InheritanceTag:
		return "DW_TAG_inheritance"
	case PtrToMemberTypeTag:
		return "DW_TAG_ptr_to_member_type"
	case ConstTypeTag:
		return "DW_TAG_const_type"
	case FriendTag:
		return "DW_TAG_friend"
	case VolatileTypeTag:
		return "DW_TAG_volatile_type"
	case RestrictTypeTag:
		return "DW_TAG_restrict_Type"
	case AtomicTypeTag:
		return "DW_TAG_atomic_type"
	case ImmutableTypeTag:
		return "DW_TAG_immutable_Type"

	default:
		panic("llvm.DerivedTagKind.String() - Not implemented")
	}
}

type DerivedTypeMeta struct {
	baseMeta

	Tag      DerivedTagKind
	Name     string
	BaseType MetaID

	Size  uint32
	Align uint32

	Offset uint32
}

// Composite Type

type CompositeTagKind uint8

const (
	ArrayTypeTag       CompositeTagKind = 1
	ClassTypeTag                        = 2
	EnumerationTypeTag                  = 4
	StructureTypeTag                    = 19
	UnionTypeTag                        = 23
)

func (c CompositeTagKind) String() string {
	switch c {
	case ArrayTypeTag:
		return "DW_TAG_array_type"
	case ClassTypeTag:
		return "DW_TAG_class_type"
	case EnumerationTypeTag:
		return "DW_TAG_enumeration_type"
	case StructureTypeTag:
		return "DW_TAG_structure_type"
	case UnionTypeTag:
		return "DW_TAG_union_type"

	default:
		panic("llvm.CompositeTagKind.String() - Not implemented")
	}
}

type CompositeTypeMeta struct {
	baseMeta

	Tag  CompositeTagKind
	Name string

	File MetaID
	Line uint32

	Size  uint32
	Align uint32

	BaseType MetaID
	Elements []MetaID
}

// Subrange

type SubrangeMeta struct {
	baseMeta

	LowerBound uint32
	Count      uint32
}

// Enumerator

type EnumeratorMeta struct {
	baseMeta

	Name  string
	Value Int
}

// Namespace

type NamespaceMeta struct {
	baseMeta

	Name string

	Scope MetaID
	File  MetaID
	Line  uint32
}

// Global Var

type GlobalVarMeta struct {
	baseMeta

	Name        string
	LinkageName string

	Scope MetaID
	File  MetaID
	Line  uint32

	Type MetaID

	Local      bool
	Definition bool
}

// GlobalVarExpr

type GlobalVarExpr struct {
	baseMeta

	Var MetaID
}

// Subprogram

type SubprogamMeta struct {
	baseMeta

	Name        string
	LinkageName string

	Scope MetaID
	File  MetaID
	Line  uint32

	Type MetaID

	Unit MetaID
}

// Lexical Block

type LexicalBlockMeta struct {
	baseMeta

	Scope MetaID
	File  MetaID
	Line  uint32
}

// Location

type LocationMeta struct {
	baseMeta

	Scope  MetaID
	Line   uint32
	Column uint32
}

// Local Var

type LocalVarMeta struct {
	baseMeta

	Name string
	Type MetaID

	Arg uint32

	Scope MetaID
	File  MetaID
	Line  uint32
}

// Expression

type ExpressionMeta struct {
	baseMeta
}
