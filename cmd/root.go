package cmd

import (
	"fmt"
	"os"

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
			Items: []string{"Convert files", "Exit"},
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . | cyan }}",
				Active:   "\U0001F449 {{ . | cyan }}",
				Inactive: "  {{ . }}",
				Selected: "\U0001F44D {{ . | green }}",
			},
		}

		idx, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed: %v\n", err)
			return
		}

		switch idx {
		case 0:
			if err := convertCmd.RunE(cmd, args); err != nil {
				fmt.Printf("Conversion failed: %v\n", err)
			}
		case 1:
			fmt.Println("Exiting...")
			os.Exit(0)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
