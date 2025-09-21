package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vaxvhbe/hostsctl/internal/hosts"
	"github.com/vaxvhbe/hostsctl/internal/profiles"
)

func TestProfileManager_SaveAndLoad(t *testing.T) {
	// Create temporary directory for test profiles
	tmpDir, err := os.MkdirTemp("", "hostsctl-profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override config directory
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := profiles.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create test profile
	profile := &hosts.Profile{
		Name:        "test-profile",
		Description: "Test profile for unit tests",
		Entries: []hosts.Entry{
			{
				ID:       1,
				IP:       "127.0.0.1",
				Names:    []string{"test.local", "test.dev"},
				Comment:  "Test entry",
				Disabled: false,
			},
			{
				ID:       2,
				IP:       "192.168.1.100",
				Names:    []string{"server.local"},
				Comment:  "Server entry",
				Disabled: true,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test saving profile
	if err := manager.SaveProfile(profile); err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Test loading profile
	loadedProfile, err := manager.LoadProfile("test-profile")
	if err != nil {
		t.Fatalf("Failed to load profile: %v", err)
	}

	// Verify profile data
	if loadedProfile.Name != profile.Name {
		t.Errorf("Expected name %s, got %s", profile.Name, loadedProfile.Name)
	}

	if loadedProfile.Description != profile.Description {
		t.Errorf("Expected description %s, got %s", profile.Description, loadedProfile.Description)
	}

	if len(loadedProfile.Entries) != len(profile.Entries) {
		t.Errorf("Expected %d entries, got %d", len(profile.Entries), len(loadedProfile.Entries))
	}

	// Verify first entry
	if len(loadedProfile.Entries) > 0 {
		entry := loadedProfile.Entries[0]
		if entry.IP != "127.0.0.1" {
			t.Errorf("Expected IP 127.0.0.1, got %s", entry.IP)
		}
		if len(entry.Names) != 2 || entry.Names[0] != "test.local" || entry.Names[1] != "test.dev" {
			t.Errorf("Expected names [test.local, test.dev], got %v", entry.Names)
		}
	}
}

func TestProfileManager_ListProfiles(t *testing.T) {
	// Create temporary directory for test profiles
	tmpDir, err := os.MkdirTemp("", "hostsctl-profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override config directory
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := profiles.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Initially should have no profiles
	profileList, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}
	if len(profileList) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profileList))
	}

	// Create test profiles
	profile1 := &hosts.Profile{
		Name:        "dev",
		Description: "Development environment",
		Entries:     []hosts.Entry{{ID: 1, IP: "127.0.0.1", Names: []string{"dev.local"}}},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	profile2 := &hosts.Profile{
		Name:        "prod",
		Description: "Production environment",
		Entries:     []hosts.Entry{{ID: 1, IP: "10.0.0.1", Names: []string{"api.prod.com"}}},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save profiles
	if err := manager.SaveProfile(profile1); err != nil {
		t.Fatalf("Failed to save profile1: %v", err)
	}
	if err := manager.SaveProfile(profile2); err != nil {
		t.Fatalf("Failed to save profile2: %v", err)
	}

	// List profiles
	profileList, err = manager.ListProfiles()
	if err != nil {
		t.Fatalf("Failed to list profiles: %v", err)
	}

	if len(profileList) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profileList))
	}

	// Verify profiles are sorted by name
	if profileList[0].Name != "dev" || profileList[1].Name != "prod" {
		t.Errorf("Profiles not sorted correctly: got %s, %s", profileList[0].Name, profileList[1].Name)
	}

	// Verify metadata
	devProfile := profileList[0]
	if devProfile.Description != "Development environment" {
		t.Errorf("Expected dev description 'Development environment', got %s", devProfile.Description)
	}
	if devProfile.EntryCount != 1 {
		t.Errorf("Expected dev entry count 1, got %d", devProfile.EntryCount)
	}
}

