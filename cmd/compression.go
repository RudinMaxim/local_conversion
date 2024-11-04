package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/RudinMaxim/local_conversion/internal/converter"
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

	viper.BindPFlag("compression.quality", compressCmd.Flags().Lookup("quality"))
	viper.BindPFlag("compression.width", compressCmd.Flags().Lookup("width"))
	viper.BindPFlag("compression.height", compressCmd.Flags().Lookup("height"))
	viper.BindPFlag("compression.format", compressCmd.Flags().Lookup("format"))
	viper.BindPFlag("compression.skip-existing", compressCmd.Flags().Lookup("skip-existing"))
}

func runCompress(cmd *cobra.Command, args []string) error {
	// Получаем параметры из командной строки или конфига
	quality := viper.GetInt("compression.quality")
	width := viper.GetInt("compression.width")
	height := viper.GetInt("compression.height")
	format := viper.GetString("compression.format")
	skipExisting := viper.GetBool("compression.skip-existing")

	// Если параметры не указаны, запрашиваем их интерактивно
	if !cmd.Flags().Changed("quality") {
		quality = promptQuality()
	}

	if !cmd.Flags().Changed("format") {
		format = promptFormat()
	}

	opts := converter.ConversionOptions{
		SourceDir:    viper.GetString("sourceDir"),
		TargetDir:    viper.GetString("targetDir"),
		SourceFormat: "auto",
		TargetFormat: format,
		Width:        width,
		Height:       height,
		NumWorkers:   viper.GetInt("numWorkers"),
		Quality:      quality,
		SkipExisting: skipExisting,
		ErrorCallback: func(file string, err error) {
			fmt.Printf("Error processing %s: %v\n", filepath.Base(file), err)
		},
	}

	ctx := context.Background()
	if err := converter.ConvertImages(ctx, opts); err != nil {
		return fmt.Errorf("compression failed: %v", err)
	}

	return nil
}

func promptQuality() int {
	prompt := promptui.Prompt{
		Label:   "Enter compression quality (1-100)",
		Default: "80",
		Validate: func(input string) error {
			var quality int
			if _, err := fmt.Sscanf(input, "%d", &quality); err != nil {
				return fmt.Errorf("invalid number")
			}
			if quality < 1 || quality > 100 {
				return fmt.Errorf("quality must be between 1 and 100")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
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
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . | cyan }}",
			Active:   "\U0001F449 {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "\U0001F44D {{ . | green }}",
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "auto"
	}

	return result
}
