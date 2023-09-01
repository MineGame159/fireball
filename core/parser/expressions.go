package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
)

func (p *parser) expression() (ast.Expr, *core.Diagnostic) {
	return p.assignment()
}

func (p *parser) assignment() (ast.Expr, *core.Diagnostic) {
	// Cascade
	expr, err := p.or()
	if err != nil {
		return nil, err
	}

	// + += -= *= /= %=
	if p.match(scanner.Equal, scanner.PlusEqual, scanner.MinusEqual, scanner.StarEqual, scanner.SlashEqual, scanner.PercentageEqual) {
		op := p.current

		// Cascade
		value, err := p.assignment()
		if err != nil {
			return nil, err
		}

		// Return
		expr := &ast.Assignment{
			Assignee: expr,
			Op:       op,
			Value:    value,
		}

		expr.SetRangeNode(expr.Assignee, expr.Value)
		return expr, nil
	}

	// Return
	return expr, nil
}

func (p *parser) or() (ast.Expr, *core.Diagnostic) {
	// Cascade
	expr, err := p.and()
	if err != nil {
		return nil, err
	}

	// ||
	for p.match(scanner.Or) {
		op := p.current

		// Cascade
		right, err := p.and()
		if err != nil {
			return nil, err
		}

		// Set
		left := expr

		expr = &ast.Logical{
			Left:  left,
			Op:    op,
			Right: right,
		}

		expr.SetRangeNode(left, right)
	}

	// Return
	return expr, nil
}

func (p *parser) and() (ast.Expr, *core.Diagnostic) {
	// Cascade
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	// &&
	for p.match(scanner.And) {
		op := p.current

		// Cascade
		right, err := p.equality()
		if err != nil {
			return nil, err
		}

		// Set
		left := expr

		expr = &ast.Logical{
			Left:  left,
			Op:    op,
			Right: right,
		}

		expr.SetRangeNode(left, right)
	}

	// Return
	return expr, nil
}

func (p *parser) equality() (ast.Expr, *core.Diagnostic) {
	// Cascade
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	// == !=
	for p.match(scanner.EqualEqual, scanner.BangEqual) {
		op := p.current

		// Cascade
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}

		// Set
		left := expr

		expr = &ast.Binary{
			Left:  left,
			Op:    op,
			Right: right,
		}

		expr.SetRangeNode(left, right)
	}

	// Return
	return expr, nil
}

func (p *parser) comparison() (ast.Expr, *core.Diagnostic) {
	// Cascade
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	// < <= > >=
	for p.match(scanner.Less, scanner.LessEqual, scanner.Greater, scanner.GreaterEqual) {
		op := p.current

		// Cascade
		right, err := p.term()
		if err != nil {
			return nil, err
		}

		// Set
		left := expr

		expr = &ast.Binary{
			Left:  left,
			Op:    op,
			Right: right,
		}

		expr.SetRangeNode(left, right)
	}

	// Return
	return expr, nil
}

func (p *parser) term() (ast.Expr, *core.Diagnostic) {
	// Cascade
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	// + -
	for p.match(scanner.Plus, scanner.Minus) {
		op := p.current

		// Cascade
		right, err := p.factor()
		if err != nil {
			return nil, err
		}

		// Set
		left := expr

		expr = &ast.Binary{
			Left:  left,
			Op:    op,
			Right: right,
		}

		expr.SetRangeNode(left, right)
	}

	// Return
	return expr, nil
}

func (p *parser) factor() (ast.Expr, *core.Diagnostic) {
	// Cascade
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	// * / %
	for p.match(scanner.Star, scanner.Slash, scanner.Percentage) {
		op := p.current

		// Cascade
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		// Set
		left := expr

		expr = &ast.Binary{
			Left:  left,
			Op:    op,
			Right: right,
		}

		expr.SetRangeNode(left, right)
	}

	// Return
	return expr, nil
}

