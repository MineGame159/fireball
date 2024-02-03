package abi

type cLayout struct {
	biggestAlign uint32
	offset       uint32
}

func (l *cLayout) add(size, align uint32) uint32 {
	l.biggestAlign = max(l.biggestAlign, align)

	offset := alignBytes(l.offset, align)
	l.offset = offset + size

	return offset
}

func (l *cLayout) size() uint32 {
	if l.offset == 0 {
		return 0
	}

	return alignBytes(l.offset, l.biggestAlign)
}
