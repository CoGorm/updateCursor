package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for the updateCursor tool
type Config struct {
	DownloadDir     string `yaml:"download_dir"`
	FileNamePattern string `yaml:"file_name_pattern"`
	LatestSymlink   string `yaml:"latest_symlink"`
	LedgerPath      string `yaml:"ledger_path"`
}

// NewConfig creates a new config with default values
func NewConfig() *Config {
	return &Config{
		DownloadDir:     "~/Downloads/Cursor",
		FileNamePattern: "Cursor-<version>-x86_64.AppImage",
		LatestSymlink:   "~/Downloads/Cursor/Cursor.AppImage",
		LedgerPath:      "~/.config/updateCursor/cursor-versions.log",
	}
}

// LoadFromFile loads configuration from a YAML file
func (c *Config) LoadFromFile(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	return nil
}

// SaveToFile saves configuration to a YAML file
func (c *Config) SaveToFile(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.DownloadDir == "" {
		return fmt.Errorf("download_dir cannot be empty")
	}

	if c.FileNamePattern == "" {
		return fmt.Errorf("file_name_pattern cannot be empty")
	}

	if !strings.Contains(c.FileNamePattern, "<version>") {
		return fmt.Errorf("file_name_pattern must contain <version> placeholder")
	}

	if c.LatestSymlink == "" {
		return fmt.Errorf("latest_symlink cannot be empty")
	}

	if c.LedgerPath == "" {
		return fmt.Errorf("ledger_path cannot be empty")
	}

	return nil
}

// ExpandPaths expands all paths that contain ~ to absolute paths
func (c *Config) ExpandPaths() error {
	var err error

	c.DownloadDir, err = expandHomeDir(c.DownloadDir)
	if err != nil {
		return fmt.Errorf("failed to expand download_dir: %v", err)
	}

	c.LatestSymlink, err = expandHomeDir(c.LatestSymlink)
	if err != nil {
		return fmt.Errorf("failed to expand latest_symlink: %v", err)
	}

	c.LedgerPath, err = expandHomeDir(c.LedgerPath)
	if err != nil {
		return fmt.Errorf("failed to expand ledger_path: %v", err)
	}

	return nil
}

// GenerateFileName generates a filename based on the pattern and version
func (c *Config) GenerateFileName(version string) string {
	return strings.ReplaceAll(c.FileNamePattern, "<version>", version)
}

// FindConfigFile finds the config file in the default location
func (c *Config) FindConfigFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".config", "updateCursor", "config.yaml")
	return configPath, nil
}

// CreateDefaultConfigFile creates a default config file at the specified path
func (c *Config) CreateDefaultConfigFile(configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Create default config
	defaultConfig := NewConfig()
	return defaultConfig.SaveToFile(configPath)
}

// LoadOrCreateDefault loads config from file or creates default if none exists
func (c *Config) LoadOrCreateDefault() error {
	configPath, err := c.FindConfigFile()
	if err != nil {
		return fmt.Errorf("failed to find config file: %v", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		err = c.CreateDefaultConfigFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to create default config: %v", err)
		}
		// Load the created default config
		return c.LoadFromFile(configPath)
	}

	// Load existing config
	return c.LoadFromFile(configPath)
}

// expandHomeDir expands ~ to the user's home directory
func expandHomeDir(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}