func (p *parser) unary() (ast.Expr, *core.Diagnostic) {
	// ! - &
	if p.match(scanner.Bang, scanner.Minus, scanner.Ampersand) {
		op := p.current

		// Cascade
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		// Return
		expr := &ast.Unary{
			Op:    op,
			Right: right,
		}

		expr.SetRangeToken(op, p.current)
		return expr, nil
	}

	// Return cascade
	return p.call()
}

func (p *parser) call() (ast.Expr, *core.Diagnostic) {
	// Cascade
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(scanner.LeftParen) {
			// (
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(scanner.LeftBracket) {
			// {
			expr, err = p.finishIndex(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(scanner.Dot) {
			// .
			expr, err = p.finishMember(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(scanner.As) {
			// as
			expr, err = p.finishCast(expr)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	// Return
	return expr, nil
}

func (p *parser) finishCall(callee ast.Expr) (ast.Expr, *core.Diagnostic) {
	// Arguments
	args := make([]ast.Expr, 0, 4)

	for !p.check(scanner.RightParen) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		p.match(scanner.Comma)

		args = append(args, expr)
	}

	if _, err := p.consume(scanner.RightParen, "Expected ')' after call arguments."); err != nil {
		return nil, err
	}

	// Return
	expr := &ast.Call{
		Token_: p.current,
		Callee: callee,
		Args:   args,
	}

	expr.SetRangePos(callee.Range().Start, core.TokenToPos(p.current, true))
	return expr, nil
}

func (p *parser) finishIndex(value ast.Expr) (ast.Expr, *core.Diagnostic) {
	token := p.current

	// Index expression
	index, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Right bracket
	if _, err := p.consume(scanner.RightBracket, "Expected ']' after index expression."); err != nil {
		return nil, err
	}

	// Return
	expr := &ast.Index{
		Token_: token,
		Value:  value,
		Index:  index,
	}

	expr.SetRangePos(value.Range().Start, core.TokenToPos(p.current, true))
	return expr, nil
}

func (p *parser) finishMember(value ast.Expr) (ast.Expr, *core.Diagnostic) {
	// Name
	name, err := p.consume(scanner.Identifier, "Expected member name.")
	if err != nil {
		return nil, err
	}

	// Return
	expr := &ast.Member{
		Value: value,
		Name:  name,
	}

	expr.SetRangePos(value.Range().Start, core.TokenToPos(p.current, true))
	return expr, nil
}

func (p *parser) finishCast(value ast.Expr) (ast.Expr, *core.Diagnostic) {
	token := p.current

	// Type
	type_, err := p.parseType()
	if err != nil {
		return nil, err
	}

	// Return
	cast := &ast.Cast{
		Token_: token,
		Expr:   value,
	}

	cast.SetRangePos(value.Range().Start, core.TokenToPos(p.current, true))
	cast.SetType(type_)
	return cast, nil
}

func (p *parser) primary() (ast.Expr, *core.Diagnostic) {
	// nil true false 0.0 'c' "str"
	if p.match(scanner.Nil, scanner.True, scanner.False, scanner.Number, scanner.Character, scanner.String) {
		expr := &ast.Literal{
			Value: p.current,
		}

		expr.SetRangeToken(p.current, p.current)
		return expr, nil
	}

	// abc
	if p.match(scanner.Identifier) {
		expr := &ast.Identifier{
			Identifier: p.current,
		}

		expr.SetRangeToken(p.current, p.current)
		return expr, nil
	}

	// (
	if p.match(scanner.LeftParen) {
		token := p.current

		// Expression
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Right paren
		_, err = p.consume(scanner.RightParen, "Expected ')' after expression.")
		if err != nil {
			return nil, err
		}

		// Return
		group := &ast.Group{
			Token_: token,
			Expr:   expr,
		}

		group.SetRangeToken(token, p.current)
		return group, nil
	}

	// Error
	return nil, p.error(p.next, "Expected expression.")
}
