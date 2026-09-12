package eval

import "math"

const FuncAddrFlag uint64 = 1 << 63

type Register struct {
	Scalar    uint64
	Aggregate []Register
}

func Float32Reg(val float32) Register {
	return Register{Scalar: uint64(math.Float32bits(val))}
}

func Float64Reg(val float64) Register {
	return Register{Scalar: math.Float64bits(val)}
}

func (r Register) Float32() float32 {
	return math.Float32frombits(uint32(r.Scalar))
}

func (r Register) Float64() float64 {
	return math.Float64frombits(r.Scalar)
}

func (r Register) ShallowClone() Register {
	if r.Aggregate == nil {
		return Register{Scalar: r.Scalar}
	}

	aggregate := make([]Register, len(r.Aggregate))

	for i, reg := range r.Aggregate {
		aggregate[i] = reg
	}

	return Register{Aggregate: aggregate}
}
