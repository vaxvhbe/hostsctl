package cli

import (
	"testing"

	"github.com/vaxvhbe/hostsctl/internal/hosts"
)

func TestCLI_calculateDiff(t *testing.T) {
	cli := NewCLI()

	current := []hosts.Entry{
		{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
		{ID: 2, IP: "192.168.1.1", Names: []string{"server"}, Comment: "Server", Disabled: false},
		{ID: 3, IP: "192.168.1.2", Names: []string{"old.local"}, Comment: "Old", Disabled: false},
	}

	profile := []hosts.Entry{
		{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},   // Same
		{ID: 2, IP: "192.168.1.1", Names: []string{"server"}, Comment: "Modified", Disabled: false}, // Modified comment
		{ID: 4, IP: "192.168.1.3", Names: []string{"new.local"}, Comment: "New", Disabled: false},   // Added
		// old.local removed
	}

	diff := cli.calculateDiff(current, profile)

	// Check same entries
	if len(diff.Same) != 1 {
		t.Errorf("Expected 1 same entry, got %d", len(diff.Same))
	}
	if len(diff.Same) > 0 && diff.Same[0].IP != "127.0.0.1" {
		t.Errorf("Expected same entry IP 127.0.0.1, got %s", diff.Same[0].IP)
	}

	// Check added entries
	if len(diff.Added) != 1 {
		t.Errorf("Expected 1 added entry, got %d", len(diff.Added))
	}
	if len(diff.Added) > 0 && diff.Added[0].IP != "192.168.1.3" {
		t.Errorf("Expected added entry IP 192.168.1.3, got %s", diff.Added[0].IP)
	}

	// Check removed entries
	if len(diff.Removed) != 1 {
		t.Errorf("Expected 1 removed entry, got %d", len(diff.Removed))
	}
	if len(diff.Removed) > 0 && diff.Removed[0].IP != "192.168.1.2" {
		t.Errorf("Expected removed entry IP 192.168.1.2, got %s", diff.Removed[0].IP)
	}

	// Check modified entries
	if len(diff.Modified) != 1 {
		t.Errorf("Expected 1 modified entry, got %d", len(diff.Modified))
	}
	if len(diff.Modified) > 0 {
		mod := diff.Modified[0]
		if mod.Old.Comment != "Server" {
			t.Errorf("Expected old comment 'Server', got %s", mod.Old.Comment)
		}
		if mod.New.Comment != "Modified" {
			t.Errorf("Expected new comment 'Modified', got %s", mod.New.Comment)
		}
	}
}

func TestCLI_entryKey(t *testing.T) {
	cli := NewCLI()

	tests := []struct {
		entry    hosts.Entry
		expected string
	}{
		{
			entry:    hosts.Entry{IP: "127.0.0.1", Names: []string{"localhost"}},
			expected: "127.0.0.1:localhost",
		},
		{
			entry:    hosts.Entry{IP: "192.168.1.1", Names: []string{"server", "api"}},
			expected: "192.168.1.1:server,api",
		},
		{
			entry:    hosts.Entry{IP: "10.0.0.1", Names: []string{"single"}},
			expected: "10.0.0.1:single",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := cli.entryKey(tt.entry)
			if got != tt.expected {
				t.Errorf("entryKey() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCLI_entriesEqual(t *testing.T) {
	cli := NewCLI()

	tests := []struct {
		name string
		a    hosts.Entry
		b    hosts.Entry
		want bool
	}{
		{
			name: "identical entries",
			a:    hosts.Entry{IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
			b:    hosts.Entry{IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
			want: true,
		},
		{
			name: "different IP",
			a:    hosts.Entry{IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
			b:    hosts.Entry{IP: "192.168.1.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
			want: false,
		},
		{
			name: "different names",
			a:    hosts.Entry{IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
			b:    hosts.Entry{IP: "127.0.0.1", Names: []string{"local"}, Comment: "Local", Disabled: false},
			want: false,
		},
		{
			name: "different comment",
			a:    hosts.Entry{IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
			b:    hosts.Entry{IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Different", Disabled: false},
			want: false,
		},
		{
			name: "different disabled status",
			a:    hosts.Entry{IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
			b:    hosts.Entry{IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: true},
			want: false,
		},
		{
			name: "different name order",
			a:    hosts.Entry{IP: "127.0.0.1", Names: []string{"a", "b"}, Comment: "Local", Disabled: false},
			b:    hosts.Entry{IP: "127.0.0.1", Names: []string{"b", "a"}, Comment: "Local", Disabled: false},
			want: false,
		},
		{
			name: "empty entries",
			a:    hosts.Entry{},
			b:    hosts.Entry{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cli.entriesEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("entriesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLI_buildProfileCommands(t *testing.T) {
	cli := NewCLI()

	// Test main profile command
	profileCmd := cli.buildProfileCommand()
	if profileCmd == nil {
		t.Error("buildProfileCommand() returned nil")
		return
	}

	if profileCmd.Use != "profile" {
		t.Errorf("buildProfileCommand() Use = %v, want %v", profileCmd.Use, "profile")
	}

	// Test that all subcommands exist
	expectedSubcommands := []string{"list", "save", "apply", "delete", "show", "diff", "export", "import"}
	subcommands := make(map[string]bool)

	for _, cmd := range profileCmd.Commands() {
		subcommands[cmd.Name()] = true
	}

	for _, expected := range expectedSubcommands {
		if !subcommands[expected] {
			t.Errorf("buildProfileCommand() missing subcommand: %s", expected)
		}
	}

	// Test individual command builders
	tests := []struct {
		name    string
		builder func() interface{}
	}{
		{"list", func() interface{} { return cli.buildProfileListCommand() }},
		{"save", func() interface{} { return cli.buildProfileSaveCommand() }},
		{"apply", func() interface{} { return cli.buildProfileApplyCommand() }},
		{"delete", func() interface{} { return cli.buildProfileDeleteCommand() }},
		{"show", func() interface{} { return cli.buildProfileShowCommand() }},
		{"diff", func() interface{} { return cli.buildProfileDiffCommand() }},
		{"export", func() interface{} { return cli.buildProfileExportCommand() }},
		{"import", func() interface{} { return cli.buildProfileImportCommand() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.builder()
			if cmd == nil {
				t.Errorf("buildProfile%sCommand() returned nil", tt.name)
			}
		})
	}
}

func TestCLI_profileCommandFlags(t *testing.T) {
	cli := NewCLI()

	// Test save command flags
	saveCmd := cli.buildProfileSaveCommand()
	if saveCmd.Flags().Lookup("name") == nil {
		t.Error("buildProfileSaveCommand() missing --name flag")
	}
	if saveCmd.Flags().Lookup("description") == nil {
		t.Error("buildProfileSaveCommand() missing --description flag")
	}
	if saveCmd.Flags().Lookup("overwrite") == nil {
		t.Error("buildProfileSaveCommand() missing --overwrite flag")
	}

	// Test apply command flags
	applyCmd := cli.buildProfileApplyCommand()
	if applyCmd.Flags().Lookup("merge") == nil {
		t.Error("buildProfileApplyCommand() missing --merge flag")
	}
	if applyCmd.Flags().Lookup("backup") == nil {
		t.Error("buildProfileApplyCommand() missing --backup flag")
	}

	// Test delete command flags
	deleteCmd := cli.buildProfileDeleteCommand()
	if deleteCmd.Flags().Lookup("force") == nil {
		t.Error("buildProfileDeleteCommand() missing --force flag")
	}

	// Test export command flags
	exportCmd := cli.buildProfileExportCommand()
	if exportCmd.Flags().Lookup("output") == nil {
		t.Error("buildProfileExportCommand() missing --output flag")
	}
	if exportCmd.Flags().Lookup("format") == nil {
		t.Error("buildProfileExportCommand() missing --format flag")
	}

	// Test import command flags
	importCmd := cli.buildProfileImportCommand()
	if importCmd.Flags().Lookup("format") == nil {
		t.Error("buildProfileImportCommand() missing --format flag")
	}
	if importCmd.Flags().Lookup("overwrite") == nil {
		t.Error("buildProfileImportCommand() missing --overwrite flag")
	}
}

func TestDiffResult_Structure(t *testing.T) {
	// Test that DiffResult has the expected structure
	diff := &DiffResult{
		Added:   []hosts.Entry{{IP: "127.0.0.1", Names: []string{"added"}}},
		Removed: []hosts.Entry{{IP: "192.168.1.1", Names: []string{"removed"}}},
		Modified: []DiffEntry{{
			Old: hosts.Entry{IP: "10.0.0.1", Names: []string{"old"}},
			New: hosts.Entry{IP: "10.0.0.1", Names: []string{"new"}},
		}},
		Same: []hosts.Entry{{IP: "172.16.0.1", Names: []string{"same"}}},
	}

	if len(diff.Added) != 1 {
		t.Error("DiffResult.Added not set correctly")
	}
	if len(diff.Removed) != 1 {
		t.Error("DiffResult.Removed not set correctly")
	}
	if len(diff.Modified) != 1 {
		t.Error("DiffResult.Modified not set correctly")
	}
	if len(diff.Same) != 1 {
		t.Error("DiffResult.Same not set correctly")
	}
}

func TestDiffEntry_Structure(t *testing.T) {
	// Test that DiffEntry has the expected structure
	diffEntry := DiffEntry{
		Old: hosts.Entry{IP: "127.0.0.1", Names: []string{"old"}},
		New: hosts.Entry{IP: "127.0.0.1", Names: []string{"new"}},
	}

	if diffEntry.Old.Names[0] != "old" {
		t.Error("DiffEntry.Old not set correctly")
	}
	if diffEntry.New.Names[0] != "new" {
		t.Error("DiffEntry.New not set correctly")
	}
}

func TestCLI_EmptyDiff(t *testing.T) {
	cli := NewCLI()

	// Test diff with identical entries
	entries := []hosts.Entry{
		{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
		{ID: 2, IP: "192.168.1.1", Names: []string{"server"}, Comment: "Server", Disabled: false},
	}

	diff := cli.calculateDiff(entries, entries)

	if len(diff.Added) != 0 {
		t.Errorf("Expected 0 added entries, got %d", len(diff.Added))
	}
	if len(diff.Removed) != 0 {
		t.Errorf("Expected 0 removed entries, got %d", len(diff.Removed))
	}
	if len(diff.Modified) != 0 {
		t.Errorf("Expected 0 modified entries, got %d", len(diff.Modified))
	}
	if len(diff.Same) != 2 {
		t.Errorf("Expected 2 same entries, got %d", len(diff.Same))
	}
}

func TestCLI_CompletelyDifferentDiff(t *testing.T) {
	cli := NewCLI()

	current := []hosts.Entry{
		{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local", Disabled: false},
		{ID: 2, IP: "192.168.1.1", Names: []string{"server"}, Comment: "Server", Disabled: false},
	}

	profile := []hosts.Entry{
		{ID: 1, IP: "10.0.0.1", Names: []string{"api"}, Comment: "API", Disabled: false},
		{ID: 2, IP: "10.0.0.2", Names: []string{"web"}, Comment: "Web", Disabled: false},
	}

	diff := cli.calculateDiff(current, profile)

	if len(diff.Added) != 2 {
		t.Errorf("Expected 2 added entries, got %d", len(diff.Added))
	}
	if len(diff.Removed) != 2 {
		t.Errorf("Expected 2 removed entries, got %d", len(diff.Removed))
	}
	if len(diff.Modified) != 0 {
		t.Errorf("Expected 0 modified entries, got %d", len(diff.Modified))
	}
	if len(diff.Same) != 0 {
		t.Errorf("Expected 0 same entries, got %d", len(diff.Same))
	}
}
