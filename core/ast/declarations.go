package ast

import "fireball/core/scanner"
import "fireball/core/types"

//go:generate go run ../../gen/ast.go

type DeclVisitor interface {
	VisitFunc(decl *Func)
}

type Decl interface {
	Node

	Accept(visitor DeclVisitor)
}

type Func struct {
	Name    scanner.Token
	Params  []Param
	Returns types.Type
	Body    []Stmt
}

func (f *Func) Token() scanner.Token {
	return f.Name
}

func (f *Func) Accept(visitor DeclVisitor) {
	visitor.VisitFunc(f)
}

type Param struct {
	Name scanner.Token
	Type types.Type
}
