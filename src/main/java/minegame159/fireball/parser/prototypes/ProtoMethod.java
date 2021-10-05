package minegame159.fireball.parser.prototypes;

import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;

import java.util.List;

public class ProtoMethod extends ProtoFunction {
    public ProtoStruct owner;

    public ProtoMethod(Token name, ProtoType returnType, List<ProtoParameter> params, Stmt body) {
        super(name, returnType, params, body);
    }
}
