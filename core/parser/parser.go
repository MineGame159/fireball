package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fmt"
)

type parser struct {
	scanner *scanner.Scanner

	previous scanner.Token
	current  scanner.Token
	next     scanner.Token

	extern bool

	reporter core.ErrorReporter
}

func Parse(reporter core.ErrorReporter, scanner *scanner.Scanner) []ast.Decl {
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

func (p *parser) parseType() (types.Type, *core.Error) {
	ident, err := p.consume(scanner.Identifier, "Expected type name.")
	if err != nil {
		return nil, err
	}

	switch ident.Lexeme {
	case "void":
		return types.Primitive(types.Void), nil
	case "bool":
		return types.Primitive(types.Bool), nil

	case "u8":
		return types.Primitive(types.U8), nil
	case "u16":
		return types.Primitive(types.U16), nil
	case "u32":
		return types.Primitive(types.U32), nil
	case "u64":
		return types.Primitive(types.U64), nil

	case "i8":
		return types.Primitive(types.I8), nil
	case "i16":
		return types.Primitive(types.I16), nil
	case "i32":
		return types.Primitive(types.I32), nil
	case "i64":
		return types.Primitive(types.I64), nil

	case "f32":
		return types.Primitive(types.F32), nil
	case "f64":
		return types.Primitive(types.F64), nil

	default:
		return nil, p.error(ident, "Unknown type '%s'.", ident)
	}
}

func (p *parser) consume(kind scanner.TokenKind, msg string) (scanner.Token, *core.Error) {
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

func (p *parser) error(token scanner.Token, format string, args ...any) *core.Error {
	return &core.Error{
		Message: fmt.Sprintf(format, args...),
		Line:    token.Line,
		Column:  token.Column,
	}
}

func (p *parser) error2(token scanner.Token, format string, args ...any) {
	p.reporter.Report(core.Error{
		Message: fmt.Sprintf(format, args...),
		Line:    token.Line,
		Column:  token.Column,
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
