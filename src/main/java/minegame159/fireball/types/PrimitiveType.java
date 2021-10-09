package minegame159.fireball.types;

public class PrimitiveType extends Type {
    public final PrimitiveTypes type;

    public PrimitiveType(String name, PrimitiveTypes type) {
        super(name);
        this.type = type;
    }

    @Override
    public boolean canBeAssignedTo(Type to) {
        // Assigning to primitive type
        if (to instanceof PrimitiveType primitiveType) {
            return primitiveType.type.size >= type.size;
        }
        // Assigning to struct type
        else if (to instanceof StructType) {
            return to.isPointer() && type == PrimitiveTypes.Void && isPointer();
        }

        return false;
    }

    @Override
    protected Type copy() {
        return new PrimitiveType(name, type);
    }

    @Override
    public boolean equals(Type type) {
        if (!super.equals(type)) return false;
        return this.type == ((PrimitiveType) type).type;
    }
}
