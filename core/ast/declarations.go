package ast

import (
	"fireball/core/cst"
	"fireball/core/scanner"
)

// Visitor

type DeclVisitor interface {
	VisitNamespace(decl *Namespace)
	VisitUsing(decl *Using)
	VisitStruct(decl *Struct)
	VisitEnum(decl *Enum)
	VisitImpl(decl *Impl)
	VisitInterface(decl *Interface)
	VisitFunc(decl *Func)
	VisitGlobalVar(decl *GlobalVar)
}

type Decl interface {
	Node

	AcceptDecl(visitor DeclVisitor)
}

// Namespace

type Namespace struct {
	cst    cst.Node
	parent Node

	Name *NamespaceName
}

func NewNamespace(node cst.Node, name *NamespaceName) *Namespace {
	if name == nil {
		return nil
	}

	n := &Namespace{
		cst:  node,
		Name: name,
	}

	if name != nil {
		name.SetParent(n)
	}

	return n
}

func (n *Namespace) Cst() *cst.Node {
	if n.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &n.cst
}

func (n *Namespace) Token() scanner.Token {
	return scanner.Token{}
}

func (n *Namespace) Parent() Node {
	return n.parent
}

func (n *Namespace) SetParent(parent Node) {
	if parent != nil && n.parent != nil {
		panic("ast.Namespace.SetParent() - Parent is already set")
	}

	n.parent = parent
}

func (n *Namespace) AcceptChildren(visitor Visitor) {
	if n.Name != nil {
		visitor.VisitNode(n.Name)
	}
}

func (n *Namespace) Clone() Node {
	n2 := &Namespace{
		cst: n.cst,
	}

	if n.Name != nil {
		n2.Name = n.Name.Clone().(*NamespaceName)
		n2.Name.SetParent(n2)
	}

	return n2
}

func (n *Namespace) String() string {
	return ""
}

func (n *Namespace) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitNamespace(n)
}

// Using

type Using struct {
	cst    cst.Node
	parent Node

	Name *NamespaceName
}

func NewUsing(node cst.Node, name *NamespaceName) *Using {
	if name == nil {
		return nil
	}

	u := &Using{
		cst:  node,
		Name: name,
	}

	if name != nil {
		name.SetParent(u)
	}

	return u
}

func (u *Using) Cst() *cst.Node {
	if u.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &u.cst
}

func (u *Using) Token() scanner.Token {
	return scanner.Token{}
}

func (u *Using) Parent() Node {
	return u.parent
}

func (u *Using) SetParent(parent Node) {
	if parent != nil && u.parent != nil {
		panic("ast.Using.SetParent() - Parent is already set")
	}

	u.parent = parent
}

func (u *Using) AcceptChildren(visitor Visitor) {
	if u.Name != nil {
		visitor.VisitNode(u.Name)
	}
}

func (u *Using) Clone() Node {
	u2 := &Using{
		cst: u.cst,
	}

	if u.Name != nil {
		u2.Name = u.Name.Clone().(*NamespaceName)
		u2.Name.SetParent(u2)
	}

	return u2
}

func (u *Using) String() string {
	return ""
}

func (u *Using) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitUsing(u)
}

// Struct

type Struct struct {
	cst    cst.Node
	parent Node

	Attributes   []*Attribute
	Name         *Token
	Fields       []*Field
	StaticFields []*Field
}

