package hosts

import (
	"testing"
	"time"
)

func TestEntry_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
		want  bool
	}{
		{
			name: "valid entry",
			entry: Entry{
				IP:    "127.0.0.1",
				Names: []string{"localhost"},
			},
			want: true,
		},
		{
			name: "valid entry with multiple names",
			entry: Entry{
				IP:    "192.168.1.1",
				Names: []string{"server", "api"},
			},
			want: true,
		},
		{
			name: "invalid entry - no IP",
			entry: Entry{
				Names: []string{"localhost"},
			},
			want: false,
		},
		{
			name: "invalid entry - no names",
			entry: Entry{
				IP: "127.0.0.1",
			},
			want: false,
		},
		{
			name: "invalid entry - empty names",
			entry: Entry{
				IP:    "127.0.0.1",
				Names: []string{},
			},
			want: false,
		},
		{
			name:  "invalid entry - both empty",
			entry: Entry{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.IsValid(); got != tt.want {
				t.Errorf("Entry.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntry_String(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
		want  string
	}{
		{
			name: "simple entry",
			entry: Entry{
				IP:    "127.0.0.1",
				Names: []string{"localhost"},
			},
			want: "127.0.0.1\tlocalhost",
		},
		{
			name: "entry with multiple names",
			entry: Entry{
				IP:    "192.168.1.100",
				Names: []string{"server.local", "web.local"},
			},
			want: "192.168.1.100\tserver.local\tweb.local",
		},
		{
			name: "entry with comment",
			entry: Entry{
				IP:      "192.168.1.100",
				Names:   []string{"server.local"},
				Comment: "Main server",
			},
			want: "192.168.1.100\tserver.local\t# Main server",
		},
		{
			name: "disabled entry",
			entry: Entry{
				IP:       "192.168.1.200",
				Names:    []string{"disabled.local"},
				Comment:  "Disabled entry",
				Disabled: true,
			},
			want: "# 192.168.1.200\tdisabled.local\t# Disabled entry",
		},
		{
			name: "disabled entry without comment",
			entry: Entry{
				IP:       "10.0.0.1",
				Names:    []string{"test.com"},
				Disabled: true,
			},
			want: "# 10.0.0.1\ttest.com",
		},
		{
			name: "entry with multiple names and comment",
			entry: Entry{
				IP:      "192.168.1.1",
				Names:   []string{"api.local", "web.local", "app.local"},
				Comment: "Multi-service server",
			},
			want: "192.168.1.1\tapi.local\tweb.local\tapp.local\t# Multi-service server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.String(); got != tt.want {
				t.Errorf("Entry.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProfile_AddEntry(t *testing.T) {
	profile := &Profile{
		Name:      "test",
		UpdatedAt: time.Unix(0, 0), // Set to epoch for testing
	}

	entry := Entry{
		IP:    "127.0.0.1",
		Names: []string{"localhost"},
	}

	originalTime := profile.UpdatedAt
	profile.AddEntry(entry)

	if len(profile.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(profile.Entries))
	}

	if profile.Entries[0].IP != "127.0.0.1" {
		t.Errorf("Expected IP 127.0.0.1, got %s", profile.Entries[0].IP)
	}

	if !profile.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should be updated after adding entry")
	}
}

func TestProfile_RemoveEntry(t *testing.T) {
	profile := &Profile{
		Name: "test",
		Entries: []Entry{
			{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}},
			{ID: 2, IP: "192.168.1.1", Names: []string{"server"}},
		},
		UpdatedAt: time.Unix(0, 0),
	}

	// Test successful removal
	originalTime := profile.UpdatedAt
	result := profile.RemoveEntry(1)

	if !result {
		t.Error("RemoveEntry should return true for existing entry")
	}

	if len(profile.Entries) != 1 {
		t.Errorf("Expected 1 entry after removal, got %d", len(profile.Entries))
	}

	if profile.Entries[0].ID != 2 {
		t.Errorf("Expected remaining entry ID 2, got %d", profile.Entries[0].ID)
	}

	if !profile.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should be updated after removing entry")
	}

	// Test removal of non-existent entry
	result = profile.RemoveEntry(999)
	if result {
		t.Error("RemoveEntry should return false for non-existent entry")
	}
}

func TestHostsFile_FindByID(t *testing.T) {
	hostsFile := &HostsFile{
		Entries: []Entry{
			{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}},
			{ID: 2, IP: "192.168.1.1", Names: []string{"server"}},
		},
	}

	// Test finding existing entry
	entry := hostsFile.FindByID(1)
	if entry == nil {
		t.Error("FindByID should return entry for existing ID")
	} else if entry.ID != 1 {
		t.Errorf("Expected ID 1, got %d", entry.ID)
	}

	// Test finding non-existent entry
	entry = hostsFile.FindByID(999)
	if entry != nil {
		t.Error("FindByID should return nil for non-existent ID")
	}
}

func TestHostsFile_FindByName(t *testing.T) {
	hostsFile := &HostsFile{
		Entries: []Entry{
			{ID: 1, IP: "127.0.0.1", Names: []string{"localhost", "local"}},
			{ID: 2, IP: "192.168.1.1", Names: []string{"server.local"}},
			{ID: 3, IP: "192.168.1.2", Names: []string{"api.local", "web.local"}},
		},
	}

	// Test finding by existing name
	entries := hostsFile.FindByName("localhost")
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for 'localhost', got %d", len(entries))
	} else if entries[0].ID != 1 {
		t.Errorf("Expected entry ID 1, got %d", entries[0].ID)
	}

	// Test finding by name that appears in multiple entries
	entries = hostsFile.FindByName("api.local")
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for 'api.local', got %d", len(entries))
	} else if entries[0].ID != 3 {
		t.Errorf("Expected entry ID 3, got %d", entries[0].ID)
	}

	// Test finding by non-existent name
	entries = hostsFile.FindByName("nonexistent")
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for 'nonexistent', got %d", len(entries))
	}

	// Test finding name that appears multiple times in same entry
	entries = hostsFile.FindByName("local")
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for 'local', got %d", len(entries))
	}
}

