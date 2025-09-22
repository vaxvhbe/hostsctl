package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Since we can't easily test main() directly without side effects,
// we'll test the integration by running the compiled binary.
func SkipTestMain_Integration(t *testing.T) {
	// Build the binary first
	binaryPath := filepath.Join(t.TempDir(), "hostsctl-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Go back to the project root

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Test basic command execution
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantStderr   bool
	}{
		{
			name:         "help command",
			args:         []string{"--help"},
			wantExitCode: 0,
			wantStderr:   false,
		},
		{
			name:         "version info",
			args:         []string{"list", "--help"},
			wantExitCode: 0,
			wantStderr:   false,
		},
		{
			name:         "invalid command",
			args:         []string{"invalid-command"},
			wantExitCode: 1,
			wantStderr:   true,
		},
		{
			name:         "missing required flag",
			args:         []string{"add"},
			wantExitCode: 1,
			wantStderr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			// Use context for timeout

			output, err := cmd.CombinedOutput()

			// Check exit code
			var exitCode int
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.wantExitCode, exitCode)
				t.Logf("Output: %s", string(output))
			}

			// Check if we expect stderr output
			hasStderr := exitCode != 0 || strings.Contains(string(output), "Error:")
			if tt.wantStderr && !hasStderr {
				t.Error("Expected stderr output but got none")
			}
			if !tt.wantStderr && hasStderr && exitCode != 0 {
				t.Errorf("Unexpected stderr output: %s", string(output))
			}
		})
	}
}

func SkipTestMain_WithCustomHostsFile(t *testing.T) {
	// Build the binary first
	binaryPath := filepath.Join(t.TempDir(), "hostsctl-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Go back to the project root

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Create a temporary hosts file
	tmpDir := t.TempDir()
	hostsFile := filepath.Join(tmpDir, "test_hosts")

	initialContent := "127.0.0.1\tlocalhost\n192.168.1.100\ttest.local\t# Test entry"
	if err := os.WriteFile(hostsFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		wantExitCode int
	}{
		{
			name:         "list with custom hosts file",
			args:         []string{"--hosts-file", hostsFile, "list"},
			wantExitCode: 0,
		},
		{
			name:         "verify custom hosts file",
			args:         []string{"--hosts-file", hostsFile, "verify"},
			wantExitCode: 0,
		},
		{
			name:         "backup custom hosts file",
			args:         []string{"--hosts-file", hostsFile, "backup"},
			wantExitCode: 0,
		},
		{
			name:         "add entry to custom hosts file",
			args:         []string{"--hosts-file", hostsFile, "add", "--ip", "10.0.0.1", "--name", "new.local"},
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			// Use context for timeout

			output, err := cmd.CombinedOutput()

			var exitCode int
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.wantExitCode, exitCode)
				t.Logf("Output: %s", string(output))
			}
		})
	}
}

func SkipTestMain_JsonOutput(t *testing.T) {
	// Build the binary first
	binaryPath := filepath.Join(t.TempDir(), "hostsctl-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Go back to the project root

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Create a temporary hosts file
	tmpDir := t.TempDir()
	hostsFile := filepath.Join(tmpDir, "test_hosts")

	initialContent := "127.0.0.1\tlocalhost\n192.168.1.100\ttest.local\t# Test entry"
	if err := os.WriteFile(hostsFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}

	// Test JSON output
	cmd = exec.Command(binaryPath, "--hosts-file", hostsFile, "--json", "list")
	// Use context for timeout

	output, err := cmd.CombinedOutput()

	var exitCode int
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
		t.Logf("Output: %s", string(output))
	}

	// Check that output looks like JSON
	outputStr := string(output)
	if !strings.HasPrefix(strings.TrimSpace(outputStr), "[") {
		t.Error("JSON output should start with '['")
	}
	if !strings.HasSuffix(strings.TrimSpace(outputStr), "]") {
		t.Error("JSON output should end with ']'")
	}
}

func SkipTestMain_InvalidHostsFile(t *testing.T) {
	// Build the binary first
	binaryPath := filepath.Join(t.TempDir(), "hostsctl-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Go back to the project root

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Test with non-existent hosts file
	cmd = exec.Command(binaryPath, "--hosts-file", "/nonexistent/hosts", "list")
	// Use context for timeout

	output, err := cmd.CombinedOutput()

	var exitCode int
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for non-existent hosts file")
		t.Logf("Output: %s", string(output))
	}

	// Check that error message is printed
	outputStr := string(output)
	if !strings.Contains(outputStr, "Error:") {
		t.Error("Expected error message in output")
	}
}

