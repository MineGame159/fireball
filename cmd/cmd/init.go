package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var projectToml = `
Name = "%s"
Src = "src"
`

var mainFb = `
func main() {
    printf("Hello, World!\n");
}

#[Extern]
func printf(format *u8, ...) void
`

func GetInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialises a template project.",
		Args:  cobra.MaximumNArgs(1),
		Run:   initCmd,
	}

	return cmd
}

func initCmd(_ *cobra.Command, args []string) {
	projectName := "example"

	// Create directory if specified
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v\n", err)
	}

	if len(args) == 1 {
		fullPath, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatalf("Error getting absolute path: %v\n", err)
		}

		if fullPath != workingDir {
			newDirName := args[0]

			// Check if directory already exists
			if _, err := os.Stat(newDirName); err == nil {
				log.Fatalf("Directory '%s' already exists.\n", newDirName)
			}

			// Create the new directory
			if err := os.Mkdir(newDirName, os.ModePerm); err != nil {
				log.Fatalf("Error creating directory: %v\n", err)
			}

			// Change directory
			if err := os.Chdir(newDirName); err != nil {
				log.Fatalf("Error changing directory: %v\n", err)
			}

			// Set project name
			projectName = filepath.Base(newDirName)
		}
	} else {
		projectName = filepath.Base(workingDir)
	}

	// Write project.toml
	if err = writeFile("project.toml", fmt.Sprintf(projectToml, projectName)); err != nil {
		log.Fatalf("Error writing project.toml: %v\n", err)
	}

	// Create src directory
	if err = os.Mkdir("src", os.ModePerm); err != nil {
		log.Fatalf("Error creating 'src' directory: %v\n", err)
	}

	// Write src/main.fireball
	if err = writeFile("src/main.fb", mainFb); err != nil {
		log.Fatalf("Error writing src/main.fb: %v\n", err)
	}

	// Write .gitignore
	if err = writeFile(".gitignore", "build"); err != nil {
		log.Fatalf("Error writing .gitignore: %v\n", err)
	}

	_, _ = color.New(color.FgGreen).Printf("Created project '%s'.\n", projectName)
}

func writeFile(name, content string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}

	if _, err = file.WriteString(content); err != nil {
		return err
	}

	return nil
}
