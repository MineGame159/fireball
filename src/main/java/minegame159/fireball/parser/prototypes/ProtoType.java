package minegame159.fireball.parser.prototypes;

import minegame159.fireball.parser.Token;

public record ProtoType(Token name, boolean pointer) {
    public ProtoType(Token name) {
        this(name, false);
    }

    @Override
    public String toString() {
        return pointer ? name().lexeme() + '*' : name().lexeme();
    }
}
