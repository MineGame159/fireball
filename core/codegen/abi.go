package codegen

import (
	"fireball/core/abi"
	"fireball/core/ast"
	"fireball/core/ir"
)

// Parameters

func (c *codegen) valueToParams(a abi.Abi, value exprValue, valueType ast.Type, values []ir.Value) []ir.Value {
	args := a.Classify(valueType, nil)
	if len(args) == 0 {
		panic("codegen.valueToParams() - Failed to classify parameter type")
	}

	// Memory
	if args[0].Class == abi.Memory {
		value = c.toAddressable(value, valueType)
		return append(values, value.v)
	}

	// Struct
	if typeIsAbiStruct(valueType) {
		value = c.toAddressable(value, valueType)
		typ := getAbiStructType(args, false)

		for i, arg := range args {
			if arg.Class == abi.None {
				continue
			}

			switch arg.Class {
			case abi.Integer, abi.SSE:
				pointer := c.block.Add(&ir.GetElementPtrInst{
					PointerTyp: &ir.PointerType{Pointee: typ.Fields[i]},
					Typ:        typ,
					Pointer:    value.v,
					Indices: []ir.Value{
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)},
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(uint64(i))},
					},
					Inbounds: true,
				})

				values = append(values, c.block.Add(&ir.LoadInst{
					Typ:     typ.Fields[i],
					Pointer: pointer,
				}))

			default:
				panic("codegen.valueToParams() - Invalid struct argument class")
			}
		}

		return values
	}

	// Single
	value = c.load(value, valueType)
	return append(values, value.v)
}

func (c *codegen) paramsToVariable(args []abi.Arg, values []*ir.Param, variable ir.Value, variableType ast.Type) {
	if len(args) == 0 {
		panic("codegen.paramsToVariable() - Failed to classify parameter type")
	}

	if len(args) != len(values) {
		panic("codegen.paramsToVariable() - Parameter count does not match abi argument count")
	}

	// Memory
	if args[0].Class == abi.Memory {
		align := abi.GetTargetAbi().Align(variableType)

		value := c.block.Add(&ir.LoadInst{
			Typ:     c.types.get(variableType),
			Pointer: values[0],
			Align:   align,
		})

		c.block.Add(&ir.StoreInst{
			Pointer: variable,
			Value:   value,
			Align:   align,
		})

		return
	}

	// Struct
	if typeIsAbiStruct(variableType) {
		typ := getAbiStructType(args, false)
		valueI := 0

		for i, arg := range args {
			if arg.Class == abi.None {
				continue
			}

			switch arg.Class {
			case abi.Integer, abi.SSE:
				pointer := c.block.Add(&ir.GetElementPtrInst{
					PointerTyp: &ir.PointerType{Pointee: typ.Fields[i]},
					Typ:        typ,
					Pointer:    variable,
					Indices: []ir.Value{
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)},
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(uint64(i))},
					},
					Inbounds: true,
				})

				c.block.Add(&ir.StoreInst{
					Pointer: pointer,
					Value:   values[valueI],
				})

			default:
				panic("codegen.paramsToVariable() - Invalid struct argument class")
			}

			valueI++
		}

		return
	}

	// Single
	c.block.Add(&ir.StoreInst{
		Pointer: variable,
		Value:   values[0],
		Align:   abi.GetTargetAbi().Align(variableType),
	})
}

// Return values

func (c *codegen) valueToReturnValue(a abi.Abi, value exprValue, returnType ast.Type, params []*ir.Param) exprValue {
	args := a.Classify(returnType, nil)
	if len(args) == 0 {
		panic("codegen.valueToReturnValue() - Failed to classify parameter type")
	}

	// Memory
	if args[0].Class == abi.Memory {
		value = c.load(value, returnType)

		c.block.Add(&ir.StoreInst{
			Pointer: params[0],
			Value:   value.v,
			Align:   abi.GetTargetAbi().Align(returnType),
		})

		return exprValue{}
	}

	// Struct
	if typeIsAbiStruct(returnType) {
		value = c.toAddressable(value, returnType)

		typWithEmpty := getAbiStructType(args, false)
		typWithoutEmpty := getAbiStructType(args, true)

		var result ir.Value = &ir.ZeroInitConst{Typ: typWithoutEmpty}
		elementI := 0

		for i, arg := range args {
			if arg.Class == abi.None {
				continue
			}

			switch arg.Class {
			case abi.Integer, abi.SSE:
				pointer := c.block.Add(&ir.GetElementPtrInst{
					PointerTyp: &ir.PointerType{Pointee: typWithEmpty.Fields[i]},
					Typ:        typWithEmpty,
					Pointer:    value.v,
					Indices: []ir.Value{
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)},
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(uint64(i))},
					},
					Inbounds: true,
				})

				element := c.block.Add(&ir.LoadInst{
					Typ:     typWithEmpty.Fields[i],
					Pointer: pointer,
				})

				result = c.block.Add(&ir.InsertValueInst{
					Value:   result,
					Element: element,
					Indices: []uint32{uint32(elementI)},
				})

			default:
				panic("codegen.valueToReturnValue() - Invalid struct argument class")
			}

			elementI++
		}

		return exprValue{v: result}
	}

	// Single
	return value
}

