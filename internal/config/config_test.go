//go:build small

package config

import (
	"testing"
)

// TestConfig_Validate is a small test that validates config validation
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				InputFiles: []string{"file1.md", "file2.md"},
				OutputFile: "output.md",
			},
			wantErr: false,
		},
		{
			name: "empty input files",
			config: &Config{
				InputFiles: []string{},
				OutputFile: "output.md",
			},
			wantErr: true,
		},
		{
			name: "empty output file",
			config: &Config{
				InputFiles: []string{"file1.md"},
				OutputFile: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
