package main

import (
	"fireball/cmd/cmd"
	"fireball/cmd/lsp"
	"github.com/spf13/cobra"
	"log"
)

func main() {
	root := &cobra.Command{
		Use:     "fireball",
		Short:   "Tooling for the Fireball programming language",
		Version: "0.1.0",
	}

	root.AddCommand(
		cmd.GetBuildCmd(),
		cmd.GetRunCmd(),
		lsp.GetCmd(),
	)

	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}
