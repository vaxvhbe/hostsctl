package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/vaxvhbe/hostsctl/internal/hosts"
	"github.com/vaxvhbe/hostsctl/internal/lock"
	"github.com/vaxvhbe/hostsctl/internal/profiles"
)

// buildProfileCommand creates the main profile command with subcommands.
func (c *CLI) buildProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage hostname profiles",
		Long: `Manage hostname profiles for different environments.

Profiles allow you to save sets of hosts entries and quickly switch between them.
This is useful for development, testing, and production environments.`,
	}

	cmd.AddCommand(c.buildProfileListCommand())
	cmd.AddCommand(c.buildProfileSaveCommand())
	cmd.AddCommand(c.buildProfileApplyCommand())
	cmd.AddCommand(c.buildProfileDeleteCommand())
	cmd.AddCommand(c.buildProfileShowCommand())
	cmd.AddCommand(c.buildProfileDiffCommand())
	cmd.AddCommand(c.buildProfileExportCommand())
	cmd.AddCommand(c.buildProfileImportCommand())

	return cmd
}

// buildProfileListCommand creates the profile list subcommand.
func (c *CLI) buildProfileListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runProfileList()
		},
	}

	return cmd
}

// buildProfileSaveCommand creates the profile save subcommand.
func (c *CLI) buildProfileSaveCommand() *cobra.Command {
	var name, description string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current hosts entries as a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runProfileSave(name, description, overwrite)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Profile name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Profile description")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing profile")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// buildProfileApplyCommand creates the profile apply subcommand.
func (c *CLI) buildProfileApplyCommand() *cobra.Command {
	var merge, backup bool

	cmd := &cobra.Command{
		Use:   "apply [profile-name]",
		Short: "Apply a saved profile to hosts file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runProfileApply(args[0], merge, backup)
		},
	}

	cmd.Flags().BoolVar(&merge, "merge", false, "Merge with existing entries instead of replacing")
	cmd.Flags().BoolVar(&backup, "backup", true, "Create backup before applying")

	return cmd
}

// buildProfileDeleteCommand creates the profile delete subcommand.
func (c *CLI) buildProfileDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [profile-name]",
		Short: "Delete a saved profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runProfileDelete(args[0], force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Delete without confirmation")

	return cmd
}

// buildProfileShowCommand creates the profile show subcommand.
func (c *CLI) buildProfileShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [profile-name]",
		Short: "Show profile details and entries",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runProfileShow(args[0])
		},
	}

	return cmd
}

// buildProfileDiffCommand creates the profile diff subcommand.
func (c *CLI) buildProfileDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [profile-name]",
		Short: "Compare profile with current hosts file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runProfileDiff(args[0])
		},
	}

	return cmd
}

// buildProfileExportCommand creates the profile export subcommand.
func (c *CLI) buildProfileExportCommand() *cobra.Command {
	var output, format string

	cmd := &cobra.Command{
		Use:   "export [profile-name]",
		Short: "Export profile to file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runProfileExport(args[0], output, format)
		},
	}

	cmd.Flags().StringVar(&output, "output", "", "Output file path (required)")
	cmd.Flags().StringVar(&format, "format", "json", "Export format (json|yaml)")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}

// buildProfileImportCommand creates the profile import subcommand.
func (c *CLI) buildProfileImportCommand() *cobra.Command {
	var format string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "import [file-path]",
		Short: "Import profile from file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runProfileImport(args[0], format, overwrite)
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "Import format (json|yaml)")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing profile")

	return cmd
}

// runProfileList lists all saved profiles.
func (c *CLI) runProfileList() error {
	manager, err := profiles.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	profileList, err := manager.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if c.jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(profileList)
	}

	if len(profileList) == 0 {
		fmt.Println("No profiles found. Use 'hostsctl profile save' to create one.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tDESCRIPTION\tENTRIES\tCREATED\tUPDATED")
	_, _ = fmt.Fprintln(w, "----\t-----------\t-------\t-------\t-------")

	for _, profile := range profileList {
		description := profile.Description
		if len(description) > 40 {
			description = description[:37] + "..."
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
			profile.Name,
			description,
			profile.EntryCount,
			profile.CreatedAt.Format("2006-01-02"),
			profile.UpdatedAt.Format("2006-01-02"))
	}

	_ = w.Flush()
	return nil
}

