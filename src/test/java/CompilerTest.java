import minegame159.fireball.Error;
import minegame159.fireball.context.Context;
import minegame159.fireball.parser.Parser;
import minegame159.fireball.passes.Checker;
import minegame159.fireball.passes.Compiler;
import minegame159.fireball.passes.TypeResolver;
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
                struct Point {
                    i32 x;
                    i32 y;
                }
                
                i32 main() {
                    // Basic
                    var b = getNumber();
                    b = 6 / 2;
                    
                    print(b);
                    
                    // Structs
                    
                    return 0;
                }
                
                i32 getNumber() {
                    return 8;
                }
                
                void printPoint(Point p) {
                    c{ printf("[%d, %d]", p.x, p.y); }
                }
                
                void print(i32 number) {
                    c{ printf("%d", number); }
                }
                """;

        // Parse
        Parser.Result result = Parser.parse(new StringReader(source));
        Assert.assertTrue(reportErrors(result.error));

        // Resolve types
        Context context = new Context();

        Assert.assertTrue(reportErrors(context.apply(result)));
        Assert.assertTrue(reportErrors(TypeResolver.resolve(result, context)));

        // Check
        Assert.assertTrue(reportErrors(Checker.check(result, context)));

        // Compile
        try {
            Compiler.compile(result, context, new FileWriter("out/test.c"));
        } catch (IOException e) {
            e.printStackTrace();
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

    private boolean reportErrors(Error error) {
        if (error != null) System.out.println(error);

        return error == null;
    }

    private boolean reportErrors(List<Error> errors) {
        for (Error error : errors) System.out.println(error);

        return errors.isEmpty();
    }
}
