package workspace

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/checker"
	"fireball/core/parser"
	"fireball/core/scanner"
	"fireball/core/types"
	"fireball/core/utils"
	"fmt"
	"math"
	"sync"
)

type File struct {
	Project *Project
	Path    string

	Text  string
	Decls []ast.Decl

	Types     map[string]types.Type
	Functions map[string]*types.FunctionType

	Data any

	parseWaitGroup sync.WaitGroup
	checkWaitGroup sync.WaitGroup

	diagnostics []utils.Diagnostic
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
		f.Decls = parser.Parse(f, scanner.NewScanner(text))
		f.CollectTypesAndFunctions()
		f.parseWaitGroup.Done()

		// Check
		for _, file := range f.Project.Files {
			checker.Check(file, file.Project, file.Decls)
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
	typeMap := make(map[string]types.Type)
	functionMap := make(map[string]*types.FunctionType)

	for _, decl := range f.Decls {
		if v, ok := decl.(*ast.Struct); ok {
			// Struct
			fields := make([]types.Field, len(v.Fields))

			for i, field := range v.Fields {
				fields[i] = types.Field{
					Name: field.Name.Lexeme,
					Type: field.Type,
				}
			}

			if _, ok := typeMap[v.Name.Lexeme]; ok {
				f.Report(utils.Diagnostic{
					Kind:    utils.ErrorKind,
					Range:   core.TokenToRange(v.Name),
					Message: fmt.Sprintf("Type with the name '%s' aleady exists.", v.Name),
				})
			} else {
				typeMap[v.Name.Lexeme] = types.Struct(v.Name.Lexeme, fields, v.Range())
			}
		} else if v, ok := decl.(*ast.Enum); ok {
			// Enum
			if v.Type == nil {
				minValue := math.MaxInt
				maxValue := math.MinInt

				for _, case_ := range v.Cases {
					minValue = min(minValue, case_.Value)
					maxValue = max(maxValue, case_.Value)
				}

				var kind types.PrimitiveKind

				if minValue >= 0 {
					// Unsigned
					if maxValue <= math.MaxUint8 {
						kind = types.U8
					} else if maxValue <= math.MaxUint16 {
						kind = types.U16
					} else if maxValue <= math.MaxUint32 {
						kind = types.U32
					} else {
						kind = types.U64
					}
				} else {
					// Signed
					if minValue >= math.MinInt8 && maxValue <= math.MaxInt8 {
						kind = types.I8
					} else if minValue >= math.MinInt16 && maxValue <= math.MaxInt16 {
						kind = types.I16
					} else if minValue >= math.MinInt32 && maxValue <= math.MaxInt32 {
						kind = types.I32
					} else {
						kind = types.I64
					}
				}

				v.Type = types.Primitive(kind, core.Range{})
			}

			cases := make([]types.EnumCase, len(v.Cases))

			for i, case_ := range v.Cases {
				cases[i] = types.EnumCase{
					Name:  case_.Name.Lexeme,
					Value: case_.Value,
				}
			}

			if _, ok := typeMap[v.Name.Lexeme]; ok {
				f.Report(utils.Diagnostic{
					Kind:    utils.ErrorKind,
					Range:   core.TokenToRange(v.Name),
					Message: fmt.Sprintf("Type with the name '%s' aleady exists.", v.Name),
				})
			} else {
				typeMap[v.Name.Lexeme] = types.Enum(v.Name.Lexeme, v.Type, cases, v.Range())
			}
		} else if v, ok := decl.(*ast.Func); ok {
			// Function
			params := make([]types.Type, len(v.Params))

			for i, param := range v.Params {
				params[i] = param.Type
			}

			if _, ok := functionMap[v.Name.Lexeme]; ok {
				f.Report(utils.Diagnostic{
					Kind:    utils.ErrorKind,
					Range:   core.TokenToRange(v.Name),
					Message: fmt.Sprintf("Function with the name '%s' already exists.", v.Name),
				})
			} else {
				functionMap[v.Name.Lexeme] = types.Function(params, v.Variadic, v.Returns, v.Range())
			}
		}
	}

	f.Types = typeMap
	f.Functions = functionMap
}

func (f *File) Report(diag utils.Diagnostic) {
	f.diagnostics = append(f.diagnostics, diag)
}

func (f *File) FlushDiagnostics() []utils.Diagnostic {
	diagnostics := f.diagnostics
	f.diagnostics = make([]utils.Diagnostic, 0)
	return diagnostics
}
