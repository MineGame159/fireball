package ast

type CastKind uint8

const (
	None CastKind = iota

	Truncate
	Extend

	Int2Float
	Float2Int
)

func GetCast(from, to Type) (CastKind, bool) {
	if from.Equals(to) {
		return None, true
	}

	switch from := from.Resolved().(type) {
	// Primitive -> ...
	case *Primitive:
		switch to := to.Resolved().(type) {
		// Primitive -> Primitive
		case *Primitive:
			if (IsInteger(from.Kind) && IsInteger(to.Kind)) || (IsFloating(from.Kind) && IsFloating(to.Kind)) {
				if to.Size() > from.Size() {
					return Extend, true
				} else if to.Size() < from.Size() {
					return Truncate, true
				} else {
					return None, true
				}
			} else if IsInteger(from.Kind) && IsFloating(to.Kind) {
				return Int2Float, true
			} else if IsFloating(from.Kind) && IsInteger(to.Kind) {
				return Float2Int, true
			}

		// Primitive (Integer) -> Enum
		case *Enum:
			if IsInteger(from.Kind) {
				if to.Size() > from.Size() {
					return Extend, true
				} else if to.Size() < from.Size() {
					return Truncate, true
				} else {
					return None, true
				}
			}
		}

	// Pointer -> Pointer, Func
	case *Pointer:
		switch to.Resolved().(type) {
		case *Pointer, *Func:
			return None, true
		}

	// Enum -> Primitive (Integer)
	case *Enum:
		if to, ok := As[*Primitive](to); ok && IsInteger(to.Kind) {
			if to.Size() > from.Size() {
				return Extend, true
			} else if to.Size() < from.Size() {
				return Truncate, true
			} else {
				return None, true
			}
		}
	}

	return None, false
}

func GetImplicitCast(from, to Type) (CastKind, bool) {
	if from.Equals(to) {
		return None, true
	}

	switch from := from.Resolved().(type) {
	// Primitive -> Primitive
	case *Primitive:
		if to, ok := As[*Primitive](to); ok {
			// Primitive (smaller integer / floating) -> Primitive (bigger integer / floating)
			if ((IsInteger(from.Kind) && IsInteger(to.Kind)) || (IsFloating(from.Kind) && IsFloating(to.Kind))) && to.Size() > from.Size() {
				return Extend, true
			}

			// Primitive (same integer) -> Primitive (same floating)
			// TODO: Allow converting a smaller integer to the next bigger floating (eg. i16 -> f32)
			if IsInteger(from.Kind) && IsFloating(to.Kind) && to.Size() == from.Size() {
				return Int2Float, true
			}
		}

	// Pointer -> Pointer (*void)
	case *Pointer:
		if to, ok := As[*Pointer](to); ok && IsPrimitive(to.Pointee, Void) {
			return None, true
		}
	}

	return None, false
}
