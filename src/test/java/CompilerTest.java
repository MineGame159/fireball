import minegame159.fireball.*;
import minegame159.fireball.Compiler;
import minegame159.fireball.Error;
import org.junit.Test;

import java.io.FileWriter;
import java.io.IOException;
import java.io.StringReader;

public class CompilerTest {
    @Test
    public void main() {
        String source = """
                int main() {
                    i32 b = 8;
                    b = 6 / 2;
                    
                    print(b);
                    
                    return 0;
                }
                
                void print(i32 number) {
                    c{
                        printf("%d", number);
                    }
                }
                """;

        Context context = new Context();

        // Parse
        Parser parser = new Parser(context, new StringReader(source));
        parser.parse();

        if (parser.errors.size() > 0) {
            for (Error error : parser.errors) {
                System.out.printf("Error [line %d]: %s%n", error.token.line(), error.getMessage());
            }

            return;
        }

        // Check
        Checker checker = new Checker(context);
        checker.check(parser.stmts);

        if (checker.errors.size() > 0) {
            for (Error error : checker.errors) {
                System.out.printf("Error [line %d]: %s%n", error.token.line(), error.getMessage());
            }

            return;
        }

        // Compile
        try {
            Compiler compiler = new Compiler(context, new FileWriter("out/test.c"));
            compiler.compile(parser.stmts);
        } catch (IOException e) {
            e.printStackTrace();
        }

        // Run
        try {
            new ProcessBuilder().command("gcc", "-o", "out/test.exe", "out/test.c").inheritIO().start().waitFor();
            new ProcessBuilder().command("out/test.exe").inheritIO().start().waitFor();
        } catch (IOException | InterruptedException e) {
            e.printStackTrace();
        }
    }
}
