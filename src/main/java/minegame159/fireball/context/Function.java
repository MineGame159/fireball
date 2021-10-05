package minegame159.fireball.context;

import minegame159.fireball.parser.Token;
import minegame159.fireball.types.Type;

import java.util.List;

public class Function {
    public record Param(Type type, Token name) {}

    public final Token name;
    public final Type returnType;
    public final List<Param> params;

    public Function(Token name, Type returnType, List<Param> params) {
        this.name = name;
        this.returnType = returnType;
        this.params = params;
    }

    public String getOutputName() {
        return name.lexeme();
    }
}
