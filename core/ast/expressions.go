package ast

import "log"
import "fireball/core"
import "fireball/core/types"
import "fireball/core/scanner"

//go:generate go run ../../gen/ast.go

type ExprVisitor interface {
	VisitGroup(expr *Group)
	VisitLiteral(expr *Literal)
	VisitInitializer(expr *Initializer)
	VisitUnary(expr *Unary)
	VisitBinary(expr *Binary)
	VisitLogical(expr *Logical)
	VisitIdentifier(expr *Identifier)
	VisitAssignment(expr *Assignment)
	VisitCast(expr *Cast)
	VisitCall(expr *Call)
	VisitIndex(expr *Index)
	VisitMember(expr *Member)
}

type Expr interface {
	Node

	Accept(visitor ExprVisitor)

	Type() types.Type
	SetType(type_ types.Type)
}

// Group

type Group struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Token_ scanner.Token
	Expr   Expr
}

func (g *Group) Token() scanner.Token {
	return g.Token_
}

func (g *Group) Range() core.Range {
	return g.range_
}

func (g *Group) SetRangeToken(start, end scanner.Token) {
	g.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (g *Group) SetRangePos(start, end core.Pos) {
	g.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (g *Group) SetRangeNode(start, end Node) {
	g.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (g *Group) Parent() Node {
	return g.parent
}

func (g *Group) SetParent(parent Node) {
	if g.parent != nil && parent != nil {
		log.Fatalln("Group.SetParent() - Node already has a parent")
	}
	g.parent = parent
}

func (g *Group) Accept(visitor ExprVisitor) {
	visitor.VisitGroup(g)
}

func (g *Group) AcceptChildren(visitor Acceptor) {
	if g.Expr != nil {
		visitor.AcceptExpr(g.Expr)
	}
}

func (g *Group) AcceptTypes(visitor types.Visitor) {
	if g.type_ != nil {
		visitor.VisitType(g.type_)
	}
}

func (g *Group) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&g.type_)
}

func (g *Group) Leaf() bool {
	return false
}

func (g *Group) String() string {
	return g.Token().Lexeme
}

func (g *Group) Type() types.Type {
	return g.type_
}

func (g *Group) SetType(type_ types.Type) {
	g.type_ = type_
}

func (g *Group) SetChildrenParent() {
	if g.Expr != nil {
		g.Expr.SetParent(g)
	}
}

// Literal

type Literal struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Value scanner.Token
}

func (l *Literal) Token() scanner.Token {
	return l.Value
}

func (l *Literal) Range() core.Range {
	return l.range_
}

func (l *Literal) SetRangeToken(start, end scanner.Token) {
	l.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (l *Literal) SetRangePos(start, end core.Pos) {
	l.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (l *Literal) SetRangeNode(start, end Node) {
	l.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (l *Literal) Parent() Node {
	return l.parent
}

func (l *Literal) SetParent(parent Node) {
	if l.parent != nil && parent != nil {
		log.Fatalln("Literal.SetParent() - Node already has a parent")
	}
	l.parent = parent
}

func (l *Literal) Accept(visitor ExprVisitor) {
	visitor.VisitLiteral(l)
}

func (l *Literal) AcceptChildren(visitor Acceptor) {
}

func (l *Literal) AcceptTypes(visitor types.Visitor) {
	if l.type_ != nil {
		visitor.VisitType(l.type_)
	}
}

func (l *Literal) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&l.type_)
}

func (l *Literal) Leaf() bool {
	return true
}

func (l *Literal) String() string {
	return l.Token().Lexeme
}

func (l *Literal) Type() types.Type {
	return l.type_
}

func (l *Literal) SetType(type_ types.Type) {
	l.type_ = type_
}

func (l *Literal) SetChildrenParent() {
}

// Initializer

type Initializer struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Name   scanner.Token
	Fields []InitField
}

func (i *Initializer) Token() scanner.Token {
	return i.Name
}

func (i *Initializer) Range() core.Range {
	return i.range_
}

