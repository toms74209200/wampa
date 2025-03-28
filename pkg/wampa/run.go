// Package wampa provides core functionality for the Wampa application
package wampa

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/toms74209200/wampa/pkg/config"
	"github.com/toms74209200/wampa/pkg/formatter"
	"github.com/toms74209200/wampa/pkg/watcher"
)

// Constants for size limits (in bytes)
const (
	KB = 1000
	MB = 1000 * KB
)

// Maximum file size for remote files (100MB)
const maxFileSize = 100 * MB

// Run executes the main application logic
func Run(ctx context.Context, args []string) error {
	// Cache for remote file contents
	remoteContents := make(map[string]string)

	// Check for help flag first
	if config.CheckHelpFlag(args) {
		fmt.Println(config.HelpMessage)
		return nil
	}

	// Parse command line arguments
	cliOpts, err := config.ParseFlags(nil, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n\n", err)
		fmt.Println(config.HelpMessage)
		return fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	var cfg *config.Config

	// Check if config file exists and load it
	configFile := cliOpts.ConfigFile
	if configFile == "wampa.json" && len(args) == 0 {
		// When no arguments are provided and using default config
		_, err := os.Stat(configFile)
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Configuration file wampa.json not found. Please specify -i and -o options or create a configuration file.\n\n")
			fmt.Println(config.HelpMessage)
			return fmt.Errorf("config file not found")
		}
	}

	if cliOpts.ConfigFile != "" {
		data, err := os.ReadFile(cliOpts.ConfigFile)
		if err == nil {
			// Config file found and loaded successfully
			fileCfg, err := config.Parse(data)
			if err != nil {
				return fmt.Errorf("failed to parse config file: %w", err)
			}
			cfg = fileCfg
		} else if !os.IsNotExist(err) {
			// Config file exists but there was an error loading it
			return fmt.Errorf("failed to load config file: %w", err)
		}
		// If file doesn't exist, continue with CLI options only
	}

	// If no config was loaded from file, create from CLI options
	if cfg == nil {
		cfg, err = config.LoadWithCLIOptions(cliOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n\n", err)
			fmt.Println(config.HelpMessage)
			return fmt.Errorf("invalid configuration: %w", err)
		}
	} else {
		// Override file config with CLI options if provided
		if len(cliOpts.InputFiles) > 0 {
			cfg.InputFiles = cliOpts.InputFiles
		}
		if cliOpts.OutputFile != "" {
			cfg.OutputFile = cliOpts.OutputFile
		}
	}

	// Validate final config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create formatter
	formatter := formatter.NewDefaultFormatter()

	// Create and initialize watcher
	w, err := watcher.NewLocalWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create watcher: %v\n\n", err)
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer w.Close()

	// Create channel for file change events
	events := make(chan watcher.Event)

	// Start watching files
	log.Printf("Watching files: %v", cfg.InputFiles)
	log.Printf("Output file: %s", cfg.OutputFile)

	go func() {
		if err := w.Watch(ctx, cfg.InputFiles, events); err != nil {
			log.Printf("Error watching files: %v", err)
		}
	}()

	// Generate initial output
	{
		// Read all input files
		contents := make(map[string]string)
		for _, file := range cfg.InputFiles {
			// Check if the file is a remote URL
			if u, err := url.Parse(file); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
				// Create HTTP request
				req, err := watcher.CreateRemoteFileRequest(ctx, file, nil)
				if err != nil {
					log.Printf("Error creating request for %s: %v", file, err)
					break
				}

				// Execute request
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Printf("Error fetching %s: %v", file, err)
					break
				}
				defer resp.Body.Close()

				// Process response
				data, _, err := watcher.ProcessRemoteFileResponse(resp, file, maxFileSize)
				if err != nil {
					log.Printf("Error processing response from %s: %v", file, err)
					break
				}
				contents[file] = string(data)
				// Cache remote content
				remoteContents[file] = string(data)
			} else {
				// Handle local file
				data, err := os.ReadFile(file)
				if err != nil {
					log.Printf("Error generating initial output - failed to read file %s: %v", file, err)
					break
				}
				contents[file] = string(data)
			}
		}

		// Format contents
		output, err := formatter.Format(cfg.InputFiles, contents)
		if err != nil {
			log.Printf("Error generating initial output - failed to format content: %v", err)
		} else {
			// Write to output file
			if err := os.WriteFile(cfg.OutputFile, []byte(output), 0644); err != nil {
				log.Printf("Error generating initial output - failed to write to output file: %v", err)
			} else {
				log.Printf("Output file updated: %s", cfg.OutputFile)
			}
		}
	}

	// Process events
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-events:
			log.Printf("File changed: %s", e.FilePath)

			// Read all input files
			contents := make(map[string]string)
			for _, file := range cfg.InputFiles {
				// Skip remote files during change events
				if u, err := url.Parse(file); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
					// Use cached remote content
					if content, ok := remoteContents[file]; ok {
						contents[file] = content
					}
					continue
				}

				// Handle local file
				data, err := os.ReadFile(file)
				if err != nil {
					log.Printf("Error processing files - failed to read file %s: %v", file, err)
					continue
				}
				contents[file] = string(data)
			}

			// Format contents
			output, err := formatter.Format(cfg.InputFiles, contents)
			if err != nil {
				log.Printf("Error processing files - failed to format content: %v", err)
				continue
			}

			// Write to output file
			if err := os.WriteFile(cfg.OutputFile, []byte(output), 0644); err != nil {
				log.Printf("Error processing files - failed to write to output file: %v", err)
				continue
			}

			log.Printf("Output file updated: %s", cfg.OutputFile)
		}
	}
}
