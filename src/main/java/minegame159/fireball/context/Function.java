package minegame159.fireball.context;

import minegame159.fireball.parser.Token;
import minegame159.fireball.types.Type;

import java.util.List;

public record Function(Token name, Type returnType, List<Param> params) {
    public record Param(Type type, Token name) {}
}
