package minegame159.fireball.parser;

import minegame159.fireball.Error;
import minegame159.fireball.parser.prototypes.*;

import java.io.Reader;
import java.util.ArrayList;
import java.util.List;

public class Parser {
    public static class Result {
        public final List<ProtoStruct> structs = new ArrayList<>();
        public final List<ProtoFunction> functions = new ArrayList<>();

        public Error error;

        private Result() {}

        public void accept(Stmt.Visitor visitor) {
            for (ProtoStruct struct : structs) struct.accept(visitor);
            for (ProtoFunction function : functions) function.accept(visitor);
        }
    }

    private final Result result = new Result();

    private final Scanner scanner;
    private Token next, current, previous;

    private Parser(Reader reader) {
        this.scanner = new Scanner(reader);

        advance();
        advance();
    }

    public static Result parse(Reader reader) {
        Parser parser = new Parser(reader);
        parser.parse();
        return parser.result;
    }

    private void parse() {
        try {
            while (peek().type() != TokenType.Eof) {
                topLevelDeclaration();
            }
        } catch (Error error) {
            result.error = error;
        }
    }

    // Declaration

    private void topLevelDeclaration() {
        if (match(TokenType.Struct)) structDeclaration();
        else functionDeclaration();
    }

    private void structDeclaration() {
        Token name = consume(TokenType.Identifier, "Expected struct name.");

        consume(TokenType.LeftBrace, "Expected '{' after struct name.");
        List<ProtoParameter> fields = new ArrayList<>();
        List<ProtoMethod> constructors = new ArrayList<>();
        ProtoMethod destructor = null;
        List<ProtoMethod> methods = new ArrayList<>();

        while (!check(TokenType.RightBrace)) {
            if (match(TokenType.Tilde)) {
                // Destructor
                Token identifier = consume(TokenType.Identifier, "Expected struct name.");

                if (!identifier.lexeme().equals(name.lexeme())) throw error(identifier, "Destructor name must be the same as the struct name.");
                if (destructor != null) throw error(identifier, "Struct '" + identifier + "' already contains a destructor.");

                consume(TokenType.LeftParen, "Expected '('  after destructor.");
                consume(TokenType.RightParen, "Expected ')'  after '('.");

                Stmt body = statement();

                destructor = new ProtoMethod(identifier, new ProtoType(new Token(TokenType.Identifier, "void", 0, 0)), new ArrayList<>(0), body);
            }
            else {
                Token identifier = consume(TokenType.Identifier, "Expected identifier.");

                if (identifier.lexeme().equals(name.lexeme()) && match(TokenType.LeftParen)) {
                    // Constructor
                    List<ProtoParameter> params = new ArrayList<>();

                    boolean first = true;
                    while (!check(TokenType.RightParen) && (first || match(TokenType.Comma))) {
                        params.add(consumeParameter("parameter"));
                        first = false;
                    }

                    consume(TokenType.RightParen, "Expected ')' after constructor parameters.");
                    Stmt body = statement();

                    Token thisToken = new Token(TokenType.Identifier, "this", 0, 0);
                    Token initToken = new Token(TokenType.Identifier, "_init_" + methods.size(), 0, 0);

                    {
                        // Hidden initialize function
                        List<ProtoParameter> params2 = new ArrayList<>(params);
                        params2.add(0, new ProtoParameter(new ProtoType(identifier, true), thisToken));
                        methods.add(new ProtoMethod(initToken, new ProtoType(new Token(TokenType.Identifier, "void", 0, 0)), params2, body));
                    }

                    {
                        // Malloc constructor
                        Stmt.Block bod = new Stmt.Block(new ArrayList<>());

                        List<Expr> args = new ArrayList<>(params.size());
                        for (ProtoParameter param : params) args.add(new Expr.Variable(param.name()));

                        bod.statements.add(new Stmt.Variable(new ProtoType(identifier, true), thisToken, new Expr.CBlock("malloc(sizeof(" + identifier + "))")));
                        bod.statements.add(new Stmt.Expression(new Expr.Call(thisToken, new Expr.Get(new Expr.Variable(thisToken), initToken), args)));
                        bod.statements.add(new Stmt.Return(thisToken, new Expr.Variable(thisToken)));

                        constructors.add(new ProtoMethod(identifier, new ProtoType(identifier, true), params, bod));
                    }
                    {
                        // Stack constructor
                        Stmt.Block bod = new Stmt.Block(new ArrayList<>());

                        List<Expr> args = new ArrayList<>(params.size());
                        for (ProtoParameter param : params) args.add(new Expr.Variable(param.name()));

                        bod.statements.add(new Stmt.Variable(new ProtoType(identifier), thisToken, null));
                        bod.statements.add(new Stmt.Expression(new Expr.Call(thisToken, new Expr.Get(new Expr.Variable(thisToken), initToken), args)));
                        bod.statements.add(new Stmt.Return(thisToken, new Expr.Variable(thisToken)));

                        constructors.add(new ProtoMethod(identifier, new ProtoType(identifier), params, bod));
                    }
                } else {
                    // Field or method
                    boolean pointer = match(TokenType.Star);
                    ProtoType type = new ProtoType(identifier, pointer);

                    Token name2 = consume(TokenType.Identifier, "Expected field or method name.");

                    if (match(TokenType.Semicolon)) {
                        // Field
                        fields.add(new ProtoParameter(type, name2));
                    } else {
                        // Method
                        consume(TokenType.LeftParen, "Expected '(' after method name");
                        List<ProtoParameter> params = new ArrayList<>();

                        params.add(new ProtoParameter(new ProtoType(name, true), new Token(TokenType.Identifier, "this", 0, 0)));

                        boolean first = true;
                        while (!check(TokenType.RightParen) && (first || match(TokenType.Comma))) {
                            params.add(consumeParameter("parameter"));
                            first = false;
                        }

                        consume(TokenType.RightParen, "Expected ')' after method parameters.");
                        Stmt body = statement();

                        methods.add(new ProtoMethod(name2, type, params, body));
                    }
                }
            }
        }

        consume(TokenType.RightBrace, "Expected '}' after struct body.");

        result.structs.add(new ProtoStruct(name, fields, constructors, destructor, methods));
    }

