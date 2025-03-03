//go:build small

package config

import (
	"testing"
)

// TestParse is a small test that validates config parsing
func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    *Config
		wantErr bool
	}{
		{
			name:  "valid config",
			input: []byte(`{"input_files":["file1.md","file2.md"],"output_file":"output.md"}`),
			want: &Config{
				InputFiles: []string{"file1.md", "file2.md"},
				OutputFile: "output.md",
			},
			wantErr: false,
		},
		{
			name: "valid config with whitespace",
			input: []byte(`{
				"input_files": ["file1.md", "file2.md"],
				"output_file": "output.md"
			}`),
			want: &Config{
				InputFiles: []string{"file1.md", "file2.md"},
				OutputFile: "output.md",
			},
			wantErr: false,
		},
		{
			name: "valid config with JSONC line comments",
			input: []byte(`{
				// Input files to be combined
				"input_files": ["file1.md", "file2.md"],
				// Output file path
				"output_file": "output.md"
			}`),
			want: &Config{
				InputFiles: []string{"file1.md", "file2.md"},
				OutputFile: "output.md",
			},
			wantErr: true, // 標準のJSONパーサーはコメントを許可しない
		},
		{
			name: "valid config with JSONC block comments",
			input: []byte(`{
				/* Input files to be combined */
				"input_files": ["file1.md", "file2.md"],
				/* Output file path */
				"output_file": "output.md"
			}`),
			want: &Config{
				InputFiles: []string{"file1.md", "file2.md"},
				OutputFile: "output.md",
			},
			wantErr: true, // 標準のJSONパーサーはコメントを許可しない
		},
		{
			name: "valid config with JSONC comments containing brackets",
			input: []byte(`{
				/* Example config:
				{
					"input_files": ["example.md"]
				}
				*/
				"input_files": ["file1.md"],
				"output_file": "output.md"
			}`),
			want: &Config{
				InputFiles: []string{"file1.md"},
				OutputFile: "output.md",
			},
			wantErr: true, // 標準のJSONパーサーはコメントを許可しない
		},
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "invalid json - unclosed object",
			input:   []byte(`{"input_files":["file1.md"]`),
			wantErr: true,
		},
		{
			name:    "invalid json - unclosed array",
			input:   []byte(`{"input_files":["file1.md","file2.md"`),
			wantErr: true,
		},
		{
			name:    "invalid json - missing comma",
			input:   []byte(`{"input_files":["file1.md" "file2.md"],"output_file":"output.md"}`),
			wantErr: true,
		},
		{
			name:    "invalid json - trailing comma in array",
			input:   []byte(`{"input_files":["file1.md",],"output_file":"output.md"}`),
			wantErr: true,
		},
		{
			name:    "invalid json - trailing comma in object",
			input:   []byte(`{"input_files":["file1.md"],"output_file":"output.md",}`),
			wantErr: true,
		},
		{
			name:    "invalid json - single quotes",
			input:   []byte(`{'input_files':['file1.md'],'output_file':'output.md'}`),
			wantErr: true,
		},
		{
			name:    "invalid json - unquoted keys",
			input:   []byte(`{input_files:["file1.md"],output_file:"output.md"}`),
			wantErr: true,
		},
		{
			name:    "invalid json - unquoted string values",
			input:   []byte(`{"input_files":[file1.md],"output_file":output.md}`),
			wantErr: true,
		},
		{
			name:    "invalid type - input_files is number array",
			input:   []byte(`{"input_files":[1,2,3],"output_file":"output.md"}`),
			wantErr: true,
		},
		{
			name:    "invalid type - input_files is string",
			input:   []byte(`{"input_files":"file1.md","output_file":"output.md"}`),
			wantErr: true,
		},
		{
			name:    "invalid type - output_file is array",
			input:   []byte(`{"input_files":["file1.md"],"output_file":["output.md"]}`),
			wantErr: true,
		},
		{
			name:    "invalid type - output_file is number",
			input:   []byte(`{"input_files":["file1.md"],"output_file":123}`),
			wantErr: true,
		},
		{
			name:    "missing required field - input_files",
			input:   []byte(`{"output_file":"output.md"}`),
			wantErr: true,
		},
		{
			name:    "missing required field - output_file",
			input:   []byte(`{"input_files":["file1.md"]}`),
			wantErr: true,
		},
		{
			name:    "null value - input_files",
			input:   []byte(`{"input_files":null,"output_file":"output.md"}`),
			wantErr: true,
		},
		{
			name:    "null value - output_file",
			input:   []byte(`{"input_files":["file1.md"],"output_file":null}`),
			wantErr: true,
		},
		{
			name:    "wrong key - inputFiles instead of input_files",
			input:   []byte(`{"inputFiles":["file1.md"],"output_file":"output.md"}`),
			wantErr: true,
		},
		{
			name:    "wrong key - outputFile instead of output_file",
			input:   []byte(`{"input_files":["file1.md"],"outputFile":"output.md"}`),
			wantErr: true,
		},
		{
			name:  "extra key",
			input: []byte(`{"input_files":["file1.md"],"output_file":"output.md","extra":"value"}`),
			want: &Config{
				InputFiles: []string{"file1.md"},
				OutputFile: "output.md",
			},
			wantErr: false, // 余分なキーは無視する
		},
		{
			name:    "empty object",
			input:   []byte(`{}`),
			wantErr: true,
		},
		{
			name:    "empty array",
			input:   []byte(`[]`),
			wantErr: true,
		},
		{
			name:    "null",
			input:   []byte(`null`),
			wantErr: true,
		},
		{
			name:    "only whitespace",
			input:   []byte(`    `),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil {
				if len(got.InputFiles) != len(tt.want.InputFiles) {
					t.Errorf("Parse() InputFiles length = %v, want %v", len(got.InputFiles), len(tt.want.InputFiles))
					return
				}
				for i, f := range got.InputFiles {
					if f != tt.want.InputFiles[i] {
						t.Errorf("Parse() InputFiles[%d] = %v, want %v", i, f, tt.want.InputFiles[i])
					}
				}
				if got.OutputFile != tt.want.OutputFile {
					t.Errorf("Parse() OutputFile = %v, want %v", got.OutputFile, tt.want.OutputFile)
				}
			}
		})
	}
}

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
