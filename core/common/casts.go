package common

import (
	"fireball/core/abi"
	"fireball/core/ast"
)

type CastKind uint8

const (
	None CastKind = iota

	Truncate
	Extend

	Int2Float
	Float2Int

	Pointer2Interface
)

func GetCast(from, to ast.Type) (CastKind, bool) {
	if from.Equals(to) {
		return None, true
	}

	switch from := from.Resolved().(type) {
	// Primitive -> ...
	case *ast.Primitive:
		switch to := to.Resolved().(type) {
		// Primitive -> Primitive
		case *ast.Primitive:
			if (ast.IsInteger(from.Kind) && ast.IsInteger(to.Kind)) || (ast.IsFloating(from.Kind) && ast.IsFloating(to.Kind)) {
				fromSize := abi.GetTargetAbi().Size(from)
				toSize := abi.GetTargetAbi().Size(to)

				if toSize > fromSize {
					return Extend, true
				} else if toSize < fromSize {
					return Truncate, true
				} else {
					return None, true
				}
			} else if ast.IsInteger(from.Kind) && ast.IsFloating(to.Kind) {
				return Int2Float, true
			} else if ast.IsFloating(from.Kind) && ast.IsInteger(to.Kind) {
				return Float2Int, true
			}

		// Primitive (Integer) -> Enum
		case *ast.Enum:
			if ast.IsInteger(from.Kind) {
				fromSize := abi.GetTargetAbi().Size(from)
				toSize := abi.GetTargetAbi().Size(to)

				if toSize > fromSize {
					return Extend, true
				} else if toSize < fromSize {
					return Truncate, true
				} else {
					return None, true
				}
			}
		}

	// Pointer -> Interface, Pointer, Func
	case *ast.Pointer:
		switch to := to.Resolved().(type) {
		case *ast.Interface:
			if implements(from, to) {
				return Pointer2Interface, true
			}

		case *ast.Pointer, *ast.Func:
			return None, true
		}

	// Enum -> Primitive (Integer)
	case *ast.Enum:
		if to, ok := ast.As[*ast.Primitive](to); ok && ast.IsInteger(to.Kind) {
			fromSize := abi.GetTargetAbi().Size(from)
			toSize := abi.GetTargetAbi().Size(to)

			if toSize > fromSize {
				return Extend, true
			} else if toSize < fromSize {
				return Truncate, true
			} else {
				return None, true
			}
		}
	}

	return None, false
}

func GetImplicitCast(from, to ast.Type) (CastKind, bool) {
	if from.Equals(to) {
		return None, true
	}

	switch from := from.Resolved().(type) {
	// Primitive -> Primitive
	case *ast.Primitive:
		if to, ok := ast.As[*ast.Primitive](to); ok {
			fromSize := abi.GetTargetAbi().Size(from)
			toSize := abi.GetTargetAbi().Size(to)

			// Primitive (smaller integer / floating) -> Primitive (bigger integer / floating)
			if ((ast.IsInteger(from.Kind) && ast.IsInteger(to.Kind)) || (ast.IsFloating(from.Kind) && ast.IsFloating(to.Kind))) && toSize > fromSize {
				return Extend, true
			}

			// Primitive (smaller or equal integer) -> Primitive (bigger or equal floating)
			if ast.IsInteger(from.Kind) && ast.IsFloating(to.Kind) && toSize >= fromSize {
				return Int2Float, true
			}
		}

	// Pointer -> Interface, Pointer (*void)
	case *ast.Pointer:
		switch to := to.Resolved().(type) {
		case *ast.Interface:
			if implements(from, to) {
				return Pointer2Interface, true
			}

		case *ast.Pointer:
			if ast.IsPrimitive(to.Pointee, ast.Void) {
				return None, true
			}
		}
	}

	return None, false
}

func implements(type_ ast.Type, inter *ast.Interface) bool {
	// Check struct pointee
	if pointer, ok := type_.(*ast.Pointer); ok {
		if s, ok := ast.As[*ast.Struct](pointer.Pointee); ok {
			type_ = s
		} else {
			return false
		}
	} else {
		return false
	}

	// Get resolver
	resolver := ast.GetParent[*ast.File](type_).Resolver

	// Check impl
	if resolver.GetImpl(type_, inter) != nil {
		return true
	}

	return false
}
