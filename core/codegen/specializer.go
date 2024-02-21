package codegen

import "fireball/core/ast"

type directReplacement struct {
	ptr      *ast.Type
	genericI int
	original ast.Type
}

type structReplacement struct {
	s     *ast.Struct
	types []ast.Type
	ptr   *ast.Type
}

type specializer struct {
	generics []*ast.Generic

	direct  []directReplacement
	structs []structReplacement
}

func (s *specializer) prepare(node ast.Node, generics []*ast.Generic) {
	finder := specializationFinder{generics: generics}
	finder.VisitNode(node)

	s.generics = generics
	s.direct = finder.direct
	s.structs = finder.structs
}

func (s *specializer) specialize(types []ast.Type) {
	// Direct
	for _, rep := range s.direct {
		*rep.ptr = types[rep.genericI].Resolved()
	}

	// Structs
	for _, rep := range s.structs {
		// Create new specialized types slice
		specTypes := make([]ast.Type, len(rep.types))
		copy(specTypes, rep.types)

		// Replace generics
		for typeI, type_ := range specTypes {
			for genericI, generic := range s.generics {
				if generic.Equals(type_) {
					specTypes[typeI] = types[genericI]
					break
				}
			}
		}

		// Specialize
		*rep.ptr = rep.s.Specialize(specTypes)
	}
}

func (s *specializer) finish() {
	// Direct
	for _, rep := range s.direct {
		*rep.ptr = rep.original
	}

	// Structs
	for _, rep := range s.structs {
		*rep.ptr = nil
	}
}

// Finder

type specializationFinder struct {
	generics []*ast.Generic

	direct  []directReplacement
	structs []structReplacement
}

func (s *specializationFinder) VisitNode(node ast.Node) {
	if ast.IsNil(node) {
		return
	}

	switch node := node.(type) {
	case *ast.Resolvable:
		if i := s.getGenericI(node.Type); i != -1 {
			s.addDirect(&node.Type, i)
			return
		}

		if struct_, ok := node.Type.(*ast.Struct); ok {
			if s.visitStruct(struct_) {
				return
			}
		}

		if spec, ok := node.Type.(*ast.SpecializedStruct); ok {
			if s.visitSpecStruct(spec) {
				return
			}
		}

		s.VisitNode(node.Type)

	case *ast.Generic:
		if i := s.getGenericI(node); i != -1 {
			s.addDirect(&node.Type, i)
			return
		}

		s.VisitNode(node.Type)

	case *ast.Struct:
		s.visitStruct(node)

	case *ast.SpecializedStruct:
		s.visitSpecStruct(node)

	case ast.Expr:
		s.VisitNode(node.Result().Type)

		node.AcceptChildren(s)

	default:
		node.AcceptChildren(s)
	}
}

func (s *specializationFinder) visitStruct(struct_ *ast.Struct) bool {
	var types []ast.Type

	for _, field := range struct_.Fields {
		var type_ ast.Type

		switch fieldType := field.Type_.(type) {
		case *ast.Pointer:
			type_ = fieldType.Pointee
		case *ast.Array:
			type_ = fieldType.Base
		case *ast.Resolvable:
			type_ = fieldType.Type
		default:
			type_ = fieldType
		}

		if type_ != nil {
			for i, generic := range s.generics {
				if generic.Equals(type_) {
					if types == nil {
						types = make([]ast.Type, len(s.generics))
					}

					types[i] = generic
					break
				}
			}
		}
	}

	if types != nil {
		s.addStruct(struct_, types)
		return true
	}

	return false
}

func (s *specializationFinder) visitSpecStruct(spec *ast.SpecializedStruct) bool {
	for _, type_ := range spec.Types {
		for _, generic := range s.generics {
			if generic.Equals(type_) {
				s.addSpecStruct(spec)
				return true
			}
		}
	}

	return false
}

func (s *specializationFinder) getGenericI(type_ ast.Type) int {
	for i, generic := range s.generics {
		if generic.Equals(type_) {
			return i
		}
	}

	return -1
}

func (s *specializationFinder) addDirect(ptr *ast.Type, genericI int) {
	for _, rep := range s.direct {
		if rep.ptr == ptr {
			return
		}
	}

	s.direct = append(s.direct, directReplacement{
		ptr:      ptr,
		genericI: genericI,
		original: *ptr,
	})
}

func (s *specializationFinder) addStruct(struct_ *ast.Struct, types []ast.Type) {
	for _, rep := range s.structs {
		if rep.ptr == &struct_.Type {
			return
		}
	}

	s.structs = append(s.structs, structReplacement{
		s:     struct_,
		types: types,
		ptr:   &struct_.Type,
	})
}

func (s *specializationFinder) addSpecStruct(spec *ast.SpecializedStruct) {
	for _, rep := range s.structs {
		if rep.ptr == &spec.Type {
			return
		}
	}

	s.structs = append(s.structs, structReplacement{
		s:     spec.Underlying(),
		types: spec.Types,
		ptr:   &spec.Type,
	})
}
