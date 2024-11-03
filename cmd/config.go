package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config/")

	if err := viper.ReadInConfig(); err == nil {
		return // Конфигурация загружена
	}

	prompt := promptui.Prompt{
		Label:   "Enter source directory (default: ./test/input)",
		Default: "./test/input",
	}
	sourceDir, err := prompt.Run()
	if err != nil {
		fmt.Println("Prompt failed:", err)
		return
	}

	prompt = promptui.Prompt{
		Label:   "Enter target directory (default: ./test/output)",
		Default: "./test/output",
	}
	targetDir, err := prompt.Run()
	if err != nil {
		fmt.Println("Prompt failed:", err)
		return
	}

	prompt = promptui.Prompt{
		Label:   "Enter number of workers (default: 2)",
		Default: "2",
	}
	numWorkersStr, err := prompt.Run()
	if err != nil {
		fmt.Println("Prompt failed:", err)
		return
	}

	numWorkers := 2
	if numWorkersStr != "" {
		fmt.Sscanf(numWorkersStr, "%d", &numWorkers)
	}

	viper.Set("sourceDir", sourceDir)
	viper.Set("targetDir", targetDir)
	viper.Set("numWorkers", numWorkers)
	viper.Set("formats", []string{"jpg", "png", "gif", "bmp", "webp"})
	viper.Set("width", 800)
	viper.Set("height", 600)

	// Сохраняем конфигурацию в файл
	if err := viper.WriteConfigAs("config.yaml"); err != nil {
		fmt.Printf("Error creating config file: %v\n", err)
	} else {
		fmt.Println("Config file created successfully.")
	}
}
