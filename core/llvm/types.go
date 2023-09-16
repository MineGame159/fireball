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

	Size() int
}

type voidType struct {
}

func (v *voidType) isType() {}

func (v *voidType) Size() int {
	return 0
}

type primitiveType struct {
	name     string
	bitSize  int
	encoding Encoding
}

func (v *primitiveType) isType() {}

func (v *primitiveType) Size() int {
	return v.bitSize
}

type arrayType struct {
	name  string
	count int
	base  Type
}

func (v *arrayType) isType() {}

func (v *arrayType) Size() int {
	return v.count * v.base.Size()
}

type pointerType struct {
	name    string
	pointee Type
}

func (v *pointerType) isType() {}

func (v *pointerType) Size() int {
	return 64
}

type functionType struct {
	name       string
	parameters []Type
	variadic   bool
	returns    Type
}

func (v *functionType) isType() {}

func (v *functionType) Size() int {
	return 64
}

type Field struct {
	Name string
	Type Type
}

type structType struct {
	name   string
	fields []Field
}

func (v *structType) isType() {}

func (v *structType) Size() int {
	size := 0

	for _, field := range v.fields {
		size += field.Type.Size()
	}

	return size
}

type aliasType struct {
	name       string
	underlying Type
}

func (v *aliasType) isType() {}

func (v *aliasType) Size() int {
	return v.underlying.Size()
}

// Utils

func isSigned(type_ Type) bool {
	if v, ok := type_.(*primitiveType); ok {
		return v.encoding == SignedEncoding
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
