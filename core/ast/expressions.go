package ast

import (
	"fireball/core/cst"
	"fireball/core/scanner"
)

// Visitor

type ExprVisitor interface {
	VisitParen(expr *Paren)
	VisitUnary(expr *Unary)
	VisitBinary(expr *Binary)
	VisitLogical(expr *Logical)
	VisitAssignment(expr *Assignment)
	VisitMember(expr *Member)
	VisitIndex(expr *Index)
	VisitCast(expr *Cast)
	VisitCall(expr *Call)
	VisitTypeCall(expr *TypeCall)
	VisitStructInitializer(expr *StructInitializer)
	VisitArrayInitializer(expr *ArrayInitializer)
	VisitAllocateArray(expr *AllocateArray)
	VisitIdentifier(expr *Identifier)
	VisitLiteral(expr *Literal)
}

type Expr interface {
	Node

	Result() *ExprResult

	AcceptExpr(visitor ExprVisitor)
}

// Paren

type Paren struct {
	cst    cst.Node
	parent Node

	Expr Expr

	result ExprResult
}

func NewParen(node cst.Node, expr Expr) *Paren {
	if expr == nil {
		return nil
	}

	p := &Paren{
		cst:  node,
		Expr: expr,
	}

	if expr != nil {
		expr.SetParent(p)
	}

	return p
}

func (p *Paren) Cst() *cst.Node {
	if p.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &p.cst
}

func (p *Paren) Token() scanner.Token {
	return scanner.Token{}
}

func (p *Paren) Parent() Node {
	return p.parent
}

func (p *Paren) SetParent(parent Node) {
	if parent != nil && p.parent != nil {
		panic("ast.Paren.SetParent() - Parent is already set")
	}

	p.parent = parent
}

func (p *Paren) AcceptChildren(visitor Visitor) {
	if p.Expr != nil {
		visitor.VisitNode(p.Expr)
	}
}

func (p *Paren) String() string {
	return ""
}

func (p *Paren) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitParen(p)
}

func (p *Paren) Result() *ExprResult {
	return &p.result
}

// Unary

type Unary struct {
	cst    cst.Node
	parent Node

	Prefix   bool
	Operator *Token
	Value    Expr

	result ExprResult
}

func NewUnary(node cst.Node, prefix bool, operator *Token, value Expr) *Unary {
	if operator == nil && value == nil {
		return nil
	}

	u := &Unary{
		cst:      node,
		Prefix:   prefix,
		Operator: operator,
		Value:    value,
	}

	if operator != nil {
		operator.SetParent(u)
	}
	if value != nil {
		value.SetParent(u)
	}

	return u
}

func (u *Unary) Cst() *cst.Node {
	if u.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &u.cst
}

func (u *Unary) Token() scanner.Token {
	return scanner.Token{}
}

func (u *Unary) Parent() Node {
	return u.parent
}

func (u *Unary) SetParent(parent Node) {
	if parent != nil && u.parent != nil {
		panic("ast.Unary.SetParent() - Parent is already set")
	}

	u.parent = parent
}

func (u *Unary) AcceptChildren(visitor Visitor) {
	if u.Operator != nil {
		visitor.VisitNode(u.Operator)
	}
	if u.Value != nil {
		visitor.VisitNode(u.Value)
	}
}

func (u *Unary) String() string {
	return ""
}

func (u *Unary) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitUnary(u)
}

func (u *Unary) Result() *ExprResult {
	return &u.result
}

// Binary

type Binary struct {
	cst    cst.Node
	parent Node

	Left     Expr
	Operator *Token
	Right    Expr

	result ExprResult
}

func NewBinary(node cst.Node, left Expr, operator *Token, right Expr) *Binary {
	if left == nil && operator == nil && right == nil {
		return nil
	}

	b := &Binary{
		cst:      node,
		Left:     left,
		Operator: operator,
		Right:    right,
	}

	if left != nil {
		left.SetParent(b)
	}
	if operator != nil {
		operator.SetParent(b)
	}
	if right != nil {
		right.SetParent(b)
	}

	return b
}

func (b *Binary) Cst() *cst.Node {
	if b.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &b.cst
}

func (b *Binary) Token() scanner.Token {
	return scanner.Token{}
}

