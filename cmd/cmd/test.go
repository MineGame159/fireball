package cmd

import (
	"bytes"
	"errors"
	"fireball/cmd/build"
	"fireball/core/ast"
	"fireball/core/ir"
	"fireball/core/workspace"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test project.",
		Run:   testCmd,
	}

	return cmd
}

func testCmd(_ *cobra.Command, _ []string) {
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
	build.Report(project)

	// Build
	tests := getTests(project)

	output, err := build.Build(project, generateTestsEntrypoint(tests), 0, fmt.Sprintf("%s_tests", project.Config.Name))
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Run
	failed, err := runTests(output, tests)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Print info
	took := time.Now().Sub(start)

	namespaceStyle := color.New(color.FgWhite)
	testStyle := color.New(color.Underline)

	for _, test := range failed {
		file := ast.GetParent[*ast.File](test)

		for _, part := range file.Namespace.Name.Parts {
			_, _ = namespaceStyle.Printf("%s.", part)
		}

		if receiver, ok := test.Receiver().(ast.StructType); ok {
			_, _ = namespaceStyle.Printf("%s.", receiver.Underlying().Name)
		}

		_, _ = testStyle.Print(test.TestName())

		color.Red(" failed")
		fmt.Println()
	}

	if len(failed) == 0 {
		_, _ = color.New(color.FgGreen).Printf("%d tests succeeded", len(tests))
	} else {
		_, _ = color.New(color.FgYellow).Printf("%d out of %d tests failed", len(failed), len(tests))
	}

	fmt.Printf(", took %s\n", took)
}

// Run tests

func runTests(path string, tests []*ast.Func) ([]*ast.Func, error) {
	// Run
	cmd := exec.Command(path)

	stdout := bytes.Buffer{}
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil || !cmd.ProcessState.Success() {
		return nil, errors.New("failed to run: " + path)
	}

	// Parse output
	var failed []*ast.Func

	for _, line := range strings.Split(stdout.String(), "\n") {
		index, err := strconv.ParseInt(line, 10, 32)
		if err != nil {
			continue
		}

		failed = append(failed, tests[index])
	}

	return failed, nil
}

// Entrypoint generation

func generateTestsEntrypoint(tests []*ast.Func) *ir.Module {
	m := &ir.Module{Path: "__entrypoint"}

	// Run
	printf := m.Declare("printf", &ir.FuncType{
		Params: []*ir.Param{{
			Typ:   &ir.PointerType{Pointee: ir.I8},
			Name_: "format",
		}},
		Variadic: true,
	})

	testType := &ir.FuncType{Returns: ir.I1}

	testParam := &ir.Param{Typ: testType, Name_: "test"}
	testIndexParam := &ir.Param{Typ: ir.I32, Name_: "index"}

	run := m.Define("__run_test", &ir.FuncType{Params: []*ir.Param{testParam, testIndexParam}}, 0)

	runBlock := run.Block("entry")
	failedBlock := run.Block("failed")
	exitBlock := run.Block("exit")

	successful := runBlock.Add(&ir.CallInst{Callee: testParam})
	runBlock.Add(&ir.BrInst{Condition: successful, True: exitBlock, False: failedBlock})

	failedMsg := m.Constant("failed", &ir.StringConst{Length: 4, Value: []byte("%d\n\000")})

	failedBlock.Add(&ir.CallInst{
		Callee: printf,
		Args: []ir.Value{
			failedMsg,
			testIndexParam,
		},
	})
	failedBlock.Add(&ir.RetInst{})

	exitBlock.Add(&ir.RetInst{})

	// Main
	main := m.Define("main", &ir.FuncType{Returns: ir.I32}, 0)
	mainBlock := main.Block("entry")

	for i, test := range tests {
		var name strings.Builder
		test.MangledName(&name)

		mainBlock.Add(&ir.CallInst{
			Callee: run,
			Args: []ir.Value{
				m.Declare(name.String(), testType),
				&ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(uint64(i))},
			},
		})
	}

	mainBlock.Add(&ir.RetInst{Value: &ir.IntConst{Typ: ir.I32, Value: ir.Unsigned(0)}})

	return m
}

// Test collection

func getTests(project *workspace.Project) []*ast.Func {
	collector := testCollector{}

	for _, file := range project.Files {
		collector.VisitNode(file.Ast)
	}

	return collector.tests
}

type testCollector struct {
	tests []*ast.Func
}

func (t *testCollector) VisitNode(node ast.Node) {
	switch node := node.(type) {
	case *ast.Func:
		name := node.TestName()

		if name != "" {
			t.tests = append(t.tests, node)
		}

	case ast.Decl, *ast.File:
		node.AcceptChildren(t)
	}
}
