import minegame159.fireball.Scanner;
import minegame159.fireball.Token;
import minegame159.fireball.TokenType;

import java.io.StringReader;

public class LexerTest {
    public static void main(String[] args) {
        String source = "+*;    ()\n// comment\n<>=\"Hello\"\n23 && 6.23\nMain true false, for f";
        Scanner scanner = new Scanner(new StringReader(source));

        Token token;
        while ((token = scanner.next()).type() != TokenType.Eof) {
            System.out.printf("%s '%s' %d%n", token.type(), token.lexeme(), token.line());
        }
    }
}