func NewStruct(node cst.Node, attributes []*Attribute, name *Token, fields []*Field, staticfields []*Field) *Struct {
	if attributes == nil && name == nil && fields == nil && staticfields == nil {
		return nil
	}

	s := &Struct{
		cst:          node,
		Attributes:   attributes,
		Name:         name,
		Fields:       fields,
		StaticFields: staticfields,
	}

	for _, child := range attributes {
		child.SetParent(s)
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
	for _, child := range s.Attributes {
		visitor.VisitNode(child)
	}
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

func (s *Struct) Clone() Node {
	s2 := &Struct{
		cst: s.cst,
	}

	s2.Attributes = make([]*Attribute, len(s.Attributes))
	for i, child := range s2.Attributes {
		s2.Attributes[i] = child.Clone().(*Attribute)
		s2.Attributes[i].SetParent(s2)
	}
	if s.Name != nil {
		s2.Name = s.Name.Clone().(*Token)
		s2.Name.SetParent(s2)
	}
	s2.Fields = make([]*Field, len(s.Fields))
	for i, child := range s2.Fields {
		s2.Fields[i] = child.Clone().(*Field)
		s2.Fields[i].SetParent(s2)
	}
	s2.StaticFields = make([]*Field, len(s.StaticFields))
	for i, child := range s2.StaticFields {
		s2.StaticFields[i] = child.Clone().(*Field)
		s2.StaticFields[i].SetParent(s2)
	}

	return s2
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

func (e *Enum) Clone() Node {
	e2 := &Enum{
		cst:        e.cst,
		ActualType: e.ActualType,
	}

	if e.Name != nil {
		e2.Name = e.Name.Clone().(*Token)
		e2.Name.SetParent(e2)
	}
	if e.Type != nil {
		e2.Type = e.Type.Clone().(Type)
		e2.Type.SetParent(e2)
	}
	if e.ActualType != nil {
		e2.ActualType = e.ActualType.Clone().(Type)
		e2.ActualType.SetParent(e2)
	}
	e2.Cases = make([]*EnumCase, len(e.Cases))
	for i, child := range e2.Cases {
		e2.Cases[i] = child.Clone().(*EnumCase)
		e2.Cases[i].SetParent(e2)
	}

	return e2
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

	Struct     *Token
	Type       Type
	Implements Type
	Methods    []*Func
}

func NewImpl(node cst.Node, struct_ *Token, implements Type, methods []*Func) *Impl {
	if struct_ == nil && implements == nil && methods == nil {
		return nil
	}

	i := &Impl{
		cst:        node,
		Struct:     struct_,
		Implements: implements,
		Methods:    methods,
	}

	if struct_ != nil {
		struct_.SetParent(i)
	}
	if implements != nil {
		implements.SetParent(i)
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
	if i.Implements != nil {
		visitor.VisitNode(i.Implements)
	}
	for _, child := range i.Methods {
		visitor.VisitNode(child)
	}
}

func (i *Impl) Clone() Node {
	i2 := &Impl{
		cst:  i.cst,
		Type: i.Type,
	}

	if i.Struct != nil {
		i2.Struct = i.Struct.Clone().(*Token)
		i2.Struct.SetParent(i2)
	}
	if i.Type != nil {
		i2.Type = i.Type.Clone().(Type)
		i2.Type.SetParent(i2)
	}
	if i.Implements != nil {
		i2.Implements = i.Implements.Clone().(Type)
		i2.Implements.SetParent(i2)
	}
	i2.Methods = make([]*Func, len(i.Methods))
	for i, child := range i2.Methods {
		i2.Methods[i] = child.Clone().(*Func)
		i2.Methods[i].SetParent(i2)
	}

	return i2
}

func (i *Impl) String() string {
	return ""
}

func (i *Impl) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitImpl(i)
}

// Interface

type Interface struct {
	cst    cst.Node
	parent Node

	Name    *Token
	Methods []*Func
}

func NewInterface(node cst.Node, name *Token, methods []*Func) *Interface {
	if name == nil && methods == nil {
		return nil
	}

	i := &Interface{
		cst:     node,
		Name:    name,
		Methods: methods,
	}

	if name != nil {
		name.SetParent(i)
	}
	for _, child := range methods {
		child.SetParent(i)
	}

	return i
}

func (i *Interface) Cst() *cst.Node {
	if i.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &i.cst
}

func (i *Interface) Token() scanner.Token {
	return scanner.Token{}
}

func (i *Interface) Parent() Node {
	return i.parent
}

func (i *Interface) SetParent(parent Node) {
	if parent != nil && i.parent != nil {
		panic("ast.Interface.SetParent() - Parent is already set")
	}

	i.parent = parent
}

func (i *Interface) AcceptChildren(visitor Visitor) {
	if i.Name != nil {
		visitor.VisitNode(i.Name)
	}
	for _, child := range i.Methods {
		visitor.VisitNode(child)
	}
}

func (i *Interface) Clone() Node {
	i2 := &Interface{
		cst: i.cst,
	}

	if i.Name != nil {
		i2.Name = i.Name.Clone().(*Token)
		i2.Name.SetParent(i2)
	}
	i2.Methods = make([]*Func, len(i.Methods))
	for i, child := range i2.Methods {
		i2.Methods[i] = child.Clone().(*Func)
		i2.Methods[i].SetParent(i2)
	}

	return i2
}

func (i *Interface) String() string {
	return ""
}

func (i *Interface) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitInterface(i)
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

func (f *Func) Clone() Node {
	f2 := &Func{
		cst:   f.cst,
		Flags: f.Flags,
	}

	f2.Attributes = make([]*Attribute, len(f.Attributes))
	for i, child := range f2.Attributes {
		f2.Attributes[i] = child.Clone().(*Attribute)
		f2.Attributes[i].SetParent(f2)
	}
	if f.Name != nil {
		f2.Name = f.Name.Clone().(*Token)
		f2.Name.SetParent(f2)
	}
	f2.Params = make([]*Param, len(f.Params))
	for i, child := range f2.Params {
		f2.Params[i] = child.Clone().(*Param)
		f2.Params[i].SetParent(f2)
	}
	if f.Returns != nil {
		f2.Returns = f.Returns.Clone().(Type)
		f2.Returns.SetParent(f2)
	}
	f2.Body = make([]Stmt, len(f.Body))
	for i, child := range f2.Body {
		f2.Body[i] = child.Clone().(Stmt)
		f2.Body[i].SetParent(f2)
	}

	return f2
}

func (f *Func) String() string {
	return ""
}

func (f *Func) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitFunc(f)
}

// GlobalVar

type GlobalVar struct {
	cst    cst.Node
	parent Node

	Name *Token
	Type Type
}

func NewGlobalVar(node cst.Node, name *Token, type_ Type) *GlobalVar {
	if name == nil && type_ == nil {
		return nil
	}

	g := &GlobalVar{
		cst:  node,
		Name: name,
		Type: type_,
	}

	if name != nil {
		name.SetParent(g)
	}
	if type_ != nil {
		type_.SetParent(g)
	}

	return g
}

func (g *GlobalVar) Cst() *cst.Node {
	if g.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &g.cst
}

func (g *GlobalVar) Token() scanner.Token {
	return scanner.Token{}
}

func (g *GlobalVar) Parent() Node {
	return g.parent
}

func (g *GlobalVar) SetParent(parent Node) {
	if parent != nil && g.parent != nil {
		panic("ast.GlobalVar.SetParent() - Parent is already set")
	}

	g.parent = parent
}

func (g *GlobalVar) AcceptChildren(visitor Visitor) {
	if g.Name != nil {
		visitor.VisitNode(g.Name)
	}
	if g.Type != nil {
		visitor.VisitNode(g.Type)
	}
}

func (g *GlobalVar) Clone() Node {
	g2 := &GlobalVar{
		cst: g.cst,
	}

	if g.Name != nil {
		g2.Name = g.Name.Clone().(*Token)
		g2.Name.SetParent(g2)
	}
	if g.Type != nil {
		g2.Type = g.Type.Clone().(Type)
		g2.Type.SetParent(g2)
	}

	return g2
}

func (g *GlobalVar) String() string {
	return ""
}

func (g *GlobalVar) AcceptDecl(visitor DeclVisitor) {
	visitor.VisitGlobalVar(g)
}
