package minegame159.fireball.parser.prototypes;

import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;

import java.util.List;

public record ProtoFunction(Token name, ProtoType returnType, List<ProtoParameter> params, Stmt body) {
    public void accept(Stmt.Visitor visitor) {
        visitor.visitFunctionStart(this);
        body.accept(visitor);
        visitor.visitFunctionEnd(this);
    }
}
