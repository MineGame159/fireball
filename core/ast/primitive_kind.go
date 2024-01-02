package ast

import "math"

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
