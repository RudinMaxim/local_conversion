package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err == nil {
		numCPUs := runtime.NumCPU()
		configNumWorkers := viper.GetInt("numworkers")

		if configNumWorkers > 0 {
			if configNumWorkers > numCPUs {
				configNumWorkers = configNumWorkers / 2
				fmt.Printf("Configured number of workers exceeds available CPUs, adjusting to: %d\n", configNumWorkers)
			}
		} else {
			configNumWorkers = 2
		}
		viper.Set("numWorkers", configNumWorkers)

		sourceDir := viper.GetString("sourceDir")
		if err := checkDirectory(sourceDir); err != nil {
			fmt.Printf("Source directory error: %v\n", err)
			return
		}

		targetDir := viper.GetString("targetDir")
		if err := checkDirectory(targetDir); err != nil {
			fmt.Printf("Target directory error: %v\n", err)
			return
		}

		return
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
	numWorkers := selectNumWorkers()

	fmt.Printf("Number of workers set to: %d\n", numWorkers)

	viper.Set("sourceDir", sourceDir)
	viper.Set("targetDir", targetDir)
	viper.Set("numWorkers", numWorkers)
	viper.Set("formats", []string{"jpg", "png", "gif", "bmp", "tiff", "webp"})
	viper.SetDefault("quality", 80)
	viper.SetDefault("skipExisting", false)
	viper.SetDefault("compression.quality", 80)
	viper.SetDefault("compression.width", 0)
	viper.SetDefault("compression.height", 0)
	viper.SetDefault("compression.format", "auto")
	viper.SetDefault("compression.skip-existing", false)

	if err := viper.WriteConfigAs("config.yaml"); err != nil {
		fmt.Printf("Error creating config file: %v\n", err)
	} else {
		fmt.Println("Config file created successfully.")
	}
}

func configurationMenu() {
	if err := ShowConfig(); err != nil {
		fmt.Printf("Failed to show configuration: %v\n", err)
		return
	}

	prompt := promptui.Select{
		Label: "Configuration Options",
		Items: []string{"Update", "Exit"},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . | cyan }}",
			Active:   "\U0001F449 {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "\U0001F44D {{ . | green }}",
		},
	}

	for {
		idx, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed: %v\n", err)
			return
		}

		switch idx {
		case 0:
			clearConsole()
			if err := UpdateConfig(); err != nil {
				fmt.Printf("Failed to update configuration: %v\n", err)
			}
		case 1:
			clearConsole()
			fmt.Println("Exiting to main menu...")
			return
		}
	}
}

func ShowConfig() error {
	sourceDir := viper.GetString("sourceDir")
	targetDir := viper.GetString("targetDir")
	numWorkers := viper.GetInt("numWorkers")
	formats := viper.GetStringSlice("formats")

	fmt.Println("Current Configuration:")
	fmt.Printf("Source Directory: %s\n", sourceDir)
	fmt.Printf("Target Directory: %s\n", targetDir)
	fmt.Printf("Number of Workers: %d\n", numWorkers)
	fmt.Printf("Formats: %v\n", formats)

	return nil
}

func UpdateConfig() error {
	prompt := promptui.Prompt{
		Label:   "Enter source directory",
		Default: viper.GetString("sourceDir"),
	}
	sourceDir, err := prompt.Run()
	if err != nil {
		return err
	}
	viper.Set("sourceDir", sourceDir)

	prompt = promptui.Prompt{
		Label:   "Enter target directory",
		Default: viper.GetString("targetDir"),
	}
	targetDir, err := prompt.Run()
	if err != nil {
		return err
	}
	viper.Set("targetDir", targetDir)

	prompt = promptui.Prompt{
		Label:   "Enter number of workers",
		Default: fmt.Sprintf("%d", viper.GetInt("numWorkers")),
	}

	numWorkers := selectNumWorkers()

	fmt.Printf("Number of workers set to: %d\n", numWorkers)

	viper.Set("numWorkers", numWorkers)

	prompt = promptui.Prompt{
		Label:   "Default compression quality (1-100)",
		Default: fmt.Sprintf("%d", viper.GetInt("compression.quality")),
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

	if quality, err := prompt.Run(); err == nil {
		var qualityInt int
		fmt.Sscanf(quality, "%d", &qualityInt)
		viper.Set("compression.quality", qualityInt)
	}

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	fmt.Println("Configuration updated successfully!")
	return nil
}

func selectNumWorkers() int {
	numCPUs := getNumCPUs()

	prompt := promptui.Prompt{
		Label: fmt.Sprintf("Enter number of workers (max: %d)", numCPUs),
		Validate: func(input string) error {
			var numWorkers int
			_, err := fmt.Sscanf(input, "%d", &numWorkers)
			if err != nil || numWorkers <= 0 {
				return fmt.Errorf("invalid number, must be a positive integer")
			}
			if numWorkers > numCPUs {
				return fmt.Errorf("number of workers cannot exceed available CPUs (%d)", numCPUs)
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return 1
	}

	var numWorkers int
	fmt.Sscanf(result, "%d", &numWorkers)
	return numWorkers
}

func checkDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("could not read directory: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("directory is empty: %s", dir)
	}

	return nil
}
