package parser

import (
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
)

func (p *parser) declaration() ast.Decl {
	if p.match(scanner.Func) {
		return p.function()
	}

	// Error
	p.error2(p.next, "Expected declaration, got '%s'.", p.next)
	p.syncToDecl()

	return nil
}

func (p *parser) function() ast.Decl {
	// Name
	name := p.consume2(scanner.Identifier)

	if name.IsError() {
		p.error2(p.next, "Expected function name.")
		if !p.syncToDecl() {
			return nil
		}
		return nil
	}

	// Parameters
	if paren := p.consume2(scanner.LeftParen); paren.IsError() {
		p.error2(p.next, "Expected '(' after function name.")
		if !p.syncToDecl() {
			return nil
		}
		return nil
	}

	params := make([]ast.Param, 0, 4)

	for !p.check(scanner.RightParen) {
		name := p.consume2(scanner.Identifier)
		if name.IsError() {
			p.error2(p.next, "Expected parameter name.")
			if !p.syncToDecl() {
				return nil
			}
			return nil
		}

		type_, err := p.parseType()
		if err != nil {
			p.reporter.Report(*err)
			if !p.syncToDecl() {
				return nil
			}
			return nil
		}

		p.match(scanner.Comma)

		params = append(params, ast.Param{
			Name: name,
			Type: type_,
		})
	}

	if paren := p.consume2(scanner.RightParen); paren.IsError() {
		p.error2(p.next, "Expected ')' after function parameters.")
		if !p.syncToDecl() {
			return nil
		}
		return nil
	}

	// Returns
	var returns types.Type = types.Primitive(types.Void)

	if !p.check(scanner.LeftBrace) {
		type_, err := p.parseType()
		if err != nil {
			p.reporter.Report(*err)
			if !p.syncToDecl() {
				return nil
			}
			return nil
		}

		returns = type_
	}

	// Body
	if brace := p.consume2(scanner.LeftBrace); brace.IsError() {
		p.error2(p.next, "Expected '{' before function body.")
		if !p.syncToDecl() {
			return nil
		}
		return nil
	}

	body := make([]ast.Stmt, 0, 8)

	for !p.check(scanner.RightBrace) {
		stmt, err := p.statement()

		if err != nil {
			p.reporter.Report(*err)

			if !p.syncToStmt() {
				return nil
			}
		} else {
			body = append(body, stmt)
		}
	}

	if brace := p.consume2(scanner.RightBrace); brace.IsError() {
		p.error2(p.next, "Expected '}' after function body.")
		if !p.syncToDecl() {
			return nil
		}
		return nil
	}

	// Return
	return &ast.Func{
		Name:    name,
		Params:  params,
		Returns: returns,
		Body:    body,
	}
}

func (p *parser) syncToDecl() bool {
	// Advance until we hit a token that stars a declaration or EOF
	for {
		// Handle EOF
		if p.isAtEnd() {
			return false
		}

		// Check token
		switch p.next.Kind {
		case scanner.Func:
			return true

		default:
			p.advance()
		}
	}
}

func (p *parser) syncToStmt() bool {
	// Advance until we hit a semicolon or EOF
	for !p.check(scanner.Semicolon) {
		// Handle EOF
		if p.isAtEnd() {
			return false
		}
		// Advance
		p.advance()
	}

	// Skip semicolon
	p.advance()

	return true
}
