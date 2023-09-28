package architecture

import (
	"fireball/core/types"
)

type Layout interface {
	Add(type_ types.Type) (offset, size int)
	Size() int
}

type CLayout struct {
	biggestAlign int
	offset       int
}

func (l *CLayout) Add(type_ types.Type) int {
	l.biggestAlign = max(l.biggestAlign, type_.Align())

	offset := align(l.offset, l.biggestAlign)
	l.offset = offset + type_.Size()

	return offset
}

func (l *CLayout) Size() int {
	if l.offset == 0 {
		return 0
	}

	return align(l.offset, l.biggestAlign)
}

func align(value, align int) int {
	if value%align != 0 {
		value += align - (value % align)
	}

	return value
}
