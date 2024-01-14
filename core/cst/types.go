package cst

import "fireball/core/scanner"

var canStartType = []scanner.TokenKind{
	scanner.Identifier,
	scanner.Star,
	scanner.LeftBracket,
	scanner.Fn,
}

func parseType(p *parser) Node {
	switch p.peek() {
	case scanner.Identifier:
		return parseIdentifierType(p)
	case scanner.Star:
		return parsePointerType(p)
	case scanner.LeftBracket:
		return parseArrayType(p)
	case scanner.Fn:
		return parseFuncType(p)

	default:
		return p.error("Cannot start a type")
	}
}

func parseIdentifierType(p *parser) Node {
	p.begin(IdentifierTypeNode)

	if p.consume(scanner.Identifier) {
		return p.end()
	}
	for p.optional(scanner.Dot) {
		if p.consume(scanner.Identifier) {
			return p.end()
		}
	}

	return p.end()
}

func parsePointerType(p *parser) Node {
	p.begin(PointerTypeNode)

	if p.consume(scanner.Star) {
		return p.end()
	}
	if p.child(parseType) {
		return p.end()
	}

	return p.end()
}

func parseArrayType(p *parser) Node {
	p.begin(ArrayTypeNode)

	if p.consume(scanner.LeftBracket) {
		return p.end()
	}
	if p.consume(scanner.Number) {
		return p.end()
	}
	if p.consume(scanner.RightBracket) {
		return p.end()
	}
	if p.child(parseType) {
		return p.end()
	}

	return p.end()
}

var canStartFuncTypeParam = []scanner.TokenKind{scanner.Identifier}

func parseFuncType(p *parser) Node {
	p.begin(FuncTypeNode)

	if p.consume(scanner.Fn) {
		return p.end()
	}
	if p.consume(scanner.LeftParen) {
		return p.end()
	}
	if p.repeatSeparated(parseFuncTypeParam, canStartFuncTypeParam, scanner.Comma) {
		return p.end()
	}
	if p.consume(scanner.RightParen) {
		return p.end()
	}
	if p.child(parseType) {
		return p.end()
	}

	return p.end()
}

func parseFuncTypeParam(p *parser) Node {
	p.begin(FuncTypeParamNode)

	if p.optional(scanner.DotDotDot) {
		return p.end()
	}

	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.child(parseType) {
		return p.end()
	}

	return p.end()
}