func (c *codegen) returnValueToValue(a abi.Abi, value exprValue, returnType ast.Type) exprValue {
	args := a.Classify(returnType, nil)
	if len(args) == 0 {
		panic("codegen.returnValueToValue() - Failed to classify parameter type")
	}

	// Memory
	if args[0].Class == abi.Memory {
		value := c.block.Add(&ir.LoadInst{
			Typ:     c.types.get(returnType),
			Pointer: value.v,
			Align:   abi.GetTargetAbi().Align(returnType),
		})

		return exprValue{v: value}
	}

	// Struct
	if typeIsAbiStruct(returnType) {
		result := c.allocas.get(returnType, "")
		typ := getAbiStructType(args, false)

		elementI := 0

		for i, arg := range args {
			if arg.Class == abi.None {
				continue
			}

			switch arg.Class {
			case abi.Integer, abi.SSE:
				element := c.block.Add(&ir.ExtractValueInst{
					Value:   value.v,
					Indices: []uint32{uint32(elementI)},
				})

				pointer := c.block.Add(&ir.GetElementPtrInst{
					PointerTyp: &ir.PointerType{Pointee: typ.Fields[i]},
					Typ:        typ,
					Pointer:    result,
					Indices: []ir.Value{
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)},
						&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(uint64(i))},
					},
					Inbounds: true,
				})

				c.block.Add(&ir.StoreInst{
					Pointer: pointer,
					Value:   element,
				})

			default:
				panic("codegen.returnValueToValue() - Invalid struct argument class")
			}

			elementI++
		}

		return exprValue{
			v:           result,
			addressable: true,
		}
	}

	// Single
	return value
}

// Helpers

func (c *codegen) toAddressable(value exprValue, valueType ast.Type) exprValue {
	if !value.addressable {
		pointer := c.allocas.get(valueType, "")

		c.block.Add(&ir.StoreInst{
			Pointer: pointer,
			Value:   value.v,
			Align:   abi.GetTargetAbi().Align(valueType),
		})

		return exprValue{
			v:           pointer,
			addressable: true,
		}
	}

	return value
}

// Type helpers

func typeIsAbiStruct(type_ ast.Type) bool {
	switch type_.Resolved().(type) {
	case *ast.Struct, *ast.Interface:
		return true
	default:
		return false
	}
}

func getAbiStructType(args []abi.Arg, skipEmpty bool) *ir.StructType {
	fields := make([]ir.Type, 0, len(args))

	for _, arg := range args {
		if arg.Class == abi.None && skipEmpty {
			continue
		}

		fields = append(fields, getAbiArgType(arg))
	}

	return &ir.StructType{
		Name:   "",
		Fields: fields,
	}
}

func getAbiArgType(arg abi.Arg) ir.Type {
	switch arg.Class {
	case abi.Integer:
		return getAbiIntIrType(arg)
	case abi.SSE:
		return getAbiSseIrType(arg)

	default:
		panic("codegen.getAbiArgType() - Invalid class")
	}
}

func getAbiIntIrType(arg abi.Arg) ir.Type {
	if arg.Bits == 1 {
		return ir.I1
	}
	if arg.Bits <= 8 {
		return ir.I8
	}
	if arg.Bits <= 16 {
		return ir.I16
	}
	if arg.Bits <= 32 {
		return ir.I32
	}
	if arg.Bits <= 64 {
		return ir.I64
	}

	panic("codegen.getAbiIntIrType() - Invalid size")
}

func getAbiSseIrType(arg abi.Arg) ir.Type {
	switch arg.Bits {
	case 32:
		return ir.F32
	case 64:
		return ir.F64

	default:
		panic("codegen.getAbiSseIrType() - Invalid size")
	}
}
