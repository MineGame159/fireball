package minegame159.fireball.context;

import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;
import minegame159.fireball.types.Type;

import java.util.List;

public class Destructor extends Method {
    public Destructor(Struct owner, Token name, Type returnType, List<Param> params, Stmt body) {
        super(owner, name, returnType, params, body);
    }

    @Override
    public String getOutputName() {
        return "_" + owner.name + "___destructor";
    }
}
