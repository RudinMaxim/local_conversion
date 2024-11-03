package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/RudinMaxim/local_conversion/internal/converter"
	"github.com/h2non/filetype"
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
			if len(args) == 0 {
				return fmt.Errorf("no input file provided for format detection")
			}

			sourceFormat, err = detectFileFormat(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Detected source format: %s\n", sourceFormat)
		} else {
			sourceFormat = formats[sourceIdx-1]
		}

		var targetFormats []string
		for _, f := range formats {
			if f != sourceFormat {
				targetFormats = append(targetFormats, f)
			}
		}

		targetPrompt := promptui.Select{
			Label: "Select target format",
			Items: targetFormats,
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

func detectFileFormat(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 261)
	if _, err := file.Read(buf); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	kind, err := filetype.Match(buf)
	if err != nil {
		return "", fmt.Errorf("failed to match file type: %w", err)
	}

	if kind == filetype.Unknown {
		return "", fmt.Errorf("unknown file format")
	}

	return kind.Extension, nil
}

func init() {
	rootCmd.AddCommand(convertCmd)
}
