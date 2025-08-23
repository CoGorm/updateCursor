package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExampleConfigFileExists(t *testing.T) {
	// Test that an example config file exists in the repo
	examplePath := filepath.Join("..", "..", "config.example.yaml")

	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Errorf("Expected example config file to exist at %s", examplePath)
	}
}

func TestExampleConfigFileContent(t *testing.T) {
	// Test that the example config file contains expected content
	examplePath := filepath.Join("..", "..", "config.example.yaml")

	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to read example config: %v", err)
	}

	contentStr := string(content)

	// Should contain all required fields
	expectedFields := []string{
		"download_dir:",
		"file_name_pattern:",
		"latest_symlink:",
		"ledger_path:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(contentStr, field) {
			t.Errorf("Expected example config to contain field: %s", field)
		}
	}

	// Should contain example paths
	expectedExamples := []string{
		"~/Documents/Cursor",
		"Cursor_<version>.AppImage",
		"~/.local/bin/Cursor.AppImage",
		"~/.config/updateCursor",
	}

	for _, example := range expectedExamples {
		if !strings.Contains(contentStr, example) {
			t.Errorf("Expected example config to contain example: %s", example)
		}
	}

	// Should contain comments explaining the configuration
	expectedComments := []string{
		"# Configuration file for updateCursor",
		"# Download directory",
		"# Filename pattern",
		"# Latest symlink",
		"# Ledger path",
	}

	for _, comment := range expectedComments {
		if !strings.Contains(contentStr, comment) {
			t.Errorf("Expected example config to contain comment: %s", comment)
		}
	}
}

func TestExampleConfigCanBeLoaded(t *testing.T) {
	// Test that the example config can be loaded and parsed
	examplePath := filepath.Join("..", "..", "config.example.yaml")

	config := NewConfig()
	err := config.LoadFromFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to load example config: %v", err)
	}

	// Validate the loaded config
	err = config.Validate()
	if err != nil {
		t.Errorf("Example config failed validation: %v", err)
	}

	// Should have non-default values to show customization
	if config.DownloadDir == "~/Downloads/Cursor" {
		t.Error("Expected example config to show customized download dir")
	}

	if config.FileNamePattern == "Cursor-<version>-x86_64.AppImage" {
		t.Error("Expected example config to show customized filename pattern")
	}
}
