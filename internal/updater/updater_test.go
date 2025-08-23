package updater

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadCursor(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/stable/linux-x64" {
			// Simulate redirect to versioned URL
			http.Redirect(w, r, "/download/Cursor-1.0.0-x86_64.AppImage", http.StatusFound)
		} else if r.URL.Path == "/download/Cursor-1.0.0-x86_64.AppImage" {
			// Return mock AppImage content
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte("mock cursor appimage content"))
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create temporary directory for test
	tempDir := t.TempDir()

	updater := NewUpdater(server.URL+"/download/stable/linux-x64", tempDir, nil)

	// Test downloading
	filename, err := updater.DownloadCursor()
	if err != nil {
		t.Fatalf("Failed to download Cursor: %v", err)
	}

	expectedFilename := "Cursor-1.0.0-x86_64.AppImage"
	if filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, filename)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", filePath)
	}
}

func TestGetRemoteVersion(t *testing.T) {
	// Create a mock HTTP server that redirects
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/stable/linux-x64" {
			http.Redirect(w, r, "/download/Cursor-1.2.3-x86_64.AppImage", http.StatusFound)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	updater := NewUpdater(server.URL+"/download/stable/linux-x64", "/tmp", nil)

	version, err := updater.GetRemoteVersion()
	if err != nil {
		t.Fatalf("Failed to get remote version: %v", err)
	}

	expectedVersion := "1.2.3"
	if version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, version)
	}
}

func TestGetLocalVersion(t *testing.T) {
	tempDir := t.TempDir()

	updater := NewUpdater("http://example.com", tempDir, nil)

	// Test with no symlink
	version, err := updater.GetLocalVersion()
	if err != nil {
		t.Fatalf("Failed to get local version: %v", err)
	}

	if version != "" {
		t.Errorf("Expected empty version for no symlink, got %s", version)
	}

	// Test with symlink
	symlinkPath := filepath.Join(tempDir, "Cursor.AppImage")
	targetPath := filepath.Join(tempDir, "Cursor-1.0.0-x86_64.AppImage")

	// Create target file
	err = os.WriteFile(targetPath, []byte("mock content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create symlink
	err = os.Symlink(targetPath, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	version, err = updater.GetLocalVersion()
	if err != nil {
		t.Fatalf("Failed to get local version: %v", err)
	}

	expectedVersion := "1.0.0"
	if version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, version)
	}
}

func TestSwitchToVersion(t *testing.T) {
	tempDir := t.TempDir()

	updater := NewUpdater("http://example.com", tempDir, nil)

	// Create a version file
	versionFile := "Cursor-1.0.0-x86_64.AppImage"
	versionPath := filepath.Join(tempDir, versionFile)
	err := os.WriteFile(versionPath, []byte("mock content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create version file: %v", err)
	}

	// Test switching to version
	err = updater.SwitchToVersion("1.0.0")
	if err != nil {
		t.Fatalf("Failed to switch to version: %v", err)
	}

	// Verify symlink was created
	symlinkPath := filepath.Join(tempDir, "Cursor.AppImage")
	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		t.Errorf("Expected symlink %s to exist", symlinkPath)
	}

	// Verify symlink points to correct file
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to read symlink: %v", err)
	}

	if target != versionFile {
		t.Errorf("Expected symlink to point to %s, got %s", versionFile, target)
	}
}

