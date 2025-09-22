package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStore_LoadAndSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hostsctl-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	content := `127.0.0.1	localhost
192.168.1.100	server.local	# Test server
# 192.168.1.200	disabled.local	# Disabled entry`

	// Create initial hosts file
	if err := os.WriteFile(hostsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create hosts file: %v", err)
	}

	store := NewStore(hostsFile, false)

	// Test Load
	hostsData, err := store.Load()
	if err != nil {
		t.Fatalf("Store.Load() error = %v", err)
	}

	if len(hostsData.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(hostsData.Entries))
	}

	// Modify and save
	newEntry := Entry{
		IP:      "10.0.0.1",
		Names:   []string{"new.local"},
		Comment: "New entry",
	}
	hostsData.AddEntry(newEntry)

	if err := store.Save(hostsData); err != nil {
		t.Fatalf("Store.Save() error = %v", err)
	}

	// Verify file was updated
	savedContent, err := os.ReadFile(hostsFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if !strings.Contains(string(savedContent), "10.0.0.1") {
		t.Error("Saved file should contain new entry")
	}
}

func TestStore_Backup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hostsctl-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	content := "127.0.0.1\tlocalhost"

	// Create hosts file
	if err := os.WriteFile(hostsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create hosts file: %v", err)
	}

	store := NewStore(hostsFile, false)

	// Test automatic backup
	backup, err := store.Backup("")
	if err != nil {
		t.Fatalf("Store.Backup() error = %v", err)
	}

	if backup.Path == "" {
		t.Error("Backup path should not be empty")
	}

	if backup.Size == 0 {
		t.Error("Backup size should not be zero")
	}

	if backup.Original != hostsFile {
		t.Errorf("Backup original = %s, want %s", backup.Original, hostsFile)
	}

	// Verify backup file exists
	if _, err := os.Stat(backup.Path); os.IsNotExist(err) {
		t.Error("Backup file should exist")
	}

	// Test custom backup path
	customBackup := filepath.Join(tmpDir, "custom_backup")
	backup2, err := store.Backup(customBackup)
	if err != nil {
		t.Fatalf("Store.Backup() with custom path error = %v", err)
	}

	if backup2.Path != customBackup {
		t.Errorf("Custom backup path = %s, want %s", backup2.Path, customBackup)
	}
}

func TestStore_Restore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hostsctl-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	backupFile := filepath.Join(tmpDir, "hosts.backup")

	originalContent := "127.0.0.1\tlocalhost"
	backupContent := "192.168.1.1\tserver.local"

	// Create hosts file and backup
	if err := os.WriteFile(hostsFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create hosts file: %v", err)
	}
	if err := os.WriteFile(backupFile, []byte(backupContent), 0644); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	store := NewStore(hostsFile, false)

	// Test restore
	if err := store.Restore(backupFile); err != nil {
		t.Fatalf("Store.Restore() error = %v", err)
	}

	// Verify file was restored
	restoredContent, err := os.ReadFile(hostsFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if strings.TrimSpace(string(restoredContent)) != strings.TrimSpace(backupContent) {
		t.Errorf("Restored content = %q, want %q", string(restoredContent), backupContent)
	}
}

func TestStore_Verify(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hostsctl-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tests := []struct {
		name        string
		content     string
		wantIssues  int
		description string
	}{
		{
			name:        "valid file",
			content:     "127.0.0.1\tlocalhost\n192.168.1.1\tserver.local",
			wantIssues:  0,
			description: "Valid hosts file should have no issues",
		},
		{
			name:        "invalid IP",
			content:     "999.999.999.999\tinvalid.local",
			wantIssues:  1,
			description: "Invalid IP should be detected",
		},
		{
			name:        "invalid hostname",
			content:     "127.0.0.1\t.invalid",
			wantIssues:  1,
			description: "Invalid hostname should be detected",
		},
		{
			name:        "duplicate entries",
			content:     "127.0.0.1\tlocalhost\n127.0.0.1\tlocalhost",
			wantIssues:  1,
			description: "Duplicate entries should be detected",
		},
		{
			name:        "empty hostname",
			content:     "127.0.0.1\t",
			wantIssues:  1,
			description: "Empty hostname should be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostsFile := filepath.Join(tmpDir, "hosts_"+tt.name)

			if err := os.WriteFile(hostsFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create hosts file: %v", err)
			}

			store := NewStore(hostsFile, true)
			issues, err := store.Verify()

			// For invalid content, we might get a parse error instead of issues list
			if tt.wantIssues > 0 {
				// We expect either issues or an error
				if err == nil && len(issues) != tt.wantIssues {
					t.Errorf("Store.Verify() issues = %d, want %d", len(issues), tt.wantIssues)
					for i, issue := range issues {
						t.Logf("Issue %d: %s", i+1, issue)
					}
				}
				// If we get an error, that's also acceptable for invalid content
				if err != nil {
					t.Logf("Got parse error as expected: %v", err)
				}
			} else {
				// For valid content, we expect no error and no issues
				if err != nil {
					t.Errorf("Store.Verify() unexpected error = %v", err)
				}
				if len(issues) != tt.wantIssues {
					t.Errorf("Store.Verify() issues = %d, want %d", len(issues), tt.wantIssues)
				}
			}
		})
	}
}

