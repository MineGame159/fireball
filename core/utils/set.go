package utils

type Set[T comparable] struct {
	data map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{
		data: make(map[T]struct{}),
	}
}

func (s Set[T]) Add(value T) bool {
	if s.Contains(value) {
		return false
	}

	s.data[value] = struct{}{}
	return true
}

func (s Set[T]) Contains(value T) bool {
	_, contains := s.data[value]
	return contains
}
