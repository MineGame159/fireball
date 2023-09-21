package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

func GetRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run project.",
		Run:   runCmd,
	}

	cmd.Flags().Uint8VarP(&opt, "opt", "O", 0, "Optimization level. [-O0, -O1, -O2, or -O3] (default = '-O0')")

	return cmd
}

func runCmd(_ *cobra.Command, _ []string) {
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
