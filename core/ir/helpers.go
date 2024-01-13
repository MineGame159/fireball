package ir

type Int struct {
	Negative bool
	Value    uint64
}

func Signed(value int64) Int {
	var v uint64

	if value < 0 {
		v = uint64(-value)
	} else {
		v = uint64(value)
	}

	return Int{
		Negative: value < 0,
		Value:    v,
	}
}

func Unsigned(value uint64) Int {
	return Int{Value: value}
}

func indexType(typ Type, indices []uint32) Type {
	if len(indices) == 0 {
		return typ
	}

	switch typ := typ.(type) {
	case *PointerType:
		return indexType(typ.Pointee, indices[1:])
	case *ArrayType:
		return indexType(typ.Base, indices[1:])
	case *StructType:
		return indexType(typ.Fields[indices[0]], indices[1:])

	default:
		return nil
	}
}
