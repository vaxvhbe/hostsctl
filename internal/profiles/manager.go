// Package profiles provides functionality for managing persistent hostname profiles.
// Profiles allow users to save, load, and switch between different sets of hosts entries.
package profiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/vaxvhbe/hostsctl/internal/hosts"
	"github.com/vaxvhbe/hostsctl/pkg"
	yaml "gopkg.in/yaml.v3"
)

// Manager handles persistent storage and retrieval of hostname profiles.
type Manager struct {
	configDir string // Directory where profiles are stored
}

// ProfileMetadata contains summary information about a profile.
type ProfileMetadata struct {
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	FilePath    string    `json:"file_path" yaml:"file_path"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
	EntryCount  int       `json:"entry_count" yaml:"entry_count"`
	Size        int64     `json:"size" yaml:"size"`
}

// NewManager creates a new profile manager instance.
// It initializes the configuration directory if it doesn't exist.
func NewManager() (*Manager, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	manager := &Manager{
		configDir: configDir,
	}

	if err := manager.ensureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return manager, nil
}

// getConfigDir returns the appropriate configuration directory for the current user.
// It follows XDG Base Directory specification on Unix systems.
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configDir, "hostsctl", "profiles"), nil
}

// ensureConfigDir creates the configuration directory if it doesn't exist.
func (m *Manager) ensureConfigDir() error {
	return os.MkdirAll(m.configDir, 0750)
}

// SaveProfile saves a profile to persistent storage.
// The profile is stored as a JSON file in the config directory.
func (m *Manager) SaveProfile(profile *hosts.Profile) error {
	if err := m.validateProfileName(profile.Name); err != nil {
		return err
	}

	profile.UpdatedAt = time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now()
	}

	filePath := m.getProfilePath(profile.Name)

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

// LoadProfile loads a profile from persistent storage.
func (m *Manager) LoadProfile(name string) (*hosts.Profile, error) {
	if err := m.validateProfileName(name); err != nil {
		return nil, err
	}

	filePath := m.getProfilePath(name)

	// Validate file path to prevent directory traversal
	if err := pkg.ValidateSecurePath(filePath); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	data, err := os.ReadFile(filePath) // #nosec G304 -- path validated above
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile hosts.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile file: %w", err)
	}

	return &profile, nil
}

// DeleteProfile removes a profile from persistent storage.
func (m *Manager) DeleteProfile(name string) error {
	if err := m.validateProfileName(name); err != nil {
		return err
	}

	filePath := m.getProfilePath(name)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("profile '%s' not found", name)
		}
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// ListProfiles returns metadata for all saved profiles.
func (m *Manager) ListProfiles() ([]*ProfileMetadata, error) {
	files, err := os.ReadDir(m.configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*ProfileMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var profiles []*ProfileMetadata

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".json")
		metadata, err := m.getProfileMetadata(name)
		if err != nil {
			continue // Skip invalid profiles
		}

		profiles = append(profiles, metadata)
	}

	// Sort profiles by name
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles, nil
}

// getProfileMetadata returns metadata for a specific profile without loading the entire profile.
func (m *Manager) getProfileMetadata(name string) (*ProfileMetadata, error) {
	filePath := m.getProfilePath(name)

	// Validate file path to prevent directory traversal
	if err := pkg.ValidateSecurePath(filePath); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Read just enough to get metadata
	data, err := os.ReadFile(filePath) // #nosec G304 -- path validated above
	if err != nil {
		return nil, err
	}

	var profile hosts.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	return &ProfileMetadata{
		Name:        profile.Name,
		Description: profile.Description,
		FilePath:    filePath,
		CreatedAt:   profile.CreatedAt,
		UpdatedAt:   profile.UpdatedAt,
		EntryCount:  len(profile.Entries),
		Size:        stat.Size(),
	}, nil
}

// ExistsProfile checks if a profile with the given name exists.
func (m *Manager) ExistsProfile(name string) bool {
	if err := m.validateProfileName(name); err != nil {
		return false
	}

	filePath := m.getProfilePath(name)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// getProfilePath returns the file system path for a profile.
func (m *Manager) getProfilePath(name string) string {
	return filepath.Join(m.configDir, name+".json")
}

// validateProfileName ensures a profile name is valid for file system storage.
func (m *Manager) validateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("profile name too long (max 100 characters)")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("profile name contains invalid character: %s", char)
		}
	}

	// Prevent hidden files and reserved names
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("profile name cannot start with dot")
	}

	reservedNames := []string{"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}
	lowercaseName := strings.ToLower(name)
	if slices.Contains(reservedNames, lowercaseName) {
		return fmt.Errorf("profile name is reserved: %s", name)
	}

	return nil
}

// ExportProfile exports a profile to a file in the specified format.
func (m *Manager) ExportProfile(name, outputPath, format string) error {
	profile, err := m.LoadProfile(name)
	if err != nil {
		return err
	}

	var data []byte
	switch strings.ToLower(format) {
	case "json":
		data, err = json.MarshalIndent(profile, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(profile)
	default:
		return fmt.Errorf("unsupported export format: %s (supported: json, yaml)", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// ImportProfile imports a profile from a file and saves it to the profile store.
func (m *Manager) ImportProfile(filePath, format string, overwrite bool) (*hosts.Profile, error) {
	// Validate file path to prevent directory traversal
	if err := pkg.ValidateSecurePath(filePath); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	data, err := os.ReadFile(filePath) // #nosec G304 -- path validated above
	if err != nil {
		return nil, fmt.Errorf("failed to read import file: %w", err)
	}

	var profile hosts.Profile
	switch strings.ToLower(format) {
	case "json":
		err = json.Unmarshal(data, &profile)
	case "yaml":
		err = yaml.Unmarshal(data, &profile)
	default:
		return nil, fmt.Errorf("unsupported import format: %s (supported: json, yaml)", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse import file: %w", err)
	}

	if profile.Name == "" {
		profile.Name = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	if !overwrite && m.ExistsProfile(profile.Name) {
		return nil, fmt.Errorf("profile '%s' already exists (use --overwrite to replace)", profile.Name)
	}

	if err := m.SaveProfile(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// CreateFromCurrentHosts creates a new profile from the current hosts file state.
func (m *Manager) CreateFromCurrentHosts(name, description, hostsFilePath string) (*hosts.Profile, error) {
	store := hosts.NewStore(hostsFilePath, false)
	hostsFile, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load current hosts file: %w", err)
	}

	profile := &hosts.Profile{
		Name:        name,
		Description: description,
		Entries:     hostsFile.Entries,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := m.SaveProfile(profile); err != nil {
		return nil, err
	}

	return profile, nil
}
