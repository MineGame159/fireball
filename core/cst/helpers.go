package cst

import "fireball/core/scanner"

func (p *parser) advanceWrap(kind NodeKind) Node {
	p.begin(kind)

	p.advanceAddChild()

	return p.end()
}

func (p *parser) repeatSeparated(parseFn func(p *parser) Node, canStart []scanner.TokenKind, separator scanner.TokenKind) bool {
	if p.peekIs(canStart) {
		if p.child(parseFn) {
			return true
		}

		for p.optional(separator) {
			if p.child(parseFn) {
				return true
			}
		}
	}

	return false
}

func (p *parser) repeatOne(parseFn func(p *parser) Node, canStartWith ...scanner.TokenKind) bool {
	if p.child(parseFn) {
		return true
	}

	return p.repeat(parseFn, canStartWith...)
}
