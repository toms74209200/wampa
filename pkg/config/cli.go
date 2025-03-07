package config

import "fmt"

// フラグの定義
const (
	InputFilesFlag = "-i"
	OutputFileFlag = "-o"
	ConfigFileFlag = "-c"
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

	if values, ok := flags[OutputFileFlag]; ok {
		if len(values) == 0 {
			return nil, fmt.Errorf("output file path is required for %s flag", OutputFileFlag)
		}
		opts.OutputFile = values[0]
	}

	if values, ok := flags[ConfigFileFlag]; ok {
		if len(values) == 0 {
			return nil, fmt.Errorf("config file path is required for %s flag", ConfigFileFlag)
		}
		opts.ConfigFile = values[0]
		if opts.ConfigFile == "" {
			opts.ConfigFile = "wampa.json"
		}
	}

	// フラグの検証
	for flag := range flags {
		if flag != InputFilesFlag && flag != OutputFileFlag && flag != ConfigFileFlag {
			return nil, fmt.Errorf("unknown flag: %s", flag)
		}
	}

	// 設定ファイルが指定されていない場合の必須フラグの検証
	if opts.ConfigFile == "" {
		if len(opts.InputFiles) == 0 {
			return nil, fmt.Errorf("input files (%s) are required when not using a config file", InputFilesFlag)
		}
		if opts.OutputFile == "" {
			return nil, fmt.Errorf("output file (%s) is required when not using a config file", OutputFileFlag)
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
