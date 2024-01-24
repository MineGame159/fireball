package ast

import (
	"fireball/core/cst"
	"fireball/core/scanner"
)

// File

type File struct {
	cst    cst.Node
	parent Node

	Path      string
	Namespace *Namespace
	Decls     []Decl
}

func NewFile(node cst.Node, path string, namespace *Namespace, decls []Decl) *File {
	f := &File{
		cst:       node,
		Path:      path,
		Namespace: namespace,
		Decls:     decls,
	}

	if namespace != nil {
		namespace.SetParent(f)
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
	if f.Namespace != nil {
		visitor.VisitNode(f.Namespace)
	}
	for _, child := range f.Decls {
		visitor.VisitNode(child)
	}
}

func (f *File) Clone() Node {
	f2 := &File{
		cst:  f.cst,
		Path: f.Path,
	}

	if f.Namespace != nil {
		f2.Namespace = f.Namespace.Clone().(*Namespace)
		f2.Namespace.SetParent(f2)
	}
	f2.Decls = make([]Decl, len(f.Decls))
	for i, child := range f2.Decls {
		f2.Decls[i] = child.Clone().(Decl)
		f2.Decls[i].SetParent(f2)
	}

	return f2
}

func (f *File) String() string {
	return ""
}

// NamespaceName

type NamespaceName struct {
	cst    cst.Node
	parent Node

	Parts []*Token
}

func NewNamespaceName(node cst.Node, parts []*Token) *NamespaceName {
	if parts == nil {
		return nil
	}

	n := &NamespaceName{
		cst:   node,
		Parts: parts,
	}

	for _, child := range parts {
		child.SetParent(n)
	}

	return n
}

func (n *NamespaceName) Cst() *cst.Node {
	if n.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &n.cst
}

func (n *NamespaceName) Token() scanner.Token {
	return scanner.Token{}
}

func (n *NamespaceName) Parent() Node {
	return n.parent
}

func (n *NamespaceName) SetParent(parent Node) {
	if parent != nil && n.parent != nil {
		panic("ast.NamespaceName.SetParent() - Parent is already set")
	}

	n.parent = parent
}

func (n *NamespaceName) AcceptChildren(visitor Visitor) {
	for _, child := range n.Parts {
		visitor.VisitNode(child)
	}
}

func (n *NamespaceName) Clone() Node {
	n2 := &NamespaceName{
		cst: n.cst,
	}

	n2.Parts = make([]*Token, len(n.Parts))
	for i, child := range n2.Parts {
		n2.Parts[i] = child.Clone().(*Token)
		n2.Parts[i].SetParent(n2)
	}

	return n2
}

func (n *NamespaceName) String() string {
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
	if name == nil && type_ == nil {
		return nil
	}

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

func (f *Field) Clone() Node {
	f2 := &Field{
		cst: f.cst,
	}

	if f.Name != nil {
		f2.Name = f.Name.Clone().(*Token)
		f2.Name.SetParent(f2)
	}
	if f.Type != nil {
		f2.Type = f.Type.Clone().(Type)
		f2.Type.SetParent(f2)
	}

	return f2
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
	if name == nil && value == nil {
		return nil
	}

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

func (i *InitField) Clone() Node {
	i2 := &InitField{
		cst: i.cst,
	}

	if i.Name != nil {
		i2.Name = i.Name.Clone().(*Token)
		i2.Name.SetParent(i2)
	}
	if i.Value != nil {
		i2.Value = i.Value.Clone().(Expr)
		i2.Value.SetParent(i2)
	}

	return i2
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
	if name == nil && value == nil {
		return nil
	}

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

func (e *EnumCase) Clone() Node {
	e2 := &EnumCase{
		cst:         e.cst,
		ActualValue: e.ActualValue,
	}

	if e.Name != nil {
		e2.Name = e.Name.Clone().(*Token)
		e2.Name.SetParent(e2)
	}
	if e.Value != nil {
		e2.Value = e.Value.Clone().(*Token)
		e2.Value.SetParent(e2)
	}

	return e2
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
	if name == nil && type_ == nil {
		return nil
	}

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

func (p *Param) Clone() Node {
	p2 := &Param{
		cst: p.cst,
	}

	if p.Name != nil {
		p2.Name = p.Name.Clone().(*Token)
		p2.Name.SetParent(p2)
	}
	if p.Type != nil {
		p2.Type = p.Type.Clone().(Type)
		p2.Type.SetParent(p2)
	}

	return p2
}

func (p *Param) String() string {
	return ""
}

// Attribute

type Attribute struct {
	cst    cst.Node
	parent Node

	Name *Token
	Args []*Token
}

func NewAttribute(node cst.Node, name *Token, args []*Token) *Attribute {
	if name == nil && args == nil {
		return nil
	}

	a := &Attribute{
		cst:  node,
		Name: name,
		Args: args,
	}

	if name != nil {
		name.SetParent(a)
	}
	for _, child := range args {
		child.SetParent(a)
	}

	return a
}

func (a *Attribute) Cst() *cst.Node {
	if a.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &a.cst
}

func (a *Attribute) Token() scanner.Token {
	return scanner.Token{}
}

func (a *Attribute) Parent() Node {
	return a.parent
}

func (a *Attribute) SetParent(parent Node) {
	if parent != nil && a.parent != nil {
		panic("ast.Attribute.SetParent() - Parent is already set")
	}

	a.parent = parent
}

func (a *Attribute) AcceptChildren(visitor Visitor) {
	if a.Name != nil {
		visitor.VisitNode(a.Name)
	}
	for _, child := range a.Args {
		visitor.VisitNode(child)
	}
}

func (a *Attribute) Clone() Node {
	a2 := &Attribute{
		cst: a.cst,
	}

	if a.Name != nil {
		a2.Name = a.Name.Clone().(*Token)
		a2.Name.SetParent(a2)
	}
	a2.Args = make([]*Token, len(a.Args))
	for i, child := range a2.Args {
		a2.Args[i] = child.Clone().(*Token)
		a2.Args[i].SetParent(a2)
	}

	return a2
}

func (a *Attribute) String() string {
	return ""
}

// Token

type Token struct {
	cst    cst.Node
	parent Node

	Token_ scanner.Token
}

func NewToken(node cst.Node, token scanner.Token) *Token {
	if token.IsEmpty() {
		return nil
	}

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

func (t *Token) Clone() Node {
	t2 := &Token{
		cst:    t.cst,
		Token_: t.Token_,
	}

	return t2
}

func (t *Token) String() string {
	return t.Token_.String()
}

func IsNil(node Node) bool {
	if node == nil {
		return true
	}

	switch node := node.(type) {
	case *Primitive:
		return node == nil
	case *Pointer:
		return node == nil
	case *Array:
		return node == nil
	case *Resolvable:
		return node == nil
	case *Namespace:
		return node == nil
	case *Using:
		return node == nil
	case *Struct:
		return node == nil
	case *Enum:
		return node == nil
	case *Impl:
		return node == nil
	case *Func:
		return node == nil
	case *GlobalVar:
		return node == nil
	case *Expression:
		return node == nil
	case *Block:
		return node == nil
	case *Var:
		return node == nil
	case *If:
		return node == nil
	case *While:
		return node == nil
	case *For:
		return node == nil
	case *Return:
		return node == nil
	case *Break:
		return node == nil
	case *Continue:
		return node == nil
	case *Paren:
		return node == nil
	case *Unary:
		return node == nil
	case *Binary:
		return node == nil
	case *Logical:
		return node == nil
	case *Assignment:
		return node == nil
	case *Member:
		return node == nil
	case *Index:
		return node == nil
	case *Cast:
		return node == nil
	case *Call:
		return node == nil
	case *TypeCall:
		return node == nil
	case *Typeof:
		return node == nil
	case *StructInitializer:
		return node == nil
	case *ArrayInitializer:
		return node == nil
	case *AllocateArray:
		return node == nil
	case *Identifier:
		return node == nil
	case *Literal:
		return node == nil
	case *File:
		return node == nil
	case *NamespaceName:
		return node == nil
	case *Field:
		return node == nil
	case *InitField:
		return node == nil
	case *EnumCase:
		return node == nil
	case *Param:
		return node == nil
	case *Attribute:
		return node == nil
	case *Token:
		return node == nil

	default:
		panic("ast.IsNil() - Not implemented")
	}
}
