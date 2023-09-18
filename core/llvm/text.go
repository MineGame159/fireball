package llvm

import (
	"fireball/core/utils"
	"fmt"
	"io"
)

type textWriter struct {
	w io.Writer

	intrinsics utils.Set[string]

	typeNames map[Type]string

	localUnnamedCount  int
	globalUnnamedCount int

	localValueNames      map[Value]string
	localValueNamesCount map[string]int

	globalValueNames      map[Value]string
	globalValueNamesCount map[string]int

	metadataCount int
}

func WriteText(module *Module, writer io.Writer) {
	w := &textWriter{
		w: writer,

		intrinsics: utils.NewSet[string](),

		typeNames: make(map[Type]string),

		globalValueNames:      make(map[Value]string),
		globalValueNamesCount: make(map[string]int),
	}

	// Source
	w.fmt("source_filename = \"%s\"\n", module.source)
	w.line()

	// Types
	types := 0

	for _, t := range module.types {
		if v, ok := t.(*structType); ok {
			name := fmt.Sprintf("%%struct.%s", v.name)
			w.typeNames[t] = name

			w.fmt("%s = type { ", name)

			for i, field := range v.fields {
				if i > 0 {
					w.raw(", ")
				}

				w.raw(w.type_(field.Type))
			}

			w.raw(" }\n")
			types++
		}
	}

	if types > 0 {
		w.line()
	}

	// Constants
	for _, c := range module.constants {
		data, length := escape(c.data)
		w.fmt("%s = private unnamed_addr constant [%d x i8] c\"%s\"\n", w.value(c), length, data)
	}

	if len(module.constants) > 0 {
		w.line()
	}

	// Defines
	for _, define := range module.defines {
		w.beginFunction()
		w.fmt("define %s @%s(", w.type_(define.type_.returns), define.type_.name)

		for i, parameter := range define.parameters {
			if i > 0 {
				w.raw(", ")
			}

			w.fmt("%s %s", w.type_(parameter.Type()), w.value(parameter))
		}

		if define.type_.variadic {
			if len(define.parameters) > 0 {
				w.raw(", ...")
			} else {
				w.raw("...")
			}
		}

		w.fmt(") !dbg !%d {\n", define.metadata)
		w.body(define.blocks)
		w.raw("}\n\n")
	}

	// Declares
	for _, declare := range module.declares {
		w.beginFunction()
		w.fmt("declare %s @%s(", w.type_(declare.returns), declare.name)

		for i, parameter := range declare.parameters {
			if i > 0 {
				w.raw(", ")
			}

			w.raw(w.type_(parameter))
		}

		if declare.variadic {
			if len(declare.parameters) > 0 {
				w.raw(", ...")
			} else {
				w.raw("...")
			}
		}

		w.raw(")\n")
	}

	for intrinsic := range w.intrinsics.Data {
		w.fmt("declare %s\n", intrinsic)
	}

	if len(module.declares) > 0 {
		w.line()
	}

	// Debug
	w.debug(module)
}

// Functions

func (w *textWriter) beginFunction() {
	w.localUnnamedCount = 0

	w.localValueNames = make(map[Value]string)
	w.localValueNamesCount = make(map[string]int)
}

func (w *textWriter) body(blocks []*Block) {
	// Gather names of blocks and instructions
	for _, block := range blocks {
		w.value(block)

		for _, inst := range block.instructions {
			if _, ok := inst.Type().(*voidType); inst.Type() != nil && !ok {
				w.value(inst)
			}
		}
	}

	// Emit blocks and instructions
	for _, block := range blocks {
		w.fmt("%s:\n", w.value(block)[1:])

		for _, inst := range block.instructions {
			w.raw("    ")

			if w.instruction(inst) {
				break
			}
		}
	}
}

