package cst

import "fireball/core/scanner"

var canStartDecl = []scanner.TokenKind{
	scanner.Struct,
	scanner.Impl,
	scanner.Enum,
	scanner.Func,

	scanner.Hashtag,
}

func parseDecl(p *parser) Node {
	var attributes Node

	if p.peek() == scanner.Hashtag {
		attributes = parseAttributes(p)
		if p.recovering() {
			return Node{}
		}
	}

	switch p.peek() {
	case scanner.Struct:
		return parseStruct(p, attributes)
	case scanner.Impl:
		return parseImpl(p, attributes)
	case scanner.Enum:
		return parseEnum(p, attributes)
	case scanner.Func:
		return parseFunc(p, attributes)

	default:
		return p.error("Cannot start a declaration")
	}
}

// Struct

func parseStruct(p *parser, attributes Node) Node {
	p.begin(StructNode)

	p.childAdd(attributes)
	if p.consume(scanner.Struct) {
		return p.end()
	}
	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.consume(scanner.LeftBrace) {
		return p.end()
	}
	if p.repeatSync(parseStructField, scanner.RightBrace, scanner.Static, scanner.Identifier) {
		return p.end()
	}
	if p.consume(scanner.RightBrace) {
		return p.end()
	}

	return p.end()
}

func parseStructField(p *parser) Node {
	p.begin(StructFieldNode)

	p.optional(scanner.Static)
	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.child(parseType) {
		return p.end()
	}
	if p.consume(scanner.Comma) {
		return p.end()
	}

	return p.end()
}

// Impl

func parseImpl(p *parser, attributes Node) Node {
	p.begin(ImplNode)

	p.childAdd(attributes)
	if p.consume(scanner.Impl) {
		return p.end()
	}
	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.consume(scanner.LeftBrace) {
		return p.end()
	}
	if p.repeatSync(parseImplFunc, scanner.RightBrace, scanner.Hashtag, scanner.Static, scanner.Func) {
		return p.end()
	}
	if p.consume(scanner.RightBrace) {
		return p.end()
	}

	return p.end()
}

func parseImplFunc(p *parser) Node {
	var attributes Node

	if p.peek() == scanner.Hashtag {
		attributes = parseAttributes(p)
		if p.recovering() {
			return Node{}
		}
	}

	return parseFunc(p, attributes)
}

// Enum

func parseEnum(p *parser, attributes Node) Node {
	p.begin(EnumNode)

	p.childAdd(attributes)
	if p.consume(scanner.Enum) {
		return p.end()
	}
	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.optional(scanner.Colon) {
		if p.child(parseType) {
			return p.end()
		}
	}
	if p.consume(scanner.LeftBrace) {
		return p.end()
	}
	if p.repeatSync(parseEnumCase, scanner.RightBrace, scanner.Identifier) {
		return p.end()
	}
	if p.consume(scanner.RightBrace) {
		return p.end()
	}

	return p.end()
}

func parseEnumCase(p *parser) Node {
	p.begin(EnumCaseNode)

	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.optional(scanner.Equal) {
		if p.consume(scanner.Number) {
			return p.end()
		}
	}
	if p.consume(scanner.Comma) {
		return p.end()
	}

	return p.end()
}

// Func

var canStartParam = []scanner.TokenKind{scanner.Identifier, scanner.Dot}

func parseFunc(p *parser, attributes Node) Node {
	p.begin(FuncNode)

	p.childAdd(attributes)
	p.optional(scanner.Static)
	if p.consume(scanner.Func) {
		return p.end()
	}
	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.consume(scanner.LeftParen) {
		return p.end()
	}
	if p.repeatSeparated(parseFuncParam, canStartParam, scanner.Comma) {
		return p.end()
	}
	if p.consume(scanner.RightParen) {
		return p.end()
	}
	if p.peekIs(canStartType) {
		if p.child(parseType) {
			return p.end()
		}
	}
	if p.optional(scanner.LeftBrace) {
		if p.repeatSync(parseStmt, scanner.RightBrace, canStartStmt...) {
			return p.end()
		}
		if p.consume(scanner.RightBrace) {
			return p.end()
		}
	}

	return p.end()
}

func parseFuncParam(p *parser) Node {
	p.begin(FuncParamNode)

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

// Attribute

var canStartAttributeArg = []scanner.TokenKind{scanner.String}

func parseAttributes(p *parser) Node {
	p.begin(AttributesNode)

	if p.consume(scanner.Hashtag) {
		return p.end()
	}
	if p.consume(scanner.LeftBracket) {
		return p.end()
	}
	if p.repeatSync(parseAttribute, scanner.RightBracket, scanner.Identifier) {
		return p.end()
	}
	if p.consume(scanner.RightBracket) {
		return p.end()
	}

	return p.end()
}

func parseAttribute(p *parser) Node {
	p.begin(AttributeNode)

	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.optional(scanner.LeftParen) {
		if p.repeatSeparated(parseAttributeArg, canStartAttributeArg, scanner.Comma) {
			return p.end()
		}
		if p.consume(scanner.RightParen) {
			return p.end()
		}
	}

	return p.end()
}

func parseAttributeArg(p *parser) Node {
	if p.peek() == scanner.String {
		return p.advanceGetLeaf()
	}

	return p.error("Attribute argument needs to be a string")
}
