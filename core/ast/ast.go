package ast

import (
	"fireball/core"
	"fireball/core/cst"
	"fireball/core/scanner"
)

type Node interface {
	Cst() *cst.Node
	Token() scanner.Token

	Parent() Node
	SetParent(parent Node)

	AcceptChildren(visitor Visitor)

	Clone() Node

	String() string
}

type Visitor interface {
	VisitNode(node Node)
}

// GetLeaf()

func GetLeaf(node Node, pos core.Pos) Node {
	g := getLeaf{
		pos: pos,
	}

	g.VisitNode(node)
	return g.node
}

type getLeaf struct {
	pos  core.Pos
	node Node
}

func (g *getLeaf) VisitNode(node Node) {
	// Propagate node up the tree
	if g.node != nil {
		return
	}

	// Check if node contains target position
	if node.Cst() == nil || node.Cst().Range.Contains(g.pos) {
		// Check if node is a leaf node
		if !node.Token().IsEmpty() && node.Cst() != nil && node.Cst().Range.Contains(g.pos) {
			g.node = node
			return
		}

		// Children
		node.AcceptChildren(g)

		// Propagate node up the tree
		if g.node != nil {
			return
		}
	}
}

// Get()

func Get(node Node, pos core.Pos) Node {
	g := get{
		pos: pos,
	}

	g.VisitNode(node)
	return g.node
}

type get struct {
	pos  core.Pos
	node Node

	hasChild bool
}

func (g *get) VisitNode(node Node) {
	// Propagate node up the tree
	if g.node != nil {
		return
	}

	// Check if node contains target position
	if node.Cst() == nil || g.contains(node) {
		// Children
		node.AcceptChildren(g)

		// Propagate node up the tree
		if g.node != nil {
			return
		}

		// Set node if we are going back up the tree without a found node and the pos is inside the current range
		if g.contains(node) {
			g.node = node
			return
		}
	}
}

func (g *get) contains(node Node) bool {
	if node.Cst() == nil {
		return false
	}
	range_ := node.Cst().Range

	if g.pos.Line == range_.End.Line && g.pos.Column >= range_.End.Column && node.Token().IsEmpty() {
		if node.Parent() != nil && node.Parent().Cst() != nil && (node.Parent().Cst().Range.End.Line > range_.End.Line || node.Parent().Cst().Range.End.Column == range_.End.Column) {
			next := GetNextSibling(node)

			if next == nil || (next.Cst() != nil && next.Cst().Range.Start.Line > g.pos.Line) {
				return true
			}
		}
	}

	return range_.Contains(g.pos)
}

// GetNextSibling()

func GetNextSibling(node Node) Node {
	for node.Parent() != nil {
		g := getNextSibling{target: node}
		node.Parent().AcceptChildren(&g)

		if g.next != nil {
			return g.next
		}

		node = node.Parent()
	}

	return nil
}

type getNextSibling struct {
	target Node

	seenTarget bool
	next       Node
}

func (g *getNextSibling) VisitNode(node Node) {
	if node == g.target {
		g.seenTarget = true
	} else if g.seenTarget && g.next == nil {
		g.next = node
	}
}

// GetParent()

func GetParent[T Node](node Node) T {
	for !IsNil(node) {
		if n, ok := node.(T); ok {
			return n
		}

		node = node.Parent()
	}

	var empty T
	return empty
}
