package minegame159.fireball;

import minegame159.fireball.context.Struct;
import minegame159.fireball.parser.Expr;
import minegame159.fireball.parser.Token;
import minegame159.fireball.parser.prototypes.ProtoType;
import minegame159.fireball.types.Type;

import java.util.List;

public class Errors {
    // Types

    public static Error unknownType(Token token, ProtoType type) {
        return new Error(token, "Unknown type '" + type + "'.");
    }

    public static Error mismatchedType(Token token, Type expected, Type got) {
        return new Error(token, "Mismatched type, expected '" + expected + "' but got '" + got + "'.");
    }

    public static Error couldNotGetType(Token name) {
        return new Error(name, "Could not get type of '" + name + "'.");
    }

    // Variables

    public static Error undeclared(Token name) {
        return new Error(name, "Undeclared variable '" + name + "'.");
    }

    public static Error undefined(Token name) {
        return new Error(name, "Undefined variable '" + name + "'.");
    }

    // Functions

    public static Error invalidCallTarget(Token token) {
        return new Error(token, "Invalid call target.");
    }

    public static Error wrongArgumentCount(Token token, int expected, int got) {
        return new Error(token, "Wrong number of arguments, expected " + expected + " but got " + got + ".");
    }

    public static Error missingReturn(Token function) {
        return new Error(function, "Function '" + function + "' is missing a top level return statement.");
    }

    // Structs

    public static Error invalidFieldTarget(Token token) {
        return new Error(token, "Can only use fields on structs.");
    }

    public static Error unknownField(Token struct, Token field) {
        return new Error(field, "Struct '" + struct + "' does not contain field '" + field + "'.");
    }

    public static Error unknownConstructor(Struct struct, Token name, List<Expr> arguments) {
        StringBuilder sb = new StringBuilder("Struct '").append(struct.name()).append("' does not contain constructor with arguments '");

        for (int i = 0; i < arguments.size(); i++) {
            if (i > 0) sb.append(", ");
            sb.append(arguments.get(i).getType());
        }

        return new Error(name, sb.append("'.").toString());
    }

    // Other

    public static Error duplicate(Token name, String construct) {
        return new Error(name, "Duplicate " + construct + " '" + name + "'.");
    }

    public static Error wrongOperands(Token token, String operation, String expected, boolean plural) {
        if (plural) return new Error(token, "Operands of " + operation + " operation must be " + expected + "s.");
        return new Error(token, "Operand of " + operation + " operation must be " + expected + ".");
    }

    public static Error invalidPointerTarget(Token token) {
        return new Error(token, "Invalid pointer target.");
    }
}
