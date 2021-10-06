package minegame159.fireball.parser.prototypes;

import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;

import java.util.List;

public record ProtoStruct(Token name, List<ProtoParameter> fields, List<ProtoMethod> constructors, List<ProtoMethod> methods) {
    public ProtoStruct {
        for (ProtoMethod method : methods) method.owner = this;
        for (ProtoMethod constructor : constructors) constructor.owner = this;
    }

    public void accept(Stmt.Visitor visitor) {
        for (ProtoFunction method : methods) method.accept(visitor);
        for (ProtoMethod constructor : constructors) constructor.accept(visitor);
    }
}
