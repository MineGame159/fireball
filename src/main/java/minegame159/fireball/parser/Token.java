package minegame159.fireball.parser;

public record Token(TokenType type, String lexeme, int line, int character) {
    @Override
    public String toString() {
        return lexeme;
    }
}
