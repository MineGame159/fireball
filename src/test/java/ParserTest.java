import minegame159.fireball.Context;
import minegame159.fireball.Error;
import minegame159.fireball.Parser;

import java.io.StringReader;

public class ParserTest {
    public static void main(String[] args) {
        String source = """
                void hi() {
                    print("Hi");
                }
                
                void hello(i32 number, String string) {
                    return;
                    print("No cope");
                }
                
                void main() {
                    i32 number = 159;
                    number = 9;
                    
                    {
                        String name = "MineGame159";
                        var me = name;
                    }
                    
                    if (true && number == 159) {
                        hi();
                        hello(5, "no");
                    }
                    
                    while (cope || notCope) {}
                    
                    for (i32 i = 0; i < 5; i = i + 1) {}
                    for (;;) {}
                }
                """;

        Parser parser = new Parser(new Context(), new StringReader(source));
        parser.parse();

        System.out.printf("Statements: %d%n", parser.stmts.size());

        for (Error error : parser.errors) {
            System.out.printf("Error [line %d]: %s%n", error.token.line(), error.getMessage());
        }
    }
}
