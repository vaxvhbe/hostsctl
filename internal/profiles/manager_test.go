package profiles

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vaxvhbe/hostsctl/internal/hosts"
)

func TestNewManager(t *testing.T) {
	// Set up temporary config directory
	tmpDir, err := os.MkdirTemp("", "profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override config directory
	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager == nil {
		t.Error("NewManager() should not return nil")
	}

	// Verify config directory was created
	expectedDir := filepath.Join(tmpDir, "hostsctl", "profiles")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Error("Config directory should be created")
	}
}

func TestManager_SaveAndLoadProfile(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create test profile
	profile := &hosts.Profile{
		Name:        "test-profile",
		Description: "Test profile",
		Entries: []hosts.Entry{
			{ID: 1, IP: "127.0.0.1", Names: []string{"localhost"}, Comment: "Local"},
			{ID: 2, IP: "192.168.1.1", Names: []string{"server.local"}, Comment: "Server"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test save
	if err := manager.SaveProfile(profile); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}

	// Test load
	loadedProfile, err := manager.LoadProfile("test-profile")
	if err != nil {
		t.Fatalf("LoadProfile() error = %v", err)
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
}

func TestManager_ListProfiles(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Initially should have no profiles
	profiles, err := manager.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}

	// Create test profiles
	profile1 := &hosts.Profile{
		Name:        "dev",
		Description: "Development",
		Entries:     []hosts.Entry{{ID: 1, IP: "127.0.0.1", Names: []string{"dev.local"}}},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	profile2 := &hosts.Profile{
		Name:        "prod",
		Description: "Production",
		Entries:     []hosts.Entry{{ID: 1, IP: "10.0.0.1", Names: []string{"api.prod.com"}}},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save profiles
	if err := manager.SaveProfile(profile1); err != nil {
		t.Fatalf("SaveProfile(profile1) error = %v", err)
	}
	if err := manager.SaveProfile(profile2); err != nil {
		t.Fatalf("SaveProfile(profile2) error = %v", err)
	}

	// List profiles
	profiles, err = manager.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}

	// Verify profiles are sorted by name
	if len(profiles) >= 2 {
		if profiles[0].Name > profiles[1].Name {
			t.Error("Profiles should be sorted by name")
		}
	}
}

func TestManager_DeleteProfile(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create and save test profile
	profile := &hosts.Profile{
		Name:        "test-delete",
		Description: "Profile to delete",
		Entries:     []hosts.Entry{{ID: 1, IP: "127.0.0.1", Names: []string{"test.local"}}},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := manager.SaveProfile(profile); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}

	// Verify profile exists
	if !manager.ExistsProfile("test-delete") {
		t.Error("Profile should exist before deletion")
	}

	// Delete profile
	if err := manager.DeleteProfile("test-delete"); err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
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

func TestManager_ExistsProfile(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Test non-existent profile
	if manager.ExistsProfile("non-existent") {
		t.Error("ExistsProfile should return false for non-existent profile")
	}

	// Create profile
	profile := &hosts.Profile{
		Name:      "existing",
		Entries:   []hosts.Entry{{ID: 1, IP: "127.0.0.1", Names: []string{"test.local"}}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := manager.SaveProfile(profile); err != nil {
		t.Fatalf("SaveProfile() error = %v", err)
	}

	// Test existing profile
	if !manager.ExistsProfile("existing") {
		t.Error("ExistsProfile should return true for existing profile")
	}
}

func TestManager_ProfileValidation(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "profiles-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	_ = os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() { _ = os.Setenv("XDG_CONFIG_HOME", originalXDGConfig) }()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Test invalid profile names
	invalidNames := []string{
		"",          // empty
		".hidden",   // starts with dot
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

	for _, name := range invalidNames {
		profile := &hosts.Profile{
			Name:      name,
			Entries:   []hosts.Entry{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := manager.SaveProfile(profile); err == nil {
			t.Errorf("SaveProfile should reject invalid name: %q", name)
		}
	}

	// Test valid profile names
	validNames := []string{
		"dev",
		"production",
		"test-env",
		"env_2023",
		"staging123",
	}

	for _, name := range validNames {
		profile := &hosts.Profile{
			Name:      name,
			Entries:   []hosts.Entry{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := manager.SaveProfile(profile); err != nil {
			t.Errorf("SaveProfile should accept valid name %q: %v", name, err)
		}
	}
}

func TestProfileMetadata_Structure(t *testing.T) {
	now := time.Now()
	metadata := ProfileMetadata{
		Name:        "test",
		Description: "Test profile",
		FilePath:    "/path/to/profile.json",
		CreatedAt:   now,
		UpdatedAt:   now,
		EntryCount:  5,
		Size:        1024,
	}

	if metadata.Name != "test" {
		t.Error("ProfileMetadata.Name not set correctly")
	}
	if metadata.Description != "Test profile" {
		t.Error("ProfileMetadata.Description not set correctly")
	}
	if metadata.FilePath != "/path/to/profile.json" {
		t.Error("ProfileMetadata.FilePath not set correctly")
	}
	if metadata.CreatedAt != now {
		t.Error("ProfileMetadata.CreatedAt not set correctly")
	}
	if metadata.UpdatedAt != now {
		t.Error("ProfileMetadata.UpdatedAt not set correctly")
	}
	if metadata.EntryCount != 5 {
		t.Error("ProfileMetadata.EntryCount not set correctly")
	}
	if metadata.Size != 1024 {
		t.Error("ProfileMetadata.Size not set correctly")
	}
}
