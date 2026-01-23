package config

import (
	"fmt"
)

// Flag definitions
const (
	InputFilesFlag     = "-i"
	InputFilesFlagLong = "--input"
	OutputFileFlag     = "-o"
	OutputFileFlagLong = "--output"
	ConfigFileFlag     = "-c"
	ConfigFileFlagLong = "--config"
	HelpFlag           = "-h"
	HelpFlagLong       = "--help"
)

// Help message definition
const HelpMessage = `Usage: wampa [options]

Options:
  -i, --input   Specify input file(s) (can be specified multiple times)
  -o, --output  Specify output file
  -h, --help    Display this help message`

// CheckHelpFlag checks if help flag is present in arguments
func CheckHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == HelpFlag || arg == HelpFlagLong {
			return true
		}
	}
	return false
}

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

// ParseFlags parses command line arguments and returns CLIOptions
func ParseFlags(_ interface{}, args []string) (*CLIOptions, error) {
	// 第1段階: フラグとその値を連想配列に分類
	flags := make(map[string][]string)
	var currentFlag string

	for _, arg := range args {
		if len(arg) == 0 {
			continue
		}

		if arg[0] == '-' {
			currentFlag = arg
			flags[currentFlag] = []string{}
		} else if currentFlag != "" {
			flags[currentFlag] = append(flags[currentFlag], arg)
		}
	}

	// 第2段階: 連想配列から必要なフラグの値を取り出してCLIOptionsを構築
	opts := NewCLIOptions()

	if values, ok := flags[InputFilesFlag]; ok {
		opts.InputFiles = values
	}
	if values, ok := flags[InputFilesFlagLong]; ok {
		opts.InputFiles = append(opts.InputFiles, values...)
	}

	if values, ok := flags[OutputFileFlag]; ok {
		if len(values) == 0 {
			return nil, fmt.Errorf("Output file path not specified: %s", OutputFileFlag)
		}
		opts.OutputFile = values[0]
	}
	if values, ok := flags[OutputFileFlagLong]; ok {
		if len(values) == 0 {
			return nil, fmt.Errorf("Output file path not specified: %s", OutputFileFlagLong)
		}
		opts.OutputFile = values[0]
	}

	if values, ok := flags[ConfigFileFlag]; ok {
		if len(values) == 0 {
			return nil, fmt.Errorf("Config file path not specified: %s", ConfigFileFlag)
		}
		opts.ConfigFile = values[0]
		if opts.ConfigFile == "" {
			opts.ConfigFile = "wampa.json"
		}
	}

	// Flag validation
	for flag := range flags {
		if flag != InputFilesFlag && flag != InputFilesFlagLong &&
			flag != OutputFileFlag && flag != OutputFileFlagLong &&
			flag != ConfigFileFlag && flag != ConfigFileFlagLong {
			return nil, fmt.Errorf("Unknown option: %s", flag)
		}
	}

	// Required flag validation when no config file is specified
	if opts.ConfigFile == "" {
		if len(opts.InputFiles) == 0 {
			return nil, fmt.Errorf("Configuration file wampa.json not found. Please specify -i and -o options or create a configuration file.")
		}
		if opts.OutputFile == "" {
			return nil, fmt.Errorf("Output file not specified. Please specify -o option or create a configuration file.")
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
