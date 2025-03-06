//go:build small

package config

import (
	"testing"
)

// TestNewCLIOptions tests the creation of new CLIOptions with default values
func TestNewCLIOptions(t *testing.T) {
	opts := NewCLIOptions()
	if opts == nil {
		t.Error("NewCLIOptions() returned nil")
		return
	}

	if len(opts.InputFiles) != 0 {
		t.Errorf("NewCLIOptions().InputFiles = %v, want empty slice", opts.InputFiles)
	}

	if opts.OutputFile != "" {
		t.Errorf("NewCLIOptions().OutputFile = %q, want empty string", opts.OutputFile)
	}

	if opts.ConfigFile != "wampa.json" {
		t.Errorf("NewCLIOptions().ConfigFile = %q, want \"wampa.json\"", opts.ConfigFile)
	}
}

// TestLoadWithCLIOptions tests configuration loading with CLI options
func TestLoadWithCLIOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    *CLIOptions
		want    *Config
		wantErr bool
	}{
		{
			name: "valid cli options",
			opts: &CLIOptions{
				InputFiles: []string{"file1.md", "file2.md"},
				OutputFile: "output.md",
			},
			want: &Config{
				InputFiles: []string{"file1.md", "file2.md"},
				OutputFile: "output.md",
			},
			wantErr: false,
		},
		{
			name: "empty input files",
			opts: &CLIOptions{
				InputFiles: []string{},
				OutputFile: "output.md",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty output file",
			opts: &CLIOptions{
				InputFiles: []string{"file1.md"},
				OutputFile: "",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadWithCLIOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadWithCLIOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got.InputFiles) != len(tt.want.InputFiles) {
					t.Errorf("LoadWithCLIOptions() InputFiles = %v, want %v", got.InputFiles, tt.want.InputFiles)
				} else {
					for i, file := range got.InputFiles {
						if file != tt.want.InputFiles[i] {
							t.Errorf("LoadWithCLIOptions() InputFiles[%d] = %v, want %v", i, file, tt.want.InputFiles[i])
						}
					}
				}
				if got.OutputFile != tt.want.OutputFile {
					t.Errorf("LoadWithCLIOptions() OutputFile = %v, want %v", got.OutputFile, tt.want.OutputFile)
				}
			}
		})
	}
}