func (i *Initializer) SetRangeToken(start, end scanner.Token) {
	i.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (i *Initializer) SetRangePos(start, end core.Pos) {
	i.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (i *Initializer) SetRangeNode(start, end Node) {
	i.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (i *Initializer) Parent() Node {
	return i.parent
}

func (i *Initializer) SetParent(parent Node) {
	if i.parent != nil && parent != nil {
		log.Fatalln("Initializer.SetParent() - Node already has a parent")
	}
	i.parent = parent
}

func (i *Initializer) Accept(visitor ExprVisitor) {
	visitor.VisitInitializer(i)
}

func (i *Initializer) AcceptChildren(visitor Acceptor) {
	for i_ := range i.Fields {
		if i.Fields[i_].Value != nil {
			visitor.AcceptExpr(i.Fields[i_].Value)
		}
	}
}

func (i *Initializer) AcceptTypes(visitor types.Visitor) {
	if i.type_ != nil {
		visitor.VisitType(i.type_)
	}
}

func (i *Initializer) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&i.type_)
}

func (i *Initializer) Leaf() bool {
	return false
}

func (i *Initializer) String() string {
	return i.Token().Lexeme
}

func (i *Initializer) Type() types.Type {
	return i.type_
}

func (i *Initializer) SetType(type_ types.Type) {
	i.type_ = type_
}

func (i *Initializer) SetChildrenParent() {
	for i_ := range i.Fields {
		if i.Fields[i_].Value != nil {
			i.Fields[i_].Value.SetParent(i)
		}
	}
}

// InitField

type InitField struct {
	Name  scanner.Token
	Value Expr
}

// Unary

type Unary struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Op    scanner.Token
	Right Expr
}

func (u *Unary) Token() scanner.Token {
	return u.Op
}

func (u *Unary) Range() core.Range {
	return u.range_
}

func (u *Unary) SetRangeToken(start, end scanner.Token) {
	u.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (u *Unary) SetRangePos(start, end core.Pos) {
	u.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (u *Unary) SetRangeNode(start, end Node) {
	u.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (u *Unary) Parent() Node {
	return u.parent
}

func (u *Unary) SetParent(parent Node) {
	if u.parent != nil && parent != nil {
		log.Fatalln("Unary.SetParent() - Node already has a parent")
	}
	u.parent = parent
}

func (u *Unary) Accept(visitor ExprVisitor) {
	visitor.VisitUnary(u)
}

func (u *Unary) AcceptChildren(visitor Acceptor) {
	if u.Right != nil {
		visitor.AcceptExpr(u.Right)
	}
}

func (u *Unary) AcceptTypes(visitor types.Visitor) {
	if u.type_ != nil {
		visitor.VisitType(u.type_)
	}
}

func (u *Unary) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&u.type_)
}

func (u *Unary) Leaf() bool {
	return false
}

func (u *Unary) String() string {
	return u.Token().Lexeme
}

func (u *Unary) Type() types.Type {
	return u.type_
}

func (u *Unary) SetType(type_ types.Type) {
	u.type_ = type_
}

func (u *Unary) SetChildrenParent() {
	if u.Right != nil {
		u.Right.SetParent(u)
	}
}

// Binary

type Binary struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Left  Expr
	Op    scanner.Token
	Right Expr
}

func (b *Binary) Token() scanner.Token {
	return b.Op
}

func (b *Binary) Range() core.Range {
	return b.range_
}

func (b *Binary) SetRangeToken(start, end scanner.Token) {
	b.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (b *Binary) SetRangePos(start, end core.Pos) {
	b.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (b *Binary) SetRangeNode(start, end Node) {
	b.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (b *Binary) Parent() Node {
	return b.parent
}

func (b *Binary) SetParent(parent Node) {
	if b.parent != nil && parent != nil {
		log.Fatalln("Binary.SetParent() - Node already has a parent")
	}
	b.parent = parent
}

func (b *Binary) Accept(visitor ExprVisitor) {
	visitor.VisitBinary(b)
}

func (b *Binary) AcceptChildren(visitor Acceptor) {
	if b.Left != nil {
		visitor.AcceptExpr(b.Left)
	}
	if b.Right != nil {
		visitor.AcceptExpr(b.Right)
	}
}

func (b *Binary) AcceptTypes(visitor types.Visitor) {
	if b.type_ != nil {
		visitor.VisitType(b.type_)
	}
}

func (b *Binary) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&b.type_)
}

func (b *Binary) Leaf() bool {
	return false
}

func (b *Binary) String() string {
	return b.Token().Lexeme
}

func (b *Binary) Type() types.Type {
	return b.type_
}

func (b *Binary) SetType(type_ types.Type) {
	b.type_ = type_
}

func (b *Binary) SetChildrenParent() {
	if b.Left != nil {
		b.Left.SetParent(b)
	}
	if b.Right != nil {
		b.Right.SetParent(b)
	}
}

// Logical

type Logical struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Left  Expr
	Op    scanner.Token
	Right Expr
}

