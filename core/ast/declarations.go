package ast

import (
	"fireball/core/cst"
	"fireball/core/scanner"
)

// Visitor

type DeclVisitor interface {
	VisitStruct(decl *Struct)
	VisitEnum(decl *Enum)
	VisitImpl(decl *Impl)
	VisitFunc(decl *Func)
}

type Decl interface {
	Node

	AcceptDecl(visitor DeclVisitor)
}

// Struct

type Struct struct {
	cst    cst.Node
	parent Node

	Name         *Token
	Fields       []*Field
	StaticFields []*Field
}

func NewStruct(node cst.Node, name *Token, fields []*Field, staticfields []*Field) *Struct {
	if name == nil && fields == nil && staticfields == nil {
		return nil
	}

	s := &Struct{
		cst:          node,
		Name:         name,
		Fields:       fields,
		StaticFields: staticfields,
	}

	if name != nil {
		name.SetParent(s)
	}
	for _, child := range fields {
		child.SetParent(s)
	}
	for _, child := range staticfields {
		child.SetParent(s)
	}

	return s
}

func (s *Struct) Cst() *cst.Node {
	if s.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &s.cst
}

func (s *Struct) Token() scanner.Token {
	return scanner.Token{}
}

func (s *Struct) Parent() Node {
	return s.parent
}

func (s *Struct) SetParent(parent Node) {
	if parent != nil && s.parent != nil {
		panic("ast.Struct.SetParent() - Parent is already set")
	}

	s.parent = parent
}

func (s *Struct) AcceptChildren(visitor Visitor) {
	if s.Name != nil {
		visitor.VisitNode(s.Name)
	}
	for _, child := range s.Fields {
		visitor.VisitNode(child)
	}
	for _, child := range s.StaticFields {
		visitor.VisitNode(child)
	}
}

func (s *Struct) String() string {
	return ""
}

func (s *Struct) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitStruct(s)
}

// Enum

type Enum struct {
	cst    cst.Node
	parent Node

	Name       *Token
	Type       Type
	ActualType Type
	Cases      []*EnumCase
}

func NewEnum(node cst.Node, name *Token, type_ Type, cases []*EnumCase) *Enum {
	if name == nil && type_ == nil && cases == nil {
		return nil
	}

	e := &Enum{
		cst:   node,
		Name:  name,
		Type:  type_,
		Cases: cases,
	}

	if name != nil {
		name.SetParent(e)
	}
	if type_ != nil {
		type_.SetParent(e)
	}
	for _, child := range cases {
		child.SetParent(e)
	}

	return e
}

func (e *Enum) Cst() *cst.Node {
	if e.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &e.cst
}

func (e *Enum) Token() scanner.Token {
	return scanner.Token{}
}

func (e *Enum) Parent() Node {
	return e.parent
}

func (e *Enum) SetParent(parent Node) {
	if parent != nil && e.parent != nil {
		panic("ast.Enum.SetParent() - Parent is already set")
	}

	e.parent = parent
}

func (e *Enum) AcceptChildren(visitor Visitor) {
	if e.Name != nil {
		visitor.VisitNode(e.Name)
	}
	if e.Type != nil {
		visitor.VisitNode(e.Type)
	}
	for _, child := range e.Cases {
		visitor.VisitNode(child)
	}
}

func (e *Enum) String() string {
	return ""
}

func (e *Enum) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitEnum(e)
}

// Impl

type Impl struct {
	cst    cst.Node
	parent Node

	Struct  *Token
	Type    Type
	Methods []*Func
}

func NewImpl(node cst.Node, struct_ *Token, methods []*Func) *Impl {
	if struct_ == nil && methods == nil {
		return nil
	}

	i := &Impl{
		cst:     node,
		Struct:  struct_,
		Methods: methods,
	}

	if struct_ != nil {
		struct_.SetParent(i)
	}
	for _, child := range methods {
		child.SetParent(i)
	}

	return i
}

func (i *Impl) Cst() *cst.Node {
	if i.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &i.cst
}

func (i *Impl) Token() scanner.Token {
	return scanner.Token{}
}

func (i *Impl) Parent() Node {
	return i.parent
}

func (i *Impl) SetParent(parent Node) {
	if parent != nil && i.parent != nil {
		panic("ast.Impl.SetParent() - Parent is already set")
	}

	i.parent = parent
}

func (i *Impl) AcceptChildren(visitor Visitor) {
	if i.Struct != nil {
		visitor.VisitNode(i.Struct)
	}
	for _, child := range i.Methods {
		visitor.VisitNode(child)
	}
}

func (i *Impl) String() string {
	return ""
}

func (i *Impl) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitImpl(i)
}

// Func

type Func struct {
	cst    cst.Node
	parent Node

	Attributes []*Attribute
	Flags      FuncFlags
	Name       *Token
	Params     []*Param
	Returns    Type
	Body       []Stmt
}

func NewFunc(node cst.Node, attributes []*Attribute, flags FuncFlags, name *Token, params []*Param, returns Type, body []Stmt) *Func {
	if attributes == nil && flags == 0 && name == nil && params == nil && returns == nil && body == nil {
		return nil
	}

	f := &Func{
		cst:        node,
		Attributes: attributes,
		Flags:      flags,
		Name:       name,
		Params:     params,
		Returns:    returns,
		Body:       body,
	}

	for _, child := range attributes {
		child.SetParent(f)
	}
	if name != nil {
		name.SetParent(f)
	}
	for _, child := range params {
		child.SetParent(f)
	}
	if returns != nil {
		returns.SetParent(f)
	}
	for _, child := range body {
		child.SetParent(f)
	}

	return f
}

func (f *Func) Cst() *cst.Node {
	if f.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &f.cst
}

func (f *Func) Token() scanner.Token {
	return scanner.Token{}
}

func (f *Func) Parent() Node {
	return f.parent
}

func (f *Func) SetParent(parent Node) {
	if parent != nil && f.parent != nil {
		panic("ast.Func.SetParent() - Parent is already set")
	}

	f.parent = parent
}

func (f *Func) AcceptChildren(visitor Visitor) {
	for _, child := range f.Attributes {
		visitor.VisitNode(child)
	}
	if f.Name != nil {
		visitor.VisitNode(f.Name)
	}
	for _, child := range f.Params {
		visitor.VisitNode(child)
	}
	if f.Returns != nil {
		visitor.VisitNode(f.Returns)
	}
	for _, child := range f.Body {
		visitor.VisitNode(child)
	}
}

func (f *Func) String() string {
	return ""
}

func (f *Func) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitFunc(f)
}