    private void functionDeclaration() {
        ProtoType returnType = consumeType("return");
        Token name = consume(TokenType.Identifier, "Expected function name");

        consume(TokenType.LeftParen, "Expected '(' after function name");
        List<ProtoParameter> params = new ArrayList<>();

        boolean first = true;
        while (!check(TokenType.RightParen) && (first || match(TokenType.Comma))) {
            params.add(consumeParameter("parameter"));
            first = false;
        }

        consume(TokenType.RightParen, "Expected ')' after function parameters.");
        Stmt body = statement();

        result.functions.add(new ProtoFunction(name, returnType, params, body));
    }

    private Stmt declaration() {
        if (check(TokenType.Var) || (check(TokenType.Identifier) && (checkNext(TokenType.Star) || checkNext(TokenType.Identifier)))) return variableDeclaration();

        return statement();
    }

    private Stmt variableDeclaration() {
        Token type = advance();
        boolean pointer = match(TokenType.Star);

        Token name = advance();

        Expr initializer = match(TokenType.Equal) ? expression() : null;

        consume(TokenType.Semicolon, "Expected ';' after variable initializer.");
        return new Stmt.Variable(new ProtoType(type, pointer), name, initializer);
    }

    // Statements

    private Stmt statement() {
        if (match(TokenType.LeftBrace)) return blockStatement();
        if (match(TokenType.If)) return ifStatement();
        if (match(TokenType.While)) return whileStatement();
        if (match(TokenType.For)) return forStatement();
        if (match(TokenType.Return)) return returnStatement();
        if (match(TokenType.CBlock)) return cBlockStatement();
        if (match(TokenType.Delete)) return deleteStatement();

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
        Token token = previous();

        Expr value = check(TokenType.Semicolon) ? null : expression();
        consume(TokenType.Semicolon, "Expected ';' after return value.");

        return new Stmt.Return(token, value);
    }

    private Stmt cBlockStatement() {
        return new Stmt.CBlock(previous().lexeme().replace("\n", "\\n"));
    }

