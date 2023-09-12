package types

import (
	"fireball/core"
	"fireball/core/scanner"
)

type UnresolvedType struct {
	range_     core.Range
	Identifier scanner.Token
}

func Unresolved(identifier scanner.Token, range_ core.Range) *UnresolvedType {
	return &UnresolvedType{
		range_:     range_,
		Identifier: identifier,
	}
}

func (u *UnresolvedType) Range() core.Range {
	return u.range_
}

func (u *UnresolvedType) Size() int {
	return 0
}

func (u *UnresolvedType) WithRange(range_ core.Range) Type {
	return &UnresolvedType{
		range_:     range_,
		Identifier: u.Identifier,
	}
}

func (u *UnresolvedType) Equals(other Type) bool {
	return false
}

func (u *UnresolvedType) CanAssignTo(other Type) bool {
	return false
}

func (u *UnresolvedType) AcceptTypes(visitor Visitor) {}

func (u *UnresolvedType) AcceptTypesPtr(visitor PtrVisitor) {}

func (u *UnresolvedType) String() string {
	return u.Identifier.Lexeme
}
