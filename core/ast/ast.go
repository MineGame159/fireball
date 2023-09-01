package ast

import (
	"fireball/core/scanner"
	"fireball/core/types"
)

type Node interface {
	Token() scanner.Token

	Range() Range
	SetRangeToken(start, end scanner.Token)
	SetRangePos(start, end Pos)
	SetRangeNode(start, end Node)

	AcceptChildren(acceptor Acceptor)
	AcceptTypes(visitor types.Visitor)

	Leaf() bool
}

type Acceptor interface {
	AcceptDecl(decl Decl)
	AcceptStmt(stmt Stmt)
	AcceptExpr(expr Expr)
}

type Range struct {
	Start Pos
	End   Pos
}

type Pos struct {
	Line   int
	Column int
}

func (r Range) Contains(pos Pos) bool {
	// Check line
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}

	// Check start column
	if pos.Line == r.Start.Line && pos.Column < r.Start.Column {
		return false
	}

	// Check end column
	if pos.Line == r.End.Line && pos.Column > r.End.Column {
		return false
	}

	// True
	return true
}

func TokenToRange(token scanner.Token) Range {
	return Range{
		Start: TokenToPos(token, false),
		End:   TokenToPos(token, true),
	}
}

func TokenToPos(token scanner.Token, end bool) Pos {
	offset := 0

	if end {
		offset = len(token.Lexeme)
	}

	return Pos{
		Line:   token.Line,
		Column: token.Column + offset,
	}
}

// Visit

func VisitStmts[T Stmt](node Node, callback func(stmt T)) {
	visitor[T]{
		expressions: false,
		callback:    callback,
	}.Accept(node)
}

func VisitExprs[T Expr](node Node, callback func(expr T)) {
	visitor[T]{
		expressions: true,
		callback:    callback,
	}.Accept(node)
}

type visitor[T Node] struct {
	expressions bool
	callback    func(node T)
}

func (v visitor[T]) Accept(node Node) {
	if v2, ok := node.(T); ok {
		v.callback(v2)
	}

	node.AcceptChildren(v)
}

func (v visitor[T]) AcceptDecl(decl Decl) {
	v.Accept(decl)
}

func (v visitor[T]) AcceptStmt(stmt Stmt) {
	v.Accept(stmt)
}

func (v visitor[T]) AcceptExpr(expr Expr) {
	if v.expressions {
		v.Accept(expr)
	}
}

// GetLeaf

func GetLeaf(node Node, pos Pos) Node {
	g := get{
		pos: pos,
	}

	g.Accept(node)
	return g.node
}

type get struct {
	pos  Pos
	node Node
}

func (v *get) Accept(node Node) {
	// Propagate node up the tree
	if v.node != nil {
		return
	}

	// Check if node contains target position
	if node.Range().Contains(v.pos) {
		// Check if node is a leaf node
		if node.Leaf() {
			v.node = node
			return
		}

		// Children
		node.AcceptChildren(v)

		// TODO: Work around the fact that some nodes store names as scanner.Token which does not inherit ast.Node
		if variable, ok := node.(*Variable); ok && TokenToRange(variable.Name).Contains(v.pos) {
			v.node = node
			return
		}

		if member, ok := node.(*Member); ok && TokenToRange(member.Name).Contains(v.pos) {
			v.node = node
			return
		}
	}
}

func (v *get) AcceptDecl(decl Decl) {
	v.Accept(decl)
}

func (v *get) AcceptStmt(stmt Stmt) {
	v.Accept(stmt)
}

func (v *get) AcceptExpr(expr Expr) {
	v.Accept(expr)
}
