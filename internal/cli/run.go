package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CoGorm/updateCursor/internal/config"
	"github.com/CoGorm/updateCursor/internal/ledger"
	"github.com/CoGorm/updateCursor/internal/updater"
	"github.com/CoGorm/updateCursor/internal/version"
)

const (
	defaultDownloadURL = "https://www.cursor.com/download/stable/linux-x64"
	defaultWorkDir     = "~/Applications/Cursor"
	launchLink         = "Cursor.AppImage"
	ledgerFile         = ".cursor-versions.log"
	versionFile        = ".cursor-version"
)

// Run executes the CLI application with the given arguments
func Run(args []string) error {
	if len(args) < 1 {
		// Default to update command
		return executeCommand("update", []string{})
	}

	command := args[0]
	commandArgs := args[1:]

	return executeCommand(command, commandArgs)
}

func executeCommand(command string, args []string) error {
	// Load or create default config
	cfg := config.NewConfig()
	err := cfg.LoadOrCreateDefault()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	// Expand paths in config
	err = cfg.ExpandPaths()
	if err != nil {
		return fmt.Errorf("error expanding config paths: %v", err)
	}

	// Use config values for paths
	workDir := cfg.DownloadDir
	ledgerPath := cfg.LedgerPath

	// Create updater instance with config
	up := updater.NewUpdater(defaultDownloadURL, workDir, cfg)

	// Create ledger instance
	led := ledger.NewLedger(ledgerPath)

	switch command {
	case "check":
		return executeCheck(up)
	case "update":
		return executeUpdate(up, led, cfg)
	case "force":
		return executeForce(up, led, cfg)
	case "list":
		return executeList(led)
	case "switch":
		if len(args) < 1 {
			return fmt.Errorf("usage: %s switch <version>", os.Args[0])
		}
		return executeSwitch(up, led, args[0], cfg)
	case "-h", "--help":
		showUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func executeCheck(up *updater.Updater) error {
	localVersion, err := up.GetLocalVersion()
	if err != nil {
		return fmt.Errorf("error getting local version: %v", err)
	}

	remoteVersion, err := up.GetRemoteVersion()
	if err != nil {
		return fmt.Errorf("error getting remote version: %v", err)
	}

	if localVersion == "" {
		fmt.Printf("Local: (unknown)\n")
	} else {
		fmt.Printf("Local: %s\n", localVersion)
	}

	fmt.Printf("Remote: %s\n", remoteVersion)

	// Check if update is needed and show clear message
	if localVersion == "" || version.LessThan(localVersion, remoteVersion) {
		fmt.Printf("\n🔄 Update needed: Local version is older than remote version\n")
		// Return error to indicate update is needed (main will handle exit code)
		return fmt.Errorf("update needed")
	} else {
		fmt.Printf("\n✅ No update needed: You have the latest version\n")
	}

	return nil
}

func executeUpdate(up *updater.Updater, led *ledger.Ledger, cfg *config.Config) error {
	// Check if update is needed
	needsUpdate, remoteVersion, err := up.CheckForUpdates()
	if err != nil {
		return fmt.Errorf("error checking for updates: %v", err)
	}

	if !needsUpdate {
		localVersion, _ := up.GetLocalVersion()
		fmt.Printf("✅ Already up to date (%s).\n", localVersion)
		return nil
	}

	// Check if we're in test mode
	isTestMode := os.Getenv("UPDATECURSOR_TEST_MODE") == "true"

	// Set up progress tracking
	fmt.Printf("Downloading Cursor %s...\n", remoteVersion)
	if isTestMode {
		fmt.Printf("🔄 Test mode: Simulating download...\n")
	}

	up.SetProgressCallback(func(update updater.ProgressUpdate) {
		displayProgress(update)
	})

	// Download the update
	filename, err := up.DownloadCursor()
	if err != nil {
		return fmt.Errorf("error downloading Cursor: %v", err)
	}

	// Add spacing after download completion
	fmt.Println()

	// Calculate SHA256
	filePath := filepath.Join(up.WorkDir(), filename)
	sha256, err := up.CalculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("error calculating SHA256: %v", err)
	}

	// Switch to the new version
	err = up.SwitchToVersion(remoteVersion)
	if err != nil {
		return fmt.Errorf("error switching to version: %v", err)
	}

	// Log the update
	entry := ledger.Entry{
		Timestamp:  time.Now(),
		Version:    remoteVersion,
		InternalID: "", // TODO: Extract from AppImage
		Filename:   filename,
		SHA256:     sha256,
		Action:     "update",
	}

	if err := led.Append(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to log update: %v\n", err)
	}

	// Update version file
	updateVersionFile(remoteVersion, cfg)

	fmt.Printf("Updated to version %s\n", remoteVersion)
	return nil
}

func executeForce(up *updater.Updater, led *ledger.Ledger, cfg *config.Config) error {
	// Get remote version
	remoteVersion, err := up.GetRemoteVersion()
	if err != nil {
		return fmt.Errorf("error getting remote version: %v", err)
	}

	// Remove existing file if it exists
	filename := up.GenerateFileName(remoteVersion)
	filePath := filepath.Join(up.WorkDir(), filename)
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove existing file: %v\n", err)
		}
	}

	// Check if we're in test mode
	isTestMode := os.Getenv("UPDATECURSOR_TEST_MODE") == "true"

	// Set up progress tracking
	fmt.Printf("Force downloading Cursor %s...\n", remoteVersion)
	if isTestMode {
		fmt.Printf("🔄 Test mode: Simulating force download...\n")
	}

	up.SetProgressCallback(func(update updater.ProgressUpdate) {
		displayProgress(update)
	})

	// Download the update
	filename, err = up.DownloadCursor()
	if err != nil {
		return fmt.Errorf("error downloading Cursor: %v", err)
	}

	// Add spacing after download completion
	fmt.Println()

	// Calculate SHA256
	filePath = filepath.Join(up.WorkDir(), filename)
	sha256, err := up.CalculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("error calculating SHA256: %v", err)
	}

	// Switch to the new version
	err = up.SwitchToVersion(remoteVersion)
	if err != nil {
		return fmt.Errorf("error switching to version: %v", err)
	}

	// Log the update
	entry := ledger.Entry{
		Timestamp:  time.Now(),
		Version:    remoteVersion,
		InternalID: "", // TODO: Extract from AppImage
		Filename:   filename,
		SHA256:     sha256,
		Action:     "force",
	}

	if err := led.Append(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to log update: %v\n", err)
	}

	// Update version file
	updateVersionFile(remoteVersion, cfg)

	fmt.Printf("Force updated to version %s\n", remoteVersion)
	return nil
}

