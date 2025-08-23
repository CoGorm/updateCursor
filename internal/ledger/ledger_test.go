package ledger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLedgerEntry(t *testing.T) {
	entry := Entry{
		Timestamp:  time.Now(),
		Version:    "1.0.0",
		InternalID: "12345",
		Filename:   "Cursor-1.0.0-x86_64.AppImage",
		SHA256:     "abc123def456",
		Action:     "download",
	}

	if entry.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", entry.Version)
	}

	if entry.Action != "download" {
		t.Errorf("Expected action download, got %s", entry.Action)
	}
}

func TestLedgerWriteAndRead(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	ledgerPath := filepath.Join(tempDir, "test-ledger.log")

	ledger := NewLedger(ledgerPath)

	// Test writing entries
	entry1 := Entry{
		Timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Version:    "1.0.0",
		InternalID: "12345",
		Filename:   "Cursor-1.0.0-x86_64.AppImage",
		SHA256:     "abc123def456",
		Action:     "download",
	}

	entry2 := Entry{
		Timestamp:  time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
		Version:    "1.1.0",
		InternalID: "67890",
		Filename:   "Cursor-1.1.0-x86_64.AppImage",
		SHA256:     "def456ghi789",
		Action:     "download",
	}

	err := ledger.Append(entry1)
	if err != nil {
		t.Fatalf("Failed to append entry1: %v", err)
	}

	err = ledger.Append(entry2)
	if err != nil {
		t.Fatalf("Failed to append entry2: %v", err)
	}

	// Test reading entries
	entries, err := ledger.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Verify first entry
	if entries[0].Version != "1.0.0" {
		t.Errorf("Expected first entry version 1.0.0, got %s", entries[0].Version)
	}

	// Verify second entry
	if entries[1].Version != "1.1.0" {
		t.Errorf("Expected second entry version 1.1.0, got %s", entries[1].Version)
	}
}

func TestLedgerFindByInternalID(t *testing.T) {
	tempDir := t.TempDir()
	ledgerPath := filepath.Join(tempDir, "test-ledger.log")

	ledger := NewLedger(ledgerPath)

	// Add test entries
	entries := []Entry{
		{
			Timestamp:  time.Now(),
			Version:    "1.0.0",
			InternalID: "12345",
			Filename:   "Cursor-1.0.0-x86_64.AppImage",
			SHA256:     "abc123",
			Action:     "download",
		},
		{
			Timestamp:  time.Now(),
			Version:    "1.1.0",
			InternalID: "67890",
			Filename:   "Cursor-1.1.0-x86_64.AppImage",
			SHA256:     "def456",
			Action:     "download",
		},
		{
			Timestamp:  time.Now(),
			Version:    "1.0.0",
			InternalID: "12345",
			Filename:   "Cursor-1.0.0-x86_64.AppImage",
			SHA256:     "abc123",
			Action:     "switch",
		},
	}

	for _, entry := range entries {
		err := ledger.Append(entry)
		if err != nil {
			t.Fatalf("Failed to append entry: %v", err)
		}
	}

	// Test finding by internal ID
	found, err := ledger.FindByInternalID("12345")
	if err != nil {
		t.Fatalf("Failed to find by internal ID: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 entries for internal ID 12345, got %d", len(found))
	}

	// Test finding non-existent internal ID
	notFound, err := ledger.FindByInternalID("99999")
	if err != nil {
		t.Fatalf("Failed to find by internal ID: %v", err)
	}

	if len(notFound) != 0 {
		t.Errorf("Expected 0 entries for non-existent internal ID, got %d", len(notFound))
	}
}

func TestLedgerGetLatestVersion(t *testing.T) {
	tempDir := t.TempDir()
	ledgerPath := filepath.Join(tempDir, "test-ledger.log")

	ledger := NewLedger(ledgerPath)

	// Add test entries with different timestamps
	entries := []Entry{
		{
			Timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Version:    "1.0.0",
			InternalID: "12345",
			Filename:   "Cursor-1.0.0-x86_64.AppImage",
			SHA256:     "abc123",
			Action:     "download",
		},
		{
			Timestamp:  time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			Version:    "1.1.0",
			InternalID: "67890",
			Filename:   "Cursor-1.1.0-x86_64.AppImage",
			SHA256:     "def456",
			Action:     "download",
		},
	}

	for _, entry := range entries {
		err := ledger.Append(entry)
		if err != nil {
			t.Fatalf("Failed to append entry: %v", err)
		}
	}

	// Test getting latest version
	latest, err := ledger.GetLatestVersion()
	if err != nil {
		t.Fatalf("Failed to get latest version: %v", err)
	}

	if latest.Version != "1.1.0" {
		t.Errorf("Expected latest version 1.1.0, got %s", latest.Version)
	}
}

func TestLedgerFileFormat(t *testing.T) {
	tempDir := t.TempDir()
	ledgerPath := filepath.Join(tempDir, "test-ledger.log")

	ledger := NewLedger(ledgerPath)

	entry := Entry{
		Timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Version:    "1.0.0",
		InternalID: "12345",
		Filename:   "Cursor-1.0.0-x86_64.AppImage",
		SHA256:     "abc123def456",
		Action:     "download",
	}

	err := ledger.Append(entry)
	if err != nil {
		t.Fatalf("Failed to append entry: %v", err)
	}

	// Read the raw file content to verify TSV format
	content, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to read ledger file: %v", err)
	}

	expectedLine := "2024-01-01T12:00:00Z\t1.0.0\t12345\tCursor-1.0.0-x86_64.AppImage\tabc123def456\tdownload\n"
	if string(content) != expectedLine {
		t.Errorf("Expected TSV format:\n%q\ngot:\n%q", expectedLine, string(content))
	}
}
