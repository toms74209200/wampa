# Wampa

Wampa is a CLI tool that monitors input files and combines them into a single output file. It's particularly useful for AI coding assistants that need context information from multiple files.

## Motivation

- Provide integrated context information and rules needed by AI coding agents when generating code
- Solve the problem of scattered rule files caused by different AI coding agents using different file formats (.clinerules, .cursorrules, etc.)
- Maintain separate contexts for project specifications and coding rules for better human readability

## Installation

```bash
go install github.com/toms74209200/wampa@latest
```

## Usage

### Basic Command

Wampa monitors `<input_files>` and updates `<output_file>` whenever input files change. You can specify multiple input files separated by spaces.

```bash
wampa -i <input_files> -o <output_file>
```

### Configuration File

When a `wampa.json` configuration file exists in the current directory, Wampa can be run without arguments:

```bash
wampa
```

Example `wampa.json`:
```json
{
    "input_files": ["input1.md", "input2.txt"],
    "output_file": "output.txt"
}
```

If both a configuration file exists and command-line arguments are provided, the command-line arguments take precedence.

### Remote Files

Wampa can also monitor files available over HTTP/HTTPS:

```bash
wampa -i https://example.com/input1.md -o output.txt
```

## File Combination Format

When combining multiple input files, Wampa creates a single output file where each section is preceded by its filename. The format is optimized for AI assistants to recognize different contexts while still being Markdown-friendly.

Example output:
```markdown
[//]: # "filepath: spec.md"
# Product Specifications
- Feature A: Does X
- Feature B: Does Y

[//]: # "filepath: rules.md"
# Coding Rules
1. Use camelCase for variables
2. Add comments for public functions
```

## Command Line Options

- `-i <input_files>`: Space-separated list of input files to monitor
- `-o <output_file>`: Path to the output file
- `-c <config_file>`: Path to the configuration file (defaults to `wampa.json`)

## Requirements

- Go 1.23.4 or higher

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

[toms74209200](https://github.com/toms74209200)