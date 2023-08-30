package types

import "fireball/core/scanner"

type UnresolvedType struct {
	Identifier scanner.Token
}

func (u *UnresolvedType) Size() int {
	return 0
}

func (u *UnresolvedType) CanAssignTo(other Type) bool {
	return false
}

func (u *UnresolvedType) AcceptTypes(visitor Visitor) {}

func (u *UnresolvedType) String() string {
	return u.Identifier.Lexeme
}