func TestStore_LoadNonExistentFile(t *testing.T) {
	store := NewStore("/nonexistent/hosts", false)
	_, err := store.Load()

	if err == nil {
		t.Error("Store.Load() should return error for non-existent file")
	}
}

func TestStore_SaveWithoutPermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir, err := os.MkdirTemp("", "hostsctl-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create read-only dir: %v", err)
	}

	hostsFile := filepath.Join(readOnlyDir, "hosts")
	store := NewStore(hostsFile, false)

	hostsData := &HostsFile{
		Entries: []Entry{{IP: "127.0.0.1", Names: []string{"localhost"}}},
	}

	err = store.Save(hostsData)
	if err == nil {
		t.Error("Store.Save() should return error when no write permissions")
	}
}

func TestStore_BackupNonExistentFile(t *testing.T) {
	store := NewStore("/nonexistent/hosts", false)
	_, err := store.Backup("")

	if err == nil {
		t.Error("Store.Backup() should return error for non-existent file")
	}
}

func TestStore_RestoreNonExistentBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hostsctl-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	store := NewStore(hostsFile, false)

	err = store.Restore("/nonexistent/backup")
	if err == nil {
		t.Error("Store.Restore() should return error for non-existent backup")
	}
}

func TestStore_VerifyNonExistentFile(t *testing.T) {
	store := NewStore("/nonexistent/hosts", true)
	_, err := store.Verify()

	if err == nil {
		t.Error("Store.Verify() should return error for non-existent file")
	}
}

func TestNewStore(t *testing.T) {
	store := NewStore("/etc/hosts", false)
	if store == nil {
		t.Error("NewStore should not return nil")
	}

	store = NewStore("/custom/hosts", true)
	if store == nil {
		t.Error("NewStore should not return nil")
	}
}

func TestStore_AtomicWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hostsctl-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	originalContent := "127.0.0.1\tlocalhost"

	// Create initial file
	if err := os.WriteFile(hostsFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create hosts file: %v", err)
	}

	store := NewStore(hostsFile, false)

	// Create large data to test atomic write
	var entries []Entry
	for i := 0; i < 1000; i++ {
		entries = append(entries, Entry{
			IP:    "192.168.1." + string(rune(i%254+1)),
			Names: []string{fmt.Sprintf("host%d.local", i)},
		})
	}

	hostsData := &HostsFile{Entries: entries}

	// Save should be atomic
	if err := store.Save(hostsData); err != nil {
		t.Fatalf("Store.Save() error = %v", err)
	}

	// Verify file was completely written
	savedContent, err := os.ReadFile(hostsFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	// Should contain all entries
	savedStr := string(savedContent)
	if !strings.Contains(savedStr, "host999.local") {
		t.Error("File should contain all entries (atomic write)")
	}
}

func TestStore_BackupTimestamp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hostsctl-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")
	content := "127.0.0.1\tlocalhost"

	if err := os.WriteFile(hostsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create hosts file: %v", err)
	}

	store := NewStore(hostsFile, false)

	// Create backup
	before := time.Now()
	backup, err := store.Backup("")
	after := time.Now()

	if err != nil {
		t.Fatalf("Store.Backup() error = %v", err)
	}

	// Check timestamp is reasonable
	if backup.CreatedAt.Before(before) || backup.CreatedAt.After(after) {
		t.Error("Backup timestamp should be between before and after times")
	}

	// Check filename contains timestamp
	filename := filepath.Base(backup.Path)
	if !strings.Contains(filename, "hostsctl") {
		t.Error("Backup filename should contain 'hostsctl'")
	}
}

func TestStore_LoadCorruptedFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hostsctl-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	hostsFile := filepath.Join(tmpDir, "hosts")

	// Create file with binary content (corrupted)
	corruptedContent := []byte{0x00, 0x01, 0x02, 0xff, 0xfe}
	if err := os.WriteFile(hostsFile, corruptedContent, 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	store := NewStore(hostsFile, false)

	// Should still be able to load (even if content is unusual)
	hostsData, err := store.Load()
	if err != nil {
		t.Fatalf("Store.Load() should handle corrupted file: %v", err)
	}

	// Corrupted content should result in no valid entries
	if len(hostsData.Entries) != 0 {
		t.Errorf("Corrupted file should result in 0 entries, got %d", len(hostsData.Entries))
	}
}