// runProfileSave saves the current hosts file as a profile.
func (c *CLI) runProfileSave(name, description string, overwrite bool) error {
	manager, err := profiles.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	if !overwrite && manager.ExistsProfile(name) {
		return fmt.Errorf("profile '%s' already exists (use --overwrite to replace)", name)
	}

	profile, err := manager.CreateFromCurrentHosts(name, description, c.hostsFile)
	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	if c.jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(profile)
	}

	action := "Created"
	if overwrite {
		action = "Updated"
	}

	fmt.Printf("%s profile '%s' with %d entries\n", action, profile.Name, len(profile.Entries))
	if profile.Description != "" {
		fmt.Printf("Description: %s\n", profile.Description)
	}

	return nil
}

// runProfileApply applies a saved profile to the hosts file.
func (c *CLI) runProfileApply(name string, merge, backup bool) error {
	manager, err := profiles.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	profile, err := manager.LoadProfile(name)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	return lock.WithQuickLock(c.hostsFile, func() error {
		store := hosts.NewStore(c.hostsFile, false)

		var hostsFile *hosts.HostsFile
		if merge {
			// Load current hosts file and merge with profile
			hostsFile, err = store.Load()
			if err != nil {
				return fmt.Errorf("failed to load current hosts file: %w", err)
			}

			// Add profile entries to current hosts
			for _, entry := range profile.Entries {
				hostsFile.AddEntry(entry)
			}
		} else {
			// Replace entire hosts file with profile
			hostsFile = &hosts.HostsFile{
				Entries: profile.Entries,
				Path:    c.hostsFile,
			}
		}

		if backup {
			if _, err := store.Backup(""); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
		}

		if err := store.Save(hostsFile); err != nil {
			return fmt.Errorf("failed to save hosts file: %w", err)
		}

		if c.jsonOutput {
			result := map[string]interface{}{
				"profile":    profile.Name,
				"entries":    len(profile.Entries),
				"merge":      merge,
				"backup":     backup,
				"applied_at": time.Now(),
			}
			return json.NewEncoder(os.Stdout).Encode(result)
		}

		action := "Applied"
		if merge {
			action = "Merged"
		}

		fmt.Printf("%s profile '%s' (%d entries)\n", action, profile.Name, len(profile.Entries))
		if profile.Description != "" {
			fmt.Printf("Description: %s\n", profile.Description)
		}

		return nil
	})
}

// runProfileDelete deletes a saved profile.
func (c *CLI) runProfileDelete(name string, force bool) error {
	manager, err := profiles.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	if !manager.ExistsProfile(name) {
		return fmt.Errorf("profile '%s' not found", name)
	}

	if !force {
		fmt.Printf("Are you sure you want to delete profile '%s'? (y/N): ", name)
		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if err := manager.DeleteProfile(name); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	if c.jsonOutput {
		result := map[string]interface{}{
			"profile":    name,
			"deleted":    true,
			"deleted_at": time.Now(),
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	fmt.Printf("Deleted profile '%s'\n", name)
	return nil
}

// runProfileShow displays detailed information about a profile.
func (c *CLI) runProfileShow(name string) error {
	manager, err := profiles.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	profile, err := manager.LoadProfile(name)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	if c.jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(profile)
	}

	fmt.Printf("Profile: %s\n", profile.Name)
	if profile.Description != "" {
		fmt.Printf("Description: %s\n", profile.Description)
	}
	fmt.Printf("Created: %s\n", profile.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", profile.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("Entries: %d\n\n", len(profile.Entries))

	if len(profile.Entries) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "STATUS\tIP\tHOSTNAMES\tCOMMENT")
		_, _ = fmt.Fprintln(w, "------\t--\t---------\t-------")

		for _, entry := range profile.Entries {
			status := "enabled"
			if entry.Disabled {
				status = "disabled"
			}

			hostnames := strings.Join(entry.Names, ", ")
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", status, entry.IP, hostnames, entry.Comment)
		}

		_ = w.Flush()
	}

	return nil
}

// runProfileDiff compares a profile with the current hosts file.
func (c *CLI) runProfileDiff(name string) error {
	manager, err := profiles.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	profile, err := manager.LoadProfile(name)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	store := hosts.NewStore(c.hostsFile, false)
	current, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load current hosts file: %w", err)
	}

	diff := c.calculateDiff(current.Entries, profile.Entries)

	if c.jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(diff)
	}

	c.printDiff(diff, profile.Name)
	return nil
}

