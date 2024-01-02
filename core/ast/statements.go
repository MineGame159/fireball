package ast

import (
	"fireball/core/cst"
	"fireball/core/scanner"
)

// Visitor

type StmtVisitor interface {
	VisitExpression(stmt *Expression)
	VisitBlock(stmt *Block)
	VisitVar(stmt *Var)
	VisitIf(stmt *If)
	VisitFor(stmt *For)
	VisitReturn(stmt *Return)
	VisitBreak(stmt *Break)
	VisitContinue(stmt *Continue)
}

type Stmt interface {
	Node

	AcceptStmt(visitor StmtVisitor)
}

// Expression

type Expression struct {
	cst    cst.Node
	parent Node

	Expr Expr
}

func NewExpression(node cst.Node, expr Expr) *Expression {
	e := &Expression{
		cst:  node,
		Expr: expr,
	}

	if expr != nil {
		expr.SetParent(e)
	}

	return e
}

func (e *Expression) Cst() *cst.Node {
	if e.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &e.cst
}

func (e *Expression) Token() scanner.Token {
	return scanner.Token{}
}

func (e *Expression) Parent() Node {
	return e.parent
}

func (e *Expression) SetParent(parent Node) {
	if parent != nil && e.parent != nil {
		panic("ast.Expression.SetParent() - Parent is already set")
	}

	e.parent = parent
}

func (e *Expression) AcceptChildren(visitor Visitor) {
	if e.Expr != nil {
		visitor.VisitNode(e.Expr)
	}
}

func (e *Expression) String() string {
	return ""
}

func (e *Expression) AcceptStmt(visitor StmtVisitor) {
	visitor.VisitExpression(e)
}

// Block

type Block struct {
	cst    cst.Node
	parent Node

	Stmts []Stmt
}

func NewBlock(node cst.Node, stmts []Stmt) *Block {
	b := &Block{
		cst:   node,
		Stmts: stmts,
	}

	for _, child := range stmts {
		child.SetParent(b)
	}

	return b
}

func (b *Block) Cst() *cst.Node {
	if b.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &b.cst
}

func (b *Block) Token() scanner.Token {
	return scanner.Token{}
}

func (b *Block) Parent() Node {
	return b.parent
}

func (b *Block) SetParent(parent Node) {
	if parent != nil && b.parent != nil {
		panic("ast.Block.SetParent() - Parent is already set")
	}

	b.parent = parent
}

func (b *Block) AcceptChildren(visitor Visitor) {
	for _, child := range b.Stmts {
		visitor.VisitNode(child)
	}
}

func (b *Block) String() string {
	return ""
}

func (b *Block) AcceptStmt(visitor StmtVisitor) {
	visitor.VisitBlock(b)
}

// Var

type Var struct {
	cst    cst.Node
	parent Node

	Name       *Token
	Type       Type
	ActualType Type
	Value      Expr
}

func NewVar(node cst.Node, name *Token, type_ Type, value Expr) *Var {
	v := &Var{
		cst:   node,
		Name:  name,
		Type:  type_,
		Value: value,
	}

	if name != nil {
		name.SetParent(v)
	}
	if type_ != nil {
		type_.SetParent(v)
	}
	if value != nil {
		value.SetParent(v)
	}

	return v
}

func (v *Var) Cst() *cst.Node {
	if v.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &v.cst
}

func (v *Var) Token() scanner.Token {
	return scanner.Token{}
}

func (v *Var) Parent() Node {
	return v.parent
}

func (v *Var) SetParent(parent Node) {
	if parent != nil && v.parent != nil {
		panic("ast.Var.SetParent() - Parent is already set")
	}

	v.parent = parent
}

func (v *Var) AcceptChildren(visitor Visitor) {
	if v.Name != nil {
		visitor.VisitNode(v.Name)
	}
	if v.Type != nil {
		visitor.VisitNode(v.Type)
	}
	if v.Value != nil {
		visitor.VisitNode(v.Value)
	}
}

func (v *Var) String() string {
	return ""
}

func (v *Var) AcceptStmt(visitor StmtVisitor) {
	visitor.VisitVar(v)
}

// If

type If struct {
	cst    cst.Node
	parent Node

	Condition Expr
	Then      Stmt
	Else      Stmt
}

func NewIf(node cst.Node, condition Expr, then Stmt, else_ Stmt) *If {
	i := &If{
		cst:       node,
		Condition: condition,
		Then:      then,
		Else:      else_,
	}

	if condition != nil {
		condition.SetParent(i)
	}
	if then != nil {
		then.SetParent(i)
	}
	if else_ != nil {
		else_.SetParent(i)
	}

	return i
}

func (i *If) Cst() *cst.Node {
	if i.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &i.cst
}

func (i *If) Token() scanner.Token {
	return scanner.Token{}
}

func (i *If) Parent() Node {
	return i.parent
}

