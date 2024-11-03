package converter

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/disintegration/imaging"
)

func ConvertImages(sourceDir, targetDir, sourceFormat, targetFormat string, width, height, numWorkers int) error {
	files, err := filepath.Glob(filepath.Join(sourceDir, "*."+sourceFormat))
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
				if err := processImage(file, targetDir, targetFormat, width, height); err != nil {
					fmt.Printf("Error processing file %s: %v\n", file, err)
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
