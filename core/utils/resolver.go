package utils

import (
	"fireball/core/ast"
	"fireball/core/types"
)

type Resolver interface {
	GetType(name string) (types.Type, string)

	GetFunction(name string) (*ast.Func, string)

	GetMethod(type_ types.Type, name string) (*ast.Func, string)
}