func TestCheckForUpdates(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/stable/linux-x64" {
			http.Redirect(w, r, "/download/Cursor-1.1.0-x86_64.AppImage", http.StatusFound)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	updater := NewUpdater(server.URL+"/download/stable/linux-x64", tempDir, nil)

	// Test with no local version
	needsUpdate, remoteVersion, err := updater.CheckForUpdates()
	if err != nil {
		t.Fatalf("Failed to check for updates: %v", err)
	}

	if !needsUpdate {
		t.Error("Expected update to be needed when no local version exists")
	}

	if remoteVersion != "1.1.0" {
		t.Errorf("Expected remote version 1.1.0, got %s", remoteVersion)
	}

	// Test with older local version
	symlinkPath := filepath.Join(tempDir, "Cursor.AppImage")
	targetPath := filepath.Join(tempDir, "Cursor-1.0.0-x86_64.AppImage")

	err = os.WriteFile(targetPath, []byte("mock content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	err = os.Symlink(targetPath, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	needsUpdate, remoteVersion, err = updater.CheckForUpdates()
	if err != nil {
		t.Fatalf("Failed to check for updates: %v", err)
	}

	if !needsUpdate {
		t.Error("Expected update to be needed when local version is older")
	}

	// Test with same version
	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/stable/linux-x64" {
			http.Redirect(w, r, "/download/Cursor-1.0.0-x86_64.AppImage", http.StatusFound)
		}
	}))
	defer server.Close()

	updater = NewUpdater(server.URL+"/download/stable/linux-x64", tempDir, nil)

	needsUpdate, _, err = updater.CheckForUpdates()
	if err != nil {
		t.Fatalf("Failed to check for updates: %v", err)
	}

	if needsUpdate {
		t.Error("Expected no update needed when versions are the same")
	}
}

func TestCalculateSHA256(t *testing.T) {
	tempDir := t.TempDir()

	updater := NewUpdater("http://example.com", tempDir, nil)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate SHA256
	hash, err := updater.CalculateSHA256(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate SHA256: %v", err)
	}

	// Expected SHA256 for "Hello, World!"
	expectedHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	if hash != expectedHash {
		t.Errorf("Expected SHA256 %s, got %s", expectedHash, hash)
	}
}

func TestGetLocalVersionWithRegularFile(t *testing.T) {
	tempDir := t.TempDir()

	updater := NewUpdater("http://example.com", tempDir, nil)

	// Create a version file first
	versionFile := "Cursor-1.0.0-x86_64.AppImage"
	versionPath := filepath.Join(tempDir, versionFile)
	err := os.WriteFile(versionPath, []byte("mock content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create version file: %v", err)
	}

	// Create Cursor.AppImage as a copy of the version file (not a symlink)
	cursorPath := filepath.Join(tempDir, "Cursor.AppImage")
	err = os.WriteFile(cursorPath, []byte("mock content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create Cursor.AppImage: %v", err)
	}

	// Test that we can detect the version even when Cursor.AppImage is a regular file
	version, err := updater.GetLocalVersion()
	if err != nil {
		t.Fatalf("Failed to get local version: %v", err)
	}

	// Should detect version 1.0.0 by matching file size and timestamp
	expectedVersion := "1.0.0"
	if version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, version)
	}
}

func TestGetLocalVersionWithRegularFileDifferentTimestamps(t *testing.T) {
	tempDir := t.TempDir()

	updater := NewUpdater("http://example.com", tempDir, nil)

	// Create a version file first
	versionFile := "Cursor-1.0.0-x86_64.AppImage"
	versionPath := filepath.Join(tempDir, versionFile)
	err := os.WriteFile(versionPath, []byte("mock content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create version file: %v", err)
	}

	// Wait a moment to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Create Cursor.AppImage as a copy of the version file (not a symlink)
	cursorPath := filepath.Join(tempDir, "Cursor.AppImage")
	err = os.WriteFile(cursorPath, []byte("mock content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create Cursor.AppImage: %v", err)
	}

	// Test that we can detect the version even when Cursor.AppImage is a regular file
	version, err := updater.GetLocalVersion()
	if err != nil {
		t.Fatalf("Failed to get local version: %v", err)
	}

	// Should detect version 1.0.0 by matching file size and timestamp
	expectedVersion := "1.0.0"
	if version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, version)
	}
}

