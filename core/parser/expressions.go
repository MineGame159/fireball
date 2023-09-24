package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
	"fireball/core/types"
)

func (p *parser) expression() ast.Expr {
	return p.assignment()
}

func (p *parser) assignment() ast.Expr {
	// Cascade
	expr := p.logicalOr()
	if expr == nil {
		return nil
	}

	// = += -= *= /= %=
	if p.match(scanner.Equal, scanner.PlusEqual, scanner.MinusEqual, scanner.StarEqual, scanner.SlashEqual, scanner.PercentageEqual) {
		op := p.current

		// Cascade
		value := p.assignment()
		if value == nil {
			return nil
		}

		// Return
		expr := &ast.Assignment{
			Assignee: expr,
			Op:       op,
			Value:    value,
		}

		expr.SetRangeNode(expr.Assignee, expr.Value)
		expr.SetChildrenParent()

		return expr
	}

	// Return
	return expr
}

func (p *parser) logicalOr() ast.Expr {
	// Cascade
	expr := p.logicalAnd()
	if expr == nil {
		return nil
	}

	// ||
	for p.match(scanner.Or) {
		op := p.current

		// Cascade
		right := p.logicalAnd()
		if right == nil {
			return nil
		}

		// Set
		logical := &ast.Logical{
			Left:  expr,
			Op:    op,
			Right: right,
		}

		logical.SetRangeNode(expr, right)
		logical.SetChildrenParent()

		expr = logical
	}

	// Return
	return expr
}

func (p *parser) logicalAnd() ast.Expr {
	// Cascade
	expr := p.bitwiseOr()
	if expr == nil {
		return nil
	}

	// &&
	for p.match(scanner.And) {
		op := p.current

		// Cascade
		right := p.bitwiseOr()
		if right == nil {
			return nil
		}

		// Set
		logical := &ast.Logical{
			Left:  expr,
			Op:    op,
			Right: right,
		}

		logical.SetRangeNode(expr, right)
		logical.SetChildrenParent()

		expr = logical
	}

	// Return
	return expr
}

func (p *parser) bitwiseOr() ast.Expr {
	// Cascade
	expr := p.bitwiseAnd()
	if expr == nil {
		return nil
	}

	// |
	for p.match(scanner.Pipe) {
		op := p.current

		// Cascade
		right := p.bitwiseAnd()
		if right == nil {
			return nil
		}

		// Set
		logical := &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}

		logical.SetRangeNode(expr, right)
		logical.SetChildrenParent()

		expr = logical
	}

	// Return
	return expr
}

func (p *parser) bitwiseAnd() ast.Expr {
	// Cascade
	expr := p.equality()
	if expr == nil {
		return nil
	}

	// &
	for p.match(scanner.Ampersand) {
		op := p.current

		// Cascade
		right := p.equality()
		if right == nil {
			return nil
		}

		// Set
		logical := &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}

		logical.SetRangeNode(expr, right)
		logical.SetChildrenParent()

		expr = logical
	}

	// Return
	return expr
}

func (p *parser) equality() ast.Expr {
	// Cascade
	expr := p.comparison()
	if expr == nil {
		return nil
	}

	// == !=
	for p.match(scanner.EqualEqual, scanner.BangEqual) {
		op := p.current

		// Cascade
		right := p.comparison()
		if right == nil {
			return nil
		}

		// Set
		binary := &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}

		binary.SetRangeNode(expr, right)
		binary.SetChildrenParent()

		expr = binary
	}

	// Return
	return expr
}

func (p *parser) comparison() ast.Expr {
	// Cascade
	expr := p.shift()
	if expr == nil {
		return nil
	}

	// < <= > >=
	for p.match(scanner.Less, scanner.LessEqual, scanner.Greater, scanner.GreaterEqual) {
		op := p.current

		// Cascade
		right := p.shift()
		if right == nil {
			return nil
		}

		// Set
		binary := &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}

		binary.SetRangeNode(expr, right)
		binary.SetChildrenParent()

		expr = binary
	}

	// Return
	return expr
}

func (p *parser) shift() ast.Expr {
	// Cascade
	expr := p.term()
	if expr == nil {
		return nil
	}

	// << >>
	for p.match(scanner.LessLess, scanner.GreaterGreater) {
		op := p.current

		// Cascade
		right := p.term()
		if right == nil {
			return nil
		}

		// Set
		binary := &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}

		binary.SetRangeNode(expr, right)
		binary.SetChildrenParent()

		expr = binary
	}

	// Return
	return expr
}

func (p *parser) term() ast.Expr {
	// Cascade
	expr := p.factor()
	if expr == nil {
		return nil
	}

	// + -
	for p.match(scanner.Plus, scanner.Minus) {
		op := p.current

		// Cascade
		right := p.factor()
		if right == nil {
			return nil
		}

		// Set
		binary := &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}

		binary.SetRangeNode(expr, right)
		binary.SetChildrenParent()

		expr = binary
	}

	// Return
	return expr
}

