package minegame159.fireball.parser;

import java.io.IOException;
import java.io.Reader;

public class Scanner {
    private final Reader reader;
    private final StringBuilder sb = new StringBuilder();

    private char current, next;
    private int line = 1, character, firstTokenCharacter;

    private boolean inString;

    public Scanner(Reader reader) {
        this.reader = reader;

        advance(false);
        advance(false);
    }

    public Token next() {
        sb.setLength(0);

        skipWhitespace();
        if (isAtEnd()) return token(TokenType.Eof);

        firstTokenCharacter = character;
        char c = advance();

        if (c == 'c' && peek() == '{') {
            advance(false);
            return cBlock();
        }
        if (isDigit(c)) return number();
        if (isAlpha(c)) return identifier();

        return switch (c) {
            case '(' -> token(TokenType.LeftParen);
            case ')' -> token(TokenType.RightParen);
            case '{' -> token(TokenType.LeftBrace);
            case '}' -> token(TokenType.RightBrace);
            case ':' -> token(TokenType.Colon);
            case ';' -> token(TokenType.Semicolon);
            case ',' -> token(TokenType.Comma);
            case '.' -> token(TokenType.Dot);

            case '!' -> token(match('=') ? TokenType.BangEqual : TokenType.Bang);
            case '=' -> token(match('=') ? TokenType.EqualEqual : TokenType.Equal);
            case '>' -> token(match('=') ? TokenType.GreaterEqual : TokenType.Greater);
            case '<' -> token(match('=') ? TokenType.LessEqual : TokenType.Less);

            case '+' -> token(match('=') ? TokenType.PlusEqual : TokenType.Plus);
            case '-' -> token(match('=') ? TokenType.MinusEqual : TokenType.Minus);
            case '*' -> token(match('=') ? TokenType.StarEqual : TokenType.Star);
            case '/' -> token(match('=') ? TokenType.SlashEqual : TokenType.Slash);
            case '%' -> token(match('=') ? TokenType.PercentageEqual : TokenType.Percentage);

            case '&' -> token(match('&') ? TokenType.And : TokenType.Ampersand);
            case '|' -> token(match('|') ? TokenType.Or : TokenType.Pipe);

            case '"' -> string();

            default -> error("Unexpected character");
        };
    }

    private Token cBlock() {
        boolean countTrimStart = true;
        int trimStart = 0;
        
        int braceCount = 0;

        while (!isAtEnd()) {
            if (countTrimStart) {
                if (isMeaninglessWhitespace(peek())) trimStart++;
                else countTrimStart = false;
            }

            if (peek() == '\n') line++;
            else if (peek() == '{') braceCount++;
            else if (peek() == '}') {
                if (braceCount > 0) braceCount--;
                else break;
            }

            advance();
        }

        if (isAtEnd()) return error("Unterminated C block.");

        advance(false);
        sb.delete(0, trimStart + 1);

        int trimEnd = 0;
        for (int i = sb.length() - 1; i >= 0; i--) {
            if (isMeaninglessWhitespace(sb.charAt(i))) trimEnd++;
            else break;
        }
        sb.delete(sb.length() - trimEnd, sb.length() + trimEnd);

        return new Token(TokenType.CBlock, sb.toString(), line, firstTokenCharacter);
    }

    private Token number() {
        while (isDigit(peek())) advance();

        if (peek() == '.' && isDigit(peekNext())) {
            advance();

            while (isDigit(peek())) advance();

            return token(TokenType.Float);
        }

        return token(TokenType.Int);
    }

    private Token string() {
        inString = true;
        while (peek() != '"' && !isAtEnd()) {
            if (peek() == '\n') line++;
            advance();
        }

        if (isAtEnd()) return error("Unterminated string.");

        advance();
        inString = false;

        return token(TokenType.String);
    }

    private Token identifier() {
        while (isAlpha(peek()) || isDigit(peek())) advance();
        return token(identifierType());
    }

    private TokenType identifierType() {
        return switch (sb.charAt(0)) {
            case 'n' -> checkKeyword(1, "ull", TokenType.Null);
            case 't' -> checkKeyword(1, "rue", TokenType.True);
            case 'f' -> {
                if (sb.length() > 1) yield switch (sb.charAt(1)) {
                    case 'a' -> checkKeyword(2, "lse", TokenType.False);
                    case 'o' -> checkKeyword(2, "r", TokenType.For);
                    default -> TokenType.Identifier;
                };
                yield TokenType.Identifier;
            }
            case 'i' -> checkKeyword(1, "f", TokenType.If);
            case 'w' -> checkKeyword(1, "hile", TokenType.While);
            case 'v' -> checkKeyword(1, "ar", TokenType.Var);
            case 'e' -> checkKeyword(1, "lse", TokenType.Else);
            case 'r' -> checkKeyword(1, "eturn", TokenType.Return);
            case 's' -> checkKeyword(1, "truct", TokenType.Struct);

            default -> TokenType.Identifier;
        };
    }

    private TokenType checkKeyword(int start, String rest, TokenType type) {
        return sb.substring(start).equals(rest) ? type : TokenType.Identifier;
    }

    // Utils

    private boolean match(char expected) {
        if (isAtEnd()) return false;
        if (peek() != expected) return false;

        advance();
        return true;
    }

    private char advance() {
        return advance(true);
    }

    private char advance(boolean append) {
        if (append) sb.append(current);

        char prev = current;
        current = next;

        try {
            int i = reader.read();
            next = i == -1 ? '\0' : (char) i;
        } catch (IOException e) {
            e.printStackTrace();
        }

        if (prev != '\0' && (inString || prev != '\n')) character++;
        return prev;
    }

    private char peek() {
        return current;
    }

    private char peekNext() {
        return next;
    }

    private boolean isAtEnd() {
        return peek() == '\0';
    }

    private boolean isMeaninglessWhitespace(char c) {
        return c == ' ' || c == '\r' || c == '\t';
    }

    private void skipWhitespace() {
        while (true) {
            if (isMeaninglessWhitespace(peek())) {
                advance(false);
                continue;
            }

            switch (peek()) {
                case '\n' -> {
                    incrementLine();
                    advance(false);
                }
                case '/' -> {
                    if (peekNext() == '/') {
                        while (peek() != '\n' && !isAtEnd()) advance(false);
                    }
                    else return;
                }
                default -> {
                    return;
                }
            }
        }
    }

    private boolean isDigit(char c) {
        return c >= '0' && c <= '9';
    }

    private boolean isAlpha(char c) {
        return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_';
    }

    private void incrementLine() {
        line++;
        character = 0;
    }

    private Token token(TokenType type) {
        return new Token(type, sb.toString(), line, firstTokenCharacter);
    }

    private Token error(String msg) {
        return new Token(TokenType.Error, msg, line, firstTokenCharacter);
    }
}
