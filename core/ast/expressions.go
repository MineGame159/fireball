package ast

import "fireball/core/scanner"
import "fireball/core/types"

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

type Group struct {
	type_ types.Type

	Token_ scanner.Token
	Expr   Expr
}

func (g *Group) Token() scanner.Token {
	return g.Token_
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
	visitor.VisitType(&g.type_)
}

func (g *Group) Type() types.Type {
	return g.type_
}

func (g *Group) SetType(type_ types.Type) {
	g.type_ = type_
}

type Literal struct {
	type_ types.Type

	Value scanner.Token
}

func (l *Literal) Token() scanner.Token {
	return l.Value
}

func (l *Literal) Accept(visitor ExprVisitor) {
	visitor.VisitLiteral(l)
}

func (l *Literal) AcceptChildren(acceptor Acceptor) {
}

func (l *Literal) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(&l.type_)
}

func (l *Literal) Type() types.Type {
	return l.type_
}

func (l *Literal) SetType(type_ types.Type) {
	l.type_ = type_
}

type Unary struct {
	type_ types.Type

	Op    scanner.Token
	Right Expr
}

func (u *Unary) Token() scanner.Token {
	return u.Op
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
	visitor.VisitType(&u.type_)
}

func (u *Unary) Type() types.Type {
	return u.type_
}

func (u *Unary) SetType(type_ types.Type) {
	u.type_ = type_
}

type Binary struct {
	type_ types.Type

	Left  Expr
	Op    scanner.Token
	Right Expr
}

func (b *Binary) Token() scanner.Token {
	return b.Op
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
	visitor.VisitType(&b.type_)
}

func (b *Binary) Type() types.Type {
	return b.type_
}

func (b *Binary) SetType(type_ types.Type) {
	b.type_ = type_
}

type Logical struct {
	type_ types.Type

	Left  Expr
	Op    scanner.Token
	Right Expr
}

func (l *Logical) Token() scanner.Token {
	return l.Op
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
	visitor.VisitType(&l.type_)
}

func (l *Logical) Type() types.Type {
	return l.type_
}

func (l *Logical) SetType(type_ types.Type) {
	l.type_ = type_
}

type Identifier struct {
	type_ types.Type

	Identifier scanner.Token
}

func (i *Identifier) Token() scanner.Token {
	return i.Identifier
}

func (i *Identifier) Accept(visitor ExprVisitor) {
	visitor.VisitIdentifier(i)
}

func (i *Identifier) AcceptChildren(acceptor Acceptor) {
}

func (i *Identifier) AcceptTypes(visitor types.Visitor) {
	visitor.VisitType(&i.type_)
}

func (i *Identifier) Type() types.Type {
	return i.type_
}

func (i *Identifier) SetType(type_ types.Type) {
	i.type_ = type_
}

type Assignment struct {
	type_ types.Type

	Assignee Expr
	Op       scanner.Token
	Value    Expr
}

func (a *Assignment) Token() scanner.Token {
	return a.Op
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
	visitor.VisitType(&a.type_)
}

func (a *Assignment) Type() types.Type {
	return a.type_
}

func (a *Assignment) SetType(type_ types.Type) {
	a.type_ = type_
}

type Cast struct {
	type_ types.Type

	Token_ scanner.Token
	Expr   Expr
}

func (c *Cast) Token() scanner.Token {
	return c.Token_
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
	visitor.VisitType(&c.type_)
}

func (c *Cast) Type() types.Type {
	return c.type_
}

func (c *Cast) SetType(type_ types.Type) {
	c.type_ = type_
}

type Call struct {
	type_ types.Type

	Token_ scanner.Token
	Callee Expr
	Args   []Expr
}

func (c *Call) Token() scanner.Token {
	return c.Token_
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
	visitor.VisitType(&c.type_)
}

func (c *Call) Type() types.Type {
	return c.type_
}

func (c *Call) SetType(type_ types.Type) {
	c.type_ = type_
}

type Index struct {
	type_ types.Type

	Token_ scanner.Token
	Value  Expr
	Index  Expr
}

func (i *Index) Token() scanner.Token {
	return i.Token_
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
	visitor.VisitType(&i.type_)
}

func (i *Index) Type() types.Type {
	return i.type_
}

func (i *Index) SetType(type_ types.Type) {
	i.type_ = type_
}

type Member struct {
	type_ types.Type

	Value Expr
	Name  scanner.Token
}

func (m *Member) Token() scanner.Token {
	return m.Name
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
	visitor.VisitType(&m.type_)
}

func (m *Member) Type() types.Type {
	return m.type_
}

func (m *Member) SetType(type_ types.Type) {
	m.type_ = type_
}
