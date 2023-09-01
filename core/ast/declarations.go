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

// Struct

type Struct struct {
	range_ Range

	Name   scanner.Token
	Fields []Field
	Type   types.Type
}

func (s *Struct) Token() scanner.Token {
	return s.Name
}

func (s *Struct) Range() Range {
	return s.range_
}

func (s *Struct) SetRangeToken(start, end scanner.Token) {
	s.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (s *Struct) SetRangePos(start, end Pos) {
	s.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (s *Struct) SetRangeNode(start, end Node) {
	s.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
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

func (s *Struct) Leaf() bool {
	return true
}

// Field

type Field struct {
	Name scanner.Token
	Type types.Type
}

// Func

type Func struct {
	range_ Range

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

func (f *Func) Range() Range {
	return f.range_
}

func (f *Func) SetRangeToken(start, end scanner.Token) {
	f.range_ = Range{
		Start: TokenToPos(start, false),
		End:   TokenToPos(end, true),
	}
}

func (f *Func) SetRangePos(start, end Pos) {
	f.range_ = Range{
		Start: start,
		End:   end,
	}
}

func (f *Func) SetRangeNode(start, end Node) {
	f.range_ = Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
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

func (f *Func) Leaf() bool {
	return false
}

// Param

type Param struct {
	Name scanner.Token
	Type types.Type
}
