package minegame159.fireball.types;

public abstract class Type implements Cloneable {
    public String name;
    private boolean pointer;

    public Type(String name) {
        this.name = name;
    }

    public abstract boolean canBeAssignedTo(Type to);

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

    public boolean isPointer() {
        return pointer;
    }

    public Type pointer() {
        if (pointer) return this;

        Type type = copy();
        type.name += '*';
        type.pointer = true;
        return type;
    }

    protected abstract Type copy();

    public boolean equals(Type type) {
        return getClass() == type.getClass() && pointer == type.pointer;
    }

    @Override
    public String toString() {
        return name;
    }
}
