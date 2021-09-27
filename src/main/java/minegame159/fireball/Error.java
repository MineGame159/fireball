package minegame159.fireball;

import minegame159.fireball.parser.Token;

public class Error extends RuntimeException {
    public final Token token;

    public Error(Token token, String message) {
        super(message);
        this.token = token;
    }

    @Override
    public String toString() {
        return String.format("Error [%d:%d]: %s", token.line(), token.character(), getMessage());
    }
}
