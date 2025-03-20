package config

import (
	"fmt"
)

// フラグの定義
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

// ヘルプメッセージの定義
const HelpMessage = `使用方法: wampa [オプション]

オプション:
  -i, --input   入力ファイルを指定（複数指定可能）
  -o, --output  出力ファイルを指定
  -h, --help    このヘルプメッセージを表示`

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
			return nil, fmt.Errorf("出力ファイルのパスが指定されていません: %s", OutputFileFlag)
		}
		opts.OutputFile = values[0]
	}
	if values, ok := flags[OutputFileFlagLong]; ok {
		if len(values) == 0 {
			return nil, fmt.Errorf("出力ファイルのパスが指定されていません: %s", OutputFileFlagLong)
		}
		opts.OutputFile = values[0]
	}

	if values, ok := flags[ConfigFileFlag]; ok {
		if len(values) == 0 {
			return nil, fmt.Errorf("設定ファイルのパスが指定されていません: %s", ConfigFileFlag)
		}
		opts.ConfigFile = values[0]
		if opts.ConfigFile == "" {
			opts.ConfigFile = "wampa.json"
		}
	}

	// フラグの検証
	for flag := range flags {
		if flag != InputFilesFlag && flag != OutputFileFlag && flag != ConfigFileFlag {
			return nil, fmt.Errorf("不明なオプション: %s", flag)
		}
	}

	// 設定ファイルが指定されていない場合の必須フラグの検証
	if opts.ConfigFile == "" {
		if len(opts.InputFiles) == 0 {
			return nil, fmt.Errorf("設定ファイル wampa.json が見つかりません。-i および -o オプションを指定するか、設定ファイルを作成してください。")
		}
		if opts.OutputFile == "" {
			return nil, fmt.Errorf("出力ファイルが指定されていません。-o オプションを指定するか、設定ファイルを作成してください。")
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
