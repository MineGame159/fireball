package cmd

import (
	"fireball/cmd/build"
	"fireball/core"
	"fireball/core/checker"
	"fireball/core/codegen"
	"fireball/core/parser"
	"fireball/core/scanner"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GetBuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build file",
		Short: "Build a single source file into an executable",
		Args:  cobra.ExactArgs(1),
		Run:   buildCmd,
	}
}

func buildCmd(_ *cobra.Command, args []string) {
	buildExecutable(args[0])
}

func buildExecutable(input string) string {
	start := time.Now()

	// Read file
	b, err := os.ReadFile(input)
	if err != nil {
		log.Fatalln("Invalid file")
	}

	text := string(b)

	// Parse and check
	reporter := &consoleReporter{
		error:   color.New(color.FgRed),
		warning: color.New(color.FgYellow),
	}

	decls := parser.Parse(reporter, scanner.NewScanner(text))
	checker.Check(reporter, decls)

	if reporter.errorCount > 0 {
		fmt.Println()
		_, _ = color.New(color.FgRed).Print("Build failed")

		if reporter.errorCount == 1 {
			fmt.Printf(", with %d error\n", reporter.errorCount)
		} else {
			fmt.Printf(", with %d errors\n", reporter.errorCount)
		}

		fmt.Println()
		os.Exit(1)
	}

	// Get output paths
	filename := getFilename(input)
	output := "build/" + filename

	// Emit LLVM IR file
	_ = os.Mkdir("build", 0750)
	irFile, _ := os.Create(output + ".ll")

	codegen.Emit(input, decls, irFile)

	// Compile to object file
	c := build.Compiler{
		OptimizationLevel: 0,
	}

	if err := c.Compile(irFile.Name(), output+".o"); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	// Link to executable
	l := build.Linker{
		Crt: true,
	}

	l.AddLibrary("c")
	l.AddInput(output + ".o")

	if err := l.Link(output); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}

	// Return
	took := time.Now().Sub(start)

	if reporter.hadDiagnostic {
		fmt.Println()
	}

	_, _ = color.New(color.FgGreen).Print("Build successful")
	fmt.Printf(", took %s\n", took)
	fmt.Println()

	return output
}

func getFilename(path string) string {
	filename := filepath.Base(path)

	i := strings.IndexByte(filename, '.')
	if i != -1 {
		filename = filename[:i]
	}

	return filename
}

type consoleReporter struct {
	error   *color.Color
	warning *color.Color

	hadDiagnostic bool
	errorCount    int
}

func (c *consoleReporter) Report(diag core.Diagnostic) {
	if diag.Kind == core.ErrorKind {
		_, _ = c.error.Fprint(os.Stderr, "ERROR   ")
		_, _ = fmt.Fprintln(os.Stderr, diag.String())

		c.errorCount++
	} else {
		_, _ = c.warning.Print("WARNING ")
		fmt.Println(diag.String())
	}

	c.hadDiagnostic = true
}
