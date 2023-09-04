package utils

import "fireball/core/types"

type Resolver interface {
	GetType(name string) types.Type

	GetFunction(name string) (*types.FunctionType, string)
}
