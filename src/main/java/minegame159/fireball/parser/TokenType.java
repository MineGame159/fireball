package minegame159.fireball.parser;

public enum TokenType {
    // Single-character tokens
    LeftParen, RightParen, LeftBrace, RightBrace,
    Comma, Dot, Colon, Semicolon, Tilde,

    // One or two character tokens
    Bang, BangEqual,
    Equal, EqualEqual,
    Greater, GreaterEqual,
    Less, LessEqual,

    Plus, PlusPlus, PlusEqual,
    Minus, MinusMinus, MinusEqual,
    Star, StarEqual,
    Slash, SlashEqual,
    Percentage, PercentageEqual,

    Ampersand, And,
    Pipe, Or,

    // Literals
    Null, Int, Float, String, Identifier,

    // Keywords
    True, False, If, Else, While, For, Var, Return, CBlock, Struct, New, Delete,

    Error, Eof
}
