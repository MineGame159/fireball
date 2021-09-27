package minegame159.fireball;

import minegame159.fireball.parser.Token;
import minegame159.fireball.types.Type;

public class Errors {
    // Types

    public static Error unknownType(Token token, Token type) {
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

    // Calls

    public static Error invalidCallTarget(Token token) {
        return new Error(token, "Invalid call target.");
    }

    public static Error wrongArgumentCount(Token token, int expected, int got) {
        return new Error(token, "Wrong number of arguments, expected " + expected + " but got " + got + ".");
    }

    // Other

    public static Error wrongOperands(Token token, String operation, String expected, boolean plural) {
        if (plural) return new Error(token, "Operands of " + operation + " operation must be " + expected + "s.");
        return new Error(token, "Operand of " + operation + " operation must be " + expected + ".");
    }

    public static Error duplicateField(Token name) {
        return new Error(name, "Duplicate field '" + name + "'.");
    }
}
