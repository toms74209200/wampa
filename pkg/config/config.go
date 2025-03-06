package config

import "fmt"

// Config represents the application configuration
type Config struct {
	InputFiles []string `json:"input_files"`
	OutputFile string   `json:"output_file"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("configuration is nil")
	}

	if len(c.InputFiles) == 0 {
		return fmt.Errorf("input_files must not be empty")
	}

	for i, file := range c.InputFiles {
		if file == "" {
			return fmt.Errorf("input_files[%d] must not be empty", i)
		}
	}

	if c.OutputFile == "" {
		return fmt.Errorf("output_file must not be empty")
	}

	return nil
}
