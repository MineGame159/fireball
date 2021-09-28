package minegame159.fireball.parser.prototypes;

import minegame159.fireball.parser.Stmt;
import minegame159.fireball.parser.Token;

import java.util.List;

public record ProtoStruct(Token name, List<ProtoParameter> fields) {
    public void accept(Stmt.Visitor visitor) {}
}
