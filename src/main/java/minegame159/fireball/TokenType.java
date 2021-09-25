package minegame159.fireball;

public enum TokenType {
    // Single-character tokens
    LeftParen, RightParen, LeftBrace, RightBrace,
    Comma, Dot, Colon, Semicolon,

    // One or two character tokens
    Bang, BangEqual,
    Equal, EqualEqual,
    Greater, GreaterEqual,
    Less, LessEqual,

    Plus, PlusEqual,
    Minus, MinusEqual,
    Star, StarEqual,
    Slash, SlashEqual,
    Percentage, PercentageEqual,

    And, Or,

    // Literals
    Null, Int, Float, String, Identifier,

    // Keywords
    True, False, If, Else, While, For, Var, Return,

    Error, Eof
}