// runProfileExport exports a profile to an external file.
func (c *CLI) runProfileExport(name, output, format string) error {
	manager, err := profiles.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	if err := manager.ExportProfile(name, output, format); err != nil {
		return fmt.Errorf("failed to export profile: %w", err)
	}

	if c.jsonOutput {
		result := map[string]interface{}{
			"profile":     name,
			"output":      output,
			"format":      format,
			"exported_at": time.Now(),
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	fmt.Printf("Exported profile '%s' to %s (%s format)\n", name, output, format)
	return nil
}

// runProfileImport imports a profile from an external file.
func (c *CLI) runProfileImport(filePath, format string, overwrite bool) error {
	manager, err := profiles.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	profile, err := manager.ImportProfile(filePath, format, overwrite)
	if err != nil {
		return fmt.Errorf("failed to import profile: %w", err)
	}

	if c.jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(profile)
	}

	action := "Imported"
	if overwrite {
		action = "Updated"
	}

	fmt.Printf("%s profile '%s' (%d entries) from %s\n", action, profile.Name, len(profile.Entries), filePath)
	if profile.Description != "" {
		fmt.Printf("Description: %s\n", profile.Description)
	}

	return nil
}

// DiffResult represents the result of comparing two sets of hosts entries.
type DiffResult struct {
	Added    []hosts.Entry `json:"added"`
	Removed  []hosts.Entry `json:"removed"`
	Modified []DiffEntry   `json:"modified"`
	Same     []hosts.Entry `json:"same"`
}

// DiffEntry represents a modified entry in a diff.
type DiffEntry struct {
	Old hosts.Entry `json:"old"`
	New hosts.Entry `json:"new"`
}

// calculateDiff compares current hosts entries with profile entries.
func (c *CLI) calculateDiff(current, profile []hosts.Entry) *DiffResult {
	result := &DiffResult{
		Added:    []hosts.Entry{},
		Removed:  []hosts.Entry{},
		Modified: []DiffEntry{},
		Same:     []hosts.Entry{},
	}

	// Create maps for efficient lookup
	currentMap := make(map[string]hosts.Entry)
	profileMap := make(map[string]hosts.Entry)

	for _, entry := range current {
		key := c.entryKey(entry)
		currentMap[key] = entry
	}

	for _, entry := range profile {
		key := c.entryKey(entry)
		profileMap[key] = entry
	}

	// Find added and modified entries
	for key, profileEntry := range profileMap {
		if currentEntry, exists := currentMap[key]; exists {
			if c.entriesEqual(currentEntry, profileEntry) {
				result.Same = append(result.Same, profileEntry)
			} else {
				result.Modified = append(result.Modified, DiffEntry{
					Old: currentEntry,
					New: profileEntry,
				})
			}
		} else {
			result.Added = append(result.Added, profileEntry)
		}
	}

	// Find removed entries
	for key, currentEntry := range currentMap {
		if _, exists := profileMap[key]; !exists {
			result.Removed = append(result.Removed, currentEntry)
		}
	}

	return result
}

// entryKey generates a unique key for a hosts entry for comparison.
func (c *CLI) entryKey(entry hosts.Entry) string {
	return fmt.Sprintf("%s:%s", entry.IP, strings.Join(entry.Names, ","))
}

// entriesEqual compares two hosts entries for equality.
func (c *CLI) entriesEqual(a, b hosts.Entry) bool {
	return a.IP == b.IP &&
		strings.Join(a.Names, ",") == strings.Join(b.Names, ",") &&
		a.Comment == b.Comment &&
		a.Disabled == b.Disabled
}

// printDiff prints a human-readable diff report.
func (c *CLI) printDiff(diff *DiffResult, profileName string) {
	fmt.Printf("Comparing current hosts file with profile '%s':\n\n", profileName)

	if len(diff.Added) > 0 {
		fmt.Printf("Added entries (%d):\n", len(diff.Added))
		for _, entry := range diff.Added {
			fmt.Printf("  + %s %s", entry.IP, strings.Join(entry.Names, " "))
			if entry.Comment != "" {
				fmt.Printf(" # %s", entry.Comment)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	if len(diff.Removed) > 0 {
		fmt.Printf("Removed entries (%d):\n", len(diff.Removed))
		for _, entry := range diff.Removed {
			fmt.Printf("  - %s %s", entry.IP, strings.Join(entry.Names, " "))
			if entry.Comment != "" {
				fmt.Printf(" # %s", entry.Comment)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	if len(diff.Modified) > 0 {
		fmt.Printf("Modified entries (%d):\n", len(diff.Modified))
		for _, mod := range diff.Modified {
			fmt.Printf("  ~ %s %s", mod.Old.IP, strings.Join(mod.Old.Names, " "))
			if mod.Old.Comment != "" {
				fmt.Printf(" # %s", mod.Old.Comment)
			}
			fmt.Println()
			fmt.Printf("    %s %s", mod.New.IP, strings.Join(mod.New.Names, " "))
			if mod.New.Comment != "" {
				fmt.Printf(" # %s", mod.New.Comment)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	fmt.Printf("Summary: %d added, %d removed, %d modified, %d unchanged\n",
		len(diff.Added), len(diff.Removed), len(diff.Modified), len(diff.Same))
}
