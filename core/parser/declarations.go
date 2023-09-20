package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"strconv"
)

func (p *parser) declaration() ast.Decl {
	start := p.next

	if p.match(scanner.Struct) {
		return p.struct_()
	}
	if p.match(scanner.Impl) {
		return p.impl()
	}
	if p.match(scanner.Enum) {
		return p.enum()
	}
	if p.match(scanner.Func) {
		return p.function(start, 0)
	}

	if flags := p.functionFlags(); flags != 0 {
		if p.match(scanner.Func) {
			return p.function(start, flags)
		}
	}

	// Error
	p.error(p.next, "Expected declaration, got '%s'.", p.next)
	p.syncToDecl()

	return nil
}

func (p *parser) functionFlags() ast.FuncFlags {
	flags := ast.FuncFlags(0)

	for {
		if p.match(scanner.Static) {
			flags |= ast.Static
		} else if p.match(scanner.Extern) {
			flags |= ast.Extern
		} else {
			break
		}
	}

	return flags
}

func (p *parser) struct_() ast.Decl {
	start := p.current

	// Name
	name := p.consume(scanner.Identifier, "Expected struct name.")

	if name.IsError() {
		p.syncToDecl()
		return nil
	}

	// Left brace
	if brace := p.consume(scanner.LeftBrace, "Expected '{' after struct name."); brace.IsError() {
		p.syncToDecl()
		return nil
	}

	// Fields
	fields := make([]ast.Field, 0, 4)

	for p.canLoopAdvanced(scanner.RightBrace, scanner.Identifier) {
		// Name
		name := p.consume(scanner.Identifier, "Expected field name.")

		if name.IsError() {
			if !p.syncBeforeFieldOrDecl() {
				return nil
			}
			continue
		}

		// Type
		type_ := p.parseType()

		if type_ == nil {
			if !p.syncBeforeFieldOrDecl() {
				return nil
			}
			continue
		}

		// Add
		fields = append(fields, ast.Field{
			Name: name,
			Type: type_,
		})

		// Comma
		if token := p.consume(scanner.Comma, "Expected ',' after field type."); token.IsError() {
			if p.check(scanner.RightBrace) {
				break
			}
			if !p.syncBeforeFieldOrDecl() {
				return nil
			}

			continue
		}
	}

	// Right brace
	if brace := p.consume(scanner.RightBrace, "Expected '}' after struct fields."); brace.IsError() {
		p.syncToDecl()
	}

	// Return
	decl := &ast.Struct{
		Name:   name,
		Fields: fields,
	}

	decl.SetRangeToken(start, p.current)
	decl.SetChildrenParent()

	return decl
}

func (p *parser) impl() ast.Decl {
	start := p.current

	// Name
	struct_ := p.consume(scanner.Identifier, "Expected struct name.")

	if struct_.IsError() {
		p.syncToDecl()
		return nil
	}

	// Left brace
	if brace := p.consume(scanner.LeftBrace, "Expected '{' after struct name."); brace.IsError() {
		p.syncToDecl()
		return nil
	}

	// Functions
	functions := make([]ast.Decl, 0, 8)

	for p.canLoopAdvanced(scanner.RightBrace, scanner.Static, scanner.Extern, scanner.Func) {
		flags := p.functionFlags()

		p.advance()

		function := p.function(p.current, flags)
		if function == nil {
			p.syncToDecl()
			return nil
		}

		functions = append(functions, function)
	}

	// Right brace
	if brace := p.consume(scanner.RightBrace, "Expected '}' after struct methods."); brace.IsError() {
		p.syncToDecl()
	}

	// Return
	decl := &ast.Impl{
		Struct:    struct_,
		Functions: functions,
	}

	decl.SetRangeToken(start, p.current)
	decl.SetChildrenParent()

	return decl
}

