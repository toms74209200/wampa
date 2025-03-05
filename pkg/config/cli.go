package config

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
