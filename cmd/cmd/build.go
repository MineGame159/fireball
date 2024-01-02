package cmd

import (
	"fireball/cmd/build"
	"fireball/core/ast"
	"fireball/core/codegen"
	"fireball/core/llvm"
	"fireball/core/utils"
	"fireball/core/workspace"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var opt uint8

func GetBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build project.",
		Run:   buildCmd,
	}

	cmd.Flags().Uint8VarP(&opt, "opt", "O", 0, "Optimization level. [-O0, -O1, -O2, or -O3] (default = '-O0')")

	return cmd
}

func buildCmd(_ *cobra.Command, _ []string) {
	buildProject()
}

//goland:noinspection GoBoolExpressions
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
		codegen.Emit(file.Path, project, file.Ast, irFile)
		_ = irFile.Close()

		irPaths = append(irPaths, path)
	}

	entrypointPath := filepath.Join(project.Path, "build", "__entrypoint.ll")
	irPaths = append(irPaths, entrypointPath)

	err = generateEntrypoint(project, entrypointPath)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Compile
	c := build.Compiler{
		OptimizationLevel: min(max(int(opt), 0), 3),
	}

	for _, irPath := range irPaths {
		c.AddInput(irPath)
	}

	if runtime.GOOS == "darwin" {
		c.AddLibrary("System")
	} else {
		c.AddLibrary("m")
		c.AddLibrary("c")
	}

	output := filepath.Join(project.Path, "build", project.Config.Name)
	err = c.Compile(output)

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

func generateEntrypoint(project *workspace.Project, path string) error {
	// Create module
	m := llvm.NewModule()
	m.Source("__entrypoint")

	function, _ := project.GetFunction("main")

	void := m.Void()
	i32 := m.Primitive("i32", 32, llvm.SignedEncoding)

	main := m.Define(m.Function("main", []llvm.Type{}, false, i32), "_fireball_entrypoint")
	main.PushScope()
	mainBlock := main.Block("")

	if function != nil {
		var fbMain llvm.Value

		if ast.IsPrimitive(function.Returns, ast.I32) {
			fbMain = m.Declare(m.Function(function.MangledName(), []llvm.Type{}, false, i32))
		} else {
			fbMain = m.Declare(m.Function(function.MangledName(), []llvm.Type{}, false, void))
		}

		if ast.IsPrimitive(function.Returns, ast.I32) {
			call := mainBlock.Call(fbMain, []llvm.Value{}, i32)
			call.SetLocation(nil)

			mainBlock.Ret(call)
		} else {
			call := mainBlock.Call(fbMain, []llvm.Value{}, void)
			call.SetLocation(nil)

			mainBlock.Ret(main.Literal(i32, llvm.Literal{Signed: 0}))
		}
	} else {
		mainBlock.Ret(main.Literal(i32, llvm.Literal{Signed: 0}))
	}

	// Write module
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	llvm.WriteText(m, file)

	_ = file.Close()
	return nil
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
