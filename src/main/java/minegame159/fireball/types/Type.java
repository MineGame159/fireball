package minegame159.fireball.types;

public abstract class Type {
    public final String name;

    public Type(String name) {
        this.name = name;
    }

    public boolean isBool() {
        return this instanceof PrimitiveType t && t.type == PrimitiveTypes.Bool;
    }

    public boolean isNumber() {
        if (this instanceof PrimitiveType t) {
            return switch (t.type) {
                case U8, U16, U32, U64, I8, I16, I32, I64 -> true;
                default -> false;
            };
        }

        return false;
    }

    public abstract boolean equals(Type type);

    @Override
    public String toString() {
        return name;
    }
}
