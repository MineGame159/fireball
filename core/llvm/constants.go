package llvm

import (
	"fireball/core/ir"
	"math"
)

func (w *textWriter) writeConst(c ir.Const) {
	switch c := c.(type) {
	case *ir.NullConst:
		w.writeString("null")

	case *ir.IntConst:
		if c.Type().(*ir.IntType).BitSize == 1 {
			if c.Value.Value == 1 {
				w.writeString("true")
			} else {
				w.writeString("false")
			}
		} else {
			if c.Value.Negative {
				w.writeRune('-')
			}

			w.writeUint(c.Value.Value, 10)
		}

	case *ir.FloatConst:
		switch c.Type().(*ir.FloatType).BitSize {
		case 32:
			bits := math.Float64bits(c.Value)

			w.writeString("0x")
			w.writeUint(bits, 16)

		case 64:
			bits := math.Float64bits(c.Value)

			w.writeString("0x")
			w.writeUint(bits, 16)

		default:
			panic("ir.writeConst() - FloatConst - Not implemented")
		}

	case *ir.StringConst:
		w.writeString("c\"")

		for _, b := range c.Value {
			switch b {
			case '\000':
				w.writeString("\\00")
			case '\n':
				w.writeString("\\0A")
			case '\r':
				w.writeString("\\0D")
			case '\t':
				w.writeString("\\09")

			default:
				w.writeByte(b)
			}
		}

		w.writeRune('"')

	case *ir.ZeroInitConst:
		w.writeString("zeroinitializer")

	case *ir.ArrayConst:
		w.writeString("[ ")

		for i, value := range c.Values {
			if i > 0 {
				w.writeString(", ")
			}

			w.writeValue(value)
		}

		w.writeString(" ]")

	case *ir.StructConst:
		w.writeString("{ ")

		for i, field := range c.Fields {
			if i > 0 {
				w.writeString(", ")
			}

			w.writeValue(field)
		}

		w.writeString(" }")

	default:
		panic("ir.writeConst() - Not implemented")
	}
}
