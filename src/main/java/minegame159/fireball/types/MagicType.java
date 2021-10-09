package minegame159.fireball.types;

public class MagicType extends Type {
    public static final MagicType INSTANCE = new MagicType("mAgIc OwO");

    private MagicType(String name) {
        super(name);
    }

    @Override
    public boolean canBeAssignedTo(Type to) {
        return true;
    }

    @Override
    protected Type copy() {
        return INSTANCE;
    }

    @Override
    public boolean equals(Type type) {
        return true;
    }
}
