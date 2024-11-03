package cmd

import (
	"fmt"
	"log"

	"github.com/RudinMaxim/local_conversion/internal/converter"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert files from one format to another",
	RunE: func(cmd *cobra.Command, args []string) error {
		formats := viper.GetStringSlice("formats")

		sourceFormats := append([]string{"Auto"}, formats...)

		sourcePrompt := promptui.Select{
			Label: "Select source format",
			Items: sourceFormats,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . | cyan }}",
				Active:   "\U0001F449 {{ . | cyan }}",
				Inactive: "  {{ . }}",
				Selected: "\U0001F44D {{ . | green }}",
			},
		}

		sourceIdx, _, err := sourcePrompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %v", err)
		}

		var sourceFormat string
		if sourceIdx == 0 {
			sourceFormat = "auto"
		} else {
			sourceFormat = formats[sourceIdx-1]
		}

		targetPrompt := promptui.Select{
			Label: "Select target format",
			Items: formats,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . | cyan }}",
				Active:   "\U0001F449 {{ . | cyan }}",
				Inactive: "  {{ . }}",
				Selected: "\U0001F44D {{ . | green }}",
			},
		}

		_, targetFormat, err := targetPrompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed: %v", err)
		}

		sourceDir := viper.GetString("sourceDir")
		targetDir := viper.GetString("targetDir")
		numWorkers := viper.GetInt("numWorkers")
		width := viper.GetInt("width")
		height := viper.GetInt("height")

		fmt.Printf("Starting conversion from %s to %s with %d workers...\n", sourceFormat, targetFormat, numWorkers)

		if err := converter.ConvertImages(sourceDir, targetDir, sourceFormat, targetFormat, width, height, numWorkers); err != nil {
			log.Fatalf("Failed to convert images: %v", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
