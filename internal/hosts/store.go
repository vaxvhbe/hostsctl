package hosts

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Store handles atomic reading and writing of hosts files with safety features.
// It provides backup creation, atomic writes, and permission checking.
type Store struct {
	path   string  // Path to the hosts file
	parser *Parser // Parser instance for reading/writing
}

// NewStore creates a new Store instance for the specified hosts file path.
// The strict parameter controls whether parsing errors should fail or be skipped.
func NewStore(path string, strict bool) *Store {
	return &Store{
		path:   path,
		parser: NewParser(strict),
	}
}

// Load reads and parses the hosts file from disk.
// Returns a HostsFile containing all parsed entries.
func (s *Store) Load() (*HostsFile, error) {
	file, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open hosts file: %w", err)
	}
	defer file.Close()

	hostsFile, err := s.parser.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hosts file: %w", err)
	}

	hostsFile.Path = s.path
	return hostsFile, nil
}

// Save writes a HostsFile to disk atomically with automatic backup.
// It checks permissions, creates a backup, writes to a temporary file,
// and then atomically renames it to replace the original.
func (s *Store) Save(hostsFile *HostsFile) error {
	if err := s.requiresRoot(); err != nil {
		return err
	}

	content := s.parser.Serialize(hostsFile)

	if err := s.createBackup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	tempPath := s.path + ".tmp"
	if err := s.writeTemp(tempPath, content); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := os.Rename(tempPath, s.path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to atomically replace hosts file: %w", err)
	}

	return nil
}

// createBackup creates a timestamped backup of the current hosts file.
// The backup file is named with the current timestamp.
func (s *Store) createBackup() error {
	backupPath := fmt.Sprintf("%s.hostsctl.%s.bak", s.path, time.Now().Format("20060102-150405"))

	sourceFile, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	backupFile, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer backupFile.Close()

	_, err = io.Copy(backupFile, sourceFile)
	return err
}

// writeTemp writes content to a temporary file with fsync for durability.
func (s *Store) writeTemp(tempPath, content string) error {
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		return err
	}

	return tempFile.Sync()
}

// requiresRoot checks if the operation requires root privileges.
// Returns an error if trying to modify /etc/hosts without root access.
func (s *Store) requiresRoot() error {
	if s.path == "/etc/hosts" && os.Geteuid() != 0 {
		return fmt.Errorf("modifying /etc/hosts requires root privileges (run with sudo)")
	}
	return nil
}

// Backup creates a manual backup of the hosts file to the specified path.
// If outputPath is empty, generates a timestamped filename.
// Returns BackupInfo with metadata about the created backup.
func (s *Store) Backup(outputPath string) (*BackupInfo, error) {
	if outputPath == "" {
		outputPath = fmt.Sprintf("%s.hostsctl.%s.bak", s.path, time.Now().Format("20060102-150405"))
	}

	sourceFile, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	stat, err := sourceFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get source file info: %w", err)
	}

	backupFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()

	_, err = io.Copy(backupFile, sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file contents: %w", err)
	}

	return &BackupInfo{
		Path:      outputPath,
		Original:  s.path,
		CreatedAt: time.Now(),
		Size:      stat.Size(),
	}, nil
}

// Restore replaces the current hosts file with content from a backup.
// The backup file is validated before restoration.
func (s *Store) Restore(backupPath string) error {
	if err := s.requiresRoot(); err != nil {
		return err
	}

	backupFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer backupFile.Close()

	hostsFile, err := s.parser.Parse(backupFile)
	if err != nil {
		return fmt.Errorf("failed to parse backup file: %w", err)
	}

	hostsFile.Path = s.path
	return s.Save(hostsFile)
}

// ListBackups finds and returns information about all backup files.
// Searches for files matching the hostsctl backup naming pattern.
func (s *Store) ListBackups() ([]BackupInfo, error) {
	dir := filepath.Dir(s.path)
	base := filepath.Base(s.path)
	pattern := base + ".hostsctl.*.bak"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	var backups []BackupInfo
	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Path:      match,
			Original:  s.path,
			CreatedAt: stat.ModTime(),
			Size:      stat.Size(),
		})
	}

	return backups, nil
}

// Verify checks the hosts file for syntax errors and inconsistencies.
// Returns a list of issues found, or an empty slice if the file is valid.
func (s *Store) Verify() ([]string, error) {
	hostsFile, err := s.Load()
	if err != nil {
		return nil, err
	}

	var issues []string
	seenNames := make(map[string][]int)

	for _, entry := range hostsFile.Entries {
		if !entry.IsValid() {
			issues = append(issues, fmt.Sprintf("entry %d: invalid entry (missing IP or names)", entry.ID))
			continue
		}

		if !s.parser.isValidIP(entry.IP) {
			issues = append(issues, fmt.Sprintf("entry %d: invalid IP address: %s", entry.ID, entry.IP))
		}

		for _, name := range entry.Names {
			if !s.parser.isValidHostname(name) {
				issues = append(issues, fmt.Sprintf("entry %d: invalid hostname: %s", entry.ID, name))
			}

			seenNames[name] = append(seenNames[name], entry.ID)
		}
	}

	for name, ids := range seenNames {
		if len(ids) > 1 {
			issues = append(issues, fmt.Sprintf("duplicate hostname '%s' found in entries: %v", name, ids))
		}
	}

	return issues, nil
}