    private Stmt deleteStatement() {
        Token token = previous();
        Expr expr = expression();

        consume(TokenType.Semicolon, "Expected ';' after expression.");
        return new Stmt.Delete(token, expr);
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

        if (match(TokenType.Equal, TokenType.PlusEqual, TokenType.MinusEqual, TokenType.StarEqual, TokenType.SlashEqual, TokenType.PercentageEqual)) {
            Token operator = previous();
            Expr value = assignment();

            if (expr instanceof Expr.Variable) {
                Token name = ((Expr.Variable)expr).name;
                return new Expr.Assign(name, operator, value);
            }
            else if (expr instanceof Expr.Get get) {
                return new Expr.Set(get.object, get.name, operator, value);
            }

            throw error(operator, "Invalid assignment target.");
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
        Expr expr = newE();

        while (match(TokenType.Slash, TokenType.Star, TokenType.Percentage)) {
            Token operator = previous();
            Expr right = newE();
            expr = new Expr.Binary(expr, operator, right);
        }

        return expr;
    }

    private Expr newE() {
        if (match(TokenType.New)) {
            Token name = consume(TokenType.Identifier, "Expected type.");
            List<Expr> arguments = new ArrayList<>();

            consume(TokenType.LeftParen, "Expected '(' before arguments.");
            if (!check(TokenType.RightParen)) {
                do {
                    arguments.add(expression());
                } while (match(TokenType.Comma));
            }

            consume(TokenType.RightParen, "Expected ')' after arguments.");
            return new Expr.New(name, arguments);
        }

        return cast();
    }

    private Expr cast() {
        if (match(TokenType.LeftParen)) {
            ProtoType type = consumeType("target");
            consume(TokenType.RightParen, "Expected ')' after cast target.");
            return new Expr.Cast(cast(), type);
        }

        return unary();
    }

    private Expr unary() {
        if (match(TokenType.Bang, TokenType.Minus, TokenType.Ampersand, TokenType.PlusPlus, TokenType.MinusMinus)) {
            Token operator = previous();
            Expr right = unary();
            return new Expr.Unary(operator, right);
        }

        return unaryPost();
    }

    private Expr unaryPost() {
        Expr expr = call();

        while (match(TokenType.PlusPlus, TokenType.MinusMinus)) {
            Token operator = previous();
            expr = new Expr.UnaryPost(operator, expr);
        }

        return expr;
    }

    private Expr call() {
        Expr expr = primary();

        while (true) {
            if (match(TokenType.LeftParen)) expr = finishCall(expr);
            else if (match(TokenType.Dot)) {
                Token name = consume(TokenType.Identifier, "Expected field name after '.'.");
                expr = new Expr.Get(expr, name);
            }
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

        Token token = consume(TokenType.RightParen, "Expected ')' after arguments.");
        return new Expr.Call(token, callee, arguments);
    }

    private Expr primary() {
        if (match(TokenType.Null)) return new Expr.Null();
        if (match(TokenType.True)) return new Expr.Bool(true);
        if (match(TokenType.False)) return new Expr.Bool(false);
        if (match(TokenType.Int)) return new Expr.Int(4, Integer.parseInt(previous().lexeme()));
        if (match(TokenType.Float)) {
            char c = previous().lexeme().charAt(previous().lexeme().length() - 1);
            return new Expr.Float(c != 'f' && c != 'F', Double.parseDouble(previous().lexeme()));
        }
        if (match(TokenType.String)) return new Expr.String(previous.lexeme().substring(1, previous.lexeme().length() - 1));

        if (match(TokenType.LeftParen)) {
            Expr expr = expression();
            consume(TokenType.RightParen, "Expected ')' after expression.");
            return new Expr.Grouping(expr);
        }
        if (match(TokenType.Identifier)) return new Expr.Variable(previous());

        throw error(peek(), "Expected expression.");
    }

    // Prototypes

    private ProtoParameter consumeParameter(String constructName) {
        ProtoType type = consumeType(constructName);
        Token name = consume(TokenType.Identifier, "Expected " + constructName + " name.");

        return new ProtoParameter(type, name);
    }

    private ProtoType consumeType(String constructName) {
        Token type = consume(TokenType.Identifier, "Expected " + constructName + " type.");
        boolean pointer = match(TokenType.Star);

        return new ProtoType(type, pointer);
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

    private Error error(Token token, String message) {
        return new Error(token, message);
    }
}
