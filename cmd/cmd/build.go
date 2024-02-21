package cmd

import (
	"fireball/cmd/build"
	"fireball/core/ast"
	"fireball/core/ir"
	"fireball/core/workspace"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"log"
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
	build.Report(project)

	// Build
	output, err := build.Build(project, generateEntrypoint(project), opt, project.Config.Name)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Print info
	took := time.Now().Sub(start)

	_, _ = color.New(color.FgGreen).Print("Build successful")
	fmt.Printf(", took %s\n", took)
	fmt.Println()

	// Return
	return output
}

func generateEntrypoint(project *workspace.Project) *ir.Module {
	m := &ir.Module{Path: "__entrypoint"}

	resolver := project.GetResolverName(project.Config.Namespace)
	function := resolver.GetFunction("main")

	main := m.Define("main", &ir.FuncType{Returns: ir.I32}, 0)
	mainBlock := main.Block("")

	if function != nil {
		f := function.Underlying()
		var fbMain ir.Value

		var name strings.Builder
		f.MangledName(&name)

		if ast.IsPrimitive(f.Returns(), ast.I32) {
			fbMain = m.Declare(name.String(), &ir.FuncType{Returns: ir.I32})
		} else {
			fbMain = m.Declare(name.String(), &ir.FuncType{})
		}

		call := mainBlock.Add(&ir.CallInst{
			Callee: fbMain,
			Args:   nil,
		})

		if ast.IsPrimitive(f.Returns(), ast.I32) {
			mainBlock.Add(&ir.RetInst{Value: call})
		} else {
			mainBlock.Add(&ir.RetInst{Value: &ir.IntConst{
				Typ:   ir.I32,
				Value: ir.Unsigned(0),
			}})
		}
	} else {
		mainBlock.Add(&ir.RetInst{Value: &ir.IntConst{
			Typ:   ir.I32,
			Value: ir.Unsigned(0),
		}})
	}

	return m
}
