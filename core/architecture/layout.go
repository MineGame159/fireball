package architecture

import (
	"fireball/core/types"
)

type Layout interface {
	Add(type_ types.Type) (offset, size int)
}

type CLayout struct {
	biggestAlign int
	Offset       int
}

func (l *CLayout) Add(type_ types.Type) (offset, size int) {
	size = type_.Size()
	align := type_.Align()

	// Align member
	l.biggestAlign = max(l.biggestAlign, align)

	offset = l.Offset

	if offset%align != 0 {
		offset += align - (offset % align)
	}

	l.Offset = offset + size

	// Calculate total size
	size = l.Offset

	if size%l.biggestAlign != 0 {
		size += l.biggestAlign - (size % l.biggestAlign)
	}

	return offset, size
}
