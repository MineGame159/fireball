package minegame159.fireball;

import java.util.List;

public class FunctionPrototype {
    public final Token name;
    public final Token returnType;
    public final List<TokenPair> params;

    public FunctionPrototype(Token name, Token returnType, List<TokenPair> params) {
        this.name = name;
        this.returnType = returnType;
        this.params = params;
    }
}
