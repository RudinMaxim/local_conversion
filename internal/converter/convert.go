package converter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
)

type ConversionOptions struct {
	SourceDir     string
	TargetDir     string
	SourceFormat  string
	TargetFormat  string
	Width         int
	Height        int
	NumWorkers    int
	Quality       int
	SkipExisting  bool
	ErrorCallback func(string, error)
}

var (
	ErrInvalidFormat    = errors.New("invalid format specified")
	ErrInvalidDimension = errors.New("invalid dimensions specified")
	ErrInvalidWorkers   = errors.New("invalid number of workers")
)

func ConvertImages(ctx context.Context, opts ConversionOptions) error {
	if err := validateOptions(&opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	files, err := getSourceFiles(opts.SourceDir, opts.SourceFormat)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no matching files found in source directory")
	}

	start := time.Now()
	bar := pb.StartNew(len(files))

	errChan := make(chan error, len(files))

	var wg sync.WaitGroup
	fileCh := make(chan string, opts.NumWorkers)

	for i := 0; i < opts.NumWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileCh {
				select {
				case <-ctx.Done():
					return
				default:
					if err := processFileWithRetry(ctx, file, &opts); err != nil {
						if opts.ErrorCallback != nil {
							opts.ErrorCallback(file, err)
						}
						errChan <- fmt.Errorf("error processing %s: %w", file, err)
					}
					bar.Increment()
				}
			}
		}()
	}

	go func() {
		for _, file := range files {
			select {
			case <-ctx.Done():
				close(fileCh)
				return
			case fileCh <- file:
			}
		}
		close(fileCh)
	}()

	wg.Wait()
	bar.Finish()

	close(errChan)
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during conversion: %v", len(errors), errors[0])
	}

	duration := time.Since(start)
	fmt.Printf("Successfully converted %d images in %v\n", len(files)-len(errors), duration)

	return nil
}

func validateOptions(opts *ConversionOptions) error {
	if opts.NumWorkers < 1 {
		return ErrInvalidWorkers
	}
	if opts.Width < 0 || opts.Height < 0 {
		return ErrInvalidDimension
	}
	if opts.Quality < 0 || opts.Quality > 100 {
		opts.Quality = 80
	}
	return nil
}

func getSourceFiles(sourceDir, sourceFormat string) ([]string, error) {
	var pattern string
	if sourceFormat == "auto" {
		pattern = "*.*"
	} else {
		pattern = "*." + sourceFormat
	}

	files, err := filepath.Glob(filepath.Join(sourceDir, pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to read source directory: %w", err)
	}
	return files, nil
}

func processFileWithRetry(ctx context.Context, filePath string, opts *ConversionOptions) error {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := processImage(filePath, opts); err != nil {
				lastErr = err
				time.Sleep(time.Duration(attempt*100) * time.Millisecond)
				continue
			}
			return nil
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func processImage(filePath string, opts *ConversionOptions) error {
	outputFile := filepath.Join(opts.TargetDir,
		strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))+
			"."+opts.TargetFormat)

	if opts.SkipExisting {
		if _, err := os.Stat(outputFile); err == nil {
			return nil
		}
	}

	if opts.SourceFormat == "auto" {
		format, err := DetectFileFormat(filePath)
		if err != nil {
			return fmt.Errorf("format detection failed: %w", err)
		}
		if format == opts.TargetFormat {
			return nil
		}
	}

	img, err := imaging.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	if opts.Width > 0 || opts.Height > 0 {
		img = imaging.Resize(img, opts.Width, opts.Height, imaging.Lanczos)
	}

	switch opts.TargetFormat {
	case "jpg", "jpeg":
		err = imaging.Save(img, outputFile, imaging.JPEGQuality(opts.Quality))
	case "png":
		err = imaging.Save(img, outputFile, imaging.PNGCompressionLevel(9))
	case "gif", "bmp":
		err = imaging.Save(img, outputFile)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidFormat, opts.TargetFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}

func DetectFileFormat(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 261)
	if _, err := file.Read(buf); err != nil {
		return "", fmt.Errorf("failed to read file header: %w", err)
	}

	kind, err := filetype.Match(buf)
	if err != nil || kind == filetype.Unknown {
		return "", fmt.Errorf("unknown or unsupported file format")
	}

	return kind.Extension, nil
}
