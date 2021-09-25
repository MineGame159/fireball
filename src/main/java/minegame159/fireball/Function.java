package minegame159.fireball;

import minegame159.fireball.types.Type;

import java.util.List;

public record Function(Token name, Type returnType, List<Param> params) {
    public record Param(Type type, Token name) {}
}