func TestHostsFile_AddEntry(t *testing.T) {
	hostsFile := &HostsFile{}

	// Test adding first entry
	entry1 := Entry{IP: "127.0.0.1", Names: []string{"localhost"}}
	hostsFile.AddEntry(entry1)

	if len(hostsFile.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(hostsFile.Entries))
	}

	if hostsFile.Entries[0].ID != 1 {
		t.Errorf("Expected first entry ID 1, got %d", hostsFile.Entries[0].ID)
	}

	// Test adding second entry
	entry2 := Entry{IP: "192.168.1.1", Names: []string{"server"}}
	hostsFile.AddEntry(entry2)

	if len(hostsFile.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(hostsFile.Entries))
	}

	if hostsFile.Entries[1].ID != 2 {
		t.Errorf("Expected second entry ID 2, got %d", hostsFile.Entries[1].ID)
	}

	// Test adding entry when there are gaps in IDs
	hostsFile.Entries = []Entry{
		{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}},
		{ID: 5, IP: "192.168.1.1", Names: []string{"server"}},
	}

	entry3 := Entry{IP: "10.0.0.1", Names: []string{"api"}}
	hostsFile.AddEntry(entry3)

	if hostsFile.Entries[2].ID != 6 {
		t.Errorf("Expected new entry ID 6 (max+1), got %d", hostsFile.Entries[2].ID)
	}
}

func TestHostsFile_RemoveEntry(t *testing.T) {
	hostsFile := &HostsFile{
		Entries: []Entry{
			{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}},
			{ID: 2, IP: "192.168.1.1", Names: []string{"server"}},
		},
	}

	// Test successful removal
	result := hostsFile.RemoveEntry(1)
	if !result {
		t.Error("RemoveEntry should return true for existing entry")
	}

	if len(hostsFile.Entries) != 1 {
		t.Errorf("Expected 1 entry after removal, got %d", len(hostsFile.Entries))
	}

	if hostsFile.Entries[0].ID != 2 {
		t.Errorf("Expected remaining entry ID 2, got %d", hostsFile.Entries[0].ID)
	}

	// Test removal of non-existent entry
	result = hostsFile.RemoveEntry(999)
	if result {
		t.Error("RemoveEntry should return false for non-existent entry")
	}
}

func TestHostsFile_EnableDisableEntry(t *testing.T) {
	hostsFile := &HostsFile{
		Entries: []Entry{
			{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Disabled: false},
			{ID: 2, IP: "192.168.1.1", Names: []string{"server"}, Disabled: true},
		},
	}

	// Test disabling enabled entry
	result := hostsFile.DisableEntry(1)
	if !result {
		t.Error("DisableEntry should return true for existing entry")
	}

	entry := hostsFile.FindByID(1)
	if entry == nil || !entry.Disabled {
		t.Error("Entry should be disabled after DisableEntry")
	}

	// Test enabling disabled entry
	result = hostsFile.EnableEntry(2)
	if !result {
		t.Error("EnableEntry should return true for existing entry")
	}

	entry = hostsFile.FindByID(2)
	if entry == nil || entry.Disabled {
		t.Error("Entry should be enabled after EnableEntry")
	}

	// Test operations on non-existent entry
	result = hostsFile.DisableEntry(999)
	if result {
		t.Error("DisableEntry should return false for non-existent entry")
	}

	result = hostsFile.EnableEntry(999)
	if result {
		t.Error("EnableEntry should return false for non-existent entry")
	}
}

func TestExportFormat_Constants(t *testing.T) {
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}

	if FormatYAML != "yaml" {
		t.Errorf("FormatYAML = %q, want %q", FormatYAML, "yaml")
	}
}

func TestBackupInfo_Structure(t *testing.T) {
	now := time.Now()
	backup := BackupInfo{
		Path:      "/tmp/backup.txt",
		Original:  "/etc/hosts",
		CreatedAt: now,
		Size:      1024,
	}

	if backup.Path != "/tmp/backup.txt" {
		t.Error("BackupInfo.Path not set correctly")
	}
	if backup.Original != "/etc/hosts" {
		t.Error("BackupInfo.Original not set correctly")
	}
	if backup.CreatedAt != now {
		t.Error("BackupInfo.CreatedAt not set correctly")
	}
	if backup.Size != 1024 {
		t.Error("BackupInfo.Size not set correctly")
	}
}

func TestProfile_Structure(t *testing.T) {
	now := time.Now()
	profile := Profile{
		Name:        "test",
		Description: "Test profile",
		Entries:     []Entry{{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}}},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if profile.Name != "test" {
		t.Error("Profile.Name not set correctly")
	}
	if profile.Description != "Test profile" {
		t.Error("Profile.Description not set correctly")
	}
	if len(profile.Entries) != 1 {
		t.Error("Profile.Entries not set correctly")
	}
	if profile.CreatedAt != now {
		t.Error("Profile.CreatedAt not set correctly")
	}
	if profile.UpdatedAt != now {
		t.Error("Profile.UpdatedAt not set correctly")
	}
}