func (p *parser) enum() ast.Decl {
	start := p.current

	// Name
	name := p.consume(scanner.Identifier, "Expected enum name.")

	if name.IsError() {
		p.syncToDecl()
		return nil
	}

	// Type
	var type_ types.Type

	if !p.check(scanner.LeftBrace) {
		type__ := p.parseType()

		if type__ == nil {
			p.syncToDecl()
			return nil
		}

		type_ = type__
	}

	// Left brace
	if brace := p.consume(scanner.LeftBrace, "Expected '{' before enum cases."); brace.IsError() {
		p.syncToDecl()
		return nil
	}

	// Enum cases
	cases := make([]ast.EnumCase, 0, 8)
	lastValue := -1

	for p.canLoopAdvanced(scanner.RightBrace, scanner.Identifier) {
		// Name
		name := p.consume(scanner.Identifier, "Expected enum case name.")

		if name.IsError() {
			if !p.syncBeforeFieldOrDecl() {
				return nil
			}
			continue
		}

		// Value
		value := lastValue + 1
		inferValue := true
		validValue := true

		if p.match(scanner.Equal) {
			literal := p.consume(scanner.Number, "Enum case values can only be integers.")

			if literal.IsError() {
				if !p.syncBeforeFieldOrDecl() {
					return nil
				}
				continue
			}

			number, err := strconv.Atoi(literal.Lexeme)

			if err != nil {
				p.error(literal, "Invalid integer.")
				validValue = false
			}

			value = number
			inferValue = false
		}

		lastValue = value

		// Add
		if validValue {
			cases = append(cases, ast.EnumCase{
				Name:       name,
				Value:      value,
				InferValue: inferValue,
			})
		}

		// Comma
		if comma := p.consume(scanner.Comma, "Expected ',' before next enum case."); comma.IsError() {
			if p.check(scanner.RightBrace) {
				break
			}
			if !p.syncBeforeFieldOrDecl() {
				return nil
			}

			continue
		}
	}

	// Right brace
	if brace := p.consume(scanner.RightBrace, "Expected '}' after enum cases."); brace.IsError() {
		p.syncToDecl()
	}

	// Return
	decl := &ast.Enum{
		Name:      name,
		Type:      type_,
		InferType: type_ == nil,
		Cases:     cases,
	}

	decl.SetRangeToken(start, p.current)
	decl.SetChildrenParent()

	return decl
}

func (p *parser) function(start scanner.Token, flags ast.FuncFlags) ast.Decl {
	// Name
	name := p.consume(scanner.Identifier, "Expected function name.")

	if name.IsError() {
		p.syncToDecl()
		return nil
	}

	// Parameters
	if paren := p.consume(scanner.LeftParen, "Expected '(' after function name."); paren.IsError() {
		p.syncToDecl()
		return nil
	}

	params := make([]ast.Param, 0, 4)

	for p.canLoop(scanner.RightParen, scanner.Dot) {
		name := p.consume(scanner.Identifier, "Expected parameter name.")
		if name.IsError() {
			p.syncToDecl()
			return nil
		}

		type_ := p.parseType()
		if type_ == nil {
			p.syncToDecl()
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
		p.syncToDecl()
		return nil
	}

	// Returns
	var returns types.Type

	if !p.check(scanner.LeftBrace) {
		type_ := p.parseType()
		if type_ == nil {
			p.syncToDecl()
			return nil
		}

		returns = type_
	} else {
		returns = types.Primitive(types.Void, core.Range{})
	}

	// Body
	var body []ast.Stmt

	if flags&ast.Extern == 0 {
		body = make([]ast.Stmt, 0, 8)

		if brace := p.consume(scanner.LeftBrace, "Expected '{' before function body."); brace.IsError() {
			p.syncToDecl()
			return nil
		}

		for p.canLoop(scanner.RightBrace) {
			stmt := p.statement()

			if stmt == nil {
				if !p.syncToStmt() {
					break
				}
				continue
			}

			body = append(body, stmt)
		}

		if brace := p.consume(scanner.RightBrace, "Expected '}' after function body."); brace.IsError() {
			p.syncToDecl()
		}
	}

	// Return
	decl := &ast.Func{
		Flags:   flags,
		Name:    name,
		Params:  params,
		Returns: returns,
		Body:    body,
	}

	decl.SetRangeToken(start, p.current)
	decl.SetChildrenParent()

	return decl
}

// Helpers

func (p *parser) syncBeforeFieldOrDecl() bool {
	for !p.isAtEnd() {
		switch p.next.Kind {
		case scanner.Struct, scanner.Enum, scanner.Extern, scanner.Func:
			return false

		case scanner.Comma:
			p.advance()
			return true

		default:
			p.advance()
		}
	}

	return false
}
