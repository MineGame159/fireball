package llvm

import (
	"fireball/core/ir"
	"path/filepath"
)

func (w *textWriter) writeMetadata() {
	for name, meta := range w.m.NamedMetadata {
		w.writeRune('!')
		w.writeString(name)
		w.writeString(" = ")
		w.writeMeta(&meta)
		w.writeRune('\n')
	}

	if len(w.m.NamedMetadata) > 0 {
		w.writeRune('\n')
	}

	for _, meta := range w.m.Metadata {
		if !ir.IsMetaInline(meta) {
			w.writeRune('!')
			w.writeUint(uint64(meta.Index()), 10)
			w.writeString(" = ")
			w.writeMeta(meta)
			w.writeRune('\n')
		}
	}
}

func (w *textWriter) writeMeta(meta ir.Meta) {
	switch meta := meta.(type) {
	case *ir.GroupMeta:
		w.writeString("!{")

		for i, childId := range meta.Metadata {
			if i > 0 {
				w.writeString(", ")
			}

			switch child := w.m.Metadata[childId-1].(type) {
			case *ir.IntMeta:
				w.writeString("i32 ")
				w.writeInt(int64(child.Value))

			case *ir.StringMeta:
				w.writeRune('!')
				w.writeQuotedString(child.Value)

			default:
				w.writeMetaRef(childId)
			}
		}

		w.writeRune('}')

	case *ir.CompileUnitMeta:
		w.writeString("distinct !DICompileUnit(")

		w.writeString("file: ")
		w.writeMetaRef(meta.File)

		w.writeString(", producer: ")
		w.writeQuotedString(meta.Producer)

		w.writeString(", emissionKind: ")
		w.writeString(meta.Emission.String())

		w.writeString(", nameTableKind: ")
		w.writeString(meta.NameTable.String())

		if len(meta.Globals) > 0 {
			w.writeString(", globals: !{")

			for i, global := range meta.Globals {
				if i > 0 {
					w.writeString(", ")
				}

				w.writeMetaRef(global)
			}

			w.writeRune('}')
		}

		w.writeString(", language: ")
		w.writeString("DW_LANG_C")

		w.writeString(", isOptimized: ")
		w.writeBool(false)

		w.writeString(", runtimeVersion: ")
		w.writeUint(0, 10)

		w.writeString(", splitDebugInlining: ")
		w.writeBool(false)

		w.writeRune(')')

	case *ir.FileMeta:
		w.writeString("distinct !DIFile(")

		w.writeString("filename: ")
		w.writeQuotedString(filepath.Base(meta.Path))

		w.writeString(", directory: ")
		w.writeQuotedString(filepath.Dir(meta.Path))

		w.writeRune(')')

	case *ir.BasicTypeMeta:
		w.writeString("!DIBasicType(")

		w.writeString("name: ")
		w.writeQuotedString(meta.Name)

		w.writeString(", encoding: ")
		w.writeString(meta.Encoding.String())

		w.writeString(", size: ")
		w.writeUint(uint64(meta.Size), 10)

		w.writeString(", align: ")
		w.writeUint(uint64(meta.Align), 10)

		w.writeRune(')')

	case *ir.SubroutineTypeMeta:
		w.writeString("!DISubroutineType(types: !{")

		w.writeMetaRef(meta.Returns)

		for _, param := range meta.Params {
			w.writeString(", ")
			w.writeMetaRef(param)
		}

		w.writeString("})")

	case *ir.DerivedTypeMeta:
		w.writeString("!DIDerivedType(")

		w.writeString("tag: ")
		w.writeString(meta.Tag.String())

		w.writeString(", baseType: ")
		w.writeMetaRef(meta.BaseType)

		if meta.Size != 0 {
			w.writeString(", size: ")
			w.writeUint(uint64(meta.Size), 10)
		}

		if meta.Align != 0 {
			w.writeString(", align: ")
			w.writeUint(uint64(meta.Align), 10)
		}

		if meta.Offset != 0 {
			w.writeString(", offset: ")
			w.writeUint(uint64(meta.Offset), 10)
		}

		w.writeRune(')')

	case *ir.CompositeTypeMeta:
		w.writeString("!DICompositeType(")

		w.writeString("tag: ")
		w.writeString(meta.Tag.String())

		w.writeString(", name: ")
		w.writeQuotedString(meta.Name)

		w.writeString(", file: ")
		w.writeMetaRef(meta.File)

		w.writeString(", line: ")
		w.writeUint(uint64(meta.Line), 10)

		w.writeString(", size: ")
		w.writeUint(uint64(meta.Size), 10)

		w.writeString(", align: ")
		w.writeUint(uint64(meta.Align), 10)

		if meta.BaseType.Valid() {
			w.writeString(", baseType: ")
			w.writeMetaRef(meta.BaseType)
		}
		if len(meta.Elements) != 0 {
			w.writeString(", elements: !{")

			for i, element := range meta.Elements {
				if i > 0 {
					w.writeString(", ")
				}

				w.writeMetaRef(element)
			}

			w.writeRune('}')
		}

		w.writeRune(')')

	case *ir.SubrangeMeta:
		w.writeString("!DISubrange(")

		w.writeString("lowerBound: ")
		w.writeUint(uint64(meta.LowerBound), 10)

		w.writeString(", count: ")
		w.writeUint(uint64(meta.Count), 10)

		w.writeRune(')')

	case *ir.EnumeratorMeta:
		w.writeString("!DIEnumerator(")

		w.writeString("name: ")
		w.writeQuotedString(meta.Name)

		w.writeString(", value: ")
		if meta.Value.Negative {
			w.writeRune('-')
		}
		w.writeUint(meta.Value.Value, 10)

		w.writeRune(')')

	case *ir.NamespaceMeta:
		w.writeString("!DINamespace(")

		w.writeString("name: ")
		w.writeQuotedString(meta.Name)

		w.writeString(", scope: ")
		w.writeMetaRef(meta.Scope)

		w.writeString(", file: ")
		w.writeMetaRef(meta.File)

		w.writeString(", line: ")
		w.writeUint(uint64(meta.Line), 10)

		w.writeRune(')')

	case *ir.GlobalVarMeta:
		w.writeString("distinct !DIGlobalVariable(")

		w.writeString("name: ")
		w.writeQuotedString(meta.Name)

		w.writeString(", linkageName: ")
		w.writeQuotedString(meta.LinkageName)

		w.writeString(", scope: ")
		w.writeMetaRef(meta.Scope)

		w.writeString(", file: ")
		w.writeMetaRef(meta.File)

		w.writeString(", line: ")
		w.writeUint(uint64(meta.Line), 10)

		w.writeString(", type: ")
		w.writeMetaRef(meta.Type)

		w.writeString(", isLocal: ")
		w.writeBool(meta.Local)

		w.writeString(", isDefinition: ")
		w.writeBool(meta.Definition)

		w.writeRune(')')

	case *ir.GlobalVarExpr:
		w.writeString("!DIGlobalVariableExpression(")

		w.writeString("var: ")
		w.writeMetaRef(meta.Var)

		w.writeString(", expr: ")
		w.writeString("!DIExpression()")

		w.writeRune(')')

	case *ir.SubprogamMeta:
		w.writeString("distinct !DISubprogram(")

		w.writeString("name: ")
		w.writeQuotedString(meta.Name)

		w.writeString(", linkageName: ")
		w.writeQuotedString(meta.LinkageName)

		w.writeString(", scope: ")
		w.writeMetaRef(meta.Scope)

		w.writeString(", file: ")
		w.writeMetaRef(meta.File)

		w.writeString(", line: ")
		w.writeUint(uint64(meta.Line), 10)

		w.writeString(", type: ")
		w.writeMetaRef(meta.Type)

		w.writeString(", unit: ")
		w.writeMetaRef(meta.Unit)

		w.writeString(", spFlags: ")
		w.writeString("DISPFlagDefinition")

		w.writeRune(')')

	case *ir.LexicalBlockMeta:
		w.writeString("!DILexicalBlock(")

		w.writeString("scope: ")
		w.writeMetaRef(meta.Scope)

		w.writeString(", file: ")
		w.writeMetaRef(meta.File)

		w.writeString(", line: ")
		w.writeUint(uint64(meta.Line), 10)

		w.writeRune(')')

	case *ir.LocationMeta:
		w.writeString("!DILocation(")

		w.writeString("scope: ")
		w.writeMetaRef(meta.Scope)

		w.writeString(", line: ")
		w.writeUint(uint64(meta.Line), 10)

		w.writeString(", column: ")
		w.writeUint(uint64(meta.Column), 10)

		w.writeRune(')')

	case *ir.LocalVarMeta:
		w.writeString("distinct !DILocalVariable(")

		w.writeString("name: ")
		w.writeQuotedString(meta.Name)

		w.writeString(", type: ")
		w.writeMetaRef(meta.Type)

		if meta.Arg != 0 {
			w.writeString(", arg: ")
			w.writeUint(uint64(meta.Arg), 10)
		}

		w.writeString(", scope: ")
		w.writeMetaRef(meta.Scope)

		w.writeString(", file: ")
		w.writeMetaRef(meta.File)

		w.writeString(", line: ")
		w.writeUint(uint64(meta.Line), 10)

		w.writeRune(')')

	case *ir.ExpressionMeta:
		w.writeString("!DIExpression()")

	default:
		panic("ir.writeMeta() - Not implemented")
	}
}

func (w *textWriter) writeMetaRef(id ir.MetaID) {
	if id == 0 {
		w.writeString("null")
		return
	}

	w.writeRune('!')
	w.writeUint(uint64(w.m.Metadata[id-1].Index()), 10)
}
