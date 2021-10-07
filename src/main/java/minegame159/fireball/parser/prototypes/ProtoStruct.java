package minegame159.fireball.parser.prototypes;

import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;

import java.util.List;

public class ProtoStruct {
    public final Token name;
    public final List<ProtoParameter> fields;
    public final List<ProtoMethod> constructors;
    public final ProtoMethod destructor;
    public final List<ProtoMethod> methods;

    public boolean skip;

    public ProtoStruct(Token name, List<ProtoParameter> fields, List<ProtoMethod> constructors, ProtoMethod destructor, List<ProtoMethod> methods) {
        this.name = name;
        this.fields = fields;
        this.constructors = constructors;
        this.destructor = destructor;
        this.methods = methods;

        for (ProtoMethod method : methods) method.owner = this;
        for (ProtoMethod constructor : constructors) constructor.owner = this;
        if (destructor != null) destructor.owner = this;
    }

    public void accept(Stmt.Visitor visitor) {
        for (ProtoFunction method : methods) method.accept(visitor);
        for (ProtoMethod constructor : constructors) constructor.accept(visitor);
        if (destructor != null) destructor.accept(visitor);
    }
}