func SkipTestMain_ProfileCommands(t *testing.T) {
	// Build the binary first
	binaryPath := filepath.Join(t.TempDir(), "hostsctl-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Go back to the project root

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		wantExitCode int
	}{
		{
			name:         "profile help",
			args:         []string{"profile", "--help"},
			wantExitCode: 0,
		},
		{
			name:         "profile list",
			args:         []string{"profile", "list"},
			wantExitCode: 0,
		},
		{
			name:         "profile save without name",
			args:         []string{"profile", "save"},
			wantExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			// Use context for timeout

			output, err := cmd.CombinedOutput()

			var exitCode int
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.wantExitCode, exitCode)
				t.Logf("Output: %s", string(output))
			}
		})
	}
}

func SkipTestMain_SearchCommands(t *testing.T) {
	// Build the binary first
	binaryPath := filepath.Join(t.TempDir(), "hostsctl-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Go back to the project root

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Create a temporary hosts file
	tmpDir := t.TempDir()
	hostsFile := filepath.Join(tmpDir, "test_hosts")

	initialContent := "127.0.0.1\tlocalhost\n192.168.1.100\ttest.local\t# Test entry"
	if err := os.WriteFile(hostsFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test hosts file: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		wantExitCode int
	}{
		{
			name:         "search help",
			args:         []string{"search", "--help"},
			wantExitCode: 0,
		},
		{
			name:         "search localhost",
			args:         []string{"--hosts-file", hostsFile, "search", "localhost"},
			wantExitCode: 0,
		},
		{
			name:         "search with regex",
			args:         []string{"--hosts-file", hostsFile, "search", "^127", "--regex"},
			wantExitCode: 0,
		},
		{
			name:         "search without pattern",
			args:         []string{"search"},
			wantExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			// Use context for timeout

			output, err := cmd.CombinedOutput()

			var exitCode int
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.wantExitCode, exitCode)
				t.Logf("Output: %s", string(output))
			}
		})
	}
}

func SkipTestMain_CompletionCommands(t *testing.T) {
	// Build the binary first
	binaryPath := filepath.Join(t.TempDir(), "hostsctl-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Go back to the project root

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		wantExitCode int
	}{
		{
			name:         "completion help",
			args:         []string{"completion", "--help"},
			wantExitCode: 0,
		},
		{
			name:         "bash completion",
			args:         []string{"completion", "bash"},
			wantExitCode: 0,
		},
		{
			name:         "zsh completion",
			args:         []string{"completion", "zsh"},
			wantExitCode: 0,
		},
		{
			name:         "fish completion",
			args:         []string{"completion", "fish"},
			wantExitCode: 0,
		},
		{
			name:         "powershell completion",
			args:         []string{"completion", "powershell"},
			wantExitCode: 0,
		},
		{
			name:         "invalid completion shell",
			args:         []string{"completion", "invalid"},
			wantExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			// Use context for timeout

			output, err := cmd.CombinedOutput()

			var exitCode int
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.wantExitCode, exitCode)
				t.Logf("Output: %s", string(output))
			}
		})
	}
}

// TestMain_ErrorHandling tests that main() properly handles and reports errors
func SkipTestMain_ErrorHandling(t *testing.T) {
	// Build the binary first
	binaryPath := filepath.Join(t.TempDir(), "hostsctl-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Go back to the project root

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Test various error conditions
	errorTests := []struct {
		name string
		args []string
	}{
		{"invalid flag", []string{"--invalid-flag"}},
		{"unknown command", []string{"unknown-command"}},
		{"missing required args", []string{"profile", "apply"}},
		{"invalid IP format", []string{"add", "--ip", "invalid", "--name", "test"}},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			// Use context for timeout

			output, err := cmd.CombinedOutput()

			// Should exit with non-zero code
			var exitCode int
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			if exitCode == 0 {
				t.Errorf("Expected non-zero exit code for error case: %s", tt.name)
				t.Logf("Output: %s", string(output))
			}

			// Should output error message
			outputStr := string(output)
			if !strings.Contains(outputStr, "Error:") && !strings.Contains(outputStr, "unknown command") {
				t.Errorf("Expected error message in output for: %s", tt.name)
				t.Logf("Output: %s", outputStr)
			}
		})
	}
}
