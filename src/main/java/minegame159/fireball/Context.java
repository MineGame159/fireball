package minegame159.fireball;

import java.util.HashMap;
import java.util.Map;

public class Context {
    private final Map<String, Function> functions = new HashMap<>();

    public void declareFunction(Token name) {
        functions.put(name.lexeme(), new Function());
    }

    public Function getFunction(Token name) {
        return functions.get(name.lexeme());
    }
}
