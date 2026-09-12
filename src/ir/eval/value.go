package eval

import (
	"fireball/ir"
	"math"
)

type Value interface {
	isValue()
}

type IntValue struct {
	Value uint64
}

func (i *IntValue) isValue() {}

type FloatValue struct {
	Value float64
}

func (f *FloatValue) isValue() {}

type StringValue struct {
	Value *ir.String
}

func (s *StringValue) isValue() {}

type NullValue struct{}

func (n *NullValue) isValue() {}

type AggregateValue struct {
	Values []Value
}

func (a *AggregateValue) isValue() {}

type GlobalValue struct {
	Name  string
	Typ   ir.Type
	Flags ir.GlobalVarFlags
	Data  any
}

func (g *GlobalValue) isValue() {}

type FuncValue struct {
	Name      string
	Signature *ir.Signature
	Data      any
}

func (f *FuncValue) isValue() {}

func (e *Evaluator) GetValue(reg Register, typ ir.Type) Value {
	switch typ := typ.(type) {
	case *ir.SimpleType:
		switch typ.Kind {
		case ir.VoidKind:
			return nil

		case ir.FloatKind:
			return &FloatValue{Value: float64(math.Float32frombits(uint32(reg.Scalar)))}

		case ir.DoubleKind:
			return &FloatValue{Value: math.Float64frombits(reg.Scalar)}

		case ir.PointerKind:
			if reg.Scalar == 0 {
				return &NullValue{}
			}

			if reg.Scalar&FuncAddrFlag != 0 {
				if f, ok := e.funcByAddr[reg.Scalar]; ok {
					return &FuncValue{
						Name:      f.Name,
						Signature: f.Signature,
						Data:      f.Data,
					}
				}

				panic("ir.eval.Evaluator.GetValue() - Invalid function pointer")
			}

			if g, ok := e.globalByAddr[reg.Scalar]; ok {
				switch init := g.Initializer.(type) {
				case *ir.String:
					return &StringValue{Value: init}

				default:
					return &GlobalValue{
						Name:  g.Name,
						Typ:   g.Type(),
						Flags: g.Flags,
						Data:  g.Data,
					}
				}
			}

			panic("ir.eval.Evaluator.GetValue() - Invalid pointer")

		default:
			panic("ir.eval.Evaluator.GetValue() - Invalid SimpleType kind")
		}

	case *ir.IntegerType:
		return &IntValue{Value: reg.Scalar}

	case *ir.VectorType:
		values := make([]Value, len(reg.Aggregate))

		for i, register := range reg.Aggregate {
			values[i] = e.GetValue(register, typ.Element)
		}

		return &AggregateValue{Values: values}

	case *ir.ArrayType:
		values := make([]Value, len(reg.Aggregate))

		for i, register := range reg.Aggregate {
			values[i] = e.GetValue(register, typ.Element)
		}

		return &AggregateValue{Values: values}

	case *ir.StructType:
		values := make([]Value, len(reg.Aggregate))

		for i, register := range reg.Aggregate {
			values[i] = e.GetValue(register, typ.Fields[i].Type)
		}

		return &AggregateValue{Values: values}

	case *ir.RefStructType:
		return e.GetValue(reg, &typ.Struct)

	default:
		panic("ir.eval.Evaluator.GetValue() - Invalid type")
	}
}
