package build

import (
	"fireball/core/utils"
	"fireball/core/workspace"
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
)

func Report(project *workspace.Project) {
	reporter := consoleReporter{
		error:   color.New(color.FgRed),
		warning: color.New(color.FgYellow),
	}

	for _, file := range project.Files {
		for _, diagnostic := range file.Diagnostics() {
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

	if reporter.hadDiagnostic {
		fmt.Println()
	}
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
