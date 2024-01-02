package cst

import (
	"fireball/core"
	"fireball/core/scanner"
	"fireball/core/utils"
	"slices"
)

type nodeInfo struct {
	kind   NodeKind
	offset int
}

type parser struct {
	scanner  *scanner.Scanner
	previous scanner.PositionedToken
	next     scanner.PositionedToken
	next2    scanner.PositionedToken

	nodes    []nodeInfo
	children []Node

	syncPoints         [][]scanner.TokenKind
	recoverToSyncPoint int

	reporter utils.Reporter
}

func Parse(reporter utils.Reporter, text string) Node {
	p := parser{
		scanner:            scanner.NewScanner(text),
		recoverToSyncPoint: -1,
		reporter:           reporter,
	}

	p.advance()
	p.advance()
	p.begin(FileNode)

	p.repeatSync(parseDecl, scanner.Eof, canStartDecl...)

	return p.end()
}

// Nodes

func (p *parser) begin(kind NodeKind) {
	p.nodes = append(p.nodes, nodeInfo{
		kind:   kind,
		offset: len(p.children),
	})
}

func (p *parser) child(parseFn func(p *parser) Node) bool {
	return p.childAdd(parseFn(p))
}

func (p *parser) childAdd(node Node) bool {
	if p.recovering() {
		return true
	}

	if node.Token.Lexeme != "" || len(node.Children) != 0 {
		p.children = append(p.children, node)
	}

	return false
}

func (p *parser) advanceAddChild() {
	p.children = append(p.children, p.advanceGetLeaf())
}

func (p *parser) advanceGetLeaf() Node {
	p.advance()

	return Node{
		Kind: NodeKindFromToken(p.previous.Token),
		Range: core.Range{
			Start: p.previous.Pos,
			End: core.Pos{
				Line:   p.previous.Pos.Line,
				Column: p.previous.Pos.Column + uint16(len(p.previous.Token.Lexeme)),
			},
		},
		Token: p.previous.Token,
	}
}

func (p *parser) end() Node {
	info := p.nodes[len(p.nodes)-1]
	p.nodes = p.nodes[:len(p.nodes)-1]

	children := p.children[info.offset:]
	exactChildren := make([]Node, len(children))
	copy(exactChildren, children)

	p.children = p.children[:info.offset]

	return Node{
		Kind: info.kind,
		Range: core.Range{
			Start: exactChildren[0].Range.Start,
			End:   exactChildren[len(exactChildren)-1].Range.End,
		},
		Children: exactChildren,
	}
}

// Syntax

func (p *parser) consume(kind scanner.TokenKind) bool {
	if p.peek() != kind {
		p.error("Expected a " + scanner.TokenKindStr(kind))
		return true
	}

	p.advanceAddChild()
	return false
}

func (p *parser) optional(kind scanner.TokenKind) bool {
	if p.peek() == kind {
		p.advanceAddChild()
		return true
	}

	return false
}

func (p *parser) repeat(parseFn func(p *parser) Node, canStartWith ...scanner.TokenKind) bool {
	for p.peekIs(canStartWith) {
		if p.child(parseFn) {
			return true
		}
	}

	return false
}

func (p *parser) repeatSync(parseFn func(p *parser) Node, after scanner.TokenKind, canStartWith ...scanner.TokenKind) bool {
	syncPointIndex := len(p.syncPoints)
	p.syncPoints = append(p.syncPoints, canStartWith)

outer:
	for p.peek() != scanner.Eof && p.peek() != after {
		if p.recovering() {
			// Bubble up if we are not the target sync point
			if p.recoverToSyncPoint < syncPointIndex {
				break
			}

			// Check all sync points in order from closest to furthest
			for i := syncPointIndex; i >= 0; i-- {
				syncPoint := p.syncPoints[i]

				if p.peekIs(syncPoint) {
					// If we already are at the right sync point then simply set recovering to false, otherwise set the target sync point index
					if i == syncPointIndex {
						p.recoverToSyncPoint = -1
						continue outer
					} else {
						p.recoverToSyncPoint = i
						break outer
					}
				}
			}

			// Skip over the next token as no sync point was able to recover
			p.advance()
		} else {
			// Try to parse the child
			p.child(parseFn)
		}
	}

	p.syncPoints = p.syncPoints[:len(p.syncPoints)-1]

	if p.peek() == after {
		p.recoverToSyncPoint = -1
	}

	return false
}

// Tokens

func (p *parser) peek() scanner.TokenKind {
	return p.next.Kind
}

func (p *parser) peekIs(kinds []scanner.TokenKind) bool {
	return slices.Contains(kinds, p.peek())
}

func (p *parser) peek2() scanner.TokenKind {
	return p.next2.Kind
}

func (p *parser) peek2Is(kinds []scanner.TokenKind) bool {
	return slices.Contains(kinds, p.peek2())
}

func (p *parser) advance() {
	p.previous = p.next
	p.next = p.next2
	p.next2 = p.scanner.Next()
}

// Errors

func (p *parser) recovering() bool {
	return p.recoverToSyncPoint != -1
}

func (p *parser) error(msg string) Node {
	p.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   p.next.Range(),
		Message: msg,
	})

	p.recoverToSyncPoint = len(p.syncPoints) - 1
	return Node{}
}
