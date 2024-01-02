package fuckoff

import (
	"fireball/core/ast"
)

type Resolver interface {
	GetType(name string) (ast.Type, string)

	GetFunction(name string) (*ast.Func, string)

	GetMethod(type_ ast.Type, name string, static bool) (*ast.Func, string)
}
