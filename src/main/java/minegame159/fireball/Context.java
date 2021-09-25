package minegame159.fireball;

import minegame159.fireball.types.PrimitiveTypes;
import minegame159.fireball.types.Type;

import java.util.*;

public class Context {
    private final Map<String, Type> types = new HashMap<>();
    private final Map<String, Function> functions = new HashMap<>();

    private final List<FunctionPrototype> functionPrototypes = new ArrayList<>();

    public Context() {
        for (PrimitiveTypes type : PrimitiveTypes.values()) types.put(type.type.name, type.type);
    }

    // General

    // Types

    public Type getType(Token name) {
        return types.get(name.lexeme());
    }

    // Functions

    public void declareFunction(Token name, Token returnType, List<TokenPair> params) {
        functionPrototypes.add(new FunctionPrototype(name, returnType, params));
    }

    public Function getFunction(Token name) {
        return functions.get(name.lexeme());
    }

    public Collection<Function> getFunctions() {
        return functions.values();
    }

    // Type resolver

    public List<Error> resolveTypes() {
        List<Error> errors = new ArrayList<>();

        // Functions
        for (FunctionPrototype prototype : functionPrototypes) {
            // Return type
            Type returnType = getType(prototype.returnType);
            if (returnType == null) {
                errors.add(new Error(prototype.returnType, "Unknown type '" + prototype.returnType + "'."));
                continue;
            }

            // Parameters
            List<Function.Param> params = new ArrayList<>();

            for (TokenPair param : prototype.params) {
                Type type = getType(param.first());
                if (type == null) {
                    errors.add(new Error(param.first(), "Unknown type '" + prototype.returnType + "'."));
                    continue;
                }

                params.add(new Function.Param(type, param.second()));
            }

            // Create
            functions.put(prototype.name.lexeme(), new Function(prototype.name, returnType, params));
        }

        functionPrototypes.clear();
        return errors;
    }
}
