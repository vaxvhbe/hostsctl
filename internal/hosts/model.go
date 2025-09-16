// Package hosts provides data models and operations for managing hosts file entries.
package hosts

import (
	"time"
)

// Entry represents a single entry in a hosts file.
// It can contain an IP address, one or more hostnames, an optional comment,
// and can be disabled (commented out in the file).
type Entry struct {
	ID       int      `json:"id" yaml:"id"`             // Unique identifier for the entry
	IP       string   `json:"ip" yaml:"ip"`             // IP address (IPv4 or IPv6)
	Names    []string `json:"names" yaml:"names"`       // List of hostnames for this IP
	Comment  string   `json:"comment" yaml:"comment"`   // Optional comment
	Disabled bool     `json:"disabled" yaml:"disabled"` // Whether the entry is commented out
	Raw      string   `json:"-" yaml:"-"`               // Original raw line from file (not exported)
}

// Profile represents a collection of hosts entries that can be imported/exported.
// Profiles allow grouping related hosts entries for different environments
// (development, production, etc.).
type Profile struct {
	Name        string    `json:"name" yaml:"name"`               // Profile name
	Description string    `json:"description" yaml:"description"` // Profile description
	Entries     []Entry   `json:"entries" yaml:"entries"`         // List of entries in this profile
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`   // Creation timestamp
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`   // Last update timestamp
}

// HostsFile represents a complete hosts file with all its entries.
type HostsFile struct {
	Entries []Entry `json:"entries" yaml:"entries"` // List of all entries in the file
	Path    string  `json:"path" yaml:"path"`       // Path to the hosts file
}

// BackupInfo contains metadata about a hosts file backup.
type BackupInfo struct {
	Path      string    `json:"path" yaml:"path"`             // Path to the backup file
	Original  string    `json:"original" yaml:"original"`     // Path to the original file
	CreatedAt time.Time `json:"created_at" yaml:"created_at"` // Backup creation time
	Size      int64     `json:"size" yaml:"size"`             // Size of the backup file in bytes
}

// ExportFormat represents the supported export formats for profiles.
type ExportFormat string

const (
	FormatJSON ExportFormat = "json" // JSON export format
	FormatYAML ExportFormat = "yaml" // YAML export format
)

// IsValid checks if the entry has the minimum required fields.
// An entry is valid if it has both an IP address and at least one hostname.
func (e *Entry) IsValid() bool {
	return e.IP != "" && len(e.Names) > 0
}

// String returns the string representation of the entry as it would appear in a hosts file.
// The format is: [# ]IP<tab>hostname1[<tab>hostname2...][<tab># comment]
// If the entry is disabled, it will be prefixed with "# ".
func (e *Entry) String() string {
	line := e.IP
	for _, name := range e.Names {
		line += "\t" + name
	}

	if e.Comment != "" {
		line += "\t# " + e.Comment
	}

	if e.Disabled {
		line = "# " + line
	}

	return line
}

// AddEntry adds a new entry to the profile and updates the modification timestamp.
func (p *Profile) AddEntry(entry Entry) {
	p.Entries = append(p.Entries, entry)
	p.UpdatedAt = time.Now()
}

// RemoveEntry removes an entry from the profile by ID.
// Returns true if the entry was found and removed, false otherwise.
func (p *Profile) RemoveEntry(id int) bool {
	for i, entry := range p.Entries {
		if entry.ID == id {
			p.Entries = append(p.Entries[:i], p.Entries[i+1:]...)
			p.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// FindByID searches for an entry by its ID and returns a pointer to it.
// Returns nil if no entry with the given ID is found.
func (h *HostsFile) FindByID(id int) *Entry {
	for i := range h.Entries {
		if h.Entries[i].ID == id {
			return &h.Entries[i]
		}
	}
	return nil
}

// FindByName searches for entries that contain the given hostname.
// Returns a slice of pointers to all matching entries.
func (h *HostsFile) FindByName(name string) []*Entry {
	var results []*Entry
	for i := range h.Entries {
		for _, entryName := range h.Entries[i].Names {
			if entryName == name {
				results = append(results, &h.Entries[i])
				break
			}
		}
	}
	return results
}

// AddEntry adds a new entry to the hosts file.
// Automatically assigns a unique ID to the entry.
func (h *HostsFile) AddEntry(entry Entry) {
	if len(h.Entries) == 0 {
		entry.ID = 1
	} else {
		maxID := 0
		for _, e := range h.Entries {
			if e.ID > maxID {
				maxID = e.ID
			}
		}
		entry.ID = maxID + 1
	}
	h.Entries = append(h.Entries, entry)
}

// RemoveEntry removes an entry from the hosts file by ID.
// Returns true if the entry was found and removed, false otherwise.
func (h *HostsFile) RemoveEntry(id int) bool {
	for i, entry := range h.Entries {
		if entry.ID == id {
			h.Entries = append(h.Entries[:i], h.Entries[i+1:]...)
			return true
		}
	}
	return false
}

// EnableEntry enables (uncomments) an entry by ID.
// Returns true if the entry was found and enabled, false otherwise.
func (h *HostsFile) EnableEntry(id int) bool {
	entry := h.FindByID(id)
	if entry != nil {
		entry.Disabled = false
		return true
	}
	return false
}

// DisableEntry disables (comments out) an entry by ID.
// Returns true if the entry was found and disabled, false otherwise.
func (h *HostsFile) DisableEntry(id int) bool {
	entry := h.FindByID(id)
	if entry != nil {
		entry.Disabled = true
		return true
	}
	return false
}
