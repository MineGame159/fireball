package main

import (
	"fireball/cmd/lsp"
	"fireball/core"
	"fireball/core/checker"
	"fireball/core/codegen"
	"fireball/core/parser"
	"fireball/core/scanner"
	"fireball/core/typeresolver"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Invalid sub-command, use either build or lsp")
	}

	switch os.Args[1] {
	case "build":
		if len(os.Args) < 3 {
			log.Fatalln("Invalid path")
		}

		build(os.Args[2])

	case "lsp":
		port := uint16(0)

		if len(os.Args) > 2 && strings.HasPrefix(os.Args[2], "-p=") {
			num, err := strconv.Atoi(os.Args[2][3:])

			if err == nil {
				port = uint16(num)
			}
		}

		lsp.Start(port)

	default:
		log.Fatalln("Invalid sub-command, use either build or lsp")
	}
}

func build(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("Invalid file")
	}

	text := string(b)

	reporter := &consoleReporter{}

	decls := parser.Parse(reporter, scanner.NewScanner(text))
	typeresolver.Resolve(reporter, decls)
	checker.Check(reporter, decls)

	if reporter.hadError {
		os.Exit(1)
	}

	/*for _, decl := range decls {
		fmt.Println()
		ast.Print(decl, os.Stdout)
		fmt.Println()
	}*/

	_ = os.Mkdir("build", 0750)
	file, _ := os.Create("build/test.ll")

	codegen.Emit(decls, file)
}

type consoleReporter struct {
	hadError bool
}

func (c *consoleReporter) Report(error core.Diagnostic) {
	fmt.Println(error)
	c.hadError = true
}
