package parser

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/scanner"
)

func (p *parser) expression() (ast.Expr, *core.Error) {
	return p.equality()
}

func (p *parser) equality() (ast.Expr, *core.Error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(scanner.EqualEqual, scanner.BangEqual) {
		op := p.current
		right, err := p.comparison()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *parser) comparison() (ast.Expr, *core.Error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(scanner.Less, scanner.LessEqual, scanner.Greater, scanner.GreaterEqual) {
		op := p.current
		right, err := p.term()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *parser) term() (ast.Expr, *core.Error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(scanner.Plus, scanner.Minus) {
		op := p.current
		right, err := p.factor()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *parser) factor() (ast.Expr, *core.Error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(scanner.Star, scanner.Slash, scanner.Percentage) {
		op := p.current
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		expr = &ast.Binary{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *parser) unary() (ast.Expr, *core.Error) {
	if p.match(scanner.Bang, scanner.Minus, scanner.Ampersand) {
		op := p.current
		right, err := p.unary()

		if err != nil {
			return nil, err
		}

		return &ast.Unary{
			Op:    op,
			Right: right,
		}, nil
	}

	return p.call()
}

func (p *parser) call() (ast.Expr, *core.Error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(scanner.Equal, scanner.PlusEqual, scanner.MinusEqual, scanner.StarEqual, scanner.SlashEqual, scanner.PercentageEqual) {
			expr, err = p.finishAssignment(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(scanner.As) {
			token := p.current

			type_, err := p.parseType()
			if err != nil {
				return nil, err
			}

			cast := &ast.Cast{
				Token_: token,
				Expr:   expr,
			}

			cast.SetType(type_)
			expr = cast
		} else if p.match(scanner.LeftParen) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(scanner.LeftBracket) {
			expr, err = p.finishIndex(expr)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *parser) finishAssignment(assignee ast.Expr) (ast.Expr, *core.Error) {
	op := p.current

	value, err := p.expression()
	if err != nil {
		return nil, err
	}

	return &ast.Assignment{
		Assignee: assignee,
		Op:       op,
		Value:    value,
	}, nil
}

func (p *parser) finishCall(callee ast.Expr) (ast.Expr, *core.Error) {
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

	return &ast.Call{
		Token_: p.current,
		Callee: callee,
		Args:   args,
	}, nil
}

func (p *parser) finishIndex(value ast.Expr) (ast.Expr, *core.Error) {
	token := p.current

	index, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(scanner.RightBracket, "Expected ']' after index expression."); err != nil {
		return nil, err
	}

	return &ast.Index{
		Token_: token,
		Value:  value,
		Index:  index,
	}, nil
}

func (p *parser) primary() (ast.Expr, *core.Error) {
	if p.match(scanner.Nil, scanner.True, scanner.False, scanner.Number, scanner.Character, scanner.String) {
		return &ast.Literal{
			Value: p.current,
		}, nil
	}

	if p.match(scanner.Identifier) {
		return &ast.Identifier{
			Identifier: p.current,
		}, nil
	}

	if p.match(scanner.LeftParen) {
		token := p.current
		expr, err := p.expression()

		if err != nil {
			return nil, err
		}

		_, err = p.consume(scanner.RightParen, "Expected ')' after expression.")

		if err != nil {
			return nil, err
		}

		return &ast.Group{
			Token_: token,
			Expr:   expr,
		}, nil
	}

	return nil, p.error(p.next, "Expected expression.")
}
