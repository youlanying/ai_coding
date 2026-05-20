package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	InputPath  string `json:"input_path"`
	OutputDir  string `json:"output_dir"`
	FilePrefix string `json:"file_prefix"`
}

func LoadConfig(configPath string) (*Config, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.InputPath == "" {
		config.InputPath = "example/1112_90.xlsx"
	}
	if config.OutputDir == "" {
		config.OutputDir = "example"
	}
	if config.FilePrefix == "" {
		config.FilePrefix = "1112_90"
	}

	return &config, nil
}

func (c *Config) GetOutputPath(suffix string) string {
	return filepath.Join(c.OutputDir, c.FilePrefix+"_"+suffix+".json")
}

func (c *Config) GetAbsInputPath() (string, error) {
	return filepath.Abs(c.InputPath)
}

func (c *Config) GetAbsOutputDir() (string, error) {
	return filepath.Abs(c.OutputDir)
}
