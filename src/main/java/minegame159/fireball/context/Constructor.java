package minegame159.fireball.context;

import minegame159.fireball.parser.Expr;
import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;
import minegame159.fireball.types.Type;

import java.util.List;

public class Constructor extends Method {
    private final int index;

    public Constructor(Struct owner, Token name, Type returnType, List<Param> params, Stmt body, int index) {
        super(owner, name, returnType, params, body);

        this.index = index;
    }

    public boolean canCall(boolean returnsPointer, List<Expr> arguments) {
        if (arguments.size() != params.size()) return false;
        if (returnsPointer != returnType.isPointer()) return false;

        for (int i = 0; i < arguments.size(); i++) {
            if (!arguments.get(i).getType().equals(params.get(i).type())) return false;
        }

        return true;
    }

    @Override
    public String getOutputName() {
        return "_" + owner.name() + "__" + index;
    }
}
