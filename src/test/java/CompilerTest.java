import minegame159.fireball.Compiler;
import minegame159.fireball.Parser;

import java.io.FileWriter;
import java.io.IOException;
import java.io.StringReader;

public class CompilerTest {
    public static void main(String[] args) {
        String source = """
                void print(i32 number) {
                    c{
                        printf("%d", number);
                    }
                }
                
                void main() {
                    print(159);
                }
                """;

        Parser parser = new Parser(new StringReader(source));
        parser.parse();

        try {
            Compiler compiler = new Compiler(new FileWriter("out/test.c"));
            compiler.compile(parser.stmts);
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
