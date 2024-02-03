package llvm

import "fireball/core/ir"

func (w *textWriter) writeFunctions() {
	// Define
	for _, function := range w.m.Functions {
		if len(function.Blocks) != 0 {
			w.resetLocalNames()

			w.writeString("define ")
			w.writeFunction(function)

			if function.Meta().Valid() {
				w.writeString(" !dbg ")
				w.writeMetaRef(function.Meta())
			}

			w.writeString(" {\n")

			for _, param := range function.Typ.Params {
				w.cacheName(param, false)
			}

			for _, block := range function.Blocks {
				w.cacheName(block, false)

				for _, inst := range block.Instructions {
					if _, ok := inst.Type().(*ir.VoidType); inst.Type() != nil && !ok {
						w.cacheName(inst, false)
					}
				}
			}

			for i, block := range function.Blocks {
				if i > 0 {
					w.writeRune('\n')
				}

				w.writeBlock(block)
			}

			w.writeString("}\n\n")
		}
	}

	// Declare
	count := 0

	for _, function := range w.m.Functions {
		if len(function.Blocks) == 0 {
			w.resetLocalNames()

			w.writeString("declare ")
			w.writeFunction(function)
			w.writeRune('\n')

			count++
		}
	}

	if count > 0 {
		w.writeRune('\n')
	}
}

func (w *textWriter) writeFunction(function *ir.Func) {
	// Return type
	w.writeType(function.Typ.Returns)
	w.writeRune(' ')

	// Name
	w.writeName(function)

	// Parameters
	w.writeRune('(')
	w.isArgument = true

	for i, param := range function.Typ.Params {
		if i > 0 {
			w.writeString(", ")
		}

		w.writeType(param.Typ)
		w.writeRune(' ')
		w.writeName(param)
	}

	if function.Typ.Variadic {
		if len(function.Typ.Params) > 0 {
			w.writeString(", ...")
		} else {
			w.writeString("...")
		}
	}

	w.isArgument = false
	w.writeRune(')')

	// Flags
	if function.Flags&ir.InlineFlag != 0 {
		w.writeString(" alwaysinline")
	}
	if function.Flags&ir.HotFlag != 0 {
		w.writeString(" hot")
	}

	// Metadata
}

func (w *textWriter) writeBlock(block *ir.Block) {
	w.skipNameChar = true
	w.writeName(block)
	w.writeString(":\n")
	w.skipNameChar = false

	for _, inst := range block.Instructions {
		w.writeString("    ")
		w.writeInstruction(inst)
		w.writeRune('\n')
	}
}
