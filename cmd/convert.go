package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/RudinMaxim/local_conversion/internal/converter"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert files from one format to another",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			<-sigCh
			fmt.Println("\nReceived interrupt signal. Gracefully shutting down...")
			cancel()
		}()

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

		opts := converter.ConversionOptions{
			SourceDir:    viper.GetString("sourceDir"),
			TargetDir:    viper.GetString("targetDir"),
			SourceFormat: sourceFormat,
			TargetFormat: targetFormat,
			Width:        viper.GetInt("width"),
			Height:       viper.GetInt("height"),
			NumWorkers:   viper.GetInt("numWorkers"),
			Quality:      viper.GetInt("quality"),
			SkipExisting: viper.GetBool("skipExisting"),
			ErrorCallback: func(file string, err error) {
				log.Printf("Error processing %s: %v", file, err)
			},
		}

		if err := validateDirectories(opts.SourceDir, opts.TargetDir); err != nil {
			return err
		}

		fmt.Printf("\nStarting conversion with the following settings:\n")
		fmt.Printf("Source format: %s\n", sourceFormat)
		fmt.Printf("Target format: %s\n", targetFormat)
		fmt.Printf("Source directory: %s\n", opts.SourceDir)
		fmt.Printf("Target directory: %s\n", opts.TargetDir)
		fmt.Printf("Workers: %d\n", opts.NumWorkers)
		if opts.Width > 0 || opts.Height > 0 {
			fmt.Printf("Resize to: %dx%d\n", opts.Width, opts.Height)
		}
		fmt.Printf("Quality: %d%%\n", opts.Quality)
		fmt.Printf("Skip existing: %v\n\n", opts.SkipExisting)

		if err := converter.ConvertImages(ctx, opts); err != nil {
			return fmt.Errorf("conversion failed: %v", err)
		}

		return nil
	},
}

func validateDirectories(sourceDir, targetDir string) error {
	sourceStat, err := os.Stat(sourceDir)
	if err != nil {
		return fmt.Errorf("source directory error: %w", err)
	}
	if !sourceStat.IsDir() {
		return fmt.Errorf("source path is not a directory")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().Int("quality", 80, "JPEG quality (1-100)")
	convertCmd.Flags().Bool("skip-existing", false, "Skip existing files")

	viper.BindPFlag("quality", convertCmd.Flags().Lookup("quality"))
	viper.BindPFlag("skipExisting", convertCmd.Flags().Lookup("skip-existing"))
}
