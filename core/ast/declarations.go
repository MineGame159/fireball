package ast

import (
	"fireball/core"
	"log"
)
import "fireball/core/types"
import "fireball/core/scanner"

//go:generate go run ../../gen/ast.go

type DeclVisitor interface {
	VisitStruct(decl *Struct)
	VisitEnum(decl *Enum)
	VisitFunc(decl *Func)
}

type Decl interface {
	Node

	Accept(visitor DeclVisitor)
}

// Struct

type Struct struct {
	range_ core.Range
	parent Node

	Name   scanner.Token
	Fields []Field
	Type   types.Type
}

func (s *Struct) Token() scanner.Token {
	return s.Name
}

func (s *Struct) Range() core.Range {
	return s.range_
}

func (s *Struct) SetRangeToken(start, end scanner.Token) {
	s.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (s *Struct) SetRangePos(start, end core.Pos) {
	s.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (s *Struct) SetRangeNode(start, end Node) {
	s.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (s *Struct) Parent() Node {
	return s.parent
}

func (s *Struct) SetParent(parent Node) {
	if s.parent != nil && parent != nil {
		log.Fatalln("Struct.SetParent() - Node already has a parent")
	}
	s.parent = parent
}

func (s *Struct) Accept(visitor DeclVisitor) {
	visitor.VisitStruct(s)
}

func (s *Struct) AcceptChildren(visitor Acceptor) {
}

func (s *Struct) AcceptTypes(visitor types.Visitor) {
	for i_ := range s.Fields {
		if s.Fields[i_].Type != nil {
			visitor.VisitType(s.Fields[i_].Type)
		}
	}
	if s.Type != nil {
		visitor.VisitType(s.Type)
	}
}

func (s *Struct) AcceptTypesPtr(visitor types.PtrVisitor) {
	for i_ := range s.Fields {
		visitor.VisitType(&s.Fields[i_].Type)
	}
	visitor.VisitType(&s.Type)
}

func (s *Struct) Leaf() bool {
	return true
}

func (s *Struct) SetChildrenParent() {
}

// Field

type Field struct {
	Name scanner.Token
	Type types.Type
}

// Enum

type Enum struct {
	range_ core.Range
	parent Node

	Name      scanner.Token
	Type      types.Type
	InferType bool
	Cases     []EnumCase
}

func (e *Enum) Token() scanner.Token {
	return e.Name
}

func (e *Enum) Range() core.Range {
	return e.range_
}

func (e *Enum) SetRangeToken(start, end scanner.Token) {
	e.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (e *Enum) SetRangePos(start, end core.Pos) {
	e.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (e *Enum) SetRangeNode(start, end Node) {
	e.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (e *Enum) Parent() Node {
	return e.parent
}

func (e *Enum) SetParent(parent Node) {
	if e.parent != nil && parent != nil {
		log.Fatalln("Enum.SetParent() - Node already has a parent")
	}
	e.parent = parent
}

func (e *Enum) Accept(visitor DeclVisitor) {
	visitor.VisitEnum(e)
}

func (e *Enum) AcceptChildren(visitor Acceptor) {
}

func (e *Enum) AcceptTypes(visitor types.Visitor) {
	if e.Type != nil {
		visitor.VisitType(e.Type)
	}
}

func (e *Enum) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&e.Type)
}

func (e *Enum) Leaf() bool {
	return true
}

func (e *Enum) SetChildrenParent() {
}

// EnumCase

type EnumCase struct {
	Name       scanner.Token
	Value      int
	InferValue bool
}

// Func

type Func struct {
	range_ core.Range
	parent Node

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

func (f *Func) Range() core.Range {
	return f.range_
}

func (f *Func) SetRangeToken(start, end scanner.Token) {
	f.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (f *Func) SetRangePos(start, end core.Pos) {
	f.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (f *Func) SetRangeNode(start, end Node) {
	f.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (f *Func) Parent() Node {
	return f.parent
}

func (f *Func) SetParent(parent Node) {
	if f.parent != nil && parent != nil {
		log.Fatalln("Func.SetParent() - Node already has a parent")
	}
	f.parent = parent
}

func (f *Func) Accept(visitor DeclVisitor) {
	visitor.VisitFunc(f)
}

func (f *Func) AcceptChildren(visitor Acceptor) {
	for i_ := range f.Body {
		if f.Body[i_] != nil {
			visitor.AcceptStmt(f.Body[i_])
		}
	}
}

func (f *Func) AcceptTypes(visitor types.Visitor) {
	for i_ := range f.Params {
		if f.Params[i_].Type != nil {
			visitor.VisitType(f.Params[i_].Type)
		}
	}
	if f.Returns != nil {
		visitor.VisitType(f.Returns)
	}
}

func (f *Func) AcceptTypesPtr(visitor types.PtrVisitor) {
	for i_ := range f.Params {
		visitor.VisitType(&f.Params[i_].Type)
	}
	visitor.VisitType(&f.Returns)
}

func (f *Func) Leaf() bool {
	return false
}

func (f *Func) SetChildrenParent() {
	for i_ := range f.Body {
		if f.Body[i_] != nil {
			f.Body[i_].SetParent(f)
		}
	}
}

// Param

type Param struct {
	Name scanner.Token
	Type types.Type
}
