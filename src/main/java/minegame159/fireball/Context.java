package minegame159.fireball;

import java.util.Collection;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class Context {
    private final Map<String, Function> functions = new HashMap<>();

    public void declareFunction(Token name, Token returnType, List<TokenPair> params) {
        functions.put(name.lexeme(), new Function(name, returnType, params));
    }

    public Function getFunction(Token name) {
        return functions.get(name.lexeme());
    }

    public Collection<Function> getFunctions() {
        return functions.values();
    }
}
