package types

import (
	"fireball/core"
	"log"
	"math"
)

// PrimitiveType

type PrimitiveType struct {
	range_ core.Range
	Kind   PrimitiveKind
}

func Primitive(kind PrimitiveKind, range_ core.Range) *PrimitiveType {
	return &PrimitiveType{
		range_: range_,
		Kind:   kind,
	}
}

func (p *PrimitiveType) Range() core.Range {
	return p.range_
}

func (p *PrimitiveType) Size() int {
	switch p.Kind {
	case Void:
		return 0

	case Bool, U8, I8:
		return 1

	case U16, I16:
		return 2

	case U32, I32, F32:
		return 4

	case U64, I64, F64:
		return 8

	default:
		log.Fatalln("PrimitiveType.Size() - Invalid primitive kind")
		return -1
	}
}

func (p *PrimitiveType) Align() int {
	return p.Size()
}

func (p *PrimitiveType) WithRange(range_ core.Range) Type {
	return &PrimitiveType{
		range_: range_,
		Kind:   p.Kind,
	}
}

func (p *PrimitiveType) Equals(other Type) bool {
	return IsPrimitive(other, p.Kind)
}

func (p *PrimitiveType) CanAssignTo(other Type) bool {
	return IsPrimitive(other, p.Kind)
}

func (p *PrimitiveType) AcceptTypes(visitor Visitor) {}

func (p *PrimitiveType) AcceptTypesPtr(visitor PtrVisitor) {}

func (p *PrimitiveType) String() string {
	switch p.Kind {
	case Void:
		return "void"
	case Bool:
		return "bool"

	case U8:
		return "u8"
	case U16:
		return "u16"
	case U32:
		return "u32"
	case U64:
		return "u64"

	case I8:
		return "i8"
	case I16:
		return "i16"
	case I32:
		return "i32"
	case I64:
		return "i64"

	case F32:
		return "f32"
	case F64:
		return "f64"

	default:
		log.Fatalln("PrimitiveType.String() - Invalid primitive kind")
		return ""
	}
}

// Helpers

func IsPrimitive(type_ Type, kind PrimitiveKind) bool {
	if v, ok := type_.(*PrimitiveType); ok {
		return v.Kind == kind
	}

	return false
}

// Kind

type PrimitiveKind = uint8

const (
	Void PrimitiveKind = iota
	Bool

	U8
	U16
	U32
	U64

	I8
	I16
	I32
	I64

	F32
	F64
)

func IsNumber(kind PrimitiveKind) bool {
	return kind != Void && kind != Bool
}

func IsFloating(kind PrimitiveKind) bool {
	return kind == F32 || kind == F64
}

func IsUnsigned(kind PrimitiveKind) bool {
	return kind == U8 || kind == U16 || kind == U32 || kind == U64
}

func IsSigned(kind PrimitiveKind) bool {
	return kind == I8 || kind == I16 || kind == I32 || kind == I64
}

func IsInteger(kind PrimitiveKind) bool {
	return IsSigned(kind) || IsUnsigned(kind)
}

func GetBitSize(kind PrimitiveKind) int {
	switch kind {
	case Void:
		return 0
	case Bool:
		return 1

	case U8, I8:
		return 8
	case U16, I16:
		return 16
	case U32, I32, F32:
		return 32
	case U64, I64, F64:
		return 64

	default:
		panic("types GetBitSize() - Invalid type")
	}
}

func EqualsPrimitiveCategory(a, b PrimitiveKind) bool {
	return (IsInteger(a) && IsInteger(b)) || (IsFloating(a) && IsFloating(b))
}

func GetUnsignedRange(kind PrimitiveKind) (min, max uint64) {
	switch kind {
	case U8:
		return 0, math.MaxUint8
	case U16:
		return 0, math.MaxUint16
	case U32:
		return 0, math.MaxUint32
	case U64:
		return 0, math.MaxUint64

	default:
		return 0, 0
	}
}

func GetSignedRange(kind PrimitiveKind) (min, max int64) {
	switch kind {
	case I8:
		return math.MinInt8, math.MaxInt8
	case I16:
		return math.MinInt16, math.MaxInt16
	case I32:
		return math.MinInt32, math.MaxInt32
	case I64:
		return math.MinInt64, math.MaxInt64

	default:
		return 0, 0
	}
}

func GetRangeTrunc(kind PrimitiveKind) (min, max int64) {
	switch kind {
	case U8:
		return 0, math.MaxUint8
	case U16:
		return 0, math.MaxUint16
	case U32:
		return 0, math.MaxUint32
	case U64:
		return 0, math.MaxInt64

	case I8:
		return math.MinInt8, math.MaxInt8
	case I16:
		return math.MinInt16, math.MaxInt16
	case I32:
		return math.MinInt32, math.MaxInt32
	case I64:
		return math.MinInt64, math.MaxInt64

	default:
		return 0, 0
	}
}
