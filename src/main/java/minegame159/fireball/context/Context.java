package minegame159.fireball.context;

import minegame159.fireball.Error;
import minegame159.fireball.Errors;
import minegame159.fireball.parser.Parser;
import minegame159.fireball.parser.Token;
import minegame159.fireball.parser.prototypes.*;
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

    // Types

    public Type getType(ProtoType proto) {
        Type type = types.get(proto.name().lexeme());
        return proto.pointer() ? type.pointer() : type;
    }

    // Structs

    public Struct getStruct(Token name) {
        return structs.get(name.lexeme());
    }

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
        for (ProtoStruct proto : result.structs) {
            Struct struct = new Struct(proto.name(), new ArrayList<>(proto.fields().size()), new ArrayList<>(proto.methods().size()));

            types.put(proto.name().lexeme(), new StructType(proto.name().lexeme(), struct));
            structs.put(proto.name().lexeme(), struct);
        }

        for (ProtoStruct proto : result.structs) {
            Struct struct = structs.get(proto.name().lexeme());
            Set<String> fieldNames = new HashSet<>(proto.fields().size());

            for (ProtoParameter field : proto.fields()) {
                Type fieldType = getType(field.type());
                if (fieldType == null) errors.add(Errors.unknownType(field.type().name(), field.type().name()));

                if (fieldNames.contains(field.name().lexeme())) errors.add(Errors.duplicateField(field.name()));
                else fieldNames.add(field.name().lexeme());

                struct.fields().add(new Field(fieldType, field.name()));
            }

            for (ProtoMethod method : proto.methods()) {
                // Return type
                Type returnType = getType(method.returnType);
                if (returnType == null) {
                    errors.add(Errors.unknownType(method.returnType.name(), method.returnType.name()));
                    continue;
                }

                // Parameters
                List<Function.Param> params = new ArrayList<>();

                for (ProtoParameter param : method.params) {
                    Type type = getType(param.type());
                    if (type == null) {
                        errors.add(Errors.unknownType(param.type().name(), method.returnType.name()));
                        continue;
                    }

                    params.add(new Function.Param(type, param.name()));
                }

                struct.methods().add(new Method(struct, method.name, returnType, params));
            }
        }

        // Functions
        for (ProtoFunction proto : result.functions) {
            // Return type
            Type returnType = getType(proto.returnType);
            if (returnType == null) {
                errors.add(Errors.unknownType(proto.returnType.name(), proto.returnType.name()));
                continue;
            }

            // Parameters
            List<Function.Param> params = new ArrayList<>();

            for (ProtoParameter param : proto.params) {
                Type type = getType(param.type());
                if (type == null) {
                    errors.add(Errors.unknownType(param.type().name(), proto.returnType.name()));
                    continue;
                }

                params.add(new Function.Param(type, param.name()));
            }

            // Create
            functions.put(proto.name.lexeme(), new Function(proto.name, returnType, params));
        }

        return errors;
    }
}
