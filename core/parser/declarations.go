package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
)

func (p *parser) declaration() ast.Decl {
	p.extern = false

	if p.match(scanner.Struct) {
		return p.struct_()
	}

	if p.match(scanner.Func) {
		return p.function()
	}

	if p.match(scanner.Extern) {
		p.extern = true

		if p.match(scanner.Func) {
			return p.function()
		}
	}

	// Error
	p.error2(p.next, "Expected declaration, got '%s'.", p.next)
	p.syncToDecl()

	return nil
}

func (p *parser) struct_() ast.Decl {
	start := p.current

	// Name
	name := p.consume2(scanner.Identifier)

	if name.IsError() {
		p.error2(p.next, "Expected struct name.")
		if !p.syncToDecl() {
			return nil
		}
		return nil
	}

	// Left brace
	if brace := p.consume2(scanner.LeftBrace); brace.IsError() {
		p.error2(p.next, "Expected '{' after struct name.")
		if !p.syncToDecl() {
			return nil
		}
		return nil
	}

	// Fields
	fields := make([]ast.Field, 0, 4)

	for !p.check(scanner.RightBrace) {
		// Name
		name := p.consume2(scanner.Identifier)

		if name.IsError() {
			p.error2(p.next, "Expected field name.")
			if !p.syncToDecl() {
				return nil
			}
			return nil
		}

		// Type
		type_, err := p.parseType()

		if err != nil {
			p.reporter.Report(*err)
			if !p.syncToDecl() {
				return nil
			}
			return nil
		}

		// Add
		fields = append(fields, ast.Field{
			Name: name,
			Type: type_,
		})
	}

	// Right brace
	if brace := p.consume2(scanner.RightBrace); brace.IsError() {
		p.error2(p.next, "Expected '}' after struct fields.")
		if !p.syncToDecl() {
			return nil
		}
		return nil
	}

	// Return
	decl := &ast.Struct{
		Name:   name,
		Fields: fields,
	}

	decl.SetRangeToken(start, p.current)
	return decl
}

func (p *parser) function() ast.Decl {
	start := p.current

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
	variadic := false

	for !p.check(scanner.RightParen) && !p.check(scanner.Dot) {
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

	if dot := p.consume2(scanner.Dot); !dot.IsError() {
		if dot := p.consume2(scanner.Dot); !dot.IsError() {
			if dot := p.consume2(scanner.Dot); !dot.IsError() {
				variadic = true

				if !p.extern {
					p.error2(dot, "Only extern functions can be variadic.")
					if !p.syncToDecl() {
						return nil
					}
					return nil
				}
			}
		}
	}

	if paren := p.consume2(scanner.RightParen); paren.IsError() {
		p.error2(p.next, "Expected ')' after function parameters.")
		if !p.syncToDecl() {
			return nil
		}
		return nil
	}

	// Returns
	var returns types.Type

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
	} else {
		returns = types.Primitive(types.Void, core.Range{})
	}

	// Body
	var body []ast.Stmt

	if !p.extern {
		body = make([]ast.Stmt, 0, 8)

		if brace := p.consume2(scanner.LeftBrace); brace.IsError() {
			p.error2(p.next, "Expected '{' before function body.")
			if !p.syncToDecl() {
				return nil
			}
			return nil
		}

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
	}

	// Return
	decl := &ast.Func{
		Extern:   p.extern,
		Name:     name,
		Params:   params,
		Variadic: variadic,
		Returns:  returns,
		Body:     body,
	}

	decl.SetRangeToken(start, p.current)
	return decl
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
		case scanner.Extern, scanner.Func:
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