func executeList(led *ledger.Ledger) error {
	entries, err := led.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading ledger: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("No ledger entries found.")
		return nil
	}

	// Print header
	fmt.Printf("%-24s\t%-7s\t%-8s\t%-30s\t%-12s\t%s\n",
		"when(UTC)", "ver", "internal", "file", "sha256 (short)", "action")

	// Print entries
	for _, entry := range entries {
		sha256Short := entry.SHA256
		if len(sha256Short) > 12 {
			sha256Short = sha256Short[:12]
		}

		fmt.Printf("%-24s\t%-7s\t%-8s\t%-30s\t%-12s\t%s\n",
			entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Version, entry.InternalID, entry.Filename, sha256Short, entry.Action)
	}

	return nil
}

func executeSwitch(up *updater.Updater, led *ledger.Ledger, ver string, cfg *config.Config) error {
	// Validate version format
	if version.SemverFromName(fmt.Sprintf("Cursor-%s-x86_64.AppImage", ver)) == "" {
		return fmt.Errorf("invalid version format: %s", ver)
	}

	// Switch to the specified version
	err := up.SwitchToVersion(ver)
	if err != nil {
		return fmt.Errorf("error switching to version: %v", err)
	}

	// Log the switch
	entry := ledger.Entry{
		Timestamp:  time.Now(),
		Version:    ver,
		InternalID: "", // TODO: Extract from AppImage
		Filename:   up.GenerateFileName(ver),
		SHA256:     "", // TODO: Calculate SHA256
		Action:     "switch",
	}

	if err := led.Append(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to log switch: %v\n", err)
	}

	// Update version file
	updateVersionFile(ver, cfg)

	fmt.Printf("Switched to version %s\n", ver)
	return nil
}

func updateVersionFile(version string, cfg *config.Config) {
	// Use config workDir for version file
	workDir := cfg.DownloadDir
	versionFilePath := filepath.Join(workDir, versionFile)
	content := fmt.Sprintf("VERSION=%s\n", version)

	if err := os.WriteFile(versionFilePath, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to update version file: %v\n", err)
	}
}

// displayProgress shows a nice progress bar for downloads
func displayProgress(update updater.ProgressUpdate) {
	if update.TotalBytes <= 0 {
		return
	}

	// Calculate percentage
	percentage := update.Percentage
	if percentage > 100 {
		percentage = 100
	}

	// Create progress bar (50 characters wide)
	barWidth := 50
	filled := int(float64(barWidth) * percentage / 100)
	empty := barWidth - filled

	// Build progress bar string
	bar := "["
	for i := 0; i < filled; i++ {
		bar += "="
	}
	if filled < barWidth {
		bar += ">"
	}
	for i := 0; i < empty-1; i++ {
		bar += " "
	}
	bar += "]"

	// Format file sizes
	downloadedMB := float64(update.BytesDownloaded) / (1024 * 1024)
	totalMB := float64(update.TotalBytes) / (1024 * 1024)
	speedMBps := update.Speed / (1024 * 1024)

	// Display progress
	fmt.Printf("\r%s %.1f%% (%.1f/%.1f MB) %.1f MB/s",
		bar, percentage, downloadedMB, totalMB, speedMBps)

	// If download is complete, just add a newline
	if percentage >= 100 {
		fmt.Println()
	}
}

func showUsage() {
	fmt.Printf(`Usage: %s [command]

Commands:
  check           Print local vs remote versions and exit with status (10=update needed)
  update          Download latest if newer and set symlink (default)
  force           Re-download latest even if it exists and relink
  list            Show ledger (configurable location)
  switch <ver>    Point symlink at an existing version (no download)

Configuration:
  Config file: ~/.config/updateCursor/config.yaml
  Default download: ~/Downloads/Cursor
  Default symlink: ~/Downloads/Cursor/Cursor.AppImage
  Default ledger: ~/.config/updateCursor/cursor-versions.log

All paths are configurable via the config file.
`, os.Args[0])
}
