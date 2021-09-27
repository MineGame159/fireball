package minegame159.fireball.types;

import minegame159.fireball.context.Struct;

public class StructType extends Type {
    public final Struct struct;

    public StructType(String name, Struct struct) {
        super(name);
        this.struct = struct;
    }

    @Override
    public boolean equals(Type type) {
        if (type instanceof StructType t) return this.name.equals(t.name);
        return false;
    }
}
