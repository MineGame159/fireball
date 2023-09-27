package llvm

import (
	"fmt"
	"math"
	"strconv"
)

type Literal struct {
	Signed   int64
	Unsigned uint64
	Floating float64
}

type Function struct {
	module *Module

	type_ *functionType

	parameters []NameableValue
	blocks     []*Block

	metadata int
}

func (f *Function) Kind() ValueKind {
	return GlobalValue
}

func (f *Function) Type() Type {
	return f.type_
}

func (f *Function) Name() string {
	return f.type_.name
}

func (f *Function) GetParameter(index int) NameableValue {
	return f.parameters[index]
}

func (f *Function) PushScope() {
	f.module.scopes = append(f.module.scopes, f.metadata)
}

func (f *Function) PopScope() {
	f.module.scopes = f.module.scopes[:len(f.module.scopes)-1]
}

func (f *Function) Block(name string) *Block {
	// TODO: If I use named blocks then LLVM for some reason reports that a terminator instruction is inside the middle
	//       of a basic block even tho it works fine when the blocks are unnamed.

	b := &Block{
		module:       f.module,
		name:         "",
		instructions: make([]Value, 0, 8),
	}

	f.blocks = append(f.blocks, b)
	return b
}

// Literals

type literal struct {
	type_ Type
	data  string
}

func (l *literal) Kind() ValueKind {
	return LiteralValue
}

func (l *literal) Type() Type {
	return l.type_
}

func (l *literal) Name() string {
	return l.data
}

func (f *Function) Literal(type_ Type, data Literal) Value {
	if isSigned(type_) {
		return &literal{
			type_: type_,
			data:  strconv.FormatInt(data.Signed, 10),
		}
	}

	if isUnsigned(type_) {
		return &literal{
			type_: type_,
			data:  strconv.FormatUint(data.Unsigned, 10),
		}
	}

	if isFloating(type_) {
		return &literal{
			type_: type_,
			data:  fmt.Sprintf("0x%X", math.Float64bits(data.Floating)),
		}
	}

	panic("llvm.Function.Literal() - Invalid literal")
}

func (f *Function) LiteralRaw(type_ Type, data string) Value {
	return &literal{
		type_: type_,
		data:  data,
	}
}

// Block

type Block struct {
	module *Module

	name         string
	instructions []Value
}

func (b *Block) Kind() ValueKind {
	return LocalValue
}

func (b *Block) Type() Type {
	panic("llvm.Block.Type() - Blocks dont have types")
}

func (b *Block) Name() string {
	return b.name
}

func (b *Block) Variable(name string, pointer Value) Instruction {
	i := &variableMetadata{
		instruction: instruction{
			module:   b.module,
			location: -1,
		},
		pointer: pointer,
		metadata: b.module.addMetadata(Metadata{
			Type: "DILocalVariable",
			Fields: []MetadataField{
				{
					Name:  "name",
					Value: stringMetadataValue(name),
				},
				{
					Name:  "type",
					Value: refMetadataValue(b.module.typeMetadata[pointer.Type().(*pointerType).pointee]),
				},
				{
					Name:  "scope",
					Value: refMetadataValue(b.module.getScope()),
				},
				{
					Name:  "file",
					Value: refMetadataValue(b.module.getFile()),
				},
			},
		}),
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Lifetime(pointer Value, start bool) Instruction {
	i := &lifetimeMetadata{
		instruction: instruction{
			module:   b.module,
			location: -1,
		},
		pointer: pointer,
		start:   start,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) FNeg(value Value) InstructionValue {
	i := &fNeg{
		instruction: instruction{
			module:   b.module,
			type_:    value.Type(),
			location: -1,
		},
		value: value,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Binary(op BinaryKind, left, right Value) InstructionValue {
	var type_ Type

	switch op {
	case Eq, Ne, Lt, Le, Gt, Ge:
		for _, t := range b.module.types {
			if t2, ok := t.(*primitiveType); ok && t2.encoding == BooleanEncoding {
				type_ = t2
			}
		}

		if type_ == nil {
			type_ = b.module.Primitive("i1", 1, BooleanEncoding)
		}
	default:
		type_ = left.Type()
	}

	i := &binary{
		instruction: instruction{
			module:   b.module,
			type_:    type_,
			location: -1,
		},
		op:    op,
		left:  left,
		right: right,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Cast(kind CastKind, value Value, to Type) InstructionValue {
	i := &cast{
		instruction: instruction{
			module:   b.module,
			type_:    to,
			location: -1,
		},
		kind:  kind,
		value: value,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) ExtractValue(value Value, index int) InstructionValue {
	var type_ Type

	if v, ok := value.Type().(*arrayType); ok {
		type_ = v.base
	} else if v, ok := value.Type().(*structType); ok {
		type_ = v.fields[index].Type
	}

	i := &extractValue{
		instruction: instruction{
			module:   b.module,
			type_:    type_,
			location: -1,
		},
		value: value,
		index: index,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) InsertValue(value, element Value, index int) InstructionValue {
	i := &insertValue{
		instruction: instruction{
			module:   b.module,
			type_:    value.Type(),
			location: -1,
		},
		value:   value,
		element: element,
		index:   index,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Alloca(type_ Type) InstructionValue {
	i := &alloca{
		instruction: instruction{
			module:   b.module,
			type_:    &pointerType{pointee: type_},
			location: -1,
		},
		type_: type_,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Load(pointer Value) InstructionValue {
	i := &load{
		instruction: instruction{
			module:   b.module,
			type_:    pointer.Type().(*pointerType).pointee,
			location: -1,
		},
		pointer: pointer,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Store(pointer Value, value Value) InstructionValue {
	i := &store{
		instruction: instruction{
			module:   b.module,
			location: -1,
		},
		pointer: pointer,
		value:   value,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) GetElementPtr(pointer Value, indices []Value, pointerType Type, type_ Type) InstructionValue {
	i := &getElementPtr{
		instruction: instruction{
			module:   b.module,
			type_:    pointerType,
			location: -1,
		},
		type_:   type_,
		pointer: pointer,
		indices: indices,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Br(condition Value, true *Block, false *Block) Instruction {
	i := &br{
		instruction: instruction{
			module:   b.module,
			location: -1,
		},
		condition: condition,
		true:      true,
		false:     false,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Phi(firstValue Value, firstBlock *Block, secondValue Value, secondBlock *Block) InstructionValue {
	i := &phi{
		instruction: instruction{
			module:   b.module,
			type_:    firstValue.Type(),
			location: -1,
		},
		firstValue:  firstValue,
		firstBlock:  firstBlock,
		secondValue: secondValue,
		secondBlock: secondBlock,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Call(value Value, arguments []Value, returns Type) InstructionValue {
	i := &call{
		instruction: instruction{
			module:   b.module,
			type_:    returns,
			location: -1,
		},
		value:     value,
		arguments: arguments,
	}

	b.instructions = append(b.instructions, i)
	return i
}

func (b *Block) Ret(value Value) Instruction {
	i := &ret{
		instruction: instruction{
			module:   b.module,
			location: -1,
		},
		value: value,
	}

	b.instructions = append(b.instructions, i)
	return i
}