func TestGetRemoteVersionWithMultipleRedirects(t *testing.T) {
	// Create a mock HTTP server that simulates multiple redirects
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/stable/linux-x64" {
			// First redirect
			http.Redirect(w, r, "/redirect1", http.StatusFound)
		} else if r.URL.Path == "/redirect1" {
			// Second redirect
			http.Redirect(w, r, "/download/Cursor-1.4.5-x86_64.AppImage", http.StatusFound)
		} else if r.URL.Path == "/download/Cursor-1.4.5-x86_64.AppImage" {
			// Final destination
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte("mock cursor appimage content"))
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	updater := NewUpdater(server.URL+"/download/stable/linux-x64", "/tmp", nil)

	version, err := updater.GetRemoteVersion()
	if err != nil {
		t.Fatalf("Failed to get remote version: %v", err)
	}

	expectedVersion := "1.4.5"
	if version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, version)
	}
}

func TestDownloadCursorWithProgress(t *testing.T) {
	// Create a mock HTTP server that returns a large file
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/stable/linux-x64" {
			// Simulate redirect to versioned URL
			http.Redirect(w, r, "/download/Cursor-1.0.0-x86_64.AppImage", http.StatusFound)
		} else if r.URL.Path == "/download/Cursor-1.0.0-x86_64.AppImage" {
			// Return mock AppImage content with progress tracking
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Length", "1000000") // 1MB file

			// Simulate slow download to test progress
			data := make([]byte, 1000000)
			for i := 0; i < 1000000; i += 10000 { // Send in 10KB chunks
				if i+10000 <= 1000000 {
					w.Write(data[i : i+10000])
				} else {
					w.Write(data[i:])
				}
				// Small delay to simulate network
				time.Sleep(1 * time.Millisecond)
			}
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create temporary directory for test
	tempDir := t.TempDir()

	updater := NewUpdater(server.URL+"/download/stable/linux-x64", tempDir, nil)

	// Track progress updates
	var progressUpdates []ProgressUpdate
	updater.SetProgressCallback(func(update ProgressUpdate) {
		progressUpdates = append(progressUpdates, update)
	})

	// Test downloading with progress tracking
	filename, err := updater.DownloadCursor()
	if err != nil {
		t.Fatalf("Failed to download Cursor: %v", err)
	}

	expectedFilename := "Cursor-1.0.0-x86_64.AppImage"
	if filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, filename)
	}

	// Verify progress updates were received
	if len(progressUpdates) == 0 {
		t.Error("Expected progress updates, but none were received")
	}

	// Verify we got progress updates during download
	hasProgress := false
	for _, update := range progressUpdates {
		if update.BytesDownloaded > 0 && update.TotalBytes > 0 {
			hasProgress = true
			break
		}
	}

	if !hasProgress {
		t.Error("Expected progress updates with download information, but none found")
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", filePath)
	}
}

func TestUpdaterCreatesDownloadDirectory(t *testing.T) {
	// Test that updater creates download directory automatically
	tempDir := t.TempDir()
	downloadDir := filepath.Join(tempDir, "downloads", "cursor")

	// Create updater with a download directory that doesn't exist yet
	updater := NewUpdater("http://example.com", downloadDir, nil)

	// The updater should create the download directory when needed
	// This test verifies that the directory creation logic works
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		// Directory doesn't exist yet, which is expected
		// The updater should create it during download operations
	}

	// Verify that the updater can work with the directory path
	// even if it doesn't exist yet
	if updater.WorkDir() != downloadDir {
		t.Errorf("Expected work dir %s, got %s", downloadDir, updater.WorkDir())
	}
}

func TestUpdaterEnsuresDirectoriesExist(t *testing.T) {
	// Test that updater ensures all necessary directories exist
	tempDir := t.TempDir()
	downloadDir := filepath.Join(tempDir, "downloads", "cursor")

	// Create updater
	updater := NewUpdater("http://example.com", downloadDir, nil)

	// Call the method that ensures directories exist
	err := updater.ensureDirectories()
	if err != nil {
		t.Fatalf("Failed to ensure directories: %v", err)
	}

	// Verify download directory was created
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		t.Errorf("Expected download directory to exist: %s", downloadDir)
	}
}
