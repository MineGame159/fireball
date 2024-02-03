package ast

import (
	"fireball/core/cst"
	"fireball/core/scanner"
)

// Visitor

type TypeVisitor interface {
	VisitPrimitive(type_ *Primitive)
	VisitPointer(type_ *Pointer)
	VisitArray(type_ *Array)
	VisitResolvable(type_ *Resolvable)
	VisitStruct(type_ *Struct)
	VisitEnum(type_ *Enum)
	VisitInterface(type_ *Interface)
	VisitFunc(type_ *Func)
}

type Type interface {
	Node

	Equals(other Type) bool

	Resolved() Type

	AcceptType(visitor TypeVisitor)
}

// Primitive

type Primitive struct {
	cst    cst.Node
	parent Node

	Kind   PrimitiveKind
	Token_ scanner.Token
}

func NewPrimitive(node cst.Node, kind PrimitiveKind, token scanner.Token) *Primitive {
	if kind == 0 && token.IsEmpty() {
		return nil
	}

	p := &Primitive{
		cst:    node,
		Kind:   kind,
		Token_: token,
	}

	return p
}

func (p *Primitive) Cst() *cst.Node {
	if p.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &p.cst
}

func (p *Primitive) Token() scanner.Token {
	return p.Token_
}

func (p *Primitive) Parent() Node {
	return p.parent
}

func (p *Primitive) SetParent(parent Node) {
	if parent != nil && p.parent != nil {
		panic("ast.Primitive.SetParent() - Parent is already set")
	}

	p.parent = parent
}

func (p *Primitive) AcceptChildren(visitor Visitor) {
}

func (p *Primitive) Clone() Node {
	p2 := &Primitive{
		cst:    p.cst,
		Kind:   p.Kind,
		Token_: p.Token_,
	}

	return p2
}

func (p *Primitive) String() string {
	return p.Token_.String()
}

func (p *Primitive) AcceptType(visitor TypeVisitor) {
	visitor.VisitPrimitive(p)
}

func (p *Primitive) Resolved() Type {
	return p
}

// Pointer

type Pointer struct {
	cst    cst.Node
	parent Node

	Pointee Type
}

func NewPointer(node cst.Node, pointee Type) *Pointer {
	if pointee == nil {
		return nil
	}

	p := &Pointer{
		cst:     node,
		Pointee: pointee,
	}

	if pointee != nil {
		pointee.SetParent(p)
	}

	return p
}

func (p *Pointer) Cst() *cst.Node {
	if p.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &p.cst
}

func (p *Pointer) Token() scanner.Token {
	return scanner.Token{}
}

func (p *Pointer) Parent() Node {
	return p.parent
}

func (p *Pointer) SetParent(parent Node) {
	if parent != nil && p.parent != nil {
		panic("ast.Pointer.SetParent() - Parent is already set")
	}

	p.parent = parent
}

func (p *Pointer) AcceptChildren(visitor Visitor) {
	if p.Pointee != nil {
		visitor.VisitNode(p.Pointee)
	}
}

func (p *Pointer) Clone() Node {
	p2 := &Pointer{
		cst: p.cst,
	}

	if p.Pointee != nil {
		p2.Pointee = p.Pointee.Clone().(Type)
		p2.Pointee.SetParent(p2)
	}

	return p2
}

func (p *Pointer) String() string {
	return ""
}

func (p *Pointer) AcceptType(visitor TypeVisitor) {
	visitor.VisitPointer(p)
}

func (p *Pointer) Resolved() Type {
	return p
}

// Array

type Array struct {
	cst    cst.Node
	parent Node

	Base  Type
	Count uint32
}

func NewArray(node cst.Node, base Type, count uint32) *Array {
	if base == nil {
		return nil
	}

	a := &Array{
		cst:   node,
		Base:  base,
		Count: count,
	}

	if base != nil {
		base.SetParent(a)
	}

	return a
}

func (a *Array) Cst() *cst.Node {
	if a.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &a.cst
}

func (a *Array) Token() scanner.Token {
	return scanner.Token{}
}

func (a *Array) Parent() Node {
	return a.parent
}

func (a *Array) SetParent(parent Node) {
	if parent != nil && a.parent != nil {
		panic("ast.Array.SetParent() - Parent is already set")
	}

	a.parent = parent
}

func (a *Array) AcceptChildren(visitor Visitor) {
	if a.Base != nil {
		visitor.VisitNode(a.Base)
	}
}

func (a *Array) Clone() Node {
	a2 := &Array{
		cst:   a.cst,
		Count: a.Count,
	}

	if a.Base != nil {
		a2.Base = a.Base.Clone().(Type)
		a2.Base.SetParent(a2)
	}

	return a2
}

func (a *Array) String() string {
	return ""
}

func (a *Array) AcceptType(visitor TypeVisitor) {
	visitor.VisitArray(a)
}

func (a *Array) Resolved() Type {
	return a
}

// Resolvable

type Resolvable struct {
	cst    cst.Node
	parent Node

	Parts []*Token
	Type  Type
}

func NewResolvable(node cst.Node, parts []*Token) *Resolvable {
	if parts == nil {
		return nil
	}

	r := &Resolvable{
		cst:   node,
		Parts: parts,
	}

	for _, child := range parts {
		child.SetParent(r)
	}

	return r
}

func (r *Resolvable) Cst() *cst.Node {
	if r.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &r.cst
}

func (r *Resolvable) Token() scanner.Token {
	return scanner.Token{}
}

func (r *Resolvable) Parent() Node {
	return r.parent
}

func (r *Resolvable) SetParent(parent Node) {
	if parent != nil && r.parent != nil {
		panic("ast.Resolvable.SetParent() - Parent is already set")
	}

	r.parent = parent
}

func (r *Resolvable) AcceptChildren(visitor Visitor) {
	for _, child := range r.Parts {
		visitor.VisitNode(child)
	}
}

func (r *Resolvable) Clone() Node {
	r2 := &Resolvable{
		cst:  r.cst,
		Type: r.Type,
	}

	r2.Parts = make([]*Token, len(r.Parts))
	for i, child := range r2.Parts {
		r2.Parts[i] = child.Clone().(*Token)
		r2.Parts[i].SetParent(r2)
	}
	if r.Type != nil {
		r2.Type = r.Type.Clone().(Type)
		r2.Type.SetParent(r2)
	}

	return r2
}

func (r *Resolvable) String() string {
	return ""
}

func (r *Resolvable) AcceptType(visitor TypeVisitor) {
	visitor.VisitResolvable(r)
}

func (r *Resolvable) Resolved() Type {
	return r.Type
}
