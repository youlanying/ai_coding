package main

import (
	"fmt"
	"os"
)

func main() {
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	fmt.Printf("Loading config from: %s\n", configPath)
	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	inputPath, err := config.GetAbsInputPath()
	if err != nil {
		fmt.Printf("Failed to get absolute input path: %v\n", err)
		os.Exit(1)
	}

	outputDir, err := config.GetAbsOutputDir()
	if err != nil {
		fmt.Printf("Failed to get absolute output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Input file: %s\n", inputPath)
	fmt.Printf("Output directory: %s\n", outputDir)
	fmt.Printf("File prefix: %s\n", config.FilePrefix)

	parser := NewParser(inputPath)
	fmt.Println("Parsing Excel file...")
	data, err := parser.Parse()
	if err != nil {
		fmt.Printf("Failed to parse Excel file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed data:\n")
	fmt.Printf("  - Config items: %d\n", len(data.Config))
	fmt.Printf("  - Pay table items: %d\n", len(data.PayTable))
	fmt.Printf("  - Prize items: %d\n", len(data.Prize))
	fmt.Printf("  - Base reel strip items: %d\n", len(data.BaseReelStrip))
	fmt.Printf("  - Free reel strip items: %d\n", len(data.FreeReelStrip))

	generator := NewGenerator(outputDir, config.FilePrefix)
	fmt.Println("Generating JSON files...")
	if err := generator.GenerateAll(data); err != nil {
		fmt.Printf("Failed to generate JSON files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All JSON files generated successfully!")
}
