package minegame159.fireball.passes;

import java.io.IOException;
import java.io.Writer;

public class CompilerWriter {
    private final Writer writer;

    private int indentLevel;
    private String indent = "";
    private boolean skipIndent;

    private boolean skipNewLine;

    public CompilerWriter(Writer writer) {
        this.writer = writer;
    }

    public void indentUp() {
        indentLevel++;

        indent = "";
        for (int i = 0; i < indentLevel; i++) indent += "    ";
    }

    public void indentDown() {
        indentLevel--;

        indent = "";
        for (int i = 0; i < indentLevel; i++) indent += "    ";
    }

    public CompilerWriter skipIndent() {
        skipIndent = true;
        return this;
    }

    public CompilerWriter indent() {
        if (skipIndent) {
            skipIndent = false;
            return this;
        }

        write(indent);
        return this;
    }

    public CompilerWriter skipNewLine() {
        skipNewLine = true;
        return this;
    }

    public CompilerWriter write(char c) {
        try {
            writer.write(c);
        } catch (IOException e) {
            e.printStackTrace();
        }
        return this;
    }

    public void writeln(char c) {
        try {
            writer.write(c);

            if (skipNewLine) skipNewLine = false;
            else writer.write('\n');
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public CompilerWriter write(Object o) {
        try {
            writer.write(o.toString());
        } catch (IOException e) {
            e.printStackTrace();
        }
        return this;
    }

    public void writeln(String s) {
        try {
            writer.write(s);

            if (skipNewLine) skipNewLine = false;
            else writer.write('\n');
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public void writeSemicolon() {
        try {
            writer.write(";");

            if (skipNewLine) skipNewLine = false;
            else writer.write('\n');
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public void close() {
        try {
            writer.close();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
