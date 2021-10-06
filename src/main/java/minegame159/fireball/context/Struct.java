package minegame159.fireball.context;

import minegame159.fireball.parser.Expr;
import minegame159.fireball.parser.Token;

import java.util.List;

public record Struct(Token name, List<Field> fields, List<Constructor> constructors, List<Method> methods) {
    public Field getField(Token name) {
        for (Field field : fields) {
            if (field.name().equals(name)) return field;
        }

        return null;
    }

    public Constructor getConstructor(List<Expr> arguments) {
        for (Constructor constructor : constructors) {
            if (constructor.canCall(arguments)) return constructor;
        }

        return null;
    }

    public Method getMethod(Token name) {
        for (Method method : methods) {
            if (method.name.equals(name)) return method;
        }

        return null;
    }
}
