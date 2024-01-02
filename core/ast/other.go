package ast

import (
	"fireball/core/cst"
	"fireball/core/scanner"
)

// File

type File struct {
	cst    cst.Node
	parent Node

	Path  string
	Decls []Decl
}

func NewFile(node cst.Node, path string, decls []Decl) *File {
	f := &File{
		cst:   node,
		Path:  path,
		Decls: decls,
	}

	for _, child := range decls {
		child.SetParent(f)
	}

	return f
}

func (f *File) Cst() *cst.Node {
	if f.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &f.cst
}

func (f *File) Token() scanner.Token {
	return scanner.Token{}
}

func (f *File) Parent() Node {
	return f.parent
}

func (f *File) SetParent(parent Node) {
	if parent != nil && f.parent != nil {
		panic("ast.File.SetParent() - Parent is already set")
	}

	f.parent = parent
}

func (f *File) AcceptChildren(visitor Visitor) {
	for _, child := range f.Decls {
		visitor.VisitNode(child)
	}
}

func (f *File) String() string {
	return ""
}

// Field

type Field struct {
	cst    cst.Node
	parent Node

	Name *Token
	Type Type
}

func NewField(node cst.Node, name *Token, type_ Type) *Field {
	f := &Field{
		cst:  node,
		Name: name,
		Type: type_,
	}

	if name != nil {
		name.SetParent(f)
	}
	if type_ != nil {
		type_.SetParent(f)
	}

	return f
}

func (f *Field) Cst() *cst.Node {
	if f.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &f.cst
}

func (f *Field) Token() scanner.Token {
	return scanner.Token{}
}

func (f *Field) Parent() Node {
	return f.parent
}

func (f *Field) SetParent(parent Node) {
	if parent != nil && f.parent != nil {
		panic("ast.Field.SetParent() - Parent is already set")
	}

	f.parent = parent
}

func (f *Field) AcceptChildren(visitor Visitor) {
	if f.Name != nil {
		visitor.VisitNode(f.Name)
	}
	if f.Type != nil {
		visitor.VisitNode(f.Type)
	}
}

func (f *Field) String() string {
	return ""
}

// InitField

type InitField struct {
	cst    cst.Node
	parent Node

	Name  *Token
	Value Expr
}

func NewInitField(node cst.Node, name *Token, value Expr) *InitField {
	i := &InitField{
		cst:   node,
		Name:  name,
		Value: value,
	}

	if name != nil {
		name.SetParent(i)
	}
	if value != nil {
		value.SetParent(i)
	}

	return i
}

func (i *InitField) Cst() *cst.Node {
	if i.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &i.cst
}

func (i *InitField) Token() scanner.Token {
	return scanner.Token{}
}

func (i *InitField) Parent() Node {
	return i.parent
}

func (i *InitField) SetParent(parent Node) {
	if parent != nil && i.parent != nil {
		panic("ast.InitField.SetParent() - Parent is already set")
	}

	i.parent = parent
}

func (i *InitField) AcceptChildren(visitor Visitor) {
	if i.Name != nil {
		visitor.VisitNode(i.Name)
	}
	if i.Value != nil {
		visitor.VisitNode(i.Value)
	}
}

func (i *InitField) String() string {
	return ""
}

// EnumCase

type EnumCase struct {
	cst    cst.Node
	parent Node

	Name        *Token
	Value       *Token
	ActualValue int64
}

func NewEnumCase(node cst.Node, name *Token, value *Token) *EnumCase {
	e := &EnumCase{
		cst:   node,
		Name:  name,
		Value: value,
	}

	if name != nil {
		name.SetParent(e)
	}
	if value != nil {
		value.SetParent(e)
	}

	return e
}

func (e *EnumCase) Cst() *cst.Node {
	if e.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &e.cst
}

func (e *EnumCase) Token() scanner.Token {
	return scanner.Token{}
}

func (e *EnumCase) Parent() Node {
	return e.parent
}

func (e *EnumCase) SetParent(parent Node) {
	if parent != nil && e.parent != nil {
		panic("ast.EnumCase.SetParent() - Parent is already set")
	}

	e.parent = parent
}

func (e *EnumCase) AcceptChildren(visitor Visitor) {
	if e.Name != nil {
		visitor.VisitNode(e.Name)
	}
	if e.Value != nil {
		visitor.VisitNode(e.Value)
	}
}

func (e *EnumCase) String() string {
	return ""
}

// Param

type Param struct {
	cst    cst.Node
	parent Node

	Name *Token
	Type Type
}

func NewParam(node cst.Node, name *Token, type_ Type) *Param {
	p := &Param{
		cst:  node,
		Name: name,
		Type: type_,
	}

	if name != nil {
		name.SetParent(p)
	}
	if type_ != nil {
		type_.SetParent(p)
	}

	return p
}

func (p *Param) Cst() *cst.Node {
	if p.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &p.cst
}

func (p *Param) Token() scanner.Token {
	return scanner.Token{}
}

func (p *Param) Parent() Node {
	return p.parent
}

func (p *Param) SetParent(parent Node) {
	if parent != nil && p.parent != nil {
		panic("ast.Param.SetParent() - Parent is already set")
	}

	p.parent = parent
}

func (p *Param) AcceptChildren(visitor Visitor) {
	if p.Name != nil {
		visitor.VisitNode(p.Name)
	}
	if p.Type != nil {
		visitor.VisitNode(p.Type)
	}
}

func (p *Param) String() string {
	return ""
}

// Token

type Token struct {
	cst    cst.Node
	parent Node

	Token_ scanner.Token
}

func NewToken(node cst.Node, token scanner.Token) *Token {
	t := &Token{
		cst:    node,
		Token_: token,
	}

	return t
}

func (t *Token) Cst() *cst.Node {
	if t.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &t.cst
}

func (t *Token) Token() scanner.Token {
	return t.Token_
}

func (t *Token) Parent() Node {
	return t.parent
}

func (t *Token) SetParent(parent Node) {
	if parent != nil && t.parent != nil {
		panic("ast.Token.SetParent() - Parent is already set")
	}

	t.parent = parent
}

func (t *Token) AcceptChildren(visitor Visitor) {
}

func (t *Token) String() string {
	return t.Token_.String()
}
