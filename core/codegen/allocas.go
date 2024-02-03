package codegen

import (
	"fireball/core/abi"
	"fireball/core/ast"
	"fireball/core/ir"
)

type allocas struct {
	c *codegen

	count int
}

func (a *allocas) reset() {
	a.count = 0
}

func (a *allocas) get(type_ ast.Type, name string) *ir.AllocaInst {
	alloca := &ir.AllocaInst{
		Typ:   a.c.types.get(type_),
		Align: abi.GetTargetAbi().Align(type_),
	}

	if name == "" {
		name = "temp"
	}
	alloca.SetName(name)

	a.c.function.Blocks[0].Insert(a.count, alloca)
	a.count++

	return alloca
}
