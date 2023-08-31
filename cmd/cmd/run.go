package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

func GetRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run file",
		Short: "Runs a single source file",
		Args:  cobra.ExactArgs(1),
		Run:   runCmd,
	}
}

func runCmd(_ *cobra.Command, args []string) {
	// Build
	output := buildExecutable(args[0])

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
