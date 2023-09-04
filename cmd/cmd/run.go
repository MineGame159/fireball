package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

func GetRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run project.",
		Run:   runCmd,
	}
}

func runCmd(_ *cobra.Command, args []string) {
	// Build
	output := buildProject()

	// Run
	cmd := exec.Command(output)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(4)
	}
}
