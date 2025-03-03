package config

import (
	"encoding/json"
	"fmt"
)

// Config represents the application configuration
type Config struct {
	InputFiles []string `json:"input_files"`
	OutputFile string   `json:"output_file"`
}

// Parse parses the configuration from a byte array
func Parse(data []byte) (*Config, error) {
	// 空の入力チェック
	if len(data) == 0 {
		return nil, fmt.Errorf("empty configuration")
	}

	// まず構造をチェック
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	// 必須フィールドの存在チェック
	inputFiles, hasInputFiles := jsonMap["input_files"]
	if !hasInputFiles {
		return nil, fmt.Errorf("missing required field: input_files")
	}

	outputFile, hasOutputFile := jsonMap["output_file"]
	if !hasOutputFile {
		return nil, fmt.Errorf("missing required field: output_file")
	}

	// input_filesの型チェック
	inputFilesSlice, ok := inputFiles.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input_files must be an array")
	}

	// input_filesの要素が全て文字列かチェック
	for i, file := range inputFilesSlice {
		if _, ok := file.(string); !ok {
			return nil, fmt.Errorf("input_files[%d] must be a string", i)
		}
	}

	// output_fileの型チェック
	if _, ok := outputFile.(string); !ok {
		return nil, fmt.Errorf("output_file must be a string")
	}

	// 実際の構造体へのパース
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// バリデーション
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
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