func (w *textWriter) instruction(i Value) bool {
	if _, ok := i.Type().(*voidType); i.Type() != nil && !ok {
		w.fmt("%s = ", w.value(i))
	}

	location := -1
	terminal := false

	switch inst := i.(type) {
	case *variable:
		w.intrinsics.Add("void @llvm.dbg.declare(metadata, metadata, metadata)")
		w.fmt("call void @llvm.dbg.declare(metadata ptr %s, metadata !%d, metadata !DIExpression())", w.value(inst.pointer), inst.metadata)

		location = inst.location

	case *lifetime:
		if inst.start {
			w.intrinsics.Add("void @llvm.lifetime.start(i64, ptr)")
			w.fmt("call void @llvm.lifetime.start(i64 %d, ptr %s)", inst.pointer.Type().Size(), w.value(inst.pointer))
		} else {
			w.intrinsics.Add("void @llvm.lifetime.end(i64, ptr)")
			w.fmt("call void @llvm.lifetime.end(i64 %d, ptr %s)", inst.pointer.Type().Size(), w.value(inst.pointer))
		}

		location = inst.location

	case *fNeg:
		w.fmt("fneg %s %s", w.type_(inst.Type()), w.value(inst.value))
		location = inst.location

	case *binary:
		a := ""

		switch inst.op {
		case Add:
			a = ternary(isFloating(inst.left.Type()), "fadd", "add")
		case Sub:
			a = ternary(isFloating(inst.left.Type()), "fsub", "sub")
		case Mul:
			a = ternary(isFloating(inst.left.Type()), "fmul", "mul")
		case Div:
			a = ternary(isFloating(inst.left.Type()), "fdiv", ternary(isSigned(inst.Type()), "sdiv", "udiv"))
		case Rem:
			a = ternary(isFloating(inst.left.Type()), "frem", ternary(isSigned(inst.Type()), "srem", "urem"))

		case Eq:
			a = ternary(isFloating(inst.left.Type()), "fcmp oeq", "icmp eq")
		case Ne:
			a = ternary(isFloating(inst.left.Type()), "fcmp one", "icmp ne")
		case Lt:
			a = ternary(isFloating(inst.left.Type()), "fcmp olt", ternary(isSigned(inst.Type()), "icmp slt", "icmp ult"))
		case Le:
			a = ternary(isFloating(inst.left.Type()), "fcmp ole", ternary(isSigned(inst.Type()), "icmp sle", "icmp ule"))
		case Gt:
			a = ternary(isFloating(inst.left.Type()), "fcmp ogt", ternary(isSigned(inst.Type()), "icmp sgt", "icmp ugt"))
		case Ge:
			a = ternary(isFloating(inst.left.Type()), "fcmp oge", ternary(isSigned(inst.Type()), "icmp sge", "icmp uge"))

		case Or:
			a = "or"
		case Xor:
			a = "xor"
		case And:
			a = "and"
		case Shl:
			a = "shl"
		case Shr:
			a = ternary(isSigned(inst.left.Type()), "ashr", "lshr")
		}

		w.fmt("%s %s %s, %s", a, w.type_(inst.left.Type()), w.value(inst.left), w.value(inst.right))
		location = inst.location

	case *cast:
		a := ""

		switch inst.kind {
		case Trunc:
			a = "trunc"
		case ZExt:
			a = "zext"
		case SExt:
			a = "sext"
		case FpTrunc:
			a = "fptrunc"
		case FpExt:
			a = "fpext"
		case FpToUi:
			a = "fptoui"
		case FpToSi:
			a = "fptosi"
		case UiToFp:
			a = "uitofp"
		case SiToFp:
			a = "sitofp"
		case PtrToInt:
			a = "ptrtoint"
		case IntToPtr:
			a = "inttoptr"
		case Bitcast:
			a = "bitcast"
		}

		w.fmt("%s %s %s to %s", a, w.type_(inst.value.Type()), w.value(inst.value), w.type_(inst.Type()))
		location = inst.location

	case *extractValue:
		w.fmt("extractvalue %s %s, %d", w.type_(inst.value.Type()), w.value(inst.value), inst.index)
		location = inst.location

	case *insertValue:
		w.fmt("insertvalue %s %s, %s %s, %d", w.type_(inst.value.Type()), w.value(inst.value), w.type_(inst.element.Type()), w.value(inst.element), inst.index)
		location = inst.location

	case *alloca:
		w.fmt("alloca %s", w.type_(inst.type_))
		location = inst.location

	case *load:
		w.fmt("load %s, ptr %s", w.type_(inst.Type()), w.value(inst.pointer))
		location = inst.location

	case *store:
		w.fmt("store %s %s, ptr %s", w.type_(inst.value.Type()), w.value(inst.value), w.value(inst.pointer))
		location = inst.location

	case *getElementPtr:
		w.fmt("getelementptr inbounds %s, %s %s", w.type_(inst.type_), w.type_(inst.pointer.Type()), w.value(inst.pointer))
		location = inst.location

		for _, index := range inst.indices {
			w.fmt(", %s %s", w.type_(index.Type()), w.value(index))
		}

	case *br:
		if inst.condition == nil {
			w.fmt("br label %s", w.value(inst.true))
		} else {
			w.fmt("br i1 %s, label %s, label %s", w.value(inst.condition), w.value(inst.true), w.value(inst.false))
		}

		location = inst.location
		terminal = true

	case *phi:
		w.fmt("phi %s [ %s, %s ], [ %s, %s ]", w.type_(inst.type_), w.value(inst.firstValue), w.value(inst.firstBlock), w.value(inst.secondValue), w.value(inst.secondBlock))
		location = inst.location

	case *call:
		w.fmt("call %s %s(", w.type_(inst.Type()), w.value(inst.value))

		for i, argument := range inst.arguments {
			if i > 0 {
				w.raw(", ")
			}

			w.fmt("%s %s", w.type_(argument.Type()), w.value(argument))
		}

		w.raw(")")
		location = inst.location

	case *ret:
		if inst.value == nil {
			w.raw("ret void")
		} else {
			w.fmt("ret %s %s", w.type_(inst.value.Type()), w.value(inst.value))
		}

		location = inst.location
		terminal = true

	default:
		panic("textWriter.instruction() - Invalid instruction")
	}

	if location != -1 {
		w.fmt(", !dbg !%d", location)
	}

	w.raw("\n")
	return terminal
}

