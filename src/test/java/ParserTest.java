import minegame159.fireball.Parser;

import java.io.StringReader;

public class ParserTest {
    public static void main(String[] args) {
        String source = "(5 + 6) * 3.6";

        Parser parser = new Parser(new StringReader(source));
        parser.parse();

        System.out.printf("Expressions: %d%n", parser.exprs.size());
        System.out.printf("Errors: %d", parser.errors.size());
    }
}