func (p *parser) factor() ast.Expr {
	// Cascade
	expr := p.unary()
	if expr == nil {
		return nil
	}

	// * / %
	for p.match(scanner.Star, scanner.Slash, scanner.Percentage) {
		op := p.current

		// Cascade
		right := p.unary()
		if right == nil {
			return nil
		}

		// Set
		binary := &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}

		binary.SetRangeNode(expr, right)
		binary.SetChildrenParent()

		expr = binary
	}

	// Return
	return expr
}

func (p *parser) unary() ast.Expr {
	// ! - & * ++ --
	if p.match(scanner.Bang, scanner.Minus, scanner.Ampersand, scanner.Star, scanner.PlusPlus, scanner.MinusMinus) {
		op := p.current

		// Cascade
		right := p.unary()
		if right == nil {
			return nil
		}

		// Return
		expr := &ast.Unary{
			Op:     op,
			Value:  right,
			Prefix: true,
		}

		expr.SetRangeToken(op, p.current)
		expr.SetChildrenParent()

		return expr
	}

	// Return cascade
	return p.postfix()
}

func (p *parser) postfix() ast.Expr {
	// Cascade
	expr := p.call()
	if expr == nil {
		return nil
	}

	// ++ --
	if p.match(scanner.PlusPlus, scanner.MinusMinus) {
		expr := &ast.Unary{
			Op:     p.current,
			Value:  expr,
			Prefix: false,
		}

		expr.SetRangePos(expr.Range().Start, core.TokenToPos(p.current, true))
		expr.SetChildrenParent()

		return expr
	}

	// Return
	return expr
}

func (p *parser) call() ast.Expr {
	// Cascade
	expr := p.primary()
	if expr == nil {
		return nil
	}

	for {
		if p.match(scanner.LeftParen) {
			// (
			expr = p.finishCall(expr)
			if expr == nil {
				return nil
			}
		} else if p.match(scanner.LeftBracket) {
			// {
			expr = p.finishIndex(expr)
			if expr == nil {
				return nil
			}
		} else if p.match(scanner.Dot) {
			// .
			expr = p.finishMember(expr)
			if expr == nil {
				return nil
			}
		} else if p.match(scanner.As) {
			// as
			expr = p.finishCast(expr)
			if expr == nil {
				return nil
			}
		} else {
			break
		}
	}

	// Return
	return expr
}

func (p *parser) finishCall(callee ast.Expr) ast.Expr {
	// Arguments
	args := make([]ast.Expr, 0, 4)

	for p.canLoop(scanner.RightParen) {
		expr := p.expression()
		if expr == nil {
			return nil
		}

		p.match(scanner.Comma)

		args = append(args, expr)
	}

	if token := p.consume(scanner.RightParen, "Expected ')' after call arguments."); token.IsError() {
		return nil
	}

	// Return
	expr := &ast.Call{
		Token_: p.current,
		Callee: callee,
		Args:   args,
	}

	expr.SetRangePos(callee.Range().Start, core.TokenToPos(p.current, true))
	expr.SetChildrenParent()

	return expr
}

func (p *parser) finishIndex(value ast.Expr) ast.Expr {
	token := p.current

	// Index expression
	index := p.expression()
	if index == nil {
		return nil
	}

	// Right bracket
	if token := p.consume(scanner.RightBracket, "Expected ']' after index expression."); token.IsError() {
		return nil
	}

	// Return
	expr := &ast.Index{
		Token_: token,
		Value:  value,
		Index:  index,
	}

	expr.SetRangePos(value.Range().Start, core.TokenToPos(p.current, true))
	expr.SetChildrenParent()

	return expr
}

func (p *parser) finishMember(value ast.Expr) ast.Expr {
	// Name
	name := p.consume(scanner.Identifier, "Expected member name.")
	if name.IsError() {
		return nil
	}

	// Return
	expr := &ast.Member{
		Value: value,
		Name:  name,
	}

	expr.SetRangePos(value.Range().Start, core.TokenToPos(p.current, true))
	expr.SetChildrenParent()

	return expr
}

func (p *parser) finishCast(value ast.Expr) ast.Expr {
	token := p.current

	// Type
	target := p.parseType()
	if target == nil {
		return nil
	}

	// Return
	expr := &ast.Cast{
		Token_: token,
		Target: target,
		Expr:   value,
	}

	expr.SetRangePos(value.Range().Start, core.TokenToPos(p.current, true))
	expr.SetChildrenParent()

	return expr
}

