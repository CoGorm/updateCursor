package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CoGorm/updateCursor/internal/config"
	"github.com/CoGorm/updateCursor/internal/version"
)

// ProgressUpdate represents a download progress update
type ProgressUpdate struct {
	BytesDownloaded int64
	TotalBytes      int64
	Percentage      float64
	Speed           float64 // bytes per second
}

// ProgressCallback is a function type for progress updates
type ProgressCallback func(ProgressUpdate)

// ProgressReader wraps an io.Reader to track download progress
type ProgressReader struct {
	Reader          io.Reader
	TotalBytes      int64
	BytesDownloaded *int64
	Callback        ProgressCallback
	lastUpdate      time.Time
}

// Read implements io.Reader interface and tracks progress
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if n > 0 {
		*pr.BytesDownloaded += int64(n)

		// Call progress callback if available
		if pr.Callback != nil {
			now := time.Now()
			// Update progress every 100ms to avoid too many callbacks
			if now.Sub(pr.lastUpdate) >= 100*time.Millisecond {
				percentage := float64(*pr.BytesDownloaded) / float64(pr.TotalBytes) * 100
				if pr.TotalBytes > 0 {
					percentage = float64(*pr.BytesDownloaded) / float64(pr.TotalBytes) * 100
				}

				// Calculate speed (bytes per second)
				speed := float64(*pr.BytesDownloaded) / now.Sub(pr.lastUpdate).Seconds()

				pr.Callback(ProgressUpdate{
					BytesDownloaded: *pr.BytesDownloaded,
					TotalBytes:      pr.TotalBytes,
					Percentage:      percentage,
					Speed:           speed,
				})

				pr.lastUpdate = now
			}
		}
	}
	return n, err
}

// Updater manages Cursor downloads and version management
type Updater struct {
	downloadURL      string
	workDir          string
	launchLink       string
	progressCallback ProgressCallback
	config           *config.Config
}

// NewUpdater creates a new updater instance
func NewUpdater(downloadURL, workDir string, cfg *config.Config) *Updater {
	return &Updater{
		downloadURL: downloadURL,
		workDir:     workDir,
		launchLink:  filepath.Join(workDir, "Cursor.AppImage"),
		config:      cfg,
	}
}

// DownloadCursor downloads the latest Cursor version and returns the filename
func (u *Updater) DownloadCursor() (string, error) {
	// Get remote version
	remoteVersion, err := u.GetRemoteVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get remote version: %v", err)
	}

	// Construct filename using config pattern
	filename := u.GenerateFileName(remoteVersion)
	filepath := u.getDownloadPath(filename)

	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		return filename, nil
	}

	// Download the file
	resp, err := http.Get(u.downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %v", err)
	}
	defer resp.Body.Close()

	// Follow redirects to get the actual file
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		location := resp.Header.Get("Location")
		if location == "" {
			return "", fmt.Errorf("redirect location not found")
		}

		// Download from the redirected location
		resp, err = http.Get(location)
		if err != nil {
			return "", fmt.Errorf("failed to download from redirect: %v", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Ensure directories exist before creating the file
	if err := u.ensureDirectories(); err != nil {
		return "", fmt.Errorf("failed to ensure directories: %v", err)
	}

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Get total file size for progress tracking
	totalBytes := resp.ContentLength
	var bytesDownloaded int64

	// Create a progress reader
	progressReader := &ProgressReader{
		Reader:          resp.Body,
		TotalBytes:      totalBytes,
		BytesDownloaded: &bytesDownloaded,
		Callback:        u.progressCallback,
	}

	// Copy content to file with progress tracking
	_, err = io.Copy(file, progressReader)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}

	// Make file executable
	if err := os.Chmod(filepath, 0755); err != nil {
		return "", fmt.Errorf("failed to make file executable: %v", err)
	}

	return filename, nil
}

// GetRemoteVersion gets the remote version by following the download URL redirect
func (u *Updater) GetRemoteVersion() (string, error) {
	// Use a client that follows redirects but allows us to capture the final URL
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow all redirects automatically
			return nil
		},
	}

	// Make HEAD request that will follow all redirects
	resp, err := client.Head(u.downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to check redirect: %v", err)
	}
	defer resp.Body.Close()

	// The response URL will be the final URL after all redirects
	finalURL := resp.Request.URL.String()

	// Extract version from final URL
	version := version.SemverFromName(filepath.Base(finalURL))
	if version == "" {
		return "", fmt.Errorf("could not extract version from URL: %s", finalURL)
	}

	return version, nil
}

