package fuckoff

import (
	"fireball/core/ast"
)

type Resolver interface {
	GetType(name string) ast.Type

	GetFunction(name string) *ast.Func

	GetMethod(type_ ast.Type, name string, static bool) *ast.Func
	GetMethods(type_ ast.Type, static bool) []*ast.Func

	GetFileNodes() []*ast.File
}
