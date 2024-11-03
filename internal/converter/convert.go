package converter

import (
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

func ConvertImages(sourceDir, targetDir, sourceFormat, targetFormat string, width, height, numWorkers int) error {
	files, err := filepath.Glob(filepath.Join(sourceDir, "*.*"))
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Запуск таймера для измерения времени выполнения
	start := time.Now()

	// Создаем прогресс-бар
	bar := pb.StartNew(len(files))

	var wg sync.WaitGroup
	fileCh := make(chan string, len(files))
	for _, file := range files {
		fileCh <- file
	}
	close(fileCh)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileCh {
				var format string
				if sourceFormat == "auto" {
					var err error
					format, err = DetectFileFormat(file)
					if err != nil {
						fmt.Printf("Skipping file %s: %v\n", file, err)
						bar.Increment()
						continue
					}
				} else {
					format = sourceFormat
				}

				if format != targetFormat {
					if err := processImage(file, targetDir, targetFormat, width, height); err != nil {
						fmt.Printf("Error processing file %s: %v\n", file, err)
					}
				} else {
					fmt.Printf("Skipping file %s: source and target formats are the same\n", file)
				}
				bar.Increment()
			}
		}()
	}

	wg.Wait()
	bar.Finish()

	duration := time.Since(start)
	fmt.Printf("Conversion completed in %v\n", duration)

	return nil
}

func processImage(filePath, targetDir, format string, width, height int) error {
	img, err := imaging.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	img = imaging.Resize(img, width, height, imaging.Lanczos)

	outputFile := filepath.Join(targetDir, strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))+"."+format)
	switch format {
	case "jpg", "jpeg":
		err = imaging.Save(img, outputFile, imaging.JPEGQuality(80))
	case "png", "gif", "bmp":
		err = imaging.Save(img, outputFile)
	default:
		return fmt.Errorf("unsupported format: %s", format)
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
