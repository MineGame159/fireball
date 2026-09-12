package comptime

import (
	"errors"
	"fireball/abi"
	"fireball/ast"
	"fireball/codegen"
	"fireball/core"
	"fireball/fb-core"
	"fireball/ir"
	"fireball/ir/eval"
	"fireball/lexer"
	"fireball/sema"
	"fireball/types"
	"fmt"
)

func Evaluate(file *ast.File, instantiations *types.InstantiationCache, typeEnv *sema.TypeEnvironment, fileDataMap map[*ast.File]codegen.FileData, builtins fb_core.Builtins) (map[ast.Expr]eval.Value, []core.Diagnostic) {
	defer core.Scope()()

	evaluations := make(map[ast.Expr]eval.Value)
	var diagnostics []core.Diagnostic

	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.Const:
			// Check
			if !fileDataMap[file].ExprInfos[decl.Value].CompTime {
				continue
			}

			seen := make(map[ast.Expr]any)
			seen[decl.Value] = nil

			prevLen := len(diagnostics)
			diagnostics = checkCompTimeValue(fileDataMap, decl.Value, seen, decl.Name(), diagnostics)

			if len(diagnostics) > prevLen {
				continue
			}

			// Get literal value directly
			value := getLiteralEvalValue(decl.Value, builtins)

			// Interpret IR
			if value == nil {
				// Generate IR module
				module := ir.NewModule()
				module.Path = "__comptime__"

				c := codegen.New(module, file, abi.AMD64, abi.SystemV, instantiations, typeEnv, fileDataMap, builtins, true, false)
				fun := generateModule(c, decl)

				// Evaluate IR module
				e := eval.NewEvaluator()
				e.Load(module)

				reg, err := e.Run(fun)

				if err != nil {
					if err, ok := errors.AsType[eval.Error](err); ok && err.File != "" {
						diagnostics = append(diagnostics, getDiagnostic(err))
						continue
					}

					panic("comptime.Evaluate() - " + err.Error())
				}

				value = e.GetValue(reg, c.Types.Get(c.NodeType(decl.Type)))
			}

			evaluations[decl.Value] = value
		}
	}

	return evaluations, diagnostics
}

func checkCompTimeValue(fileDataMap map[*ast.File]codegen.FileData, expr ast.Expr, seen map[ast.Expr]any, errNode ast.Node, diagnostics []core.Diagnostic) []core.Diagnostic {
	switch expr := expr.(type) {
	case *ast.Identifier:
		switch node := fileDataMap[ast.GetFile(expr)].ExprInfos[expr].Node.(type) {
		case *ast.Const:
			if _, ok := seen[node.Value]; ok {
				diagnostics = append(diagnostics, core.Diagnostic{
					Kind:    core.Error,
					Path:    ast.GetFile(expr).Path,
					Range:   errNode.Range(),
					Message: fmt.Sprintf("cyclic reference of constant '%s'", node.Name().Token.Text),
				})

				return diagnostics
			}

			seen[node.Value] = nil
			diagnostics = checkCompTimeValue(fileDataMap, node.Value, seen, errNode, diagnostics)
			delete(seen, node.Value)
		}

	default:
		for child := range expr.Children() {
			if child, ok := child.(ast.Expr); ok {
				diagnostics = checkCompTimeValue(fileDataMap, child, seen, errNode, diagnostics)
			}
		}
	}

	return diagnostics
}

func getLiteralEvalValue(expr ast.Expr, builtins fb_core.Builtins) eval.Value {
	switch expr := expr.(type) {
	case *ast.Bool:
		if expr.Value {
			return &eval.IntValue{Value: 1}
		}

		return &eval.IntValue{Value: 0}

	case *ast.Number:
		// Integer
		if lexer.IsInteger(expr.Token.Kind) {
			return &eval.IntValue{Value: lexer.ParseInteger(expr.Token).TwosComplement()}
		}

		// Float
		if expr.Token.Kind == lexer.Decimal32bit {
			value, err := lexer.ParseDecimal(expr.Token)
			if err != nil {
				panic("comptime.getLiteralEvalValue() - Failed to parse float '" + expr.Token.Text + "'")
			}

			return &eval.FloatValue{Value: value}
		}

		// Double
		if expr.Token.Kind == lexer.Decimal {
			value, err := lexer.ParseDecimal(expr.Token)
			if err != nil {
				panic("comptime.getLiteralEvalValue() - Failed to parse double '" + expr.Token.Text + "'")
			}

			return &eval.FloatValue{Value: value}
		}

		// Unknown
		panic("comptime.getLiteralEvalValue() - Invalid token kind")

	case *ast.Character:
		return &eval.IntValue{Value: uint64(expr.Rune)}

	case *ast.String:
		fields := abi.AMD64.Info(builtins.StringView).Fields

		ptrI := 0
		if builtins.StringView.Fields[fields[0].Index].Name == "size" {
			ptrI = 1
		}

		init := ir.NewString(expr.Runes, true)

		values := make([]eval.Value, 2)

		values[ptrI] = &eval.StringValue{Value: init}
		values[1-ptrI] = &eval.IntValue{Value: uint64(init.Size)}

		return &eval.AggregateValue{Values: values}

	default:
		return nil
	}
}

func generateModule(c *codegen.Codegen, decl *ast.Const) *ir.Function {
	evalFunc := &ast.Func{
		Name_:   &ast.Leaf{Token: lexer.Token{Kind: lexer.Identifier, Text: "__comptime__"}},
		Returns: decl.Type,
	}

	evalFile := &ast.File{
		Mod: &ast.Mod{Path: []*ast.Leaf{
			{Token: lexer.Token{Kind: lexer.Identifier, Text: "__comptime__"}},
		}},
		Decls: []ast.Decl{evalFunc},
	}

	evalFunc.SetParent(evalFile)

	funTyp := &types.Func{Returns: c.NodeType(decl.Type)}

	fun := c.CreateFunction(
		evalFunc,
		funTyp,
		false,
		nil,
	)

	c.BeginFunc(evalFunc, funTyp, fun)

	val := c.LoadImplicitCast(decl.Value, c.UnderlyingNodeType(decl.Type))
	c.Emitter.Ret(val)

	c.EndFunc()

	return fun
}

func getDiagnostic(err eval.Error) core.Diagnostic {
	return core.Diagnostic{
		Kind: core.Error,
		Path: err.File,
		Range: core.Range{
			Start: core.Pos{
				Line:   err.Line,
				Column: err.Column,
			},
			End: core.Pos{
				Line:   err.Line,
				Column: err.Column + 1,
			},
		},
		Message: err.Msg,
	}
}
