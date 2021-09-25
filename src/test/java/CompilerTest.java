import minegame159.fireball.Compiler;
import minegame159.fireball.Error;
import minegame159.fireball.*;
import org.junit.Assert;
import org.junit.Test;

import java.io.FileWriter;
import java.io.IOException;
import java.io.StringReader;
import java.util.List;

public class CompilerTest {
    @Test
    public void main() {
        String source = """
                i32 main() {
                    i32 b = getNumber();
                    b = 6 / 2;
                    
                    print(b);
                    
                    return 0;
                }
                
                i32 getNumber() {
                    return 8;
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

        Assert.assertTrue(reportErrors(parser.errors));

        // Resolve types
        TypeResolver typeResolver = new TypeResolver(context);
        typeResolver.resolve(parser.stmts);

        Assert.assertTrue(reportErrors(typeResolver.errors));

        // Check
        Checker checker = new Checker(context);
        checker.check(parser.stmts);

        Assert.assertTrue(reportErrors(checker.errors));

        // Compile
        try {
            Compiler compiler = new Compiler(context, new FileWriter("out/test.c"));
            compiler.compile(parser.stmts);
        } catch (IOException e) {
            e.printStackTrace();
            Assert.fail();
        }

        // Run
        try {
            new ProcessBuilder().command("gcc", "-o", "out/test.exe", "out/test.c").inheritIO().start().waitFor();
            new ProcessBuilder().command("out/test.exe").inheritIO().start().waitFor();
        } catch (IOException | InterruptedException e) {
            e.printStackTrace();
            Assert.fail();
        }
    }

    private boolean reportErrors(List<Error> errors) {
        for (Error error : errors) {
            System.out.printf("Error [line %d]: %s%n", error.token.line(), error.getMessage());
        }

        return errors.isEmpty();
    }
}
