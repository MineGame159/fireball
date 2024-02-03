package llvm

import (
	"fireball/core/ir"
)

func (w *textWriter) writeStructs() {
	for _, s := range w.m.Structs {
		w.writeString("%struct.")
		w.writeString(s.Name)
		w.writeString(" = type ")
		w.writeStruct(s)
		w.writeRune('\n')
	}

	if len(w.m.Structs) > 0 {
		w.writeRune('\n')
	}
}

func (w *textWriter) writeType(t ir.Type) {
	if t == nil {
		w.writeString("void")
		return
	}

	switch t := t.(type) {
	case *ir.VoidType:
		w.writeString("void")

	case *ir.IntType:
		w.writeRune('i')
		w.writeInt(int64(t.BitSize))

	case *ir.FloatType:
		switch t.BitSize {
		case 32:
			w.writeString("float")
		case 64:
			w.writeString("double")

		default:
			panic("ir.writeType() - llvm.FloatType - Not implemented")
		}

	case *ir.PointerType:
		w.writeString("ptr")

		if t.ByVal != nil && w.isArgument {
			w.writeString(" byval(")
			w.writeType(t.ByVal)
			w.writeRune(')')
		} else if t.SRet != nil && w.isArgument {
			w.writeString(" sret(")
			w.writeType(t.SRet)
			w.writeRune(')')
		}

	case *ir.ArrayType:
		w.writeRune('[')
		w.writeInt(int64(t.Count))
		w.writeString(" x ")
		w.writeType(t.Base)
		w.writeRune(']')

	case *ir.StructType:
		for _, s := range w.m.Structs {
			if s == t {
				w.writeString("%struct.")
				w.writeString(s.Name)

				return
			}
		}

		w.writeStruct(t)

	case *ir.FuncType:
		w.writeString("ptr")

	case *ir.MetaType:
		w.writeString("metadata")

	default:
		panic("ir.writeType() - Not implemented")
	}
}

func (w *textWriter) writeStruct(s *ir.StructType) {
	w.writeString("{ ")

	for i, field := range s.Fields {
		if i > 0 {
			w.writeString(", ")
		}

		w.writeType(field)
	}

	w.writeString(" }")
}
