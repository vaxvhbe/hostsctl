package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"unicode"

	"github.com/vaxvhbe/hostsctl/internal/hosts"
)

func TestCLI_matchesPattern(t *testing.T) {
	cli := NewCLI()

	tests := []struct {
		text    string
		pattern string
		want    bool
	}{
		{"test.local", "*.local", true},
		{"example.com", "*.local", false},
		{"192.168.1.1", "192.168.*", true},
		{"10.0.0.1", "192.168.*", false},
		{"development", "*dev*", true},
		{"production", "*dev*", false},
		{"localhost", "local", true},
		{"example.com", "local", false},
		{"", "", true},
		{"test", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.text+"_"+tt.pattern, func(t *testing.T) {
			got := cli.matchesPattern(tt.text, tt.pattern)
			if got != tt.want {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.text, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestCLI_matchWildcard(t *testing.T) {
	cli := NewCLI()

	tests := []struct {
		text    string
		pattern string
		want    bool
	}{
		{"test.local", "*.local", true},
		{"app.test.local", "*.local", true},
		{"localhost", "*.local", false},
		{"api.dev", "api.*", true},
		{"api.v1.com", "api.*", true},
		{"test-api.com", "api.*", false},
		{"192.168.1.1", "192.168.*", true},
		{"192.168.1.100", "192.168.*", true},
		{"10.0.0.1", "192.168.*", false},
		{"development", "*dev*", true},
		{"devtools", "*dev*", true},
		{"production", "*dev*", false},
		{"test.com", "*", true},
	}

	for _, tt := range tests {
		t.Run(tt.text+"_"+tt.pattern, func(t *testing.T) {
			got := cli.matchWildcard(tt.text, tt.pattern)
			if got != tt.want {
				t.Errorf("matchWildcard(%q, %q) = %v, want %v", tt.text, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestCLI_applyListFilters(t *testing.T) {
	cli := NewCLI()

	entries := []hosts.Entry{
		{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local machine", Disabled: false},
		{ID: 2, IP: "192.168.1.100", Names: []string{"server.local"}, Comment: "Development server", Disabled: false},
		{ID: 3, IP: "192.168.1.200", Names: []string{"disabled.local"}, Comment: "Disabled entry", Disabled: true},
		{ID: 4, IP: "10.0.0.1", Names: []string{"api.prod.com"}, Comment: "Production API", Disabled: false},
	}

	tests := []struct {
		name    string
		filters ListFilters
		wantLen int
		wantIDs []int
	}{
		{
			name:    "show all entries",
			filters: ListFilters{ShowAll: true},
			wantLen: 4,
			wantIDs: []int{1, 2, 3, 4},
		},
		{
			name:    "show only enabled",
			filters: ListFilters{ShowAll: false},
			wantLen: 3,
			wantIDs: []int{1, 2, 4},
		},
		{
			name:    "filter by IP pattern",
			filters: ListFilters{ShowAll: true, IPFilter: "192.168.*"},
			wantLen: 2,
			wantIDs: []int{2, 3},
		},
		{
			name:    "filter by status enabled",
			filters: ListFilters{ShowAll: true, StatusFilter: "enabled"},
			wantLen: 3,
			wantIDs: []int{1, 2, 4},
		},
		{
			name:    "filter by status disabled",
			filters: ListFilters{ShowAll: true, StatusFilter: "disabled"},
			wantLen: 1,
			wantIDs: []int{3},
		},
		{
			name:    "filter by comment",
			filters: ListFilters{ShowAll: true, CommentFilter: "development"},
			wantLen: 1,
			wantIDs: []int{2},
		},
		{
			name:    "filter by hostname",
			filters: ListFilters{ShowAll: true, NameFilter: "*.local"},
			wantLen: 2,
			wantIDs: []int{2, 3},
		},
		{
			name:    "filter by hostname substring",
			filters: ListFilters{ShowAll: true, NameFilter: "local"},
			wantLen: 3,
			wantIDs: []int{1, 2, 3},
		},
		{
			name:    "multiple filters",
			filters: ListFilters{ShowAll: false, IPFilter: "192.168.*", StatusFilter: "enabled"},
			wantLen: 1,
			wantIDs: []int{2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cli.applyListFilters(entries, tt.filters)
			if len(got) != tt.wantLen {
				t.Errorf("applyListFilters() len = %v, want %v", len(got), tt.wantLen)
			}

			gotIDs := make([]int, len(got))
			for i, entry := range got {
				gotIDs[i] = entry.ID
			}

			if len(gotIDs) == len(tt.wantIDs) {
				for i, id := range gotIDs {
					if id != tt.wantIDs[i] {
						t.Errorf("applyListFilters() IDs = %v, want %v", gotIDs, tt.wantIDs)
						break
					}
				}
			}
		})
	}
}

func TestCLI_runListWithFilters(t *testing.T) {
	// Create a temporary hosts file for testing
	tmpDir, err := os.MkdirTemp("", "hostsctl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	content := `127.0.0.1	localhost
192.168.1.100	server.local	# Test server
# 192.168.1.200	disabled.local	# Disabled entry`

	if err := os.WriteFile(hostsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write hosts file: %v", err)
	}

	cli := NewCLI()
	cli.hostsFile = hostsFile
	cli.jsonOutput = true

	// Test with empty filters
	filters := ListFilters{ShowAll: true}
	err = cli.runListWithFilters(filters)
	if err != nil {
		t.Errorf("runListWithFilters() error = %v", err)
	}
}

func TestCLI_buildCommands(t *testing.T) {
	cli := NewCLI()

	// Test that all commands can be built without errors
	rootCmd := cli.buildRootCommand()
	if rootCmd == nil {
		t.Error("buildRootCommand() returned nil")
	}

	// Test individual command builders
	tests := []struct {
		name    string
		builder func() interface{}
	}{
		{"list", func() interface{} { return cli.buildListCommand() }},
		{"add", func() interface{} { return cli.buildAddCommand() }},
		{"remove", func() interface{} { return cli.buildRemoveCommand() }},
		{"enable", func() interface{} { return cli.buildEnableCommand() }},
		{"disable", func() interface{} { return cli.buildDisableCommand() }},
		{"backup", func() interface{} { return cli.buildBackupCommand() }},
		{"restore", func() interface{} { return cli.buildRestoreCommand() }},
		{"import", func() interface{} { return cli.buildImportCommand() }},
		{"export", func() interface{} { return cli.buildExportCommand() }},
		{"verify", func() interface{} { return cli.buildVerifyCommand() }},
		{"profile", func() interface{} { return cli.buildProfileCommand() }},
		{"search", func() interface{} { return cli.buildSearchCommand() }},
		{"completion", func() interface{} { return cli.buildCompletionCommand() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.builder()
			if cmd == nil {
				// Simple title case conversion without strings.Title (deprecated)
				titleName := string(unicode.ToUpper(rune(tt.name[0]))) + tt.name[1:]
				t.Errorf("build%sCommand() returned nil", titleName)
			}
		})
	}
}

func TestCLI_runVerify(t *testing.T) {
	// Create a temporary hosts file for testing
	tmpDir, err := os.MkdirTemp("", "hostsctl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tests := []struct {
		name        string
		content     string
		shouldError bool
	}{
		{
			name:        "valid hosts file",
			content:     "127.0.0.1\tlocalhost\n192.168.1.1\tserver.local",
			shouldError: false,
		},
		{
			name:        "invalid IP",
			content:     "999.999.999.999\tinvalid.local",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostsFile := filepath.Join(tmpDir, "hosts_"+tt.name)
			if err := os.WriteFile(hostsFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write hosts file: %v", err)
			}

			cli := NewCLI()
			cli.hostsFile = hostsFile
			cli.jsonOutput = true

			err := cli.runVerify()
			if (err != nil) != tt.shouldError {
				t.Errorf("runVerify() error = %v, shouldError %v", err, tt.shouldError)
			}
		})
	}
}

func TestCLI_runBackup(t *testing.T) {
	// Create a temporary hosts file for testing
	tmpDir, err := os.MkdirTemp("", "hostsctl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	content := "127.0.0.1\tlocalhost"

	if err := os.WriteFile(hostsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write hosts file: %v", err)
	}

	cli := NewCLI()
	cli.hostsFile = hostsFile
	cli.jsonOutput = true

	// Test backup without specific output
	err = cli.runBackup("")
	if err != nil {
		t.Errorf("runBackup() error = %v", err)
	}

	// Test backup with specific output
	backupFile := filepath.Join(tmpDir, "backup.txt")
	err = cli.runBackup(backupFile)
	if err != nil {
		t.Errorf("runBackup() with output error = %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}
}

func TestCLI_runExport(t *testing.T) {
	// Create a temporary hosts file for testing
	tmpDir, err := os.MkdirTemp("", "hostsctl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	content := "127.0.0.1\tlocalhost\n192.168.1.100\tserver.local\t# Test server"

	if err := os.WriteFile(hostsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write hosts file: %v", err)
	}

	cli := NewCLI()
	cli.hostsFile = hostsFile

	tests := []struct {
		name   string
		format string
		valid  bool
	}{
		{"json export", "json", true},
		{"yaml export", "yaml", true},
		{"invalid format", "xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exportFile := filepath.Join(tmpDir, "export."+tt.format)
			err := cli.runExport(exportFile, tt.format)

			if tt.valid && err != nil {
				t.Errorf("runExport() error = %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("runExport() should have failed with invalid format")
			}

			if tt.valid {
				// Verify export file exists and is valid
				if _, err := os.Stat(exportFile); os.IsNotExist(err) {
					t.Error("Export file was not created")
				}

				// For JSON, verify it's valid JSON
				if tt.format == "json" {
					data, err := os.ReadFile(exportFile)
					if err != nil {
						t.Errorf("Failed to read export file: %v", err)
					}
					var profile hosts.Profile
					if err := json.Unmarshal(data, &profile); err != nil {
						t.Errorf("Export file is not valid JSON: %v", err)
					}
				}
			}
		})
	}
}
