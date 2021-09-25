package minegame159.fireball.types;

import java.util.Locale;

public enum PrimitiveTypes {
    Void,
    Bool,
    U8, U16, U32, U64,
    I8, I16, I32, I64,
    F32, F64;

    public final PrimitiveType type;

    PrimitiveTypes() {
        this.type = new PrimitiveType(name().toLowerCase(Locale.ROOT), this);
    }
}
