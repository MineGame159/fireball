package llvm

import "fireball/core/ir"

func (w *textWriter) writeGlobals() {
	// Constants
	count := 0

	for _, global := range w.m.Globals {
		if global.Constant {
			w.writeGlobalConstant(global)
			count++
		}
	}

	if count > 0 {
		w.writeRune('\n')
	}

	// Variables
	count = 0

	for _, global := range w.m.Globals {
		if !global.Constant {
			w.writeGlobalVar(global)
			count++
		}
	}

	if count > 0 {
		w.writeRune('\n')
	}
}

func (w *textWriter) writeGlobalConstant(global *ir.GlobalVar) {
	w.writeName(global)
	w.writeString(" = private unnamed_addr constant ")
	w.writeType(global.Typ)
	w.writeRune(' ')
	w.writeConst(global.Value.(ir.Const))
	w.writeRune('\n')
}

func (w *textWriter) writeGlobalVar(global *ir.GlobalVar) {
	w.writeName(global)
	w.writeString(" = ")

	if global.Value == nil {
		w.writeString("external global ")
		w.writeType(global.Typ)
	} else {
		w.writeString("global ")
		w.writeType(global.Typ)
		w.writeRune(' ')
		w.writeConst(global.Value.(ir.Const))
	}

	if global.Meta().Valid() {
		w.writeString(", !dbg ")
		w.writeMetaRef(global.Meta())
	}

	w.writeRune('\n')
}
