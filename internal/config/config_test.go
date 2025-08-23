package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	// Test default configuration when no config file exists
	config := NewConfig()

	// Verify default values
	expectedDownloadDir := "~/Downloads/Cursor"
	if config.DownloadDir != expectedDownloadDir {
		t.Errorf("Expected default download dir %s, got %s", expectedDownloadDir, config.DownloadDir)
	}

	expectedFileNamePattern := "Cursor-<version>-x86_64.AppImage"
	if config.FileNamePattern != expectedFileNamePattern {
		t.Errorf("Expected default filename pattern %s, got %s", expectedFileNamePattern, config.FileNamePattern)
	}

	expectedLatestSymlink := "~/Downloads/Cursor/Cursor.AppImage"
	if config.LatestSymlink != expectedLatestSymlink {
		t.Errorf("Expected default latest symlink %s, got %s", expectedLatestSymlink, config.LatestSymlink)
	}

	expectedLedgerPath := "~/.config/updateCursor/cursor-versions.log"
	if config.LedgerPath != expectedLedgerPath {
		t.Errorf("Expected default ledger path %s, got %s", expectedLedgerPath, config.LedgerPath)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `
download_dir: "~/Applications/Cursor"
file_name_pattern: "Cursor_<version>.AppImage"
latest_symlink: "~/.local/bin/Cursor.AppImage"
ledger_path: "~/.config/updateCursor/cursor-versions.log"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test loading config from file
	config := NewConfig()
	err = config.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	expectedDownloadDir := "~/Applications/Cursor"
	if config.DownloadDir != expectedDownloadDir {
		t.Errorf("Expected download dir %s, got %s", expectedDownloadDir, config.DownloadDir)
	}

	expectedFileNamePattern := "Cursor_<version>.AppImage"
	if config.FileNamePattern != expectedFileNamePattern {
		t.Errorf("Expected filename pattern %s, got %s", expectedFileNamePattern, config.FileNamePattern)
	}

	expectedLatestSymlink := "~/.local/bin/Cursor.AppImage"
	if config.LatestSymlink != expectedLatestSymlink {
		t.Errorf("Expected latest symlink %s, got %s", expectedLatestSymlink, config.LatestSymlink)
	}

	expectedLedgerPath := "~/.config/updateCursor/cursor-versions.log"
	if config.LedgerPath != expectedLedgerPath {
		t.Errorf("Expected ledger path %s, got %s", expectedLedgerPath, config.LedgerPath)
	}
}

func TestConfigValidation(t *testing.T) {
	config := NewConfig()

	// Test valid config
	config.DownloadDir = "~/Applications/Cursor"
	config.FileNamePattern = "Cursor_<version>.AppImage"
	config.LatestSymlink = "~/.local/bin/Cursor.AppImage"
	config.LedgerPath = "~/.config/updateCursor/cursor-versions.log"

	err := config.Validate()
	if err != nil {
		t.Errorf("Expected valid config, but got error: %v", err)
	}

	// Test invalid config - missing <version> placeholder
	config.FileNamePattern = "Cursor.AppImage"
	err = config.Validate()
	if err == nil {
		t.Error("Expected error for missing <version> placeholder, but got none")
	}

	// Test invalid config - empty download dir
	config.FileNamePattern = "Cursor_<version>.AppImage"
	config.DownloadDir = ""
	err = config.Validate()
	if err == nil {
		t.Error("Expected error for empty download dir, but got none")
	}
}

func TestExpandPaths(t *testing.T) {
	config := NewConfig()
	config.DownloadDir = "~/Applications/Cursor"
	config.LatestSymlink = "~/.local/bin/Cursor.AppImage"
	config.LedgerPath = "~/.config/updateCursor/cursor-versions.log"

	// Test path expansion
	err := config.ExpandPaths()
	if err != nil {
		t.Fatalf("Failed to expand paths: %v", err)
	}

	// Verify paths are expanded (should not start with ~)
	if config.DownloadDir[0] == '~' {
		t.Errorf("Expected expanded download dir, got %s", config.DownloadDir)
	}

	if config.LatestSymlink[0] == '~' {
		t.Errorf("Expected expanded latest symlink, got %s", config.LatestSymlink)
	}

	if config.LedgerPath[0] == '~' {
		t.Errorf("Expected expanded ledger path, got %s", config.LedgerPath)
	}
}

func TestGenerateFileName(t *testing.T) {
	config := NewConfig()
	config.FileNamePattern = "Cursor_<version>.AppImage"

	// Test filename generation
	filename := config.GenerateFileName("1.4.5")
	expected := "Cursor_1.4.5.AppImage"

	if filename != expected {
		t.Errorf("Expected filename %s, got %s", expected, filename)
	}

	// Test with different pattern
	config.FileNamePattern = "cursor-<version>-linux.AppImage"
	filename = config.GenerateFileName("1.4.5")
	expected = "cursor-1.4.5-linux.AppImage"

	if filename != expected {
		t.Errorf("Expected filename %s, got %s", expected, filename)
	}
}

func TestConfigFileDiscovery(t *testing.T) {
	// Test finding config file in default location
	config := NewConfig()

	// Should find config in ~/.config/updateCursor/config.yaml
	configPath, err := config.FindConfigFile()
	if err != nil {
		t.Fatalf("Failed to find config file: %v", err)
	}

	// Should return the expanded config path (not the tilde version)
	// The function expands paths automatically
	if configPath[0] == '~' {
		t.Errorf("Expected expanded config path, got %s", configPath)
	}

	// Should contain the expected directory structure
	if !strings.Contains(configPath, ".config") || !strings.Contains(configPath, "updateCursor") {
		t.Errorf("Expected config path to contain .config/updateCursor, got %s", configPath)
	}
}

func TestCreateDefaultConfigFile(t *testing.T) {
	// Test creating default config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	config := NewConfig()
	err := config.CreateDefaultConfigFile(configPath)
	if err != nil {
		t.Fatalf("Failed to create default config file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected config file to exist at %s", configPath)
	}

	// Verify file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Should contain default values
	contentStr := string(content)
	if !strings.Contains(contentStr, "~/Downloads/Cursor") {
		t.Error("Expected config file to contain default download directory")
	}

	if !strings.Contains(contentStr, "Cursor-<version>-x86_64.AppImage") {
		t.Error("Expected config file to contain default filename pattern")
	}
}

func TestLoadConfigWithInvalidYAML(t *testing.T) {
	// Test loading invalid YAML config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid-config.yaml")

	invalidYAML := `
download_dir: "~/Applications/Cursor"
file_name_pattern: "Cursor_<version>.AppImage
latest_symlink: "~/.local/bin/Cursor.AppImage"
ledger_path: "~/.config/updateCursor/cursor-versions.log"
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	config := NewConfig()
	err = config.LoadFromFile(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, but got none")
	}
}

func TestConfigWithCustomPaths(t *testing.T) {
	// Test config with completely custom paths
	config := NewConfig()
	config.DownloadDir = "/opt/cursor/versions"
	config.FileNamePattern = "cursor-<version>-linux-x64.AppImage"
	config.LatestSymlink = "/usr/local/bin/cursor"
	config.LedgerPath = "/var/log/cursor-updates.log"

	// Test validation
	err := config.Validate()
	if err != nil {
		t.Errorf("Expected valid custom config, but got error: %v", err)
	}

	// Test filename generation
	filename := config.GenerateFileName("1.4.5")
	expected := "cursor-1.4.5-linux-x64.AppImage"

	if filename != expected {
		t.Errorf("Expected filename %s, got %s", expected, filename)
	}
}