func (l *Logical) Token() scanner.Token {
	return l.Op
}

func (l *Logical) Range() core.Range {
	return l.range_
}

func (l *Logical) SetRangeToken(start, end scanner.Token) {
	l.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (l *Logical) SetRangePos(start, end core.Pos) {
	l.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (l *Logical) SetRangeNode(start, end Node) {
	l.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (l *Logical) Parent() Node {
	return l.parent
}

func (l *Logical) SetParent(parent Node) {
	if l.parent != nil && parent != nil {
		log.Fatalln("Logical.SetParent() - Node already has a parent")
	}
	l.parent = parent
}

func (l *Logical) Accept(visitor ExprVisitor) {
	visitor.VisitLogical(l)
}

func (l *Logical) AcceptChildren(visitor Acceptor) {
	if l.Left != nil {
		visitor.AcceptExpr(l.Left)
	}
	if l.Right != nil {
		visitor.AcceptExpr(l.Right)
	}
}

func (l *Logical) AcceptTypes(visitor types.Visitor) {
	if l.type_ != nil {
		visitor.VisitType(l.type_)
	}
}

func (l *Logical) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&l.type_)
}

func (l *Logical) Leaf() bool {
	return false
}

func (l *Logical) String() string {
	return l.Token().Lexeme
}

func (l *Logical) Type() types.Type {
	return l.type_
}

func (l *Logical) SetType(type_ types.Type) {
	l.type_ = type_
}

func (l *Logical) SetChildrenParent() {
	if l.Left != nil {
		l.Left.SetParent(l)
	}
	if l.Right != nil {
		l.Right.SetParent(l)
	}
}

// Identifier

type Identifier struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Identifier scanner.Token
	Kind       IdentifierKind
}

func (i *Identifier) Token() scanner.Token {
	return i.Identifier
}

func (i *Identifier) Range() core.Range {
	return i.range_
}

func (i *Identifier) SetRangeToken(start, end scanner.Token) {
	i.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (i *Identifier) SetRangePos(start, end core.Pos) {
	i.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (i *Identifier) SetRangeNode(start, end Node) {
	i.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (i *Identifier) Parent() Node {
	return i.parent
}

func (i *Identifier) SetParent(parent Node) {
	if i.parent != nil && parent != nil {
		log.Fatalln("Identifier.SetParent() - Node already has a parent")
	}
	i.parent = parent
}

func (i *Identifier) Accept(visitor ExprVisitor) {
	visitor.VisitIdentifier(i)
}

func (i *Identifier) AcceptChildren(visitor Acceptor) {
}

func (i *Identifier) AcceptTypes(visitor types.Visitor) {
	if i.type_ != nil {
		visitor.VisitType(i.type_)
	}
}

func (i *Identifier) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&i.type_)
}

func (i *Identifier) Leaf() bool {
	return true
}

func (i *Identifier) String() string {
	return i.Token().Lexeme
}

func (i *Identifier) Type() types.Type {
	return i.type_
}

func (i *Identifier) SetType(type_ types.Type) {
	i.type_ = type_
}

func (i *Identifier) SetChildrenParent() {
}

// IdentifierKind

type IdentifierKind uint8

const (
	FunctionKind  IdentifierKind = 0
	EnumKind      IdentifierKind = 1
	VariableKind  IdentifierKind = 2
	ParameterKind IdentifierKind = 3
)

// Assignment

type Assignment struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Assignee Expr
	Op       scanner.Token
	Value    Expr
}

func (a *Assignment) Token() scanner.Token {
	return a.Op
}

func (a *Assignment) Range() core.Range {
	return a.range_
}

