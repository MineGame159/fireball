package minegame159.fireball;

public class Error extends RuntimeException {
    public final Token token;

    public Error(Token token, String message) {
        super(message);
        this.token = token;
    }
}
