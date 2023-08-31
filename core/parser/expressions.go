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
	expr, err := p.or()
	if err != nil {
		return nil, err
	}

	if p.match(scanner.Equal, scanner.PlusEqual, scanner.MinusEqual, scanner.StarEqual, scanner.SlashEqual, scanner.PercentageEqual) {
		op := p.current

		value, err := p.assignment()
		if err != nil {
			return nil, err
		}

		return &ast.Assignment{
			Assignee: expr,
			Op:       op,
			Value:    value,
		}, nil
	}

	return expr, nil
}

func (p *parser) or() (ast.Expr, *core.Diagnostic) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}

	for p.match(scanner.Or) {
		op := p.current

		right, err := p.and()
		if err != nil {
			return nil, err
		}

		expr = &ast.Logical{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *parser) and() (ast.Expr, *core.Diagnostic) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.match(scanner.And) {
		op := p.current

		right, err := p.equality()
		if err != nil {
			return nil, err
		}

		expr = &ast.Logical{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	return expr, nil
}

func (p *parser) equality() (ast.Expr, *core.Diagnostic) {
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

func (p *parser) comparison() (ast.Expr, *core.Diagnostic) {
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

func (p *parser) term() (ast.Expr, *core.Diagnostic) {
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

func (p *parser) factor() (ast.Expr, *core.Diagnostic) {
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

func (p *parser) unary() (ast.Expr, *core.Diagnostic) {
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

func (p *parser) call() (ast.Expr, *core.Diagnostic) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(scanner.LeftParen) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(scanner.LeftBracket) {
			expr, err = p.finishIndex(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(scanner.Dot) {
			expr, err = p.finishMember(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(scanner.As) {
			expr, err = p.finishCast(expr)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *parser) finishCall(callee ast.Expr) (ast.Expr, *core.Diagnostic) {
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

func (p *parser) finishIndex(value ast.Expr) (ast.Expr, *core.Diagnostic) {
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

func (p *parser) finishMember(value ast.Expr) (ast.Expr, *core.Diagnostic) {
	name, err := p.consume(scanner.Identifier, "Expected member name.")
	if err != nil {
		return nil, err
	}

	return &ast.Member{
		Value: value,
		Name:  name,
	}, nil
}

func (p *parser) finishCast(value ast.Expr) (ast.Expr, *core.Diagnostic) {
	token := p.current

	type_, err := p.parseType()
	if err != nil {
		return nil, err
	}

	cast := &ast.Cast{
		Token_: token,
		Expr:   value,
	}

	cast.SetType(type_)
	return cast, nil
}

func (p *parser) primary() (ast.Expr, *core.Diagnostic) {
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