// Types

func (w *textWriter) type_(type_ Type) string {
	if name, ok := w.typeNames[type_]; ok {
		return name
	}

	name := ""

	if _, ok := type_.(*voidType); ok {
		// Void
		name = "void"
	} else if v, ok := type_.(*primitiveType); ok {
		// Primitive
		switch v.encoding {
		case BooleanEncoding:
			name = "i1"

		case FloatEncoding:
			if v.bitSize == 32 {
				name = "float"
			} else if v.bitSize == 64 {
				name = "double"
			}

		case SignedEncoding, UnsignedEncoding:
			name = fmt.Sprintf("i%d", v.bitSize)
		}
	} else if v, ok := type_.(*arrayType); ok {
		// Array
		name = fmt.Sprintf("[%d x %s]", v.count, w.type_(v.base))
	} else if _, ok := type_.(*pointerType); ok {
		// Pointer
		name = "ptr"
	} else if v, ok := type_.(*aliasType); ok {
		// Alias
		name = w.type_(v.underlying)
	}

	if name != "" {
		w.typeNames[type_] = name
		return name
	} else {
		panic("textWriter.type_() - Invalid type")
	}
}

// Constants

func escape(original string) (string, int) {
	data := make([]uint8, 0, len(original)+3)
	length := 1

	for i := 0; i < len(original); i++ {
		char := original[i]

		if char == '\\' {
			i++

			switch original[i] {
			case '0':
				data = append(data, '\\')
				data = append(data, '0')
				data = append(data, '0')

			case 'n':
				data = append(data, '\\')
				data = append(data, '0')
				data = append(data, 'A')

			case 'r':
				data = append(data, '\\')
				data = append(data, '0')
				data = append(data, 'D')

			case 't':
				data = append(data, '\\')
				data = append(data, '0')
				data = append(data, '9')
			}

			length++
		} else {
			data = append(data, char)
			length++
		}
	}

	data = append(data, '\\')
	data = append(data, '0')
	data = append(data, '0')

	return string(data), length
}