func (a *Assignment) SetRangeToken(start, end scanner.Token) {
	a.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (a *Assignment) SetRangePos(start, end core.Pos) {
	a.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (a *Assignment) SetRangeNode(start, end Node) {
	a.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (a *Assignment) Parent() Node {
	return a.parent
}

func (a *Assignment) SetParent(parent Node) {
	if a.parent != nil && parent != nil {
		log.Fatalln("Assignment.SetParent() - Node already has a parent")
	}
	a.parent = parent
}

func (a *Assignment) Accept(visitor ExprVisitor) {
	visitor.VisitAssignment(a)
}

func (a *Assignment) AcceptChildren(visitor Acceptor) {
	if a.Assignee != nil {
		visitor.AcceptExpr(a.Assignee)
	}
	if a.Value != nil {
		visitor.AcceptExpr(a.Value)
	}
}

func (a *Assignment) AcceptTypes(visitor types.Visitor) {
	if a.type_ != nil {
		visitor.VisitType(a.type_)
	}
}

func (a *Assignment) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&a.type_)
}

func (a *Assignment) Leaf() bool {
	return false
}

func (a *Assignment) String() string {
	return a.Token().Lexeme
}

func (a *Assignment) Type() types.Type {
	return a.type_
}

func (a *Assignment) SetType(type_ types.Type) {
	a.type_ = type_
}

func (a *Assignment) SetChildrenParent() {
	if a.Assignee != nil {
		a.Assignee.SetParent(a)
	}
	if a.Value != nil {
		a.Value.SetParent(a)
	}
}

// Cast

type Cast struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Token_ scanner.Token
	Expr   Expr
}

func (c *Cast) Token() scanner.Token {
	return c.Token_
}

func (c *Cast) Range() core.Range {
	return c.range_
}

func (c *Cast) SetRangeToken(start, end scanner.Token) {
	c.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (c *Cast) SetRangePos(start, end core.Pos) {
	c.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (c *Cast) SetRangeNode(start, end Node) {
	c.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (c *Cast) Parent() Node {
	return c.parent
}

func (c *Cast) SetParent(parent Node) {
	if c.parent != nil && parent != nil {
		log.Fatalln("Cast.SetParent() - Node already has a parent")
	}
	c.parent = parent
}

func (c *Cast) Accept(visitor ExprVisitor) {
	visitor.VisitCast(c)
}

func (c *Cast) AcceptChildren(visitor Acceptor) {
	if c.Expr != nil {
		visitor.AcceptExpr(c.Expr)
	}
}

func (c *Cast) AcceptTypes(visitor types.Visitor) {
	if c.type_ != nil {
		visitor.VisitType(c.type_)
	}
}

func (c *Cast) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&c.type_)
}

func (c *Cast) Leaf() bool {
	return false
}

func (c *Cast) String() string {
	return c.Token().Lexeme
}

func (c *Cast) Type() types.Type {
	return c.type_
}

func (c *Cast) SetType(type_ types.Type) {
	c.type_ = type_
}

func (c *Cast) SetChildrenParent() {
	if c.Expr != nil {
		c.Expr.SetParent(c)
	}
}

// Call

type Call struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Token_ scanner.Token
	Callee Expr
	Args   []Expr
}

func (c *Call) Token() scanner.Token {
	return c.Token_
}

func (c *Call) Range() core.Range {
	return c.range_
}

func (c *Call) SetRangeToken(start, end scanner.Token) {
	c.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (c *Call) SetRangePos(start, end core.Pos) {
	c.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (c *Call) SetRangeNode(start, end Node) {
	c.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (c *Call) Parent() Node {
	return c.parent
}

func (c *Call) SetParent(parent Node) {
	if c.parent != nil && parent != nil {
		log.Fatalln("Call.SetParent() - Node already has a parent")
	}
	c.parent = parent
}

func (c *Call) Accept(visitor ExprVisitor) {
	visitor.VisitCall(c)
}

func (c *Call) AcceptChildren(visitor Acceptor) {
	if c.Callee != nil {
		visitor.AcceptExpr(c.Callee)
	}
	for i_ := range c.Args {
		if c.Args[i_] != nil {
			visitor.AcceptExpr(c.Args[i_])
		}
	}
}

func (c *Call) AcceptTypes(visitor types.Visitor) {
	if c.type_ != nil {
		visitor.VisitType(c.type_)
	}
}

func (c *Call) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&c.type_)
}

func (c *Call) Leaf() bool {
	return false
}

func (c *Call) String() string {
	return c.Token().Lexeme
}

func (c *Call) Type() types.Type {
	return c.type_
}

func (c *Call) SetType(type_ types.Type) {
	c.type_ = type_
}

func (c *Call) SetChildrenParent() {
	if c.Callee != nil {
		c.Callee.SetParent(c)
	}
	for i_ := range c.Args {
		if c.Args[i_] != nil {
			c.Args[i_].SetParent(c)
		}
	}
}

// Index

type Index struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Token_ scanner.Token
	Value  Expr
	Index  Expr
}

