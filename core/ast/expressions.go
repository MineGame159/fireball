package ast

import "fireball/core"
import "fireball/core/types"
import "fireball/core/scanner"

//go:generate go run ../../gen/ast.go

type ExprVisitor interface {
	VisitGroup(expr *Group)
	VisitLiteral(expr *Literal)
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

func (g *Group) Accept(visitor ExprVisitor) {
	visitor.VisitGroup(g)
}

func (g *Group) AcceptChildren(acceptor Acceptor) {
	if g.Expr != nil {
		acceptor.AcceptExpr(g.Expr)
	}
}

func (g *Group) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(g.type_)
}

func (g *Group) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&g.type_)
}

func (g *Group) Leaf() bool {
	return false
}

func (g *Group) Type() types.Type {
	return g.type_
}

func (g *Group) SetType(type_ types.Type) {
	g.type_ = type_
}

// Literal

type Literal struct {
	range_ core.Range
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

func (l *Literal) Accept(visitor ExprVisitor) {
	visitor.VisitLiteral(l)
}

func (l *Literal) AcceptChildren(acceptor Acceptor) {
}

func (l *Literal) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(l.type_)
}

func (l *Literal) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&l.type_)
}

func (l *Literal) Leaf() bool {
	return true
}

func (l *Literal) Type() types.Type {
	return l.type_
}

func (l *Literal) SetType(type_ types.Type) {
	l.type_ = type_
}

// Unary

type Unary struct {
	range_ core.Range
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

func (u *Unary) Accept(visitor ExprVisitor) {
	visitor.VisitUnary(u)
}

func (u *Unary) AcceptChildren(acceptor Acceptor) {
	if u.Right != nil {
		acceptor.AcceptExpr(u.Right)
	}
}

func (u *Unary) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(u.type_)
}

func (u *Unary) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&u.type_)
}

func (u *Unary) Leaf() bool {
	return false
}

func (u *Unary) Type() types.Type {
	return u.type_
}

func (u *Unary) SetType(type_ types.Type) {
	u.type_ = type_
}

// Binary

type Binary struct {
	range_ core.Range
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

func (b *Binary) Accept(visitor ExprVisitor) {
	visitor.VisitBinary(b)
}

func (b *Binary) AcceptChildren(acceptor Acceptor) {
	if b.Left != nil {
		acceptor.AcceptExpr(b.Left)
	}
	if b.Right != nil {
		acceptor.AcceptExpr(b.Right)
	}
}

func (b *Binary) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(b.type_)
}

func (b *Binary) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&b.type_)
}

func (b *Binary) Leaf() bool {
	return false
}

func (b *Binary) Type() types.Type {
	return b.type_
}

func (b *Binary) SetType(type_ types.Type) {
	b.type_ = type_
}

// Logical

type Logical struct {
	range_ core.Range
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

func (l *Logical) Accept(visitor ExprVisitor) {
	visitor.VisitLogical(l)
}

func (l *Logical) AcceptChildren(acceptor Acceptor) {
	if l.Left != nil {
		acceptor.AcceptExpr(l.Left)
	}
	if l.Right != nil {
		acceptor.AcceptExpr(l.Right)
	}
}

func (l *Logical) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(l.type_)
}

func (l *Logical) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&l.type_)
}

func (l *Logical) Leaf() bool {
	return false
}

func (l *Logical) Type() types.Type {
	return l.type_
}

func (l *Logical) SetType(type_ types.Type) {
	l.type_ = type_
}

// Identifier

type Identifier struct {
	range_ core.Range
	type_  types.Type

	Identifier scanner.Token
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

func (i *Identifier) Accept(visitor ExprVisitor) {
	visitor.VisitIdentifier(i)
}

func (i *Identifier) AcceptChildren(acceptor Acceptor) {
}

func (i *Identifier) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(i.type_)
}

func (i *Identifier) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&i.type_)
}

func (i *Identifier) Leaf() bool {
	return true
}

func (i *Identifier) Type() types.Type {
	return i.type_
}

func (i *Identifier) SetType(type_ types.Type) {
	i.type_ = type_
}

// Assignment

type Assignment struct {
	range_ core.Range
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

func (a *Assignment) Accept(visitor ExprVisitor) {
	visitor.VisitAssignment(a)
}

func (a *Assignment) AcceptChildren(acceptor Acceptor) {
	if a.Assignee != nil {
		acceptor.AcceptExpr(a.Assignee)
	}
	if a.Value != nil {
		acceptor.AcceptExpr(a.Value)
	}
}

func (a *Assignment) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(a.type_)
}

func (a *Assignment) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&a.type_)
}

func (a *Assignment) Leaf() bool {
	return false
}

func (a *Assignment) Type() types.Type {
	return a.type_
}

func (a *Assignment) SetType(type_ types.Type) {
	a.type_ = type_
}

// Cast

type Cast struct {
	range_ core.Range
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

func (c *Cast) Accept(visitor ExprVisitor) {
	visitor.VisitCast(c)
}

func (c *Cast) AcceptChildren(acceptor Acceptor) {
	if c.Expr != nil {
		acceptor.AcceptExpr(c.Expr)
	}
}

func (c *Cast) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(c.type_)
}

func (c *Cast) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&c.type_)
}

func (c *Cast) Leaf() bool {
	return false
}

func (c *Cast) Type() types.Type {
	return c.type_
}

func (c *Cast) SetType(type_ types.Type) {
	c.type_ = type_
}

// Call

type Call struct {
	range_ core.Range
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

func (c *Call) Accept(visitor ExprVisitor) {
	visitor.VisitCall(c)
}

func (c *Call) AcceptChildren(acceptor Acceptor) {
	if c.Callee != nil {
		acceptor.AcceptExpr(c.Callee)
	}
	for _, v := range c.Args {
		acceptor.AcceptExpr(v)
	}
}

func (c *Call) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(c.type_)
}

func (c *Call) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&c.type_)
}

func (c *Call) Leaf() bool {
	return false
}

func (c *Call) Type() types.Type {
	return c.type_
}

func (c *Call) SetType(type_ types.Type) {
	c.type_ = type_
}

// Index

type Index struct {
	range_ core.Range
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

func (i *Index) Accept(visitor ExprVisitor) {
	visitor.VisitIndex(i)
}

func (i *Index) AcceptChildren(acceptor Acceptor) {
	if i.Value != nil {
		acceptor.AcceptExpr(i.Value)
	}
	if i.Index != nil {
		acceptor.AcceptExpr(i.Index)
	}
}

func (i *Index) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(i.type_)
}

func (i *Index) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&i.type_)
}

func (i *Index) Leaf() bool {
	return false
}

func (i *Index) Type() types.Type {
	return i.type_
}

func (i *Index) SetType(type_ types.Type) {
	i.type_ = type_
}

// Member

type Member struct {
	range_ core.Range
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

func (m *Member) Accept(visitor ExprVisitor) {
	visitor.VisitMember(m)
}

func (m *Member) AcceptChildren(acceptor Acceptor) {
	if m.Value != nil {
		acceptor.AcceptExpr(m.Value)
	}
}

func (m *Member) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(m.type_)
}

func (m *Member) AcceptTypesPtr(visitor types.PtrVisitor) {
	visitor.VisitType(&m.type_)
}

func (m *Member) Leaf() bool {
	return false
}

func (m *Member) Type() types.Type {
	return m.type_
}

func (m *Member) SetType(type_ types.Type) {
	m.type_ = type_
}
