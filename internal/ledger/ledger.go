package ledger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry represents a single entry in the update ledger
type Entry struct {
	Timestamp  time.Time
	Version    string
	InternalID string
	Filename   string
	SHA256     string
	Action     string
}

// Ledger manages the update history file
type Ledger struct {
	filepath string
}

// NewLedger creates a new ledger instance
func NewLedger(filepath string) *Ledger {
	return &Ledger{
		filepath: filepath,
	}
}

// Append adds a new entry to the ledger
func (l *Ledger) Append(entry Entry) error {
	// Ensure directory exists
	dir := filepath.Dir(l.filepath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Open file for appending, create if doesn't exist
	file, err := os.OpenFile(l.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open ledger file: %v", err)
	}
	defer file.Close()

	// Format entry as TSV line
	line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n",
		entry.Timestamp.UTC().Format(time.RFC3339),
		entry.Version,
		entry.InternalID,
		entry.Filename,
		entry.SHA256,
		entry.Action,
	)

	// Write line to file
	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to ledger file: %v", err)
	}

	return nil
}

// ReadAll reads all entries from the ledger
func (l *Ledger) ReadAll() ([]Entry, error) {
	// Check if file exists
	if _, err := os.Stat(l.filepath); os.IsNotExist(err) {
		return []Entry{}, nil
	}

	// Open file for reading
	file, err := os.Open(l.filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ledger file: %v", err)
	}
	defer file.Close()

	var entries []Entry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		entry, err := parseEntry(line)
		if err != nil {
			// Skip malformed lines but continue reading
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading ledger file: %v", err)
	}

	return entries, nil
}

// FindByInternalID finds all entries with the given internal ID
func (l *Ledger) FindByInternalID(id string) ([]Entry, error) {
	entries, err := l.ReadAll()
	if err != nil {
		return nil, err
	}

	var found []Entry
	for _, entry := range entries {
		if entry.InternalID == id {
			found = append(found, entry)
		}
	}

	return found, nil
}

// GetLatestVersion returns the most recent entry based on timestamp
func (l *Ledger) GetLatestVersion() (Entry, error) {
	entries, err := l.ReadAll()
	if err != nil {
		return Entry{}, err
	}

	if len(entries) == 0 {
		return Entry{}, fmt.Errorf("no entries found in ledger")
	}

	// Find entry with latest timestamp
	latest := entries[0]
	for _, entry := range entries[1:] {
		if entry.Timestamp.After(latest.Timestamp) {
			latest = entry
		}
	}

	return latest, nil
}

// parseEntry parses a TSV line into an Entry struct
func parseEntry(line string) (Entry, error) {
	parts := strings.Split(line, "\t")
	if len(parts) != 6 {
		return Entry{}, fmt.Errorf("invalid entry format: expected 6 parts, got %d", len(parts))
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return Entry{}, fmt.Errorf("invalid timestamp format: %v", err)
	}

	return Entry{
		Timestamp:  timestamp,
		Version:    parts[1],
		InternalID: parts[2],
		Filename:   parts[3],
		SHA256:     parts[4],
		Action:     parts[5],
	}, nil
}
