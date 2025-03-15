# Wampa Product Specification

## Overview
Wampa is a CLI application that monitors input files and updates an output file whenever the input files change.

## Motivation
- Integrates context information and rules required by AI coding agents when generating program code into a single file
- Addresses the problem of scattered rule files due to different AI coding agents using different file formats (.clinerules, .cursorrules, etc.) and individual customization needs
- Provides context-separated project specifications and coding rules for better human engineer comprehension

## Functionality
- Monitors specified input files for changes
- When changes are detected, updates the specified output file
- Can watch multiple input files (space-separated on the command line)
- Supports configuration via wampa.toml file
- Command line parameters override configuration file settings when both are present
- Supports remote files via URL (HTTP/HTTPS)
- Uses filenames as section separators in the output file
- Displays help information with `-h` or `--help` flag
- Shows help information when command line arguments contain errors

## File Combination Format
When combining multiple input files into a single output file, each input file content is preceded by its filename as a section separator. The section separators use Markdown-friendly formats that are optimized for AI coding agent context recognition.

### Example:
Given the following input files:

**spec.md**:
```markdown
# Product Specifications
- Feature A: Does X
- Feature B: Does Y
```

**rules.md**:
```markdown
# Coding Rules
1. Use camelCase for variables
2. Add comments for public functions
```

**TODO.md**:
```markdown
# Pending Tasks
- [ ] Implement feature C
- [ ] Fix bug in module D
- [ ] Update documentation
```

The combined output file would look like:

```markdown
[//]: # "filepath: spec.md"
# Product Specifications
- Feature A: Does X
- Feature B: Does Y

[//]: # "filepath: rules.md"
# Coding Rules
1. Use camelCase for variables
2. Add comments for public functions

[//]: # "filepath: TODO.md"
# Pending Tasks
- [ ] Implement feature C
- [ ] Fix bug in module D
- [ ] Update documentation
```

This format enables AI coding agents to understand the context of different sections while keeping the information in a single file, and the section separators are hidden when rendered as Markdown.

## Usage

### Basic Command Line Usage
```bash
wampa -i <input_files> -o <output_file>
```

### Configuration File Usage
When wampa.toml exists in the current directory, the application can be run without arguments:
```bash
wampa
```

### Help Information Display
To display help information, use the following:
```bash
wampa -h
# or
wampa --help
```
When command line arguments contain errors, help information will be automatically displayed.

### Configuration File Format (wampa.toml)
```toml
input_files = ["input1.md", "input2.txt"]
output_file = "output.txt"
```

### Remote File Support
```bash
wampa -i https://example.com/input1.md -o output.txt
```
