package ast

type CastKind uint8

const (
	None CastKind = iota

	Truncate
	Extend

	Int2Float
	Float2Int

	Pointer2Interface
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

	// Pointer -> Interface, Pointer, Func
	case *Pointer:
		switch to := to.Resolved().(type) {
		case *Interface:
			if implements(from, to) {
				return Pointer2Interface, true
			}

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

	// Pointer -> Interface, Pointer (*void)
	case *Pointer:
		switch to := to.Resolved().(type) {
		case *Interface:
			if implements(from, to) {
				return Pointer2Interface, true
			}

		case *Pointer:
			if IsPrimitive(to.Pointee, Void) {
				return None, true
			}
		}
	}

	return None, false
}

func implements(type_ Type, inter *Interface) bool {
	// Check struct pointee
	if pointer, ok := type_.(*Pointer); ok {
		if s, ok := As[*Struct](pointer.Pointee); ok {
			type_ = s
		} else {
			return false
		}
	} else {
		return false
	}

	// Get resolver
	resolver := GetParent[*File](type_).Resolver

	// Check impl
	if resolver.GetImpl(type_, inter) != nil {
		return true
	}

	return false
}
