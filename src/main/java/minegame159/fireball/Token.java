package minegame159.fireball;

public record Token(TokenType type, String lexeme, int line) {
    @Override
    public String toString() {
        return lexeme;
    }
}
