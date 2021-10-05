package minegame159.fireball.parser.prototypes;

import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;

import java.util.List;

public class ProtoFunction {
    public final Token name;
    public final ProtoType returnType;
    public final List<ProtoParameter> params;
    public final Stmt body;

    public ProtoFunction(Token name, ProtoType returnType, List<ProtoParameter> params, Stmt body) {
        this.name = name;
        this.returnType = returnType;
        this.params = params;
        this.body = body;
    }

    public void accept(Stmt.Visitor visitor) {
        visitor.visitFunctionStart(this);
        body.accept(visitor);
        visitor.visitFunctionEnd(this);
    }
}
