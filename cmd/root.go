package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "local_conversion",
	Short: "Welcome to the File Converter CLI!",
	Long:  `A CLI tool for converting large volumes of files with multi-threading support.`,
	Run: func(cmd *cobra.Command, args []string) {
		prompt := promptui.Select{
			Label: "Choose an option",
			Items: []string{"Convert files", "Configuration", "Exit"},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . | cyan }}",
				Active:   "\U0001F449 {{ . | cyan }}",
				Inactive: "  {{ . }}",
				Selected: "\U0001F44D {{ . | green }}",
			},
		}

		for {
			idx, _, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed: %v\n", err)
				return
			}

			switch idx {
			case 0:
				clearConsole()
				if err := convertCmd.RunE(cmd, args); err != nil {
					fmt.Printf("Conversion failed: %v\n", err)
				}
			case 1:
				clearConsole()
				configurationMenu()
			case 3:
				fmt.Println("Exiting...")
				os.Exit(0)
			}
		}

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func clearConsole() {
	var cmd *exec.Cmd
	if os.Getenv("OS") == "Windows_NT" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to clear console: %v\n", err)
	}
}

func getNumCPUs() int {
	return runtime.NumCPU()
}
