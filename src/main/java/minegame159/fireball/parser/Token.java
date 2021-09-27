package minegame159.fireball.parser;

import java.util.Objects;

public record Token(TokenType type, String lexeme, int line, int character) {
    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        Token token = (Token) o;

        return Objects.equals(lexeme, token.lexeme);
    }

    @Override
    public int hashCode() {
        return lexeme != null ? lexeme.hashCode() : 0;
    }

    @Override
    public String toString() {
        return lexeme;
    }
}
