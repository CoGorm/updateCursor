package updater

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CoGorm/updateCursor/internal/config"
)

func TestUpdaterWithConfig(t *testing.T) {
	// Test that updater uses config values correctly
	cfg := &config.Config{
		DownloadDir:     "~/Applications/Cursor",
		FileNamePattern: "Cursor_<version>.AppImage",
		LatestSymlink:   "~/.local/bin/Cursor.AppImage",
		LedgerPath:      "~/.config/updateCursor/cursor-versions.log",
	}

	// Expand paths for testing
	err := cfg.ExpandPaths()
	if err != nil {
		t.Fatalf("Failed to expand config paths: %v", err)
	}

	// Create updater with config
	up := NewUpdater("http://example.com", cfg.DownloadDir, cfg)

	// Verify updater uses config values
	if up.config == nil {
		t.Error("Expected updater to have config")
	}

	// Test that download path uses config
	expectedDownloadPath := filepath.Join(cfg.DownloadDir, "test")
	actualDownloadPath := up.getDownloadPath("test")
	if actualDownloadPath != expectedDownloadPath {
		t.Errorf("Expected download path %s, got %s", expectedDownloadPath, actualDownloadPath)
	}

	// Test that filename generation uses config
	expectedFilename := "Cursor_1.4.5.AppImage"
	actualFilename := up.GenerateFileName("1.4.5")
	if actualFilename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, actualFilename)
	}
}

func TestUpdaterCreatesConfigDirectories(t *testing.T) {
	// Test that updater creates necessary directories from config
	tempDir := t.TempDir()

	cfg := &config.Config{
		DownloadDir:     filepath.Join(tempDir, "downloads"),
		FileNamePattern: "Cursor_<version>.AppImage",
		LatestSymlink:   filepath.Join(tempDir, "bin", "Cursor.AppImage"),
		LedgerPath:      filepath.Join(tempDir, "config", "cursor-versions.log"),
	}

	up := NewUpdater("http://example.com", cfg.DownloadDir, cfg)

	// Test that directories are created when needed
	err := up.ensureDirectories()
	if err != nil {
		t.Fatalf("Failed to ensure directories: %v", err)
	}

	// Verify download directory exists
	if _, err := os.Stat(cfg.DownloadDir); os.IsNotExist(err) {
		t.Errorf("Expected download directory to exist: %s", cfg.DownloadDir)
	}

	// Verify bin directory exists
	binDir := filepath.Dir(cfg.LatestSymlink)
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		t.Errorf("Expected bin directory to exist: %s", binDir)
	}

	// Verify config directory exists
	configDir := filepath.Dir(cfg.LedgerPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Expected config directory to exist: %s", configDir)
	}
}

func TestUpdaterUsesConfigForLedgerPath(t *testing.T) {
	// Test that updater uses config ledger path
	tempDir := t.TempDir()

	cfg := &config.Config{
		DownloadDir:     filepath.Join(tempDir, "downloads"),
		FileNamePattern: "Cursor_<version>.AppImage",
		LatestSymlink:   filepath.Join(tempDir, "bin", "Cursor.AppImage"),
		LedgerPath:      filepath.Join(tempDir, "config", "cursor-versions.log"),
	}

	up := NewUpdater("http://example.com", cfg.DownloadDir, cfg)

	// Test that ledger path is used correctly
	ledgerPath := up.getLedgerPath()
	if ledgerPath != cfg.LedgerPath {
		t.Errorf("Expected ledger path %s, got %s", cfg.LedgerPath, ledgerPath)
	}
}

func TestUpdaterUsesConfigForSymlinkPath(t *testing.T) {
	// Test that updater uses config symlink path
	tempDir := t.TempDir()

	cfg := &config.Config{
		DownloadDir:     filepath.Join(tempDir, "downloads"),
		FileNamePattern: "Cursor_<version>.AppImage",
		LatestSymlink:   filepath.Join(tempDir, "bin", "Cursor.AppImage"),
		LedgerPath:      filepath.Join(tempDir, "config", "cursor-versions.log"),
	}

	up := NewUpdater("http://example.com", cfg.DownloadDir, cfg)

	// Test that symlink path is used correctly
	symlinkPath := up.getLatestSymlinkPath()
	if symlinkPath != cfg.LatestSymlink {
		t.Errorf("Expected symlink path %s, got %s", cfg.LatestSymlink, symlinkPath)
	}
}
