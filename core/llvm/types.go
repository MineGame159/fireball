package llvm

type Encoding string

const (
	BooleanEncoding  Encoding = "DW_ATE_boolean"
	FloatEncoding    Encoding = "DW_ATE_float"
	SignedEncoding   Encoding = "DW_ATE_signed"
	UnsignedEncoding Encoding = "DW_ATE_unsigned"
)

type Type interface {
	isType()

	Size() uint32
}

type voidType struct {
}

func (v *voidType) isType() {}

func (v *voidType) Size() uint32 {
	return 0
}

type primitiveType struct {
	name     string
	bitSize  uint32
	encoding Encoding
}

func (v *primitiveType) isType() {}

func (v *primitiveType) Size() uint32 {
	return v.bitSize
}

type arrayType struct {
	name  string
	count uint32
	base  Type
}

func (v *arrayType) isType() {}

func (v *arrayType) Size() uint32 {
	return v.count * v.base.Size()
}

type pointerType struct {
	name    string
	pointee Type
}

func (v *pointerType) isType() {}

func (v *pointerType) Size() uint32 {
	return 64
}

type functionType struct {
	name       string
	parameters []Type
	variadic   bool
	returns    Type
}

func (v *functionType) isType() {}

func (v *functionType) Size() uint32 {
	return 64
}

type Field struct {
	Name   string
	Type   Type
	Offset uint32
}

type structType struct {
	name   string
	size   uint32
	fields []Field
}

func (v *structType) isType() {}

func (v *structType) Size() uint32 {
	return v.size
}

type aliasType struct {
	name       string
	underlying Type
}

func (v *aliasType) isType() {}

func (v *aliasType) Size() uint32 {
	return v.underlying.Size()
}

// Utils

func isSigned(type_ Type) bool {
	if v, ok := type_.(*primitiveType); ok {
		return v.encoding == SignedEncoding || v.encoding == BooleanEncoding
	}

	return false
}

func isUnsigned(type_ Type) bool {
	if v, ok := type_.(*primitiveType); ok {
		return v.encoding == UnsignedEncoding
	}

	return false
}

func isFloating(type_ Type) bool {
	if v, ok := type_.(*primitiveType); ok {
		return v.encoding == FloatEncoding
	}

	return false
}
