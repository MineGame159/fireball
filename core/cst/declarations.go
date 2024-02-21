package cst

import "fireball/core/scanner"

var canStartDecl = []scanner.TokenKind{
	scanner.Namespace,
	scanner.Using,

	scanner.Struct,
	scanner.Impl,
	scanner.Enum,
	scanner.Interface,
	scanner.Func,
	scanner.Var,

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
	case scanner.Namespace:
		return parseNamespaceDecl(p, attributes)
	case scanner.Using:
		return parseUsingDecl(p, attributes)

	case scanner.Struct:
		return parseStructDecl(p, attributes)
	case scanner.Impl:
		return parseImplDecl(p, attributes)
	case scanner.Enum:
		return parseEnumDecl(p, attributes)
	case scanner.Interface:
		return parseInterfaceDecl(p, attributes)
	case scanner.Func:
		return parseFuncDecl(p, attributes)
	case scanner.Var:
		return parseVarDecl(p, attributes)

	default:
		return p.error("Cannot start a declaration")
	}
}

// Namespace

func parseNamespaceDecl(p *parser, attributes Node) Node {
	p.begin(NamespaceDeclNode)

	p.childAdd(attributes)
	if p.consume(scanner.Namespace) {
		return p.end()
	}
	if p.child(parseNamespaceName) {
		return p.end()
	}
	if p.consume(scanner.Semicolon) {
		return p.end()
	}

	return p.end()
}

func parseUsingDecl(p *parser, attributes Node) Node {
	p.begin(UsingDeclNode)

	p.childAdd(attributes)
	if p.consume(scanner.Using) {
		return p.end()
	}
	if p.child(parseNamespaceName) {
		return p.end()
	}
	if p.consume(scanner.Semicolon) {
		return p.end()
	}

	return p.end()
}

func parseNamespaceName(p *parser) Node {
	p.begin(NamespaceNameNode)

	if p.child(parseNamespacePart) {
		return p.end()
	}
	for p.optional(scanner.Dot) {
		if p.child(parseNamespacePart) {
			return p.end()
		}
	}

	return p.end()
}

func parseNamespacePart(p *parser) Node {
	if p.peek() != scanner.Identifier {
		p.error("Expected a " + scanner.TokenKindStr(scanner.Identifier))
		return Node{}
	}

	return p.advanceGetLeaf()
}

// Struct

func parseStructDecl(p *parser, attributes Node) Node {
	p.begin(StructDeclNode)

	p.childAdd(attributes)
	if p.consume(scanner.Struct) {
		return p.end()
	}
	if p.consume(scanner.Identifier) {
		return p.end()
	}

	if p.optional(scanner.LeftBracket) {
		if p.repeatSeparated(parseGenericParam, canStartGenericParam, scanner.Comma) {
			return p.end()
		}
		if p.consume(scanner.RightBracket) {
			return p.end()
		}
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

func parseImplDecl(p *parser, attributes Node) Node {
	p.begin(ImplDeclNode)

	p.childAdd(attributes)
	if p.consume(scanner.Impl) {
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
	if p.repeatSync(parseFuncDeclWithAttributes, scanner.RightBrace, scanner.Hashtag, scanner.Static, scanner.Func) {
		return p.end()
	}
	if p.consume(scanner.RightBrace) {
		return p.end()
	}

	return p.end()
}

// Enum

func parseEnumDecl(p *parser, attributes Node) Node {
	p.begin(EnumDeclNode)

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
		if p.consume(scanner.Number, scanner.Hex, scanner.Binary) {
			return p.end()
		}
	}
	if p.consume(scanner.Comma) {
		return p.end()
	}

	return p.end()
}

// Interface

func parseInterfaceDecl(p *parser, attributes Node) Node {
	p.begin(InterfaceDeclNode)

	p.childAdd(attributes)
	if p.consume(scanner.Interface) {
		return p.end()
	}
	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.consume(scanner.LeftBrace) {
		return p.end()
	}
	if p.repeatSync(parseFuncDeclWithAttributes, scanner.RightBrace, scanner.Hashtag, scanner.Func) {
		return p.end()
	}
	if p.consume(scanner.RightBrace) {
		return p.end()
	}

	return p.end()
}

// Func

var canStartParam = []scanner.TokenKind{scanner.Identifier, scanner.Dot}

func parseFuncDeclWithAttributes(p *parser) Node {
	var attributes Node

	if p.peek() == scanner.Hashtag {
		attributes = parseAttributes(p)
		if p.recovering() {
			return Node{}
		}
	}

	return parseFuncDecl(p, attributes)
}

func parseFuncDecl(p *parser, attributes Node) Node {
	p.begin(FuncDeclNode)

	p.childAdd(attributes)
	p.optional(scanner.Static)
	if p.consume(scanner.Func) {
		return p.end()
	}
	if p.consume(scanner.Identifier) {
		return p.end()
	}

	if p.optional(scanner.LeftBracket) {
		if p.repeatSeparated(parseGenericParam, canStartGenericParam, scanner.Comma) {
			return p.end()
		}
		if p.consume(scanner.RightBracket) {
			return p.end()
		}
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

// Var

func parseVarDecl(p *parser, attributes Node) Node {
	p.begin(VarDeclNode)

	p.childAdd(attributes)
	if p.consume(scanner.Var) {
		return p.end()
	}
	if p.consume(scanner.Identifier) {
		return p.end()
	}
	if p.child(parseType) {
		return p.end()
	}
	if p.consume(scanner.Semicolon) {
		return p.end()
	}

	return p.end()
}

// Attribute

var canStartAttribute = []scanner.TokenKind{scanner.Identifier}
var canStartAttributeArg = []scanner.TokenKind{scanner.String}

func parseAttributes(p *parser) Node {
	p.begin(AttributesNode)

	if p.consume(scanner.Hashtag) {
		return p.end()
	}
	if p.consume(scanner.LeftBracket) {
		return p.end()
	}
	if p.repeatSeparated(parseAttribute, canStartAttribute, scanner.Comma) {
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

// Generics

var canStartGenericParam = []scanner.TokenKind{scanner.Identifier}

func parseGenericParam(p *parser) Node {
	p.begin(GenericParamNode)

	p.consume(scanner.Identifier)

	return p.end()
}