// GetLocalVersion gets the current local version from the symlink or regular file
func (u *Updater) GetLocalVersion() (string, error) {
	// Check if file exists
	if _, err := os.Stat(u.launchLink); os.IsNotExist(err) {
		return "", nil
	}

	// Try to read as symlink first
	if target, err := os.Readlink(u.launchLink); err == nil {
		// It's a symlink, extract version from target
		version := version.SemverFromName(filepath.Base(target))
		return version, nil
	}

	// It's a regular file, try to find matching version by size
	// This is a fallback for when Cursor.AppImage is a copy rather than symlink
	stat, err := os.Stat(u.launchLink)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %v", err)
	}

	// Look for version files with matching size (more reliable than timestamp)
	files, err := os.ReadDir(u.workDir)
	if err != nil {
		return "", fmt.Errorf("failed to read work directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		// Use config pattern for matching, fallback to default if no config
		if u.config != nil {
			// Extract version from filename using config pattern
			if !strings.Contains(u.config.FileNamePattern, "<version>") {
				continue
			}
			// Simple check: filename should contain a version number
			if !strings.Contains(filename, ".") {
				continue
			}
		} else {
			// Default pattern check
			if !strings.HasPrefix(filename, "Cursor-") || !strings.HasSuffix(filename, "-x86_64.AppImage") {
				continue
			}
		}

		// Check if this version file matches our current Cursor.AppImage by size
		versionFilePath := filepath.Join(u.workDir, filename)
		if versionStat, err := os.Stat(versionFilePath); err == nil {
			if versionStat.Size() == stat.Size() {
				// Found matching version file by size
				version := version.SemverFromName(filename)
				return version, nil
			}
		}
	}

	// Could not determine version
	return "", nil
}

// SwitchToVersion switches the symlink to point to a specific version
func (u *Updater) SwitchToVersion(version string) error {
	// Construct filename for the version using config pattern
	filename := u.GenerateFileName(version)
	filePath := u.getDownloadPath(filename)

	// Check if the version file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("version file not found: %s", filename)
	}

	// Get symlink path from config
	symlinkPath := u.getLatestSymlinkPath()

	// Remove existing symlink if it exists
	if _, err := os.Lstat(symlinkPath); err == nil {
		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %v", err)
		}
	}

	// Create new symlink - use relative path if possible
	symlinkDir := filepath.Dir(symlinkPath)
	relativePath, err := filepath.Rel(symlinkDir, filePath)
	if err != nil {
		// Fallback to absolute path if relative path calculation fails
		relativePath = filePath
	}

	if err := os.Symlink(relativePath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	return nil
}

// CheckForUpdates checks if an update is available
func (u *Updater) CheckForUpdates() (bool, string, error) {
	// Get remote version
	remoteVersion, err := u.GetRemoteVersion()
	if err != nil {
		return false, "", err
	}

	// Get local version
	localVersion, err := u.GetLocalVersion()
	if err != nil {
		return false, "", err
	}

	// If no local version, update is needed
	if localVersion == "" {
		return true, remoteVersion, nil
	}

	// Check if remote version is newer
	needsUpdate := version.LessThan(localVersion, remoteVersion)
	return needsUpdate, remoteVersion, nil
}

// WorkDir returns the working directory of the updater
func (u *Updater) WorkDir() string {
	return u.workDir
}

// SetProgressCallback sets the callback function for progress updates
func (u *Updater) SetProgressCallback(callback ProgressCallback) {
	u.progressCallback = callback
}

// getDownloadPath returns the full path for a download file
func (u *Updater) getDownloadPath(filename string) string {
	if u.config != nil {
		return filepath.Join(u.config.DownloadDir, filename)
	}
	return filepath.Join(u.workDir, filename)
}

// GenerateFileName generates a filename based on config pattern and version
func (u *Updater) GenerateFileName(version string) string {
	if u.config != nil {
		return u.config.GenerateFileName(version)
	}
	return fmt.Sprintf("Cursor-%s-x86_64.AppImage", version)
}

// getLedgerPath returns the ledger path from config
func (u *Updater) getLedgerPath() string {
	if u.config != nil {
		return u.config.LedgerPath
	}
	return filepath.Join(u.workDir, ".cursor-versions.log")
}

// getLatestSymlinkPath returns the latest symlink path from config
func (u *Updater) getLatestSymlinkPath() string {
	if u.config != nil {
		return u.config.LatestSymlink
	}
	return u.launchLink
}

// ensureDirectories creates necessary directories
func (u *Updater) ensureDirectories() error {
	// Always ensure the work directory exists
	if err := os.MkdirAll(u.workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %v", err)
	}

	// If config exists, also ensure config-specific directories
	if u.config != nil {
		// Ensure download directory exists
		if err := os.MkdirAll(u.config.DownloadDir, 0755); err != nil {
			return fmt.Errorf("failed to create download directory: %v", err)
		}

		// Ensure symlink directory exists
		symlinkDir := filepath.Dir(u.config.LatestSymlink)
		if err := os.MkdirAll(symlinkDir, 0755); err != nil {
			return fmt.Errorf("failed to create symlink directory: %v", err)
		}

		// Ensure config directory exists
		configDir := filepath.Dir(u.config.LedgerPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %v", err)
		}
	}

	return nil
}

// CalculateSHA256 calculates the SHA256 hash of a file
func (u *Updater) CalculateSHA256(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %v", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
