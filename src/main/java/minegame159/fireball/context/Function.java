package minegame159.fireball.context;

import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;
import minegame159.fireball.types.Type;

import java.util.List;

public class Function {
    public record Param(Type type, Token name) {}

    public final Token name;
    public final Type returnType;
    public final List<Param> params;
    public final Stmt body;

    public Function(Token name, Type returnType, List<Param> params, Stmt body) {
        this.name = name;
        this.returnType = returnType;
        this.params = params;
        this.body = body;
    }

    public void accept(Stmt.Visitor visitor) {
        body.accept(visitor);
    }

    public String getOutputName() {
        return name.lexeme();
    }
}