func (b *Binary) Parent() Node {
	return b.parent
}

func (b *Binary) SetParent(parent Node) {
	if parent != nil && b.parent != nil {
		panic("ast.Binary.SetParent() - Parent is already set")
	}

	b.parent = parent
}

func (b *Binary) AcceptChildren(visitor Visitor) {
	if b.Left != nil {
		visitor.VisitNode(b.Left)
	}
	if b.Operator != nil {
		visitor.VisitNode(b.Operator)
	}
	if b.Right != nil {
		visitor.VisitNode(b.Right)
	}
}

func (b *Binary) String() string {
	return ""
}

func (b *Binary) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitBinary(b)
}

func (b *Binary) Result() *ExprResult {
	return &b.result
}

// Logical

type Logical struct {
	cst    cst.Node
	parent Node

	Left     Expr
	Operator *Token
	Right    Expr

	result ExprResult
}

func NewLogical(node cst.Node, left Expr, operator *Token, right Expr) *Logical {
	if left == nil && operator == nil && right == nil {
		return nil
	}

	l := &Logical{
		cst:      node,
		Left:     left,
		Operator: operator,
		Right:    right,
	}

	if left != nil {
		left.SetParent(l)
	}
	if operator != nil {
		operator.SetParent(l)
	}
	if right != nil {
		right.SetParent(l)
	}

	return l
}

func (l *Logical) Cst() *cst.Node {
	if l.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &l.cst
}

func (l *Logical) Token() scanner.Token {
	return scanner.Token{}
}

func (l *Logical) Parent() Node {
	return l.parent
}

func (l *Logical) SetParent(parent Node) {
	if parent != nil && l.parent != nil {
		panic("ast.Logical.SetParent() - Parent is already set")
	}

	l.parent = parent
}

func (l *Logical) AcceptChildren(visitor Visitor) {
	if l.Left != nil {
		visitor.VisitNode(l.Left)
	}
	if l.Operator != nil {
		visitor.VisitNode(l.Operator)
	}
	if l.Right != nil {
		visitor.VisitNode(l.Right)
	}
}

func (l *Logical) String() string {
	return ""
}

func (l *Logical) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitLogical(l)
}

func (l *Logical) Result() *ExprResult {
	return &l.result
}

// Assignment

type Assignment struct {
	cst    cst.Node
	parent Node

	Assignee Expr
	Operator *Token
	Value    Expr

	result ExprResult
}

func NewAssignment(node cst.Node, assignee Expr, operator *Token, value Expr) *Assignment {
	if assignee == nil && operator == nil && value == nil {
		return nil
	}

	a := &Assignment{
		cst:      node,
		Assignee: assignee,
		Operator: operator,
		Value:    value,
	}

	if assignee != nil {
		assignee.SetParent(a)
	}
	if operator != nil {
		operator.SetParent(a)
	}
	if value != nil {
		value.SetParent(a)
	}

	return a
}

func (a *Assignment) Cst() *cst.Node {
	if a.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &a.cst
}

func (a *Assignment) Token() scanner.Token {
	return scanner.Token{}
}

func (a *Assignment) Parent() Node {
	return a.parent
}

func (a *Assignment) SetParent(parent Node) {
	if parent != nil && a.parent != nil {
		panic("ast.Assignment.SetParent() - Parent is already set")
	}

	a.parent = parent
}

func (a *Assignment) AcceptChildren(visitor Visitor) {
	if a.Assignee != nil {
		visitor.VisitNode(a.Assignee)
	}
	if a.Operator != nil {
		visitor.VisitNode(a.Operator)
	}
	if a.Value != nil {
		visitor.VisitNode(a.Value)
	}
}

func (a *Assignment) String() string {
	return ""
}

func (a *Assignment) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitAssignment(a)
}

func (a *Assignment) Result() *ExprResult {
	return &a.result
}

// Member

type Member struct {
	cst    cst.Node
	parent Node

	Value Expr
	Name  *Token

	result ExprResult
}

func NewMember(node cst.Node, value Expr, name *Token) *Member {
	if value == nil && name == nil {
		return nil
	}

	m := &Member{
		cst:   node,
		Value: value,
		Name:  name,
	}

	if value != nil {
		value.SetParent(m)
	}
	if name != nil {
		name.SetParent(m)
	}

	return m
}