// Names

func (w *textWriter) value(value Value) string {
	switch value.Kind() {
	case GlobalValue:
		if name, ok := w.globalValueNames[value]; ok {
			return name
		}

		name := value.Name()

		if name == "" {
			name = fmt.Sprintf("@%d", w.globalUnnamedCount)
			w.globalUnnamedCount++
		} else {
			name = "@" + name

			if count, ok := w.globalValueNamesCount[name]; ok {
				name += fmt.Sprintf(".%d", count+1)
				w.globalValueNamesCount[name]++
			} else {
				w.globalValueNamesCount[name] = 0
			}
		}

		w.globalValueNames[value] = name
		return name

	case LocalValue:
		if name, ok := w.localValueNames[value]; ok {
			return name
		}

		name := value.Name()

		if name == "" {
			name = fmt.Sprintf("%%%d", w.localUnnamedCount)
			w.localUnnamedCount++
		} else {
			name = "%" + name

			if count, ok := w.localValueNamesCount[name]; ok {
				name += fmt.Sprintf(".%d", count+1)
				w.localValueNamesCount[name]++
			} else {
				w.localValueNamesCount[name] = 0
			}
		}

		w.localValueNames[value] = name
		return name
	}

	return value.Name()
}

// Debug

func (w *textWriter) debug(module *Module) {
	// Named metadata
	for name, metadata := range module.namedMetadata {
		w.fmt("!%s = ", name)
		w.metadata(metadata)
	}

	w.line()

	// Unnamed metadata
	for i, metadata := range module.metadata {
		w.fmt("!%d = ", i)
		w.metadata(metadata)
	}
}

func (w *textWriter) metadata(metadata Metadata) {
	if metadata.Distinct {
		w.raw("distinct ")
	}

	if metadata.Type != "" {
		w.fmt("!%s(", metadata.Type)
	} else {
		w.raw("!{")
	}

	for i, field := range metadata.Fields {
		if i > 0 {
			w.raw(", ")
		}

		if field.Name != "" {
			w.fmt("%s: ", field.Name)

			switch field.Value.Kind {
			case StringMetadataValueKind:
				w.fmt("\"%s\"", field.Value.String)

			case EnumMetadataValueKind:
				w.raw(field.Value.String)

			case NumberMetadataValueKind:
				w.fmt("%d", field.Value.Number)

			case RefMetadataValueKind:
				w.fmt("!%d", field.Value.Number)
			}
		} else {
			switch field.Value.Kind {
			case StringMetadataValueKind:
				w.fmt("!\"%s\"", field.Value.String)

			case EnumMetadataValueKind:
				w.raw(field.Value.String)

			case NumberMetadataValueKind:
				w.fmt("i32 %d", field.Value.Number)

			case RefMetadataValueKind:
				w.fmt("!%d", field.Value.Number)
			}
		}
	}

	if metadata.Type != "" {
		w.raw(")\n")
	} else {
		w.raw("}\n")
	}
}

// Utils

func ternary[T any](condition bool, true T, false T) T {
	if condition {
		return true
	}

	return false
}

// Write

func (w *textWriter) fmt(format string, args ...any) {
	_, _ = fmt.Fprintf(w.w, format, args...)
}

func (w *textWriter) raw(str string) {
	_, _ = w.w.Write([]byte(str))
}

func (w *textWriter) line() {
	_, _ = w.w.Write([]byte{'\n'})
}
