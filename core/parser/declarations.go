package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
	"strconv"
)

func (p *parser) declaration() ast.Decl {
	attributesStart := p.next
	attributes := p.parseAttributes()
	start := p.next

	if p.match(scanner.Struct) {
		if len(attributes) > 0 {
			p.error(attributesStart, "Structs cannot have attributes.")
		}

		return p.struct_()
	}

	if p.match(scanner.Impl) {
		if len(attributes) > 0 {
			p.error(attributesStart, "Implementations cannot have attributes.")
		}

		return p.impl()
	}

	if p.match(scanner.Enum) {
		if len(attributes) > 0 {
			p.error(attributesStart, "Enums cannot have attributes.")
		}

		return p.enum()
	}

	if p.match(scanner.Func) {
		return p.function(start, attributes, 0)
	}

	if p.match(scanner.Static) {
		if p.match(scanner.Func) {
			return p.function(start, attributes, ast.Static)
		}
	}

	// Error
	p.error(p.next, "Expected declaration, got '%s'.", p.next)
	p.syncToDecl()

	return nil
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
	staticFields := make([]ast.Field, 0)
	fields := make([]ast.Field, 0, 4)

	for p.canLoopAdvanced(scanner.RightBrace, scanner.Static, scanner.Identifier) {
		// Static
		static := false

		if p.match(scanner.Static) {
			static = true
		}

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
		field := ast.Field{
			Name: name,
			Type: type_,
		}

		if static {
			staticFields = append(staticFields, field)
		} else {
			fields = append(fields, field)
		}

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
		Name:         name,
		StaticFields: staticFields,
		Fields:       fields,
	}

	decl.SetRangeToken(start, p.current)
	decl.SetChildrenParent()

	for i := range decl.StaticFields {
		decl.StaticFields[i].Parent = decl
	}

	for i := range decl.Fields {
		decl.Fields[i].Parent = decl
	}

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

	for p.canLoopAdvanced(scanner.RightBrace, scanner.Hashtag, scanner.Static, scanner.Func) {
		start := p.next

		attributes := p.parseAttributes()
		flags := ast.FuncFlags(0)

		if p.match(scanner.Static) {
			flags = ast.Static
		}

		if token := p.consume(scanner.Func, "Expected 'func' to start a function."); token.IsError() {
			return nil
		}

		function := p.function(start, attributes, flags)
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

func (p *parser) function(start scanner.Token, attributes []any, flags ast.FuncFlags) ast.Decl {
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
	decl := &ast.Func{
		Attributes: attributes,
		Flags:      flags,
		Name:       name,
		Params:     params,
		Returns:    returns,
	}

	if decl.HasBody() {
		decl.Body = make([]ast.Stmt, 0, 8)

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

			decl.Body = append(decl.Body, stmt)
		}

		if brace := p.consume(scanner.RightBrace, "Expected '}' after function body."); brace.IsError() {
			p.syncToDecl()
		}
	}

	// Return
	decl.SetRangeToken(start, p.current)
	decl.SetChildrenParent()

	return decl
}

// Attributes

func (p *parser) parseAttributes() []any {
	var attributes []any

	if p.match(scanner.Hashtag) {
		// [
		if token := p.consume(scanner.LeftBracket, "Expected '[' before attributes."); token.IsError() {
			p.syncToDecl()
			return attributes
		}

		// Attributes
		for {
			// Comma
			if len(attributes) > 0 {
				if token := p.consume(scanner.Comma, "Expected ',' between attributes."); token.IsError() {
					p.syncToDecl()
					return attributes
				}
			}

			// Name
			name := p.consume(scanner.Identifier, "Expected attribute name.")
			if name.IsError() {
				p.syncToDecl()
				return attributes
			}

			// Args
			var args []string
			argsError := false

			if p.match(scanner.LeftParen) {
				for {
					// Comma
					if len(args) > 0 {
						argsError = true
						p.syncTo(scanner.RightParen)

						break
					}

					// Value
					value := p.consume(scanner.String, "Expected attribute argument.")

					if value.IsError() {
						argsError = true
						p.syncTo(scanner.RightParen)

						break
					}

					args = append(args, value.Lexeme[1:len(value.Lexeme)-1])

					// Loop
					if !p.canLoopAdvanced(scanner.RightParen, scanner.Comma) {
						break
					}
				}

				// )
				if token := p.consume(scanner.RightParen, "Expected ')' after attribute arguments."); token.IsError() {
					p.syncToDecl()
					return attributes
				}
			}

			// Attribute
			if !argsError {
				attribute := p.parseAttribute(name, args)

				if attribute == nil {
					p.error(name, "Unknown attribute with name '%s'.", name)
				} else {
					attributes = append(attributes, attribute)
				}
			}

			// Loop
			if !p.canLoopAdvanced(scanner.RightBracket, scanner.Comma) {
				break
			}
		}

		// ]
		if token := p.consume(scanner.RightBracket, "Expected ']' after attributes."); token.IsError() {
			p.syncToDecl()
			return attributes
		}
	}

	return attributes
}

func (p *parser) parseAttribute(token scanner.Token, args []string) any {
	switch token.Lexeme {
	case "Extern":
		name := ""

		if len(args) == 1 {
			name = args[0]
		} else if len(args) > 1 {
			p.error(token, "Extern attribute only has 1 optional parameter.")
		}

		return types.ExternAttribute{Name: name}

	case "Intrinsic":
		name := ""

		if len(args) == 1 {
			name = args[0]
		} else if len(args) > 1 {
			p.error(token, "Intrinsic attribute only has 1 optional parameter.")
		}

		return types.IntrinsicAttribute{Name: name}

	case "Inline":
		if len(args) != 0 {
			p.error(token, "Inline attribute doesn't have any parameters.")
		}

		return types.InlineAttribute{}

	default:
		return nil
	}
}

// Helpers

func (p *parser) syncBeforeFieldOrDecl() bool {
	for !p.isAtEnd() {
		switch p.next.Kind {
		case scanner.Struct, scanner.Enum, scanner.Static, scanner.Func:
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
