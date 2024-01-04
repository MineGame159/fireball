package workspace

import (
	"fireball/core/ast"
	"fireball/core/ast/cst2ast"
	"fireball/core/checker"
	"fireball/core/cst"
	"fireball/core/typeresolver"
	"fireball/core/utils"
	"fmt"
	"math"
	"path/filepath"
	"sync"
)

type File struct {
	Project *Project
	Path    string

	Text string

	Cst cst.Node
	Ast *ast.File

	Types     map[string]ast.Type
	Functions map[string]*ast.Func

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

		f.Cst = cst.Parse(f, text)
		f.Ast = cst2ast.Convert(f, f.AbsolutePath(), f.Cst)

		f.parseDiagnosticCount = len(f.diagnostics)

		f.CollectTypesAndFunctions()
		f.parseWaitGroup.Done()

		// Check
		for _, file := range f.Project.Files {
			file.diagnostics = file.diagnostics[:file.parseDiagnosticCount]
			typeresolver.Resolve(file, file.Project, file.Ast)
		}

		for _, file := range f.Project.Files {
			checker.Check(file, file.Project, file.Ast)
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

func (f *File) CollectTypesAndFunctions() {
	typeMap := make(map[string]ast.Type)
	functionMap := make(map[string]*ast.Func)

	for _, decl := range f.Ast.Decls {
		if struct_, ok := decl.(*ast.Struct); ok {
			// Struct
			if struct_.Name == nil {
				continue
			}

			if _, ok := typeMap[struct_.Name.String()]; ok {
				f.Report(utils.Diagnostic{
					Kind:    utils.ErrorKind,
					Range:   struct_.Name.Cst().Range,
					Message: fmt.Sprintf("Type with the name '%s' aleady exists.", struct_.Name),
				})
			} else {
				typeMap[struct_.Name.String()] = struct_
			}
		} else if enum, ok := decl.(*ast.Enum); ok {
			// Enum
			if enum.Name == nil {
				continue
			}

			if enum.Type == nil {
				minValue := int64(math.MaxInt64)
				maxValue := int64(math.MinInt64)

				for _, case_ := range enum.Cases {
					minValue = min(minValue, case_.ActualValue)
					maxValue = max(maxValue, case_.ActualValue)
				}

				var kind ast.PrimitiveKind

				if minValue >= 0 {
					// Unsigned
					if maxValue <= math.MaxUint8 {
						kind = ast.U8
					} else if maxValue <= math.MaxUint16 {
						kind = ast.U16
					} else if maxValue <= math.MaxUint32 {
						kind = ast.U32
					} else {
						kind = ast.U64
					}
				} else {
					// Signed
					if minValue >= math.MinInt8 && maxValue <= math.MaxInt8 {
						kind = ast.I8
					} else if minValue >= math.MinInt16 && maxValue <= math.MaxInt16 {
						kind = ast.I16
					} else if minValue >= math.MinInt32 && maxValue <= math.MaxInt32 {
						kind = ast.I32
					} else {
						kind = ast.I64
					}
				}

				enum.ActualType = &ast.Primitive{Kind: kind}
			} else {
				enum.ActualType = enum.Type
			}

			if _, ok := typeMap[enum.Name.String()]; ok {
				f.Report(utils.Diagnostic{
					Kind:    utils.ErrorKind,
					Range:   enum.Name.Cst().Range,
					Message: fmt.Sprintf("Type with the name '%s' aleady exists.", enum.Name),
				})
			} else {
				typeMap[enum.Name.String()] = enum
			}
		} else if function, ok := decl.(*ast.Func); ok {
			// Function
			if function.Name == nil {
				continue
			}

			if _, ok := functionMap[function.Name.String()]; ok {
				f.Report(utils.Diagnostic{
					Kind:    utils.ErrorKind,
					Range:   function.Name.Cst().Range,
					Message: fmt.Sprintf("Function with the name '%s' already exists.", function.Name),
				})
			} else {
				functionMap[function.Name.String()] = function
			}
		}
	}

	f.Types = typeMap
	f.Functions = functionMap
}

func (f *File) Report(diag utils.Diagnostic) {
	f.diagnostics = append(f.diagnostics, diag)
}

func (f *File) Diagnostics() []utils.Diagnostic {
	return f.diagnostics
}
