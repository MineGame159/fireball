import minegame159.fireball.Error;
import minegame159.fireball.Parser;

import java.io.StringReader;

public class ParserTest {
    public static void main(String[] args) {
        String source = "(5 + 6) * 3.6;\"Hi\";";

        Parser parser = new Parser(new StringReader(source));
        parser.parse();

        System.out.printf("Statements: %d%n", parser.stmts.size());

        for (Error error : parser.errors) {
            System.out.printf("Error [line %d]: %s%n", error.token.line(), error.getMessage());
        }
    }
}
