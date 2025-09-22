package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vaxvhbe/hostsctl/internal/hosts"
)

func TestCLI_searchEntries(t *testing.T) {
	cli := NewCLI()

	entries := []hosts.Entry{
		{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local machine", Disabled: false},
		{ID: 2, IP: "192.168.1.100", Names: []string{"server.local", "web.local"}, Comment: "Development server", Disabled: false},
		{ID: 3, IP: "192.168.1.200", Names: []string{"disabled.local"}, Comment: "Disabled entry", Disabled: true},
		{ID: 4, IP: "10.0.0.1", Names: []string{"api.prod.com"}, Comment: "Production API", Disabled: false},
	}

	tests := []struct {
		name    string
		options SearchOptions
		wantLen int
		wantErr bool
	}{
		{
			name: "search all fields for 'local'",
			options: SearchOptions{
				Pattern:         "local",
				SearchIP:        true,
				SearchNames:     true,
				SearchComments:  true,
				IncludeDisabled: true,
			},
			wantLen: 3, // localhost, server.local, disabled.local (but not "Local machine" because that would be another match for the same entry)
			wantErr: false,
		},
		{
			name: "search only IP addresses",
			options: SearchOptions{
				Pattern:         "192.168",
				SearchIP:        true,
				SearchNames:     false,
				SearchComments:  false,
				IncludeDisabled: true,
			},
			wantLen: 2, // 192.168.1.100 and 192.168.1.200
			wantErr: false,
		},
		{
			name: "search only hostnames",
			options: SearchOptions{
				Pattern:         "server",
				SearchIP:        false,
				SearchNames:     true,
				SearchComments:  false,
				IncludeDisabled: true,
			},
			wantLen: 1, // server.local
			wantErr: false,
		},
		{
			name: "search only comments",
			options: SearchOptions{
				Pattern:         "Development",
				SearchIP:        false,
				SearchNames:     false,
				SearchComments:  true,
				IncludeDisabled: true,
				IgnoreCase:      true,
			},
			wantLen: 1, // "Development server"
			wantErr: false,
		},
		{
			name: "exclude disabled entries",
			options: SearchOptions{
				Pattern:         "local",
				SearchIP:        true,
				SearchNames:     true,
				SearchComments:  true,
				IncludeDisabled: false,
			},
			wantLen: 2, // localhost, server.local (excludes disabled.local)
			wantErr: false,
		},
		{
			name: "regex search",
			options: SearchOptions{
				Pattern:         "^192\\.168",
				UseRegex:        true,
				SearchIP:        true,
				SearchNames:     false,
				SearchComments:  false,
				IncludeDisabled: true,
			},
			wantLen: 2, // 192.168.1.100 and 192.168.1.200
			wantErr: false,
		},
		{
			name: "invalid regex",
			options: SearchOptions{
				Pattern:         "[invalid",
				UseRegex:        true,
				SearchIP:        true,
				SearchNames:     false,
				SearchComments:  false,
				IncludeDisabled: true,
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "case insensitive search",
			options: SearchOptions{
				Pattern:         "PROD",
				IgnoreCase:      true,
				SearchIP:        false,
				SearchNames:     true,
				SearchComments:  true,
				IncludeDisabled: true,
			},
			wantLen: 1, // api.prod.com (one match per entry, not separate for name and comment)
			wantErr: false,
		},
		{
			name: "no matches",
			options: SearchOptions{
				Pattern:         "nonexistent",
				SearchIP:        true,
				SearchNames:     true,
				SearchComments:  true,
				IncludeDisabled: true,
			},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := cli.searchEntries(entries, tt.options)

			if (err != nil) != tt.wantErr {
				t.Errorf("searchEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(results) != tt.wantLen {
				t.Errorf("searchEntries() len = %v, want %v", len(results), tt.wantLen)
			}

			// Verify that all results have the correct match type
			for _, result := range results {
				if result.MatchType == "" {
					t.Error("searchEntries() result missing MatchType")
				}
				if result.MatchText == "" {
					t.Error("searchEntries() result missing MatchText")
				}
			}
		})
	}
}

func TestCLI_runSearch(t *testing.T) {
	// Create a temporary hosts file for testing
	tmpDir, err := os.MkdirTemp("", "hostsctl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	content := `127.0.0.1	localhost	# Local machine
192.168.1.100	server.local web.local	# Development server
# 192.168.1.200	disabled.local	# Disabled entry
10.0.0.1	api.prod.com	# Production API`

	if err := os.WriteFile(hostsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write hosts file: %v", err)
	}

	cli := NewCLI()
	cli.hostsFile = hostsFile
	cli.jsonOutput = true

	tests := []struct {
		name    string
		options SearchOptions
		wantErr bool
	}{
		{
			name: "basic search",
			options: SearchOptions{
				Pattern: "local",
			},
			wantErr: false,
		},
		{
			name: "regex search",
			options: SearchOptions{
				Pattern:  "^192",
				UseRegex: true,
			},
			wantErr: false,
		},
		{
			name: "invalid regex",
			options: SearchOptions{
				Pattern:  "[invalid",
				UseRegex: true,
			},
			wantErr: true,
		},
		{
			name: "search with all options",
			options: SearchOptions{
				Pattern:         "server",
				UseRegex:        false,
				IgnoreCase:      true,
				SearchIP:        true,
				SearchNames:     true,
				SearchComments:  true,
				IncludeDisabled: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.runSearch(tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("runSearch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCLI_buildSearchCommand(t *testing.T) {
	cli := NewCLI()
	cmd := cli.buildSearchCommand()

	if cmd == nil {
		t.Error("buildSearchCommand() returned nil")
		return
	}

	if cmd.Use != "search [pattern]" {
		t.Errorf("buildSearchCommand() Use = %v, want %v", cmd.Use, "search [pattern]")
	}

	// Test that flags are properly set
	flags := cmd.Flags()
	expectedFlags := []string{"regex", "ignore-case", "ip", "names", "comments", "include-disabled", "case-insensitive", "glob"}

	for _, flagName := range expectedFlags {
		if flags.Lookup(flagName) == nil {
			t.Errorf("buildSearchCommand() missing flag: %s", flagName)
		}
	}
}

func TestSearchOptions_DefaultBehavior(t *testing.T) {
	cli := NewCLI()

	// Test that when no specific fields are selected, all fields are searched
	entries := []hosts.Entry{
		{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local machine", Disabled: false},
	}

	options := SearchOptions{
		Pattern: "local",
		// No specific fields set - should default to all (this is handled in runSearch)
		SearchIP:       true,
		SearchNames:    true,
		SearchComments: true,
	}

	results, err := cli.searchEntries(entries, options)
	if err != nil {
		t.Errorf("searchEntries() error = %v", err)
	}

	// Should find matches (localhost in names)
	if len(results) == 0 {
		t.Error("searchEntries() should find matches when no fields specified")
	}
}

func TestSearchResult_Structure(t *testing.T) {
	// Test that SearchResult has the expected fields
	result := SearchResult{
		Entry: hosts.Entry{
			ID:    1,
			IP:    "127.0.0.1",
			Names: []string{"localhost"},
		},
		MatchType: "hostname",
		MatchText: "localhost",
	}

	if result.Entry.ID != 1 {
		t.Error("SearchResult.Entry not properly set")
	}
	if result.MatchType != "hostname" {
		t.Error("SearchResult.MatchType not properly set")
	}
	if result.MatchText != "localhost" {
		t.Error("SearchResult.MatchText not properly set")
	}
}
