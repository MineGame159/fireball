package build

import (
	"fireball/core/codegen"
	"fireball/core/ir"
	"fireball/core/llvm"
	"fireball/core/workspace"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Build(project *workspace.Project, entrypoint *ir.Module, optimizationLevel uint8, outputName string) (string, error) {
	_ = os.Mkdir("build", 0750)

	// Emit project IR
	ctx := codegen.Context{}
	irPaths := make([]string, 0, len(project.Files))

	for _, file := range project.Files {
		path := strings.ReplaceAll(file.Path, string(os.PathSeparator), "-")
		path = filepath.Join(project.Path, "build", path[:len(path)-3]+".ll")

		irFile, err := os.Create(path)
		if err != nil {
			return "", err
		}

		module := codegen.Emit(&ctx, file.AbsolutePath(), project.GetResolverFile(file.Ast), file.Ast)
		llvm.WriteText(module, irFile)

		_ = irFile.Close()

		irPaths = append(irPaths, path)
	}

	// Emit entrypoint IR
	entrypointPath := filepath.Join(project.Path, "build", "__entrypoint.ll")
	irPaths = append(irPaths, entrypointPath)

	entrypointFile, err := os.Create(entrypointPath)
	if err != nil {
		return "", err
	}

	llvm.WriteText(entrypoint, entrypointFile)
	_ = entrypointFile.Close()

	// Compile
	c := Compiler{
		OptimizationLevel: min(max(int(optimizationLevel), 0), 3),
	}

	for _, irPath := range irPaths {
		c.AddInput(irPath)
	}

	for _, library := range project.Config.LinkLibraries {
		c.AddLibrary(library)
	}

	output := filepath.Join(project.Path, "build", outputName)
	output = withExecutableExtension(output)

	err = c.Compile(output)
	if err != nil {
		return "", err
	}

	return output, nil
}

func withExecutableExtension(path string) string {
	if runtime.GOOS == "windows" {
		return path + ".exe"
	}

	return path
}
