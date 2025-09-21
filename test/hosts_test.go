package test

import (
	"io"
	"os"
	"testing"

	"github.com/vaxvhbe/hostsctl/internal/hosts"
)

func TestHostsFileOperations(t *testing.T) {
	hostsFile := &hosts.HostsFile{
		Entries: []hosts.Entry{},
	}

	entry1 := hosts.Entry{
		IP:      "127.0.0.1",
		Names:   []string{"localhost"},
		Comment: "Local host",
	}

	hostsFile.AddEntry(entry1)

	if len(hostsFile.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(hostsFile.Entries))
	}

	if hostsFile.Entries[0].ID != 1 {
		t.Errorf("Expected entry ID 1, got %d", hostsFile.Entries[0].ID)
	}

	entry2 := hosts.Entry{
		IP:      "192.168.1.100",
		Names:   []string{"server.local"},
		Comment: "Test server",
	}

	hostsFile.AddEntry(entry2)

	if len(hostsFile.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(hostsFile.Entries))
	}

	if hostsFile.Entries[1].ID != 2 {
		t.Errorf("Expected entry ID 2, got %d", hostsFile.Entries[1].ID)
	}
}

func TestFindByName(t *testing.T) {
	hostsFile := &hosts.HostsFile{
		Entries: []hosts.Entry{
			{
				ID:    1,
				IP:    "127.0.0.1",
				Names: []string{"localhost", "local"},
			},
			{
				ID:    2,
				IP:    "192.168.1.100",
				Names: []string{"server.local"},
			},
		},
	}

	entries := hostsFile.FindByName("localhost")
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry for 'localhost', got %d", len(entries))
	}
	if entries[0].ID != 1 {
		t.Errorf("Expected entry ID 1, got %d", entries[0].ID)
	}

	entries = hostsFile.FindByName("server.local")
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry for 'server.local', got %d", len(entries))
	}
	if entries[0].ID != 2 {
		t.Errorf("Expected entry ID 2, got %d", entries[0].ID)
	}

	entries = hostsFile.FindByName("nonexistent")
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for 'nonexistent', got %d", len(entries))
	}
}

func TestRemoveEntry(t *testing.T) {
	hostsFile := &hosts.HostsFile{
		Entries: []hosts.Entry{
			{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}},
			{ID: 2, IP: "192.168.1.100", Names: []string{"server.local"}},
		},
	}

	removed := hostsFile.RemoveEntry(1)
	if !removed {
		t.Error("Expected entry to be removed")
	}

	if len(hostsFile.Entries) != 1 {
		t.Fatalf("Expected 1 entry after removal, got %d", len(hostsFile.Entries))
	}

	if hostsFile.Entries[0].ID != 2 {
		t.Errorf("Expected remaining entry ID 2, got %d", hostsFile.Entries[0].ID)
	}

	removed = hostsFile.RemoveEntry(999)
	if removed {
		t.Error("Expected no entry to be removed for non-existent ID")
	}
}

func TestEnableDisableEntry(t *testing.T) {
	hostsFile := &hosts.HostsFile{
		Entries: []hosts.Entry{
			{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Disabled: false},
		},
	}

	disabled := hostsFile.DisableEntry(1)
	if !disabled {
		t.Error("Expected entry to be disabled")
	}

	entry := hostsFile.FindByID(1)
	if entry == nil || !entry.Disabled {
		t.Error("Expected entry to be disabled")
	}

	enabled := hostsFile.EnableEntry(1)
	if !enabled {
		t.Error("Expected entry to be enabled")
	}

	entry = hostsFile.FindByID(1)
	if entry == nil || entry.Disabled {
		t.Error("Expected entry to be enabled")
	}
}

func TestStoreWithTempFile(t *testing.T) {
	content := `127.0.0.1	localhost
192.168.1.100	server.local	# Test server`

	tempFile, err := os.CreateTemp("", "hosts_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()
	defer func() { _ = tempFile.Close() }()

	if _, err := io.WriteString(tempFile, content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	_ = tempFile.Close()

	store := hosts.NewStore(tempFile.Name(), false)

	hostsFile, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load hosts file: %v", err)
	}

	if len(hostsFile.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(hostsFile.Entries))
	}

	newEntry := hosts.Entry{
		IP:      "10.0.0.1",
		Names:   []string{"test.local"},
		Comment: "Added by test",
	}

	hostsFile.AddEntry(newEntry)

	backupInfo, err := store.Backup("")
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	defer func() { _ = os.Remove(backupInfo.Path) }()

	if backupInfo.Size == 0 {
		t.Error("Expected backup to have non-zero size")
	}
}

func TestEntryString(t *testing.T) {
	tests := []struct {
		entry    hosts.Entry
		expected string
	}{
		{
			entry: hosts.Entry{
				IP:       "127.0.0.1",
				Names:    []string{"localhost"},
				Comment:  "",
				Disabled: false,
			},
			expected: "127.0.0.1\tlocalhost",
		},
		{
			entry: hosts.Entry{
				IP:       "192.168.1.100",
				Names:    []string{"server.local", "web.local"},
				Comment:  "Main server",
				Disabled: false,
			},
			expected: "192.168.1.100\tserver.local\tweb.local\t# Main server",
		},
		{
			entry: hosts.Entry{
				IP:       "192.168.1.200",
				Names:    []string{"disabled.local"},
				Comment:  "Disabled entry",
				Disabled: true,
			},
			expected: "# 192.168.1.200\tdisabled.local\t# Disabled entry",
		},
	}

	for i, test := range tests {
		result := test.entry.String()
		if result != test.expected {
			t.Errorf("Test %d: Expected '%s', got '%s'", i+1, test.expected, result)
		}
	}
}
