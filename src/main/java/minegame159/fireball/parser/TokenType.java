package minegame159.fireball.parser;

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

    Ampersand, And,
    Pipe, Or,

    // Literals
    Null, Int, Float, String, Identifier,

    // Keywords
    True, False, If, Else, While, For, Var, Return, CBlock,

    Error, Eof
}
