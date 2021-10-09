package minegame159.fireball.types;

import java.util.Locale;

public enum PrimitiveTypes {
    Void(0),
    Bool(1),
    U8(1), U16(2), U32(4), U64(8),
    I8(1), I16(2), I32(4), I64(8),
    F32(4), F64(8);

    public final PrimitiveType type;
    public final int size;

    PrimitiveTypes(int size) {
        this.type = new PrimitiveType(name().toLowerCase(Locale.ROOT), this);
        this.size = size;
    }
}
