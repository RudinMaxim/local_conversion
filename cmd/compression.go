package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/RudinMaxim/local_conversion/internal/compression"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var compressCmd = &cobra.Command{
	Use:   "compress",
	Short: "Compress images with specified parameters",
	RunE:  runCompress,
}

func init() {
	rootCmd.AddCommand(compressCmd)

	compressCmd.Flags().IntP("quality", "q", 80, "Compression quality (1-100)")
	compressCmd.Flags().IntP("width", "w", 0, "Maximum width (0 for original)")
	compressCmd.Flags().IntP("height", "h", 0, "Maximum height (0 for original)")
	compressCmd.Flags().StringP("format", "f", "auto", "Target format (jpg, png, gif, bmp, or auto)")
	compressCmd.Flags().BoolP("skip-existing", "s", false, "Skip existing files")

	// Bind flags to configuration keys
	_ = viper.BindPFlag("compression.quality", compressCmd.Flags().Lookup("quality"))
	_ = viper.BindPFlag("compression.width", compressCmd.Flags().Lookup("width"))
	_ = viper.BindPFlag("compression.height", compressCmd.Flags().Lookup("height"))
	_ = viper.BindPFlag("compression.format", compressCmd.Flags().Lookup("format"))
	_ = viper.BindPFlag("compression.skip-existing", compressCmd.Flags().Lookup("skip-existing"))
}

func runCompress(cmd *cobra.Command, args []string) error {
	// Read configuration parameters
	opts, err := buildCompressionOptions(cmd)
	if err != nil {
		return fmt.Errorf("invalid options: %v", err)
	}

	ctx := context.Background()

	// Execute compression with provided options
	if err := compression.CompressionImages(ctx, opts); err != nil {
		return fmt.Errorf("compression failed: %v", err)
	}

	log.Println("Compression completed successfully.")
	return nil
}

func buildCompressionOptions(cmd *cobra.Command) (compression.CompressionOptions, error) {
	quality := viper.GetInt("compression.quality")
	if !isValidQuality(quality) {
		return compression.CompressionOptions{}, errors.New("quality must be between 1 and 100")
	}

	width := viper.GetInt("compression.width")
	height := viper.GetInt("compression.height")

	format := viper.GetString("compression.format")
	if !isValidFormat(format) {
		return compression.CompressionOptions{}, errors.New("unsupported format")
	}

	// Interactive prompts if not specified
	if !cmd.Flags().Changed("quality") {
		quality = promptQuality()
	}

	if !cmd.Flags().Changed("format") {
		format = promptFormat()
	}

	opts := compression.CompressionOptions{
		SourceDir:    viper.GetString("sourceDir"),
		TargetDir:    viper.GetString("targetDir"),
		SourceFormat: "auto",
		TargetFormat: format,
		Width:        width,
		Height:       height,
		NumWorkers:   viper.GetInt("numWorkers"),
		Quality:      quality,
		SkipExisting: viper.GetBool("compression.skip-existing"),
		ErrorCallback: func(file string, err error) {
			log.Printf("Error processing file %s: %v\n", filepath.Base(file), err)
		},
	}

	return opts, nil
}

func promptQuality() int {
	prompt := promptui.Prompt{
		Label:   "Enter compression quality (1-100)",
		Default: "80",
		Validate: func(input string) error {
			var quality int
			if _, err := fmt.Sscanf(input, "%d", &quality); err != nil || !isValidQuality(quality) {
				return errors.New("please enter a valid quality (1-100)")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		log.Println("Error during quality prompt, using default: 80")
		return 80
	}

	var quality int
	fmt.Sscanf(result, "%d", &quality)
	return quality
}

func promptFormat() string {
	prompt := promptui.Select{
		Label: "Select target format",
		Items: []string{"auto", "jpg", "png", "gif", "bmp"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Println("Error during format prompt, using default: auto")
		return "auto"
	}

	return result
}

// Helpers for validation
func isValidQuality(quality int) bool {
	return quality >= 1 && quality <= 100
}

func isValidFormat(format string) bool {
	supportedFormats := []string{"auto", "jpg", "png", "gif", "bmp"}
	for _, f := range supportedFormats {
		if format == f {
			return true
		}
	}
	return false
}
