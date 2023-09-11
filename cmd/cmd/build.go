package cmd

import (
	"fireball/cmd/build"
	"fireball/core/codegen"
	"fireball/core/utils"
	"fireball/core/workspace"
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
		Use:   "build",
		Short: "Build project.",
		Run:   buildCmd,
	}
}

func buildCmd(_ *cobra.Command, args []string) {
	buildProject()
}

func buildProject() string {
	start := time.Now()

	// Create project
	project, err := workspace.NewProject(".")
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Load files
	err = project.LoadFiles()
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Report errors
	reporter := consoleReporter{
		error:   color.New(color.FgRed),
		warning: color.New(color.FgYellow),
	}

	for _, file := range project.Files {
		for _, diagnostic := range file.FlushDiagnostics() {
			reporter.Report(file, diagnostic)
		}
	}

	if reporter.errorCount > 0 {
		fmt.Println()
		_, _ = color.New(color.FgRed).Print("Build failed")

		if reporter.errorCount == 1 {
			fmt.Printf(", with %d error\n", reporter.errorCount)
		} else {
			fmt.Printf(", with %d errors\n", reporter.errorCount)
		}

		os.Exit(1)
	}

	// Emit LLVM IR
	_ = os.Mkdir("build", 0750)

	irPaths := make([]string, 0, len(project.Files))

	for _, file := range project.Files {
		path := strings.ReplaceAll(file.Path, "/", "-")
		path = filepath.Join(project.Path, "build", path[:len(path)-3]+".ll")

		irFile, _ := os.Create(path)
		codegen.Emit(file.Path, project, file.Decls, irFile)
		_ = irFile.Close()

		irPaths = append(irPaths, path)
	}

	entrypointPath := filepath.Join(project.Path, "build", "__entrypoint.ll")
	irPaths = append(irPaths, entrypointPath)

	err = os.WriteFile(entrypointPath, []byte(`
define i32 @main() {
entry:
	%0 = call i32 @fb$main()
	ret i32 %0
}

declare i32 @fb$main()
`), 0750)

	if err != nil {
		log.Fatalln(err.Error())
	}

	// Compile LLVM IR files to object files
	c := build.Compiler{OptimizationLevel: 0}

	for _, irPath := range irPaths {
		path := irPath[:len(irPath)-3] + ".o"
		err := c.Compile(irPath, path)

		if err != nil {
			log.Fatalln(err.Error())
		}
	}

	// Link all object files to final executable
	l := build.Linker{Crt: true}

	l.AddLibrary("c")

	for _, irPath := range irPaths {
		l.AddInput(irPath[:len(irPath)-3] + ".o")
	}

	output := filepath.Join(project.Path, "build", project.Config.Name)
	err = l.Link(output)

	if err != nil {
		log.Fatalln(err.Error())
	}

	// Print info
	took := time.Now().Sub(start)

	if reporter.hadDiagnostic {
		fmt.Println()
	}

	_, _ = color.New(color.FgGreen).Print("Build successful")
	fmt.Printf(", took %s\n", took)
	fmt.Println()

	// Return
	return output
}

type consoleReporter struct {
	error   *color.Color
	warning *color.Color

	hadDiagnostic bool
	errorCount    int
}

func (c *consoleReporter) Report(file *workspace.File, diag utils.Diagnostic) {
	path, err := filepath.Rel(file.Project.Config.Src, file.Path)
	if err != nil {
		path = file.Path
	}

	msg := fmt.Sprintf("[%s:%d:%d] %s", path, diag.Range.Start.Line, diag.Range.Start.Column+1, diag.Message)

	if diag.Kind == utils.ErrorKind {
		_, _ = c.error.Fprint(os.Stderr, "ERROR   ")
		_, _ = fmt.Fprintln(os.Stderr, msg)

		c.errorCount++
	} else {
		_, _ = c.warning.Print("WARNING ")
		fmt.Println(msg)
	}

	c.hadDiagnostic = true
}
