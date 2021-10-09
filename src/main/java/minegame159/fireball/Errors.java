package minegame159.fireball;

import minegame159.fireball.context.Struct;
import minegame159.fireball.parser.Expr;
import minegame159.fireball.parser.Token;
import minegame159.fireball.parser.TokenType;
import minegame159.fireball.parser.prototypes.ProtoType;
import minegame159.fireball.types.Type;

import java.util.ArrayList;
import java.util.List;

public class Errors {
    private static final List<Error> list = new ArrayList<>();
    private static boolean canAddErrors;
    
    public static void clear() {
        list.clear();
        canAddErrors = true;
    }
    
    public static List<Error> get() {
        canAddErrors = false;
        return list;
    }

    private static void add(Error error) {
        if (!canAddErrors) throw new RuntimeException("Tried to add error at wrong time.");
        list.add(error);
    }

    // Types

    public static void unknownType(Token token, ProtoType type) {
        add(new Error(token, "Unknown type '" + type + "'."));
    }

    public static void mismatchedType(Token token, Type expected, Type got) {
        add(new Error(token, "Mismatched type, expected '" + expected + "' but got '" + got + "'."));
    }

    public static void couldNotGetType(Token name) {
        add(new Error(name, "Could not get type of '" + name + "'."));
    }

    // Variables

    public static void undeclared(Token name) {
        add(new Error(name, "Undeclared variable '" + name + "'."));
    }

    public static void undefined(Token name) {
        add(new Error(name, "Undefined variable '" + name + "'."));
    }

    // Functions

    public static void invalidCallTarget(Token token) {
        add(new Error(token, "Invalid call target."));
    }

    public static void wrongArgumentCount(Token token, int expected, int got) {
        add(new Error(token, "Wrong number of arguments, expected " + expected + " but got " + got + "."));
    }

    public static void missingReturn(Token function) {
        add(new Error(function, "Function '" + function + "' is missing a top level add(statement."));
    }

    // Structs

    public static void invalidFieldTarget(Token token) {
        add(new Error(token, "Can only use fields on structs."));
    }

    public static void invalidNewTarget(Token token) {
        add(new Error(token, "Can only use new on structs."));
    }

    public static void unknownField(Token struct, Token field) {
        add(new Error(field, "Struct '" + struct + "' does not contain field '" + field + "'."));
    }

    public static void unknownConstructor(Struct struct, Token name, List<Expr> arguments) {
        StringBuilder sb = new StringBuilder("Struct '").append(struct.name).append("' does not contain constructor with arguments '");

        for (int i = 0; i < arguments.size(); i++) {
            if (i > 0) sb.append(", ");
            sb.append(arguments.get(i).getType());
        }

        add(new Error(name, sb.append("'.").toString()));
    }

    // Other

    public static void duplicate(Token name, String construct) {
        add(new Error(name, "Duplicate " + construct + " '" + name + "'."));
    }

    public static void wrongOperands(Token token, String operation, String expected, boolean plural) {
        if (plural) add(new Error(token, "Operands of " + operation + " operation must be " + expected + "s."));
        add(new Error(token, "Operand of " + operation + " operation must be " + expected + "."));
    }

    public static void invalidPointerTarget(Token token) {
        add(new Error(token, "Invalid pointer target."));
    }

    public static void invalidUnaryPostTarget(Token operator) {
        add(new Error(operator, "Invalid " + (operator.type() == TokenType.PlusPlus ? "increment" : "decrement") + " target."));
    }

    public static void cannotDelete(Token token, Type type) {
        add(new Error(token, "Can only delete struct pointers. Got '" + type + "'."));
    }
}