func (m *Member) Cst() *cst.Node {
	if m.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &m.cst
}

func (m *Member) Token() scanner.Token {
	return scanner.Token{}
}

func (m *Member) Parent() Node {
	return m.parent
}

func (m *Member) SetParent(parent Node) {
	if parent != nil && m.parent != nil {
		panic("ast.Member.SetParent() - Parent is already set")
	}

	m.parent = parent
}

func (m *Member) AcceptChildren(visitor Visitor) {
	if m.Value != nil {
		visitor.VisitNode(m.Value)
	}
	if m.Name != nil {
		visitor.VisitNode(m.Name)
	}
}

func (m *Member) String() string {
	return ""
}

func (m *Member) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitMember(m)
}

func (m *Member) Result() *ExprResult {
	return &m.result
}

// Index

type Index struct {
	cst    cst.Node
	parent Node

	Value Expr
	Index Expr

	result ExprResult
}

func NewIndex(node cst.Node, value Expr, index Expr) *Index {
	if value == nil && index == nil {
		return nil
	}

	i := &Index{
		cst:   node,
		Value: value,
		Index: index,
	}

	if value != nil {
		value.SetParent(i)
	}
	if index != nil {
		index.SetParent(i)
	}

	return i
}

func (i *Index) Cst() *cst.Node {
	if i.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &i.cst
}

func (i *Index) Token() scanner.Token {
	return scanner.Token{}
}

func (i *Index) Parent() Node {
	return i.parent
}

func (i *Index) SetParent(parent Node) {
	if parent != nil && i.parent != nil {
		panic("ast.Index.SetParent() - Parent is already set")
	}

	i.parent = parent
}

func (i *Index) AcceptChildren(visitor Visitor) {
	if i.Value != nil {
		visitor.VisitNode(i.Value)
	}
	if i.Index != nil {
		visitor.VisitNode(i.Index)
	}
}

func (i *Index) String() string {
	return ""
}

func (i *Index) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitIndex(i)
}

func (i *Index) Result() *ExprResult {
	return &i.result
}

// Cast

type Cast struct {
	cst    cst.Node
	parent Node

	Value  Expr
	Target Type

	result ExprResult
}

func NewCast(node cst.Node, value Expr, target Type) *Cast {
	if value == nil && target == nil {
		return nil
	}

	c := &Cast{
		cst:    node,
		Value:  value,
		Target: target,
	}

	if value != nil {
		value.SetParent(c)
	}
	if target != nil {
		target.SetParent(c)
	}

	return c
}

func (c *Cast) Cst() *cst.Node {
	if c.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &c.cst
}

func (c *Cast) Token() scanner.Token {
	return scanner.Token{}
}

func (c *Cast) Parent() Node {
	return c.parent
}

func (c *Cast) SetParent(parent Node) {
	if parent != nil && c.parent != nil {
		panic("ast.Cast.SetParent() - Parent is already set")
	}

	c.parent = parent
}

func (c *Cast) AcceptChildren(visitor Visitor) {
	if c.Value != nil {
		visitor.VisitNode(c.Value)
	}
	if c.Target != nil {
		visitor.VisitNode(c.Target)
	}
}

func (c *Cast) String() string {
	return ""
}

func (c *Cast) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitCast(c)
}

func (c *Cast) Result() *ExprResult {
	return &c.result
}

// Call

type Call struct {
	cst    cst.Node
	parent Node

	Callee Expr
	Args   []Expr

	result ExprResult
}

func NewCall(node cst.Node, callee Expr, args []Expr) *Call {
	if callee == nil && args == nil {
		return nil
	}

	c := &Call{
		cst:    node,
		Callee: callee,
		Args:   args,
	}

	if callee != nil {
		callee.SetParent(c)
	}
	for _, child := range args {
		child.SetParent(c)
	}

	return c
}

func (c *Call) Cst() *cst.Node {
	if c.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &c.cst
}

func (c *Call) Token() scanner.Token {
	return scanner.Token{}
}

func (c *Call) Parent() Node {
	return c.parent
}

