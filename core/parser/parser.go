package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fmt"
	"strconv"
)

type parser struct {
	scanner *scanner.Scanner

	previous scanner.Token
	current  scanner.Token
	next     scanner.Token

	extern bool

	reporter core.Reporter
}

func Parse(reporter core.Reporter, scanner *scanner.Scanner) []ast.Decl {
	// Initialise parser
	p := &parser{
		scanner:  scanner,
		reporter: reporter,
	}

	p.advance()

	// Parse declarations
	decls := make([]ast.Decl, 0, 8)

	for !p.isAtEnd() {
		decl := p.declaration()

		if decl != nil {
			decls = append(decls, decl)
		}
	}

	return decls
}

func (p *parser) parseType() (types.Type, *core.Diagnostic) {
	if p.match(scanner.LeftBracket) {
		return p.parseArrayType()
	}
	if p.match(scanner.Star) {
		return p.parsePointerType()
	}

	return p.parseIdentifierType()
}

func (p *parser) parseArrayType() (types.Type, *core.Diagnostic) {
	// Count
	token, err := p.consume(scanner.Number, "Expected array size.")
	if err != nil {
		return nil, err
	}

	count, err_ := strconv.Atoi(token.Lexeme)

	if err_ != nil {
		return nil, p.error(token, "Invalid array size.")
	}
	if count < 0 {
		return nil, p.error(token, "Invalid array size.")
	}

	// Right bracket
	if _, err := p.consume(scanner.RightBracket, "Expected ']' after array size."); err != nil {
		return nil, err
	}

	// Base
	base, err := p.parseType()
	if err != nil {
		return nil, err
	}

	return &types.ArrayType{
		Count: uint32(count),
		Base:  base,
	}, nil
}

func (p *parser) parsePointerType() (types.Type, *core.Diagnostic) {
	pointee, err := p.parseType()
	if err != nil {
		return nil, err
	}

	return &types.PointerType{Pointee: pointee}, nil
}

func (p *parser) parseIdentifierType() (types.Type, *core.Diagnostic) {
	ident, err := p.consume(scanner.Identifier, "Expected type name.")
	if err != nil {
		return nil, err
	}

	var kind types.PrimitiveKind

	switch ident.Lexeme {
	case "void":
		kind = types.Void
	case "bool":
		kind = types.Bool

	case "u8":
		kind = types.U8
	case "u16":
		kind = types.U16
	case "u32":
		kind = types.U32
	case "u64":
		kind = types.U64

	case "i8":
		kind = types.I8
	case "i16":
		kind = types.I16
	case "i32":
		kind = types.I32
	case "i64":
		kind = types.I64

	case "f32":
		kind = types.F32
	case "f64":
		kind = types.F64

	default:
		return &types.UnresolvedType{Identifier: ident}, nil
	}

	return types.Primitive(kind), nil
}

func (p *parser) consume(kind scanner.TokenKind, msg string) (scanner.Token, *core.Diagnostic) {
	if p.check(kind) {
		return p.advance(), nil
	}

	return scanner.Token{}, p.error(p.next, msg)
}

func (p *parser) consume2(kind scanner.TokenKind) scanner.Token {
	if p.check(kind) {
		return p.advance()
	}

	return scanner.Token{Kind: scanner.Error}
}

func (p *parser) error(token scanner.Token, format string, args ...any) *core.Diagnostic {
	return &core.Diagnostic{
		Kind:    core.ErrorKind,
		Range:   ast.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	}
}

func (p *parser) error2(token scanner.Token, format string, args ...any) {
	p.reporter.Report(core.Diagnostic{
		Kind:    core.ErrorKind,
		Range:   ast.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *parser) match(kinds ...scanner.TokenKind) bool {
	for _, kind := range kinds {
		if p.check(kind) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *parser) check(kind scanner.TokenKind) bool {
	if p.isAtEnd() {
		return false
	}
	return p.next.Kind == kind
}

func (p *parser) advance() scanner.Token {
	if !p.isAtEnd() {
		p.previous = p.current
		p.current = p.next
		p.next = p.scanner.Next()
	}

	return p.current
}

func (p *parser) isAtEnd() bool {
	return p.next.Kind == scanner.Eof
}
