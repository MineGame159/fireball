package minegame159.fireball;

import java.io.Reader;
import java.util.ArrayList;
import java.util.List;

public class Parser {
    public final List<Stmt> stmts = new ArrayList<>();
    public final List<Error> errors = new ArrayList<>();

    private final Scanner scanner;
    private Token current, previous;

    public Parser(Reader reader) {
        scanner = new Scanner(reader);

        advance();
    }

    public void parse() {
        while (peek().type() != TokenType.Eof) {
            try {
                stmts.add(statement());
            } catch (Error error) {
                errors.add(error);
                synchronize();
            }
        }
    }

    // Statements

    private Stmt statement() {
        return expressionStatement();
    }

    private Stmt expressionStatement() {
        Expr expr = expression();
        consume(TokenType.Semicolon, "Expected ';' after expression.");
        return new Stmt.Expression(expr);
    }

    // Expressions

    private Expr expression() {
        return equality();
    }

    private Expr equality() {
        Expr expr = comparison();

        while (match(TokenType.BangEqual, TokenType.EqualEqual)) {
            Token operator = previous();
            Expr right = comparison();
            expr = new Expr.Binary(expr, operator, right);
        }

        return expr;
    }

    private Expr comparison() {
        Expr expr = term();

        while (match(TokenType.Greater, TokenType.GreaterEqual, TokenType.Less, TokenType.LessEqual)) {
            Token operator = previous();
            Expr right = term();
            expr = new Expr.Binary(expr, operator, right);
        }

        return expr;
    }

    private Expr term() {
        Expr expr = factor();

        while (match(TokenType.Minus, TokenType.Plus)) {
            Token operator = previous();
            Expr right = factor();
            expr = new Expr.Binary(expr, operator, right);
        }

        return expr;
    }

    private Expr factor() {
        Expr expr = unary();

        while (match(TokenType.Slash, TokenType.Star, TokenType.Percentage)) {
            Token operator = previous();
            Expr right = unary();
            expr = new Expr.Binary(expr, operator, right);
        }

        return expr;
    }

    private Expr unary() {
        if (match(TokenType.Bang, TokenType.Minus)) {
            Token operator = previous();
            Expr right = unary();
            return new Expr.Unary(operator, right);
        }

        return primary();
    }

    private Expr primary() {
        if (match(TokenType.Null)) return new Expr.Null();
        if (match(TokenType.True)) return new Expr.Bool(true);
        if (match(TokenType.False)) return new Expr.Bool(false);
        if (match(TokenType.Int)) return new Expr.Int(4, Integer.parseInt(previous().lexeme()));
        if (match(TokenType.Float)) return new Expr.Float(true, Double.parseDouble(previous().lexeme()));
        if (match(TokenType.String)) return new Expr.String(previous.lexeme().substring(1, previous.lexeme().length() - 1));

        if (match(TokenType.LeftParen)) {
            Expr expr = expression();
            consume(TokenType.RightParen, "Expected ')' after expression.");
            return new Expr.Grouping(expr);
        }

        throw error(peek(), "Expected expression.");
    }

    // Utils

    private boolean match(TokenType... types) {
        for (TokenType type : types) {
            if (check(type)) {
                advance();
                return true;
            }
        }

        return false;
    }

    private Token consume(TokenType type, String message) {
        if (check(type)) return advance();

        throw error(peek(), message);
    }

    private boolean check(TokenType type) {
        if (isAtEnd()) return false;
        return peek().type() == type;
    }

    private Token advance() {
        if (!isAtEnd()) {
            previous = current;
            current = scanner.next();
        }

        return previous();
    }

    private boolean isAtEnd() {
        return peek() != null && peek().type() == TokenType.Eof;
    }

    private Token peek() {
        return current;
    }

    private Token previous() {
        return previous;
    }

    private void synchronize() {
        advance();

        while (!isAtEnd()) {
            if (previous().type() == TokenType.Semicolon) return;

            switch (peek().type()) {
                case If, While, For -> {
                    return;
                }
            }

            advance();
        }
    }

    private Error error(Token token, String message) {
        return new Error(token, message);
    }
}
