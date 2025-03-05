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
		return &ValidationError{"configuration is nil"}
	}

	if len(c.InputFiles) == 0 {
		return &ValidationError{"input_files must not be empty"}
	}

	for i, file := range c.InputFiles {
		if file == "" {
			return &ValidationError{fmt.Sprintf("input_files[%d] must not be empty", i)}
		}
	}

	if c.OutputFile == "" {
		return &ValidationError{"output_file must not be empty"}
	}

	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
