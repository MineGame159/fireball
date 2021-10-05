package minegame159.fireball.context;

import minegame159.fireball.parser.Token;
import minegame159.fireball.types.Type;

import java.util.List;

public class Method extends Function {
    public final Struct owner;

    public Method(Struct owner, Token name, Type returnType, List<Param> params) {
        super(name, returnType, params);

        this.owner = owner;
    }

    @Override
    public String getOutputName() {
        return "_" + owner.name() + "__" + name;
    }
}
