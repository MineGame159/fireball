package minegame159.fireball.context;

import minegame159.fireball.Error;
import minegame159.fireball.Errors;
import minegame159.fireball.IFunction2;
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
        if (type == null) return null;

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

        // Register structs as types
        for (ProtoStruct proto : result.structs) {
            // Check for duplicate type
            if (types.containsKey(proto.name.lexeme())) {
                errors.add(Errors.duplicate(proto.name, "type"));
                proto.skip = true;
                continue;
            }

            // Register struct
            Struct struct = new Struct(proto.name, proto.constructors.size(), proto.methods.size());

            types.put(proto.name.lexeme(), new StructType(proto.name.lexeme(), struct));
            structs.put(proto.name.lexeme(), struct);
        }

        // Apply structs
        for (ProtoStruct protoStruct : result.structs) {
            // Skip if needed
            if (protoStruct.skip) continue;

            Struct struct = structs.get(protoStruct.name.lexeme());

            // Fields
            List<Field> fields = applyParams(errors, protoStruct.fields, "field", Field::new);
            if (fields != null) struct.fields = fields;

            // Constructors
            for (int i = 0; i < protoStruct.constructors.size(); i++) {
                ProtoMethod protoConstructor = protoStruct.constructors.get(i);
                int index = i;

                Constructor constr = applyFunction(errors, protoConstructor, (first, second) -> new Constructor(struct, protoConstructor.name, first, second, protoConstructor.body, index));
                if (constr != null) struct.constructors.add(constr);
            }

            // Destructor
            if (protoStruct.destructor != null) {
                Destructor destructor = applyFunction(errors, protoStruct.destructor, (first, second) -> new Destructor(struct, protoStruct.destructor.name, first, second, protoStruct.destructor.body));
                if (destructor != null) struct.destructor = destructor;
            }

            // Methods
            for (ProtoMethod protoMethod : protoStruct.methods) {
                Method method = applyFunction(errors, protoMethod, (first, second) -> new Method(struct, protoMethod.name, first, second, protoMethod.body));
                if (method != null) struct.methods.add(method);
            }
        }

        // Apply functions
        for (ProtoFunction protoFunc : result.functions) {
            Function func = applyFunction(errors, protoFunc, (first, second) -> new Function(protoFunc.name, first, second, protoFunc.body));
            if (func != null) functions.put(func.name.lexeme(), func);
        }

        return errors;
    }

    private <T> List<T> applyParams(List<Error> errors, List<ProtoParameter> proto, String construct, IFunction2<Type, Token, T> func) {
        boolean hadError = false;
        List<T> params = new ArrayList<>(proto.size());
        Set<String> names = new HashSet<>(proto.size());

        // Params
        for (ProtoParameter param : proto) {
            Type type = getType(param.type());
            if (type == null) {
                errors.add(Errors.unknownType(param.type().name(), param.type()));
                hadError = true;
            }

            if (names.contains(param.name().lexeme())) errors.add(Errors.duplicate(param.name(), construct));
            else names.add(param.name().lexeme());

            params.add(func.run(type, param.name()));
        }

        // Create
        if (hadError) return null;
        return params;
    }

    private <T> T applyFunction(List<Error> errors, ProtoFunction proto, IFunction2<Type, List<Function.Param>, T> func) {
        boolean hadError = false;

        // Return type
        Type returnType = getType(proto.returnType);
        if (returnType == null) {
            errors.add(Errors.unknownType(proto.returnType.name(), proto.returnType));
            hadError = true;
        }

        // Parameters
        List<Function.Param> params = applyParams(errors, proto.params, "parameter", Function.Param::new);
        if (params == null) hadError = true;

        // Create
        if (hadError) return null;
        return func.run(returnType, params);
    }
}
