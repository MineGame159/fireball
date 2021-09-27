package minegame159.fireball.context;

import minegame159.fireball.Error;
import minegame159.fireball.Errors;
import minegame159.fireball.TokenPair;
import minegame159.fireball.parser.Parser;
import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;
import minegame159.fireball.types.PrimitiveTypes;
import minegame159.fireball.types.StructType;
import minegame159.fireball.types.Type;

import java.util.*;

public class Context {
    private final Map<String, Type> types = new HashMap<>();

    private final Map<String, Struct> structs = new HashMap<>();
    private final Map<String, Function> functions = new HashMap<>();

    public Context() {
        for (PrimitiveTypes type : PrimitiveTypes.values()) types.put(type.type.name, type.type);
    }

    // General

    // Types

    public Type getType(Token name) {
        return types.get(name.lexeme());
    }

    // Structs

    public Collection<Struct> getStructs() {
        return structs.values();
    }

    // Functions

    public Function getFunction(Token name) {
        return functions.get(name.lexeme());
    }

    public Collection<Function> getFunctions() {
        return functions.values();
    }

    // Apply

    public List<Error> apply(Parser.Result result) {
        List<Error> errors = new ArrayList<>();

        // Structs
        for (Stmt.Struct stmt : result.structs) {
            Struct struct = new Struct(stmt.name);

            types.put(stmt.name.lexeme(), new StructType(stmt.name.lexeme(), struct));
            structs.put(stmt.name.lexeme(), struct);
        }

        // Functions
        for (Stmt.Function stmt : result.functions) {
            // Return type
            Type returnType = getType(stmt.returnType);
            if (returnType == null) {
                errors.add(Errors.unknownType(stmt.returnType, stmt.returnType));
                continue;
            }

            // Parameters
            List<Function.Param> params = new ArrayList<>();

            for (TokenPair param : stmt.params) {
                Type type = getType(param.first());
                if (type == null) {
                    errors.add(Errors.unknownType(param.first(), stmt.returnType));
                    continue;
                }

                params.add(new Function.Param(type, param.second()));
            }

            // Create
            functions.put(stmt.name.lexeme(), new Function(stmt.name, returnType, params));
        }

        return errors;
    }
}