func (i *If) SetParent(parent Node) {
	if parent != nil && i.parent != nil {
		panic("ast.If.SetParent() - Parent is already set")
	}

	i.parent = parent
}

func (i *If) AcceptChildren(visitor Visitor) {
	if i.Condition != nil {
		visitor.VisitNode(i.Condition)
	}
	if i.Then != nil {
		visitor.VisitNode(i.Then)
	}
	if i.Else != nil {
		visitor.VisitNode(i.Else)
	}
}

func (i *If) String() string {
	return ""
}

func (i *If) AcceptStmt(visitor StmtVisitor) {
	visitor.VisitIf(i)
}

// For

type For struct {
	cst    cst.Node
	parent Node

	Initializer Stmt
	Condition   Expr
	Increment   Expr
	Body        Stmt
}

func NewFor(node cst.Node, initializer Stmt, condition Expr, increment Expr, body Stmt) *For {
	f := &For{
		cst:         node,
		Initializer: initializer,
		Condition:   condition,
		Increment:   increment,
		Body:        body,
	}

	if initializer != nil {
		initializer.SetParent(f)
	}
	if condition != nil {
		condition.SetParent(f)
	}
	if increment != nil {
		increment.SetParent(f)
	}
	if body != nil {
		body.SetParent(f)
	}

	return f
}

func (f *For) Cst() *cst.Node {
	if f.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &f.cst
}

func (f *For) Token() scanner.Token {
	return scanner.Token{}
}

func (f *For) Parent() Node {
	return f.parent
}

func (f *For) SetParent(parent Node) {
	if parent != nil && f.parent != nil {
		panic("ast.For.SetParent() - Parent is already set")
	}

	f.parent = parent
}

func (f *For) AcceptChildren(visitor Visitor) {
	if f.Initializer != nil {
		visitor.VisitNode(f.Initializer)
	}
	if f.Condition != nil {
		visitor.VisitNode(f.Condition)
	}
	if f.Increment != nil {
		visitor.VisitNode(f.Increment)
	}
	if f.Body != nil {
		visitor.VisitNode(f.Body)
	}
}

func (f *For) String() string {
	return ""
}

func (f *For) AcceptStmt(visitor StmtVisitor) {
	visitor.VisitFor(f)
}

// Return

type Return struct {
	cst    cst.Node
	parent Node

	Value Expr
}

func NewReturn(node cst.Node, value Expr) *Return {
	r := &Return{
		cst:   node,
		Value: value,
	}

	if value != nil {
		value.SetParent(r)
	}

	return r
}

func (r *Return) Cst() *cst.Node {
	if r.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &r.cst
}

func (r *Return) Token() scanner.Token {
	return scanner.Token{}
}

func (r *Return) Parent() Node {
	return r.parent
}

func (r *Return) SetParent(parent Node) {
	if parent != nil && r.parent != nil {
		panic("ast.Return.SetParent() - Parent is already set")
	}

	r.parent = parent
}

func (r *Return) AcceptChildren(visitor Visitor) {
	if r.Value != nil {
		visitor.VisitNode(r.Value)
	}
}

func (r *Return) String() string {
	return ""
}

func (r *Return) AcceptStmt(visitor StmtVisitor) {
	visitor.VisitReturn(r)
}

// Break

type Break struct {
	cst    cst.Node
	parent Node
}

func NewBreak(node cst.Node) *Break {
	b := &Break{
		cst: node,
	}

	return b
}

func (b *Break) Cst() *cst.Node {
	if b.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &b.cst
}

func (b *Break) Token() scanner.Token {
	return scanner.Token{}
}

func (b *Break) Parent() Node {
	return b.parent
}

func (b *Break) SetParent(parent Node) {
	if parent != nil && b.parent != nil {
		panic("ast.Break.SetParent() - Parent is already set")
	}

	b.parent = parent
}

func (b *Break) AcceptChildren(visitor Visitor) {
}

func (b *Break) String() string {
	return ""
}

func (b *Break) AcceptStmt(visitor StmtVisitor) {
	visitor.VisitBreak(b)
}

// Continue

type Continue struct {
	cst    cst.Node
	parent Node
}

func NewContinue(node cst.Node) *Continue {
	c := &Continue{
		cst: node,
	}

	return c
}

func (c *Continue) Cst() *cst.Node {
	if c.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &c.cst
}

func (c *Continue) Token() scanner.Token {
	return scanner.Token{}
}

func (c *Continue) Parent() Node {
	return c.parent
}

func (c *Continue) SetParent(parent Node) {
	if parent != nil && c.parent != nil {
		panic("ast.Continue.SetParent() - Parent is already set")
	}

	c.parent = parent
}

func (c *Continue) AcceptChildren(visitor Visitor) {
}

func (c *Continue) String() string {
	return ""
}

func (c *Continue) AcceptStmt(visitor StmtVisitor) {
	visitor.VisitContinue(c)
}
