package minegame159.fireball.context;

import minegame159.fireball.parser.Token;

import java.util.List;

public record Struct(Token name, List<Field> fields, List<Method> methods) {
    public Field getField(Token name) {
        for (Field field : fields) {
            if (field.name().equals(name)) return field;
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
