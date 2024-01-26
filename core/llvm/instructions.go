package llvm

import "fireball/core/ir"

func (w *textWriter) writeInstruction(inst ir.Inst) {
	// Value
	if _, ok := inst.Type().(*ir.VoidType); inst.Type() != nil && !ok {
		w.writeName(inst)
		w.writeString(" = ")
	}

	// Instruction
	switch inst := inst.(type) {
	case *ir.RetInst:
		if inst.Value == nil {
			w.writeString("ret void")
		} else {
			w.writeString("ret ")
			w.writeValue(inst.Value)
		}

	case *ir.BrInst:
		if inst.Condition == nil {
			w.writeString("br ")
			w.writeValue(inst.True)
		} else {
			w.writeString("br ")
			w.writeValue(inst.Condition)
			w.writeString(", ")
			w.writeValue(inst.True)
			w.writeString(", ")
			w.writeValue(inst.False)
		}

	case *ir.FNegInst:
		w.writeString("fneg ")
		w.writeValue(inst.Value)

	case *ir.AddInst:
		if _, ok := inst.Type().(*ir.FloatType); ok {
			w.writeString("fadd ")
		} else {
			w.writeString("add ")
		}

		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.SubInst:
		if _, ok := inst.Type().(*ir.FloatType); ok {
			w.writeString("fsub ")
		} else {
			w.writeString("sub ")
		}

		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.MulInst:
		if _, ok := inst.Type().(*ir.FloatType); ok {
			w.writeString("fmul ")
		} else {
			w.writeString("mul ")
		}

		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.IDivInst:
		if inst.Signed {
			w.writeString("sdiv ")
		} else {
			w.writeString("udiv ")
		}

		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.FDivInst:
		w.writeString("fdiv ")
		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.IRemInst:
		if inst.Signed {
			w.writeString("srem ")
		} else {
			w.writeString("urem ")
		}

		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.FRemInst:
		w.writeString("frem ")
		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.ShlInst:
		w.writeString("shl ")
		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.ShrInst:
		if inst.SignExtend {
			w.writeString("ashr")
		} else {
			w.writeString("lshr")
		}

		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.AndInst:
		w.writeString("add ")
		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.OrInst:
		w.writeString("or ")
		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.XorInst:
		w.writeString("xor ")
		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.ExtractValueInst:
		w.writeString("extractvalue ")
		w.writeValue(inst.Value)

		for _, index := range inst.Indices {
			w.writeString(", ")
			w.writeUint(uint64(index), 10)
		}

	case *ir.InsertValueInst:
		w.writeString("insertvalue ")
		w.writeValue(inst.Value)
		w.writeString(", ")
		w.writeValue(inst.Element)

		for _, index := range inst.Indices {
			w.writeString(", ")
			w.writeUint(uint64(index), 10)
		}

	case *ir.AllocaInst:
		w.writeString("alloca ")
		w.writeType(inst.Typ)

		if inst.Align != 0 {
			w.writeString(", align ")
			w.writeUint(uint64(inst.Align), 10)
		}

	case *ir.LoadInst:
		w.writeString("load ")
		w.writeType(inst.Typ)
		w.writeString(", ")
		w.writeValue(inst.Pointer)

		if inst.Align != 0 {
			w.writeString(", align ")
			w.writeUint(uint64(inst.Align), 10)
		}

	case *ir.StoreInst:
		w.writeString("store ")
		w.writeValue(inst.Value)
		w.writeString(", ")
		w.writeValue(inst.Pointer)

		if inst.Align != 0 {
			w.writeString(", align ")
			w.writeUint(uint64(inst.Align), 10)
		}

	case *ir.GetElementPtrInst:
		w.writeString("getelementptr ")

		if inst.Inbounds {
			w.writeString("inbounds ")
		}

		w.writeType(inst.Typ)
		w.writeString(", ")
		w.writeValue(inst.Pointer)

		for _, index := range inst.Indices {
			w.writeString(", ")
			w.writeValue(index)
		}

	case *ir.TruncInst:
		if _, ok := inst.Typ.(*ir.FloatType); ok {
			w.writeString("fptrunc ")
		} else {
			w.writeString("trunc ")
		}

		w.writeValue(inst.Value)
		w.writeString(" to ")
		w.writeType(inst.Typ)

	case *ir.ExtInst:
		if inst.SignExtend {
			w.writeString("sext ")
		} else {
			w.writeString("zext ")
		}

		w.writeValue(inst.Value)
		w.writeString(" to ")
		w.writeType(inst.Typ)

	case *ir.FExtInst:
		w.writeString("fpext ")
		w.writeValue(inst.Value)
		w.writeString(" to ")
		w.writeType(inst.Typ)

	case *ir.F2IInst:
		if inst.Signed {
			w.writeString("fptosi ")
		} else {
			w.writeString("fptoui ")
		}

		w.writeValue(inst.Value)
		w.writeString(" to ")
		w.writeType(inst.Typ)

	case *ir.I2FInst:
		if inst.Signed {
			w.writeString("sitofp ")
		} else {
			w.writeString("uitofp ")
		}

		w.writeValue(inst.Value)
		w.writeString(" to ")
		w.writeType(inst.Typ)

	case *ir.ICmpInst:
		w.writeString("icmp ")

		if inst.Signed {
			switch inst.Kind {
			case ir.Eq:
				w.writeString("eq ")
			case ir.Ne:
				w.writeString("ne ")
			case ir.Gt:
				w.writeString("sgt ")
			case ir.Ge:
				w.writeString("sge ")
			case ir.Lt:
				w.writeString("slt ")
			case ir.Le:
				w.writeString("sle ")
			}
		} else {
			switch inst.Kind {
			case ir.Eq:
				w.writeString("eq ")
			case ir.Ne:
				w.writeString("ne ")
			case ir.Gt:
				w.writeString("ugt ")
			case ir.Ge:
				w.writeString("uge ")
			case ir.Lt:
				w.writeString("ult ")
			case ir.Le:
				w.writeString("ule ")
			}
		}

		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.FCmpInst:
		w.writeString("fcmp ")

		if inst.Ordered {
			switch inst.Kind {
			case ir.Eq:
				w.writeString("oeq ")
			case ir.Ne:
				w.writeString("one ")
			case ir.Gt:
				w.writeString("ogt ")
			case ir.Ge:
				w.writeString("oge ")
			case ir.Lt:
				w.writeString("olt ")
			case ir.Le:
				w.writeString("ole ")
			}
		} else {
			switch inst.Kind {
			case ir.Eq:
				w.writeString("ueq ")
			case ir.Ne:
				w.writeString("une ")
			case ir.Gt:
				w.writeString("ugt ")
			case ir.Ge:
				w.writeString("uge ")
			case ir.Lt:
				w.writeString("ult ")
			case ir.Le:
				w.writeString("ule ")
			}
		}

		w.writeValue(inst.Left)
		w.writeString(", ")
		w.writeValueValue(inst.Right)

	case *ir.PhiInst:
		w.writeString("phi ")
		w.writeType(inst.Type())

		for i, incoming := range inst.Incs {
			if i > 0 {
				w.writeString(", [ ")
			} else {
				w.writeString(" [ ")
			}

			w.writeValueValue(incoming.Value)
			w.writeString(", ")
			w.writeValueValue(incoming.Label)

			w.writeString(" ]")
		}

	case *ir.SelectInst:
		w.writeString("select ")
		w.writeValue(inst.Condition)
		w.writeString(", ")
		w.writeValue(inst.True)
		w.writeString(", ")
		w.writeValue(inst.False)

	case *ir.CallInst:
		var type_ *ir.FuncType

		if inst.Typ != nil {
			type_ = inst.Typ
		} else {
			type_ = inst.Callee.Type().(*ir.FuncType)
		}

		w.writeString("call ")
		w.writeType(type_.Returns)
		w.writeRune(' ')
		w.writeName(inst.Callee)
		w.writeRune('(')

		for i, arg := range inst.Args {
			if i > 0 {
				w.writeString(", ")
			}

			if i < len(type_.Params) {
				param := type_.Params[i]

				if _, ok := param.Type().(*ir.MetaType); ok {
					if _, ok := arg.(ir.MetaID); !ok {
						w.writeString("metadata ")
					}
				}
			}

			w.writeValue(arg)
		}

		w.writeRune(')')

	default:
		panic("ir.writeInstruction() - Not implemented")
	}

	// Metadata
	if inst.Meta().Valid() {
		w.writeString(", !dbg ")
		w.writeMetaRef(inst.Meta())
	}
}