func (i *Index) Token() scanner.Token {
	return i.Token_
}

func (i *Index) Range() core.Range {
	return i.range_
}

func (i *Index) SetRangeToken(start, end scanner.Token) {
	i.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (i *Index) SetRangePos(start, end core.Pos) {
	i.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (i *Index) SetRangeNode(start, end Node) {
	i.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (i *Index) Parent() Node {
	return i.parent
}

func (i *Index) SetParent(parent Node) {
	if i.parent != nil && parent != nil {
		log.Fatalln("Index.SetParent() - Node already has a parent")
	}
	i.parent = parent
}

func (i *Index) Accept(visitor ExprVisitor) {
	visitor.VisitIndex(i)
}

func (i *Index) AcceptChildren(visitor Acceptor) {
	if i.Value != nil {
		visitor.AcceptExpr(i.Value)
	}
	if i.Index != nil {
		visitor.AcceptExpr(i.Index)
	}
}

func (i *Index) AcceptTypes(visitor types.Visitor) {
	if i.type_ != nil {
		visitor.VisitType(i.type_)
	}
}

func (i *Index) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&i.type_)
}

func (i *Index) Leaf() bool {
	return false
}

func (i *Index) String() string {
	return i.Token().Lexeme
}

func (i *Index) Type() types.Type {
	return i.type_
}

func (i *Index) SetType(type_ types.Type) {
	i.type_ = type_
}

func (i *Index) SetChildrenParent() {
	if i.Value != nil {
		i.Value.SetParent(i)
	}
	if i.Index != nil {
		i.Index.SetParent(i)
	}
}

// Member

type Member struct {
	range_ core.Range
	parent Node
	type_  types.Type

	Value Expr
	Name  scanner.Token
}

func (m *Member) Token() scanner.Token {
	return m.Name
}

func (m *Member) Range() core.Range {
	return m.range_
}

func (m *Member) SetRangeToken(start, end scanner.Token) {
	m.range_ = core.Range{
		Start: core.TokenToPos(start, false),
		End:   core.TokenToPos(end, true),
	}
}

func (m *Member) SetRangePos(start, end core.Pos) {
	m.range_ = core.Range{
		Start: start,
		End:   end,
	}
}

func (m *Member) SetRangeNode(start, end Node) {
	m.range_ = core.Range{
		Start: start.Range().Start,
		End:   end.Range().End,
	}
}

func (m *Member) Parent() Node {
	return m.parent
}

func (m *Member) SetParent(parent Node) {
	if m.parent != nil && parent != nil {
		log.Fatalln("Member.SetParent() - Node already has a parent")
	}
	m.parent = parent
}

func (m *Member) Accept(visitor ExprVisitor) {
	visitor.VisitMember(m)
}

func (m *Member) AcceptChildren(visitor Acceptor) {
	if m.Value != nil {
		visitor.AcceptExpr(m.Value)
	}
}

func (m *Member) AcceptTypes(visitor types.Visitor) {
	if m.type_ != nil {
		visitor.VisitType(m.type_)
	}
}

func (m *Member) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&m.type_)
}

func (m *Member) Leaf() bool {
	return false
}

func (m *Member) String() string {
	return m.Token().Lexeme
}

func (m *Member) Type() types.Type {
	return m.type_
}

func (m *Member) SetType(type_ types.Type) {
	m.type_ = type_
}

func (m *Member) SetChildrenParent() {
	if m.Value != nil {
		m.Value.SetParent(m)
	}
}
