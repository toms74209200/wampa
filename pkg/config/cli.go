package config

import (
	"flag"
	"fmt"
	"strings"
)

// CLIOptions represents command-line arguments
type CLIOptions struct {
	InputFiles []string
	OutputFile string
	ConfigFile string
}

// NewCLIOptions creates a new CLIOptions with default values
func NewCLIOptions() *CLIOptions {
	return &CLIOptions{
		InputFiles: []string{},
		OutputFile: "",
		ConfigFile: "wampa.json",
	}
}

// ParseFlags parses command line arguments using the provided FlagSet and returns CLIOptions
func ParseFlags(flagSet *flag.FlagSet, args []string) (*CLIOptions, error) {
	opts := NewCLIOptions()

	// Define a string for input files
	var inputFiles string
	flagSet.StringVar(&inputFiles, "i", "", "Input files (space-separated)")
	flagSet.StringVar(&opts.OutputFile, "o", "", "Output file")
	flagSet.StringVar(&opts.ConfigFile, "c", "wampa.json", "Config file path")

	// Parse arguments
	if err := flagSet.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	// Split input files string into slice if provided
	if inputFiles != "" {
		opts.InputFiles = strings.Fields(inputFiles)
	}

	// Validate required flags when no config file is used
	if flagSet.Lookup("c").Value.String() == "" {
		if len(opts.InputFiles) == 0 {
			return nil, fmt.Errorf("input files (-i) are required when not using a config file")
		}
		if opts.OutputFile == "" {
			return nil, fmt.Errorf("output file (-o) is required when not using a config file")
		}
	}

	return opts, nil
}

// LoadWithCLIOptions creates a new Config from CLI options
func LoadWithCLIOptions(opts *CLIOptions) (*Config, error) {
	config := &Config{
		InputFiles: opts.InputFiles,
		OutputFile: opts.OutputFile,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}
