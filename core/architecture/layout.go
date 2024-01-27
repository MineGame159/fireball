package architecture

type Layout interface {
	Add(size, align uint32) uint32
	Size() uint32
}

type CLayout struct {
	biggestAlign uint32
	offset       uint32
}

func (l *CLayout) Add(size, align uint32) uint32 {
	l.biggestAlign = max(l.biggestAlign, align)

	offset := alignValue(l.offset, align)
	l.offset = offset + size

	return offset
}

func (l *CLayout) Size() uint32 {
	if l.offset == 0 {
		return 0
	}

	return alignValue(l.offset, l.biggestAlign)
}

func alignValue(value, align uint32) uint32 {
	if value%align != 0 {
		value += align - (value % align)
	}

	return value
}
