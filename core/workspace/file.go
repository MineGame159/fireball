package workspace

import (
	"fireball/core/ast"
	"fireball/core/ast/cst2ast"
	"fireball/core/checker"
	"fireball/core/cst"
	"fireball/core/typeresolver"
	"fireball/core/utils"
	"path/filepath"
	"sync"
)

type File struct {
	Project *Project
	Path    string

	Text string

	Cst cst.Node
	Ast *ast.File

	Data any

	parseWaitGroup sync.WaitGroup
	checkWaitGroup sync.WaitGroup

	diagnostics          []utils.Diagnostic
	parseDiagnosticCount int
}

func (f *File) AbsolutePath() string {
	return filepath.Join(f.Project.Path, f.Path)
}

func (f *File) SetText(text string, parse bool) {
	f.Text = text

	if parse {
		// Update sync groups
		f.parseWaitGroup.Add(1)

		for _, file := range f.Project.Files {
			file.checkWaitGroup.Add(1)
		}

		// Parse
		f.diagnostics = nil
		f.Project.removeFileFromNamespace(f)

		f.Cst = cst.Parse(f, text)
		f.Ast = cst2ast.Convert(f, f.AbsolutePath(), f.Cst)
		f.Ast.Resolver = f.Project.GetResolverFile(f.Ast)

		f.Project.addFileToNamespace(f)

		f.parseDiagnosticCount = len(f.diagnostics)
		f.parseWaitGroup.Done()

		// Resolve types
		for _, file := range f.Project.Files {
			if resolver := f.Project.getNamespace(file.Ast); resolver != nil {
				file.diagnostics = file.diagnostics[:file.parseDiagnosticCount]
				typeresolver.Resolve(file, resolver, file.Ast)
			}
		}

		// Specialize types
		for _, file := range f.Project.Files {
			typeresolver.Specialize(file, file.Ast)
		}

		// Check
		for _, file := range f.Project.Files {
			if resolver := f.Project.getNamespace(file.Ast); resolver != nil {
				checker.Check(file, resolver, file.Ast)
			}

			file.checkWaitGroup.Done()
		}
	}
}

func (f *File) EnsureParsed() {
	f.parseWaitGroup.Wait()
}

func (f *File) EnsureChecked() {
	f.checkWaitGroup.Wait()
}

func (f *File) Report(diag utils.Diagnostic) {
	f.diagnostics = append(f.diagnostics, diag)
}

func (f *File) Diagnostics() []utils.Diagnostic {
	return f.diagnostics
}