func TestProfileManager_DeleteProfile(t *testing.T) {
	// Create temporary directory for test profiles
	tmpDir, err := os.MkdirTemp("", "hostsctl-profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override config directory
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := profiles.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create and save test profile
	profile := &hosts.Profile{
		Name:        "test-delete",
		Description: "Profile to delete",
		Entries:     []hosts.Entry{{ID: 1, IP: "127.0.0.1", Names: []string{"delete.local"}}},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := manager.SaveProfile(profile); err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Verify profile exists
	if !manager.ExistsProfile("test-delete") {
		t.Error("Profile should exist before deletion")
	}

	// Delete profile
	if err := manager.DeleteProfile("test-delete"); err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}

	// Verify profile no longer exists
	if manager.ExistsProfile("test-delete") {
		t.Error("Profile should not exist after deletion")
	}

	// Try to delete non-existent profile
	err = manager.DeleteProfile("non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent profile")
	}
}

func TestProfileManager_ValidateProfileName(t *testing.T) {
	// Create temporary directory for test profiles
	tmpDir, err := os.MkdirTemp("", "hostsctl-profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override config directory
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := profiles.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	validNames := []string{
		"dev",
		"production",
		"test-env",
		"env_2023",
		"staging123",
	}

	invalidNames := []string{
		"",          // empty
		".hidden",   // starts with dot
		"con",       // reserved name
		"test/bad",  // contains slash
		"test\\bad", // contains backslash
		"test:bad",  // contains colon
		"test*bad",  // contains asterisk
		"test?bad",  // contains question mark
		"test\"bad", // contains quote
		"test<bad",  // contains less than
		"test>bad",  // contains greater than
		"test|bad",  // contains pipe
		"test\nbad", // contains newline
	}

	// Test valid names
	for _, name := range validNames {
		profile := &hosts.Profile{
			Name:      name,
			Entries:   []hosts.Entry{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := manager.SaveProfile(profile); err != nil {
			t.Errorf("Valid name '%s' should be accepted, got error: %v", name, err)
		}
	}

	// Test invalid names
	for _, name := range invalidNames {
		profile := &hosts.Profile{
			Name:      name,
			Entries:   []hosts.Entry{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := manager.SaveProfile(profile); err == nil {
			t.Errorf("Invalid name '%s' should be rejected", name)
		}
	}
}

func TestProfileManager_ImportExport(t *testing.T) {
	// Create temporary directory for test profiles
	tmpDir, err := os.MkdirTemp("", "hostsctl-profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override config directory
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := profiles.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create test profile
	profile := &hosts.Profile{
		Name:        "export-test",
		Description: "Profile for export testing",
		Entries: []hosts.Entry{
			{
				ID:       1,
				IP:       "127.0.0.1",
				Names:    []string{"export.local"},
				Comment:  "Export test",
				Disabled: false,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save profile
	if err := manager.SaveProfile(profile); err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Test JSON export
	jsonFile := filepath.Join(tmpDir, "test-export.json")
	if err := manager.ExportProfile("export-test", jsonFile, "json"); err != nil {
		t.Fatalf("Failed to export profile to JSON: %v", err)
	}

	// Verify JSON file exists
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		t.Error("JSON export file should exist")
	}

	// Test JSON import
	importedProfile, err := manager.ImportProfile(jsonFile, "json", true)
	if err != nil {
		t.Fatalf("Failed to import profile from JSON: %v", err)
	}

	if importedProfile.Name != profile.Name {
		t.Errorf("Expected imported name %s, got %s", profile.Name, importedProfile.Name)
	}

	// Test YAML export
	yamlFile := filepath.Join(tmpDir, "test-export.yaml")
	if err := manager.ExportProfile("export-test", yamlFile, "yaml"); err != nil {
		t.Fatalf("Failed to export profile to YAML: %v", err)
	}

	// Verify YAML file exists
	if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
		t.Error("YAML export file should exist")
	}

	// Test invalid format
	if err := manager.ExportProfile("export-test", "test.txt", "invalid"); err == nil {
		t.Error("Expected error for invalid export format")
	}
}
