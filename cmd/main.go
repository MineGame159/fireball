package main

import (
	"fireball/core"
	"fireball/core/ast"
	"fireball/core/checker"
	"fireball/core/codegen"
	"fireball/core/parser"
	"fireball/core/scanner"
	"fmt"
	"os"
)

func main() {
	b, _ := os.ReadFile("test.fb")
	text := string(b)

	reporter := &consoleReporter{}

	decls := parser.Parse(reporter, scanner.NewScanner(text))
	checker.Check(reporter, decls)

	if reporter.hadError {
		os.Exit(1)
	}

	for _, decl := range decls {
		fmt.Println()
		ast.Print(decl, os.Stdout)
		fmt.Println()
	}

	_ = os.Mkdir("build", 0750)
	file, _ := os.Create("build/test.ll")

	codegen.Emit(decls, file)
}

type consoleReporter struct {
	hadError bool
}

func (c *consoleReporter) Report(error core.Error) {
	fmt.Println(error.Error())
	c.hadError = true
}
