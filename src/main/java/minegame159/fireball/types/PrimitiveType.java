package minegame159.fireball.types;

public class PrimitiveType extends Type {
    public final PrimitiveTypes type;

    public PrimitiveType(String name, PrimitiveTypes type) {
        super(name);
        this.type = type;
    }

    @Override
    protected Type copy() {
        return new PrimitiveType(name, type);
    }

    @Override
    public boolean equals(Type type) {
        if (type instanceof PrimitiveType t) return this.type == t.type;
        return false;
    }
}