func (c *Call) SetParent(parent Node) {
	if parent != nil && c.parent != nil {
		panic("ast.Call.SetParent() - Parent is already set")
	}

	c.parent = parent
}

func (c *Call) AcceptChildren(visitor Visitor) {
	if c.Callee != nil {
		visitor.VisitNode(c.Callee)
	}
	for _, child := range c.Args {
		visitor.VisitNode(child)
	}
}

func (c *Call) String() string {
	return ""
}

func (c *Call) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitCall(c)
}

func (c *Call) Result() *ExprResult {
	return &c.result
}

// TypeCall

type TypeCall struct {
	cst    cst.Node
	parent Node

	Callee *Token
	Arg    Type

	result ExprResult
}

func NewTypeCall(node cst.Node, callee *Token, arg Type) *TypeCall {
	if callee == nil && arg == nil {
		return nil
	}

	t := &TypeCall{
		cst:    node,
		Callee: callee,
		Arg:    arg,
	}

	if callee != nil {
		callee.SetParent(t)
	}
	if arg != nil {
		arg.SetParent(t)
	}

	return t
}

func (t *TypeCall) Cst() *cst.Node {
	if t.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &t.cst
}

func (t *TypeCall) Token() scanner.Token {
	return scanner.Token{}
}

func (t *TypeCall) Parent() Node {
	return t.parent
}

func (t *TypeCall) SetParent(parent Node) {
	if parent != nil && t.parent != nil {
		panic("ast.TypeCall.SetParent() - Parent is already set")
	}

	t.parent = parent
}

func (t *TypeCall) AcceptChildren(visitor Visitor) {
	if t.Callee != nil {
		visitor.VisitNode(t.Callee)
	}
	if t.Arg != nil {
		visitor.VisitNode(t.Arg)
	}
}

func (t *TypeCall) String() string {
	return ""
}

func (t *TypeCall) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitTypeCall(t)
}

func (t *TypeCall) Result() *ExprResult {
	return &t.result
}

// StructInitializer

type StructInitializer struct {
	cst    cst.Node
	parent Node

	New    bool
	Type   Type
	Fields []*InitField

	result ExprResult
}

func NewStructInitializer(node cst.Node, new bool, type_ Type, fields []*InitField) *StructInitializer {
	if type_ == nil && fields == nil {
		return nil
	}

	s := &StructInitializer{
		cst:    node,
		New:    new,
		Type:   type_,
		Fields: fields,
	}

	if type_ != nil {
		type_.SetParent(s)
	}
	for _, child := range fields {
		child.SetParent(s)
	}

	return s
}

func (s *StructInitializer) Cst() *cst.Node {
	if s.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &s.cst
}

func (s *StructInitializer) Token() scanner.Token {
	return scanner.Token{}
}

func (s *StructInitializer) Parent() Node {
	return s.parent
}

func (s *StructInitializer) SetParent(parent Node) {
	if parent != nil && s.parent != nil {
		panic("ast.StructInitializer.SetParent() - Parent is already set")
	}

	s.parent = parent
}

func (s *StructInitializer) AcceptChildren(visitor Visitor) {
	if s.Type != nil {
		visitor.VisitNode(s.Type)
	}
	for _, child := range s.Fields {
		visitor.VisitNode(child)
	}
}

func (s *StructInitializer) String() string {
	return ""
}

func (s *StructInitializer) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitStructInitializer(s)
}

func (s *StructInitializer) Result() *ExprResult {
	return &s.result
}

// ArrayInitializer

type ArrayInitializer struct {
	cst    cst.Node
	parent Node

	Values []Expr

	result ExprResult
}

func NewArrayInitializer(node cst.Node, values []Expr) *ArrayInitializer {
	if values == nil {
		return nil
	}

	a := &ArrayInitializer{
		cst:    node,
		Values: values,
	}

	for _, child := range values {
		child.SetParent(a)
	}

	return a
}

func (a *ArrayInitializer) Cst() *cst.Node {
	if a.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &a.cst
}

func (a *ArrayInitializer) Token() scanner.Token {
	return scanner.Token{}
}

func (a *ArrayInitializer) Parent() Node {
	return a.parent
}