func (p *parser) primary() ast.Expr {
	// nil true false 0.0 'c' "str"
	if p.match(scanner.Nil, scanner.True, scanner.False, scanner.Number, scanner.Hex, scanner.Binary, scanner.Character, scanner.String) {
		expr := &ast.Literal{
			Value: p.current,
		}

		expr.SetRangeToken(p.current, p.current)
		expr.SetChildrenParent()

		return expr
	}

	// abc
	if p.match(scanner.Identifier) {
		token := p.current

		// Initializer
		if p.match(scanner.LeftBrace) {
			return p.structInitializer(false, token, types.Unresolved(token, core.TokenToRange(token)))
		}

		// Sizeof / Alignof
		if (token.Lexeme == "sizeof" || token.Lexeme == "alignof") && p.match(scanner.LeftParen) {
			return p.typeCall(token)
		}

		// New
		if token.Lexeme == "new" && p.check(scanner.Identifier) {
			type_ := p.parseType()
			if type_ == nil {
				return nil
			}

			// Struct
			if p.match(scanner.LeftBrace) {
				return p.structInitializer(true, token, type_)
			}

			// Array
			if p.match(scanner.LeftBracket) {
				return p.newArray(token, type_)
			}

			// Error
			p.error(p.next, "Expected either a '{' or '['.")
			return nil
		}

		// abc
		expr := &ast.Identifier{
			Identifier: token,
		}

		expr.SetRangeToken(token, token)
		expr.SetChildrenParent()

		return expr
	}

	// [
	if p.match(scanner.LeftBracket) {
		return p.arrayInitializer()
	}

	// (
	if p.match(scanner.LeftParen) {
		token := p.current

		// Expression
		expr := p.expression()
		if expr == nil {
			return nil
		}

		// Right paren
		if token := p.consume(scanner.RightParen, "Expected ')' after expression."); token.IsError() {
			return nil
		}

		// Return
		group := &ast.Group{
			Token_: token,
			Expr:   expr,
		}

		group.SetRangeToken(token, p.current)
		group.SetChildrenParent()

		return group
	}

	// Error
	p.error(p.next, "Expected expression.")
	return nil
}

func (p *parser) structInitializer(new bool, token scanner.Token, target types.Type) ast.Expr {
	// Fields
	fields := make([]ast.InitField, 0, 4)

	for p.canLoop(scanner.RightBrace) {
		// Comma
		if len(fields) > 0 {
			if token := p.consume(scanner.Comma, "Expected ',' between fields."); token.IsError() {
				return nil
			}
		}

		// Name
		name := p.consume(scanner.Identifier, "Expected field name.")
		if name.IsError() {
			return nil
		}

		// Colon
		if token := p.consume(scanner.Colon, "Expected ':' after field name."); token.IsError() {
			return nil
		}

		// Value
		value := p.expression()
		if value == nil {
			return nil
		}

		// Add
		fields = append(fields, ast.InitField{
			Name:  name,
			Value: value,
		})
	}

	// Right brace
	if token := p.consume(scanner.RightBrace, "Expected '}' after initializer fields."); token.IsError() {
		return nil
	}

	// Return
	expr := &ast.StructInitializer{
		Token_: token,
		New:    new,
		Target: target,
		Fields: fields,
	}

	expr.SetRangeToken(token, p.current)
	expr.SetChildrenParent()

	return expr
}

func (p *parser) arrayInitializer() ast.Expr {
	token := p.current

	// Values
	values := make([]ast.Expr, 0, 8)

	for p.canLoop(scanner.RightBracket) {
		// Comma
		if len(values) > 0 {
			if token := p.consume(scanner.Comma, "Expected ',' between array values."); token.IsError() {
				return nil
			}
		}

		// Value
		expr := p.expression()
		if expr == nil {
			return nil
		}

		values = append(values, expr)
	}

	// Right bracket
	if token := p.consume(scanner.RightBracket, "Expected ']' after array values."); token.IsError() {
		return nil
	}

	// Return
	expr := &ast.ArrayInitializer{
		Token_: token,
		Values: values,
	}

	expr.SetRangeToken(token, p.current)
	expr.SetChildrenParent()

	return expr
}

func (p *parser) newArray(token scanner.Token, type_ types.Type) ast.Expr {
	count := p.expression()
	if count == nil {
		return nil
	}

	if token := p.consume(scanner.RightBracket, "Expected ']' after array count."); token.IsError() {
		return nil
	}

	expr := &ast.NewArray{
		Token_: token,
		Type_:  type_,
		Count:  count,
	}

	expr.SetRangeToken(token, p.current)
	expr.SetChildrenParent()

	return expr
}

func (p *parser) typeCall(name scanner.Token) ast.Expr {
	// Type
	type_ := p.parseType()
	if type_ == nil {
		return nil
	}

	// Right paren
	if token := p.consume(scanner.RightParen, "Expected ')' after type."); token.IsError() {
		return nil
	}

	// Return
	expr := &ast.TypeCall{
		Name:   name,
		Target: type_,
	}

	expr.SetRangeToken(name, p.current)
	expr.SetChildrenParent()

	return expr
}
