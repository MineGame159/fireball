import minegame159.fireball.parser.Scanner;
import minegame159.fireball.parser.Token;
import minegame159.fireball.parser.TokenType;
import org.junit.Assert;
import org.junit.Test;

import java.io.StringReader;

public class LexerTest {
    @Test
    public void main() {
        String source = "+*;    ()\n// comment\n<>=\"Hello\"\n23 && 6.23\nMain true : false, for f";
        Scanner scanner = new Scanner(new StringReader(source));

        Token token;
        while ((token = scanner.next()).type() != TokenType.Eof) {
            System.out.printf("%s '%s' %d:%d%n", token.type(), token.lexeme(), token.line(), token.character());
            if (token.type() == TokenType.Error) Assert.fail();
        }
    }
}