func (a *ArrayInitializer) SetParent(parent Node) {
	if parent != nil && a.parent != nil {
		panic("ast.ArrayInitializer.SetParent() - Parent is already set")
	}

	a.parent = parent
}

func (a *ArrayInitializer) AcceptChildren(visitor Visitor) {
	for _, child := range a.Values {
		visitor.VisitNode(child)
	}
}

func (a *ArrayInitializer) String() string {
	return ""
}

func (a *ArrayInitializer) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitArrayInitializer(a)
}

func (a *ArrayInitializer) Result() *ExprResult {
	return &a.result
}

// AllocateArray

type AllocateArray struct {
	cst    cst.Node
	parent Node

	Type  Type
	Count Expr

	result ExprResult
}

func NewAllocateArray(node cst.Node, type_ Type, count Expr) *AllocateArray {
	if type_ == nil && count == nil {
		return nil
	}

	a := &AllocateArray{
		cst:   node,
		Type:  type_,
		Count: count,
	}

	if type_ != nil {
		type_.SetParent(a)
	}
	if count != nil {
		count.SetParent(a)
	}

	return a
}

func (a *AllocateArray) Cst() *cst.Node {
	if a.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &a.cst
}

func (a *AllocateArray) Token() scanner.Token {
	return scanner.Token{}
}

func (a *AllocateArray) Parent() Node {
	return a.parent
}

func (a *AllocateArray) SetParent(parent Node) {
	if parent != nil && a.parent != nil {
		panic("ast.AllocateArray.SetParent() - Parent is already set")
	}

	a.parent = parent
}

func (a *AllocateArray) AcceptChildren(visitor Visitor) {
	if a.Type != nil {
		visitor.VisitNode(a.Type)
	}
	if a.Count != nil {
		visitor.VisitNode(a.Count)
	}
}

func (a *AllocateArray) String() string {
	return ""
}

func (a *AllocateArray) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitAllocateArray(a)
}

func (a *AllocateArray) Result() *ExprResult {
	return &a.result
}

// Identifier

type Identifier struct {
	cst    cst.Node
	parent Node

	Name scanner.Token
	Kind IdentifierKind

	result ExprResult
}

func NewIdentifier(node cst.Node, name scanner.Token) *Identifier {
	if name.IsEmpty() {
		return nil
	}

	i := &Identifier{
		cst:  node,
		Name: name,
	}

	return i
}

func (i *Identifier) Cst() *cst.Node {
	if i.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &i.cst
}

func (i *Identifier) Token() scanner.Token {
	return i.Name
}

func (i *Identifier) Parent() Node {
	return i.parent
}

func (i *Identifier) SetParent(parent Node) {
	if parent != nil && i.parent != nil {
		panic("ast.Identifier.SetParent() - Parent is already set")
	}

	i.parent = parent
}

func (i *Identifier) AcceptChildren(visitor Visitor) {
}

func (i *Identifier) String() string {
	return i.Name.String()
}

func (i *Identifier) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitIdentifier(i)
}

func (i *Identifier) Result() *ExprResult {
	return &i.result
}

// Literal

type Literal struct {
	cst    cst.Node
	parent Node

	Token_ scanner.Token

	result ExprResult
}

func NewLiteral(node cst.Node, token scanner.Token) *Literal {
	if token.IsEmpty() {
		return nil
	}

	l := &Literal{
		cst:    node,
		Token_: token,
	}

	return l
}

func (l *Literal) Cst() *cst.Node {
	if l.cst.Kind == cst.UnknownNode {
		return nil
	}

	return &l.cst
}

func (l *Literal) Token() scanner.Token {
	return l.Token_
}

func (l *Literal) Parent() Node {
	return l.parent
}

func (l *Literal) SetParent(parent Node) {
	if parent != nil && l.parent != nil {
		panic("ast.Literal.SetParent() - Parent is already set")
	}

	l.parent = parent
}

func (l *Literal) AcceptChildren(visitor Visitor) {
}

func (l *Literal) String() string {
	return l.Token_.String()
}

func (l *Literal) AcceptExpr(visitor ExprVisitor) {
	visitor.VisitLiteral(l)
}

func (l *Literal) Result() *ExprResult {
	return &l.result
}
