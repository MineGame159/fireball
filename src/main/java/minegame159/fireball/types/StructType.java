package minegame159.fireball.types;

import minegame159.fireball.context.Struct;

public class StructType extends Type {
    public final Struct struct;

    public StructType(String name, Struct struct) {
        super(name);
        this.struct = struct;
    }

    @Override
    protected Type copy() {
        return new StructType(name, struct);
    }

    @Override
    public boolean equals(Type type) {
        if (!super.equals(type)) return false;
        return name.equals(type.name);
    }
}
