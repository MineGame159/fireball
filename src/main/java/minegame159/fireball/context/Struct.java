package minegame159.fireball.context;

import minegame159.fireball.parser.Expr;
import minegame159.fireball.parser.Token;

import java.util.ArrayList;
import java.util.List;

public class Struct {
    public final Token name;

    public List<Field> fields;
    public final List<Constructor> constructors;
    public Destructor destructor;
    public final List<Method> methods;

    public Struct(Token name, int constructorCount, int methodCount) {
        this.name = name;
        this.constructors = new ArrayList<>(constructorCount);
        this.methods = new ArrayList<>(methodCount);
    }

    public Field getField(Token name) {
        for (Field field : fields) {
            if (field.name().equals(name)) return field;
        }

        return null;
    }

    public Constructor getConstructor(boolean returnsPointer, List<Expr> arguments) {
        for (Constructor constructor : constructors) {
            if (constructor.canCall(returnsPointer, arguments)) return constructor;
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
