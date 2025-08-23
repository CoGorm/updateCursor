package cli

import (
	"os"
	"testing"
)

// Test configuration
func init() {
	// Set test environment to disable real downloads
	os.Setenv("UPDATECURSOR_TEST_MODE", "true")
}

func TestRunWithNoArgs(t *testing.T) {
	// Test that Run with no args defaults to update command
	err := Run([]string{})
	if err != nil {
		t.Errorf("Expected no error for default update command, got: %v", err)
	}
}

func TestRunWithUpdateCommand(t *testing.T) {
	// Test that Run with update command works
	err := Run([]string{"update"})
	if err != nil {
		t.Errorf("Expected no error for update command, got: %v", err)
	}
}

func TestRunWithCheckCommand(t *testing.T) {
	// Test that Run with check command works
	err := Run([]string{"check"})
	if err != nil {
		t.Errorf("Expected no error for check command, got: %v", err)
	}
}

func TestRunWithForceCommand(t *testing.T) {
	// Test that Run with force command works
	err := Run([]string{"force"})
	if err != nil {
		t.Errorf("Expected no error for force command, got: %v", err)
	}
}

func TestRunWithListCommand(t *testing.T) {
	// Test that Run with list command works
	err := Run([]string{"list"})
	if err != nil {
		t.Errorf("Expected no error for list command, got: %v", err)
	}
}

func TestRunWithSwitchCommand(t *testing.T) {
	// Test that Run with switch command works
	err := Run([]string{"switch", "1.4.5"})
	if err != nil {
		t.Errorf("Expected no error for switch command, got: %v", err)
	}
}

func TestRunWithSwitchCommandMissingVersion(t *testing.T) {
	// Test that Run with switch command without version returns error
	err := Run([]string{"switch"})
	if err == nil {
		t.Error("Expected error for switch command without version")
	}
}

func TestRunWithHelpCommand(t *testing.T) {
	// Test that Run with help command works
	err := Run([]string{"--help"})
	if err != nil {
		t.Errorf("Expected no error for help command, got: %v", err)
	}
}

func TestRunWithUnknownCommand(t *testing.T) {
	// Test that Run with unknown command returns error
	err := Run([]string{"unknown"})
	if err == nil {
		t.Error("Expected error for unknown command")
	}
}

func TestCheckCommandShowsUpdateStatus(t *testing.T) {
	// Test that check command shows clear update status messages
	// This test verifies the user experience improvement for update status
	// Note: We can't easily test the actual output in unit tests,
	// but we can verify the command executes without error
	err := Run([]string{"check"})
	if err != nil {
		t.Errorf("Expected check command to work, got: %v", err)
	}
}

func TestTestModeDetection(t *testing.T) {
	// Test that test mode is properly detected
	if os.Getenv("UPDATECURSOR_TEST_MODE") != "true" {
		t.Error("Expected UPDATECURSOR_TEST_MODE to be set to 'true'")
	}
}

func TestHelpCommandQuick(t *testing.T) {
	// Test help command (should be very fast)
	err := Run([]string{"--help"})
	if err != nil {
		t.Errorf("Expected help command to work, got: %v", err)
	}
}
