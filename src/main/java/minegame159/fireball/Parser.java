package minegame159.fireball;

import java.io.Reader;
import java.util.ArrayList;
import java.util.List;

public class Parser {
    public final List<Stmt> stmts = new ArrayList<>();
    public final List<Error> errors = new ArrayList<>();

    private final Scanner scanner;
    private Token next, current, previous;

    public Parser(Reader reader) {
        scanner = new Scanner(reader);

        advance();
        advance();
    }

    public void parse() {
        while (peek().type() != TokenType.Eof) {
            try {
                stmts.add(functionDeclaration());
            } catch (Error error) {
                errors.add(error);
                synchronize();
            }
        }
    }

    // Declaration

    private Stmt functionDeclaration() {
        Token returnType = consume(TokenType.Identifier, "Expected return type.");
        Token name = consume(TokenType.Identifier, "Expected function name.");

        consume(TokenType.LeftParen, "Expected '(' after function name.");
        List<TokenPair> parameters = new ArrayList<>();
        if (!check(TokenType.RightParen)) {
            do {
                parameters.add(new TokenPair(
                        consume(TokenType.Identifier, "Expected parameter type."),
                        consume(TokenType.Identifier, "Expected parameter name.")
                ));
            } while (match(TokenType.Comma));
        }
        consume(TokenType.RightParen, "Expected ')' after parameters.");

        Stmt body = statement();
        return new Stmt.Function(returnType, name, parameters, body);
    }

    private Stmt declaration() {
        if (checkNext(TokenType.Identifier) && match(TokenType.Identifier, TokenType.Var)) return variableDeclaration();

        return statement();
    }

    private Stmt variableDeclaration() {
        Token type = previous();
        Token name = advance();

        Expr initializer = null;
        if (match(TokenType.Equal)) initializer = expression();

        consume(TokenType.Semicolon, "Expected ';' after variable declaration.");
        return new Stmt.Variable(type, name, initializer);
    }

    // Statements

    private Stmt statement() {
        if (match(TokenType.LeftBrace)) return blockStatement();
        if (match(TokenType.If)) return ifStatement();
        if (match(TokenType.While)) return whileStatement();
        if (match(TokenType.For)) return forStatement();
        if (match(TokenType.Return)) return returnStatement();

        return expressionStatement();
    }

    private Stmt blockStatement() {
        List<Stmt> stmts = new ArrayList<>();

        while (!check(TokenType.RightBrace) && !isAtEnd()) {
            stmts.add(declaration());
        }

        consume(TokenType.RightBrace, "Expected '}' after block.");
        return new Stmt.Block(stmts);
    }

    private Stmt ifStatement() {
        consume(TokenType.LeftParen, "Expected '(' after 'if'.");
        Expr condition = expression();
        consume(TokenType.RightParen, "Expected ')' after if condition.");

        Stmt thenBranch = statement();
        Stmt elseBranch = match(TokenType.Else) ? statement() : null;

        return new Stmt.If(condition, thenBranch, elseBranch);
    }

    private Stmt whileStatement() {
        consume(TokenType.LeftParen, "Expected '(' after 'while'.");
        Expr condition = expression();
        consume(TokenType.RightParen, "Expected ')' after condition.");

        Stmt body = statement();
        return new Stmt.While(condition, body);
    }

    private Stmt forStatement() {
        consume(TokenType.LeftParen, "Expected '(' after 'for'.");
        Stmt initializer = match(TokenType.Semicolon) ? null : declaration();

        Expr condition = check(TokenType.Semicolon) ? null : expression();
        consume(TokenType.Semicolon, "Expected ';' after loop condition.");

        Expr increment = check(TokenType.RightParen) ? null : expression();
        consume(TokenType.RightParen, "Expected ')' after for clauses.");

        Stmt body = statement();
        return new Stmt.For(initializer, condition, increment, body);
    }

    private Stmt returnStatement() {
        Expr value = check(TokenType.Semicolon) ? null : expression();
        consume(TokenType.Semicolon, "Expected ';' after return value.");

        return new Stmt.Return(value);
    }

    private Stmt expressionStatement() {
        Expr expr = expression();
        consume(TokenType.Semicolon, "Expected ';' after expression.");
        return new Stmt.Expression(expr);
    }

    // Expressions

    private Expr expression() {
        return assignment();
    }

    private Expr assignment() {
        Expr expr = or();

        if (match(TokenType.Equal)) {
            Token equals = previous();
            Expr value = assignment();

            if (expr instanceof Expr.Variable) {
                Token name = ((Expr.Variable)expr).name;
                return new Expr.Assign(name, value);
            }

            throw error(equals, "Invalid assignment target.");
        }

        return expr;
    }

    private Expr or() {
        Expr expr = and();

        while (match(TokenType.Or)) {
            Token operator = previous();
            Expr right = and();
            expr = new Expr.Logical(expr, operator, right);
        }

        return expr;
    }

    private Expr and() {
        Expr expr = equality();

        while (match(TokenType.And)) {
            Token operator = previous();
            Expr right = equality();
            expr = new Expr.Logical(expr, operator, right);
        }

        return expr;
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

        return call();
    }

    private Expr call() {
        Expr expr = primary();

        while (true) {
            if (match(TokenType.LeftParen)) expr = finishCall(expr);
            else break;
        }

        return expr;
    }

    private Expr finishCall(Expr callee) {
        List<Expr> arguments = new ArrayList<>();

        if (!check(TokenType.RightParen)) {
            do {
                arguments.add(expression());
            } while (match(TokenType.Comma));
        }

        consume(TokenType.RightParen, "Expected ')' after arguments.");
        return new Expr.Call(callee, arguments);
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
        if (match(TokenType.Identifier)) return new Expr.Variable(previous());

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

    private boolean checkNext(TokenType type) {
        if (isAtEnd()) return false;
        return peekNext().type() == type;
    }

    private Token advance() {
        if (!isAtEnd()) {
            previous = current;
            current = next;
            next = scanner.next();
        }

        return previous();
    }

    private boolean isAtEnd() {
        return peek() != null && peek().type() == TokenType.Eof;
    }

    private Token peek() {
        return current;
    }

    private Token peekNext() {
        return next;
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
