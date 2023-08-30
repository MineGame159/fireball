package ast

import "fireball/core/scanner"
import "fireball/core/types"

//go:generate go run ../../gen/ast.go

type DeclVisitor interface {
	VisitStruct(decl *Struct)
	VisitFunc(decl *Func)
}

type Decl interface {
	Node

	Accept(visitor DeclVisitor)
}

type Struct struct {
	Range Range

	Name   scanner.Token
	Fields []Field
	Type   types.Type
}

func (s *Struct) Token() scanner.Token {
	return s.Name
}

func (s *Struct) Accept(visitor DeclVisitor) {
	visitor.VisitStruct(s)
}

func (s *Struct) AcceptChildren(acceptor Acceptor) {
}

func (s *Struct) AcceptTypes(visitor types.Visitor) {
	for i := range s.Fields {
		visitor.VisitType(&s.Fields[i].Type)
	}
	visitor.VisitType(&s.Type)
}

type Field struct {
	Name scanner.Token
	Type types.Type
}

type Func struct {
	Range Range

	Extern   bool
	Name     scanner.Token
	Params   []Param
	Variadic bool
	Returns  types.Type
	Body     []Stmt
}

func (f *Func) Token() scanner.Token {
	return f.Name
}

func (f *Func) Accept(visitor DeclVisitor) {
	visitor.VisitFunc(f)
}

func (f *Func) AcceptChildren(acceptor Acceptor) {
	for _, v := range f.Body {
		acceptor.AcceptStmt(v)
	}
}

func (f *Func) AcceptTypes(visitor types.Visitor) {
	for i := range f.Params {
		visitor.VisitType(&f.Params[i].Type)
	}
	visitor.VisitType(&f.Returns)
}

type Param struct {
	Name scanner.Token
	Type types.Type
}
