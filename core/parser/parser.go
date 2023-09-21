package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"fireball/core/utils"
	"fmt"
	"strconv"
)

type parser struct {
	scanner *scanner.Scanner

	previous scanner.Token
	current  scanner.Token
	next     scanner.Token

	reporter utils.Reporter
}

func Parse(reporter utils.Reporter, scanner *scanner.Scanner) []ast.Decl {
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

// Types

func (p *parser) parseType() types.Type {
	if p.match(scanner.LeftBracket) {
		return p.parseArrayType()
	}
	if p.match(scanner.Star) {
		return p.parsePointerType()
	}
	if p.match(scanner.LeftParen) {
		return p.parseFunctionType()
	}

	return p.parseIdentifierType()
}

func (p *parser) parseArrayType() types.Type {
	start := p.current

	// Count
	token := p.consume(scanner.Number, "Expected array size.")
	if token.IsError() {
		return nil
	}

	count, err_ := strconv.Atoi(token.Lexeme)

	if err_ != nil {
		p.error(token, "Invalid array size.")
		return nil
	}

	if count < 0 {
		p.error(token, "Invalid array size.")
		return nil
	}

	// Right bracket
	if token := p.consume(scanner.RightBracket, "Expected ']' after array size."); token.IsError() {
		return nil
	}

	// Base
	base := p.parseType()
	if base == nil {
		return nil
	}

	// Return
	return types.Array(uint32(count), base, core.TokensToRange(start, p.current))
}

func (p *parser) parsePointerType() types.Type {
	start := p.current

	// Pointee
	pointee := p.parseType()
	if pointee == nil {
		return nil
	}

	// return
	return types.Pointer(pointee, core.TokensToRange(start, p.current))
}

func (p *parser) parseFunctionType() types.Type {
	start := p.current
	flags := ast.FuncFlags(0)

	// Parameters
	params := make([]ast.Param, 0, 4)

	for p.canLoop(scanner.RightParen, scanner.Dot) {
		name := p.consume(scanner.Identifier, "Expected parameter name.")
		if name.IsError() {
			return nil
		}

		type_ := p.parseType()
		if type_ == nil {
			return nil
		}

		p.match(scanner.Comma)

		params = append(params, ast.Param{
			Name: name,
			Type: type_,
		})
	}

	if p.match(scanner.Dot) && p.match(scanner.Dot) && p.match(scanner.Dot) {
		flags |= ast.Variadic
	}

	if paren := p.consume(scanner.RightParen, "Expected ')' after function parameters."); paren.IsError() {
		return nil
	}

	// Returns
	returns := p.parseType()
	if returns == nil {
		return nil
	}

	// Return
	type_ := &ast.Func{
		Flags:   flags,
		Params:  params,
		Returns: returns,
	}

	type_.SetRangeToken(start, p.current)
	return type_
}

func (p *parser) parseIdentifierType() types.Type {
	// Name
	ident := p.consume(scanner.Identifier, "Expected type name.")
	if ident.IsError() {
		return nil
	}

	range_ := core.TokenToRange(ident)

	// Select kind
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
		// Unresolved
		return types.Unresolved(ident, range_)
	}

	// Primitive
	return types.Primitive(kind, range_)
}

// Helpers

func (p *parser) consume(kind scanner.TokenKind, msg string) scanner.Token {
	if p.check(kind) {
		return p.advance()
	}

	p.error(p.next, msg)
	return scanner.Token{Kind: scanner.Error}
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

func (p *parser) canLoop(notNext ...scanner.TokenKind) bool {
	if p.isAtEnd() {
		return false
	}

	for _, kind := range notNext {
		if p.next.Kind == kind {
			return false
		}
	}

	return true
}

func (p *parser) canLoopAdvanced(notNext scanner.TokenKind, next ...scanner.TokenKind) bool {
	if p.isAtEnd() {
		return false
	}

	ok := false

	for _, kind := range next {
		if p.next.Kind == kind {
			ok = true
			break
		}
	}

	if !ok {
		return false
	}

	return p.next.Kind != notNext
}

func (p *parser) isAtEnd() bool {
	return p.next.Kind == scanner.Eof
}

// Error handling

func (p *parser) syncToDecl() {
	p.syncTo(scanner.Struct, scanner.Enum, scanner.Extern, scanner.Func)
}

func (p *parser) syncToStmt() bool {
	for !p.isAtEnd() {
		switch p.next.Kind {
		case scanner.Semicolon:
			p.advance()
			return true

		case scanner.Struct, scanner.Enum, scanner.Extern, scanner.Func, scanner.RightBrace:
			return false

		default:
			p.advance()
		}
	}

	return false
}

func (p *parser) syncTo(kinds ...scanner.TokenKind) {
	for !p.isAtEnd() {
		for _, kind := range kinds {
			if p.next.Kind == kind {
				return
			}
		}

		p.advance()
	}
}

func (p *parser) error(token scanner.Token, format string, args ...any) {
	p.reporter.Report(utils.Diagnostic{
		Kind:    utils.ErrorKind,
		Range:   core.TokenToRange(token),
		Message: fmt.Sprintf(format, args...),
	})
}
