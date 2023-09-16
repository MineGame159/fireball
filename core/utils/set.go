package utils

type Set[T comparable] struct {
	Data map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{
		Data: make(map[T]struct{}),
	}
}

func (s Set[T]) Add(value T) bool {
	if s.Contains(value) {
		return false
	}

	s.Data[value] = struct{}{}
	return true
}

func (s Set[T]) Contains(value T) bool {
	_, contains := s.Data[value]
	return contains
}
