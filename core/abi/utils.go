package abi

var i1 = Arg{Class: Integer, Bits: 1}

var i8 = Arg{Class: Integer, Bits: 8}
var i16 = Arg{Class: Integer, Bits: 16}
var i32 = Arg{Class: Integer, Bits: 32}
var i64 = Arg{Class: Integer, Bits: 64}

var f32 = Arg{Class: SSE, Bits: 32}
var f64 = Arg{Class: SSE, Bits: 64}

var ptr = Arg{Class: Integer, Bits: 64}
var memory = Arg{Class: Memory, Bits: 64}

func alignBytes(bytes, align uint32) uint32 {
	if bytes%align != 0 {
		bytes += align - (bytes % align)
	}

	return bytes
}

func getSize(args []Arg) uint32 {
	length := uint32(len(args))

	if length == 0 {
		return 0
	}

	return (length-1)*8 + args[length-1].Bytes()
}

func getArg(args []Arg, offset uint32, arg **Arg) []Arg {
	i := offset / 8

	for i >= uint32(len(args)) {
		args = append(args, Arg{})
	}

	*arg = &args[i]

	return args
}
