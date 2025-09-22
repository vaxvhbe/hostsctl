package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/vaxvhbe/hostsctl/internal/hosts"
	"github.com/vaxvhbe/hostsctl/internal/lock"
	"github.com/vaxvhbe/hostsctl/pkg"
	yaml "gopkg.in/yaml.v3"
)

type CLI struct {
	hostsFile  string
	noColor    bool
	jsonOutput bool
}

// ListFilters contains filtering options for the list command.
type ListFilters struct {
	ShowAll       bool
	IPFilter      string
	CommentFilter string
	NameFilter    string
	StatusFilter  string
}

func NewCLI() *CLI {
	return &CLI{
		hostsFile: "/etc/hosts",
	}
}

func (c *CLI) Execute() error {
	rootCmd := c.buildRootCommand()
	return rootCmd.Execute()
}

func (c *CLI) buildRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "hostsctl",
		Short: "A CLI manager for /etc/hosts",
		Long:  "hostsctl is a command-line tool for safely managing entries in /etc/hosts files.",
	}

	rootCmd.PersistentFlags().StringVar(&c.hostsFile, "hosts-file", "/etc/hosts", "Path to hosts file")
	rootCmd.PersistentFlags().BoolVar(&c.noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&c.jsonOutput, "json", false, "Output in JSON format")

	rootCmd.AddCommand(c.buildListCommand())
	rootCmd.AddCommand(c.buildAddCommand())
	rootCmd.AddCommand(c.buildRemoveCommand())
	rootCmd.AddCommand(c.buildEnableCommand())
	rootCmd.AddCommand(c.buildDisableCommand())
	rootCmd.AddCommand(c.buildBackupCommand())
	rootCmd.AddCommand(c.buildRestoreCommand())
	rootCmd.AddCommand(c.buildImportCommand())
	rootCmd.AddCommand(c.buildExportCommand())
	rootCmd.AddCommand(c.buildVerifyCommand())
	rootCmd.AddCommand(c.buildProfileCommand())
	rootCmd.AddCommand(c.buildSearchCommand())
	rootCmd.AddCommand(c.buildCompletionCommand())

	// Setup custom completions
	c.setupCompletions(rootCmd)

	return rootCmd
}

func (c *CLI) buildListCommand() *cobra.Command {
	var showAll bool
	var filterIP, filterComment, filterName, filterStatus string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List hosts entries",
		Long: `List hosts entries with optional filtering.

Filters can be used to narrow down the results:
  --ip-filter     Show entries matching IP pattern (supports wildcards)
  --name-filter   Show entries matching hostname pattern
  --comment-filter Show entries matching comment pattern
  --status-filter Show entries with specific status (enabled|disabled)

Examples:
  hostsctl list --all                    # Show all entries
  hostsctl list --ip-filter "192.168.*" # Show local network entries
  hostsctl list --name-filter "*.local" # Show .local domains
  hostsctl list --status enabled        # Show only enabled entries`,
		RunE: func(cmd *cobra.Command, args []string) error {
			filters := ListFilters{
				ShowAll:       showAll,
				IPFilter:      filterIP,
				CommentFilter: filterComment,
				NameFilter:    filterName,
				StatusFilter:  filterStatus,
			}
			return c.runListWithFilters(filters)
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", false, "Show all entries including disabled ones")
	cmd.Flags().StringVar(&filterIP, "ip-filter", "", "Filter by IP address pattern (supports wildcards)")
	cmd.Flags().StringVar(&filterComment, "comment-filter", "", "Filter by comment pattern")
	cmd.Flags().StringVar(&filterName, "name-filter", "", "Filter by hostname pattern")
	cmd.Flags().StringVar(&filterStatus, "status-filter", "", "Filter by status (enabled|disabled)")

	return cmd
}

func (c *CLI) buildAddCommand() *cobra.Command {
	var ip, comment string
	var names []string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new hosts entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runAdd(ip, names, comment)
		},
	}

	cmd.Flags().StringVar(&ip, "ip", "", "IP address (required)")
	cmd.Flags().StringSliceVar(&names, "name", []string{}, "Hostname(s) (required, can be specified multiple times)")
	cmd.Flags().StringVar(&comment, "comment", "", "Comment for the entry")
	_ = cmd.MarkFlagRequired("ip")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (c *CLI) buildRemoveCommand() *cobra.Command {
	var name string
	var id int

	cmd := &cobra.Command{
		Use:   "rm",
		Short: "Remove hosts entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runRemove(id, name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Remove by hostname")
	cmd.Flags().IntVar(&id, "id", 0, "Remove by entry ID")

	return cmd
}

func (c *CLI) buildEnableCommand() *cobra.Command {
	var name string
	var id int

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable hosts entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runEnable(id, name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Enable by hostname")
	cmd.Flags().IntVar(&id, "id", 0, "Enable by entry ID")

	return cmd
}

func (c *CLI) buildDisableCommand() *cobra.Command {
	var name string
	var id int

	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable hosts entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runDisable(id, name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Disable by hostname")
	cmd.Flags().IntVar(&id, "id", 0, "Disable by entry ID")

	return cmd
}

func (c *CLI) buildBackupCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup of hosts file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runBackup(output)
		},
	}

	cmd.Flags().StringVar(&output, "out", "", "Output path for backup")
	return cmd
}

func (c *CLI) buildRestoreCommand() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore hosts file from backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runRestore(file)
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Backup file to restore from (required)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func (c *CLI) buildImportCommand() *cobra.Command {
	var file, format string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import hosts entries from file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runImport(file, format)
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "File to import from (required)")
	cmd.Flags().StringVar(&format, "format", "json", "Import format (json|yaml)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func (c *CLI) buildExportCommand() *cobra.Command {
	var file, format string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export hosts entries to file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runExport(file, format)
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "File to export to (required)")
	cmd.Flags().StringVar(&format, "format", "json", "Export format (json|yaml)")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func (c *CLI) buildVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify hosts file syntax and check for issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runVerify()
		},
	}

	return cmd
}

func (c *CLI) runListWithFilters(filters ListFilters) error {
	store := hosts.NewStore(c.hostsFile, false)

	hostsFile, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load hosts file: %w", err)
	}

	// Apply filters
	filteredEntries := c.applyListFilters(hostsFile.Entries, filters)

	if c.jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(filteredEntries)
	}

	c.printEntriesFiltered(filteredEntries, filters)
	return nil
}

func (c *CLI) runAdd(ip string, names []string, comment string) error {
	if err := pkg.ValidateIP(ip); err != nil {
		return fmt.Errorf("invalid IP: %s", err.Message)
	}

	if errs := pkg.ValidateHostnames(names); len(errs) > 0 {
		return fmt.Errorf("invalid hostnames: %s", errs[0].Message)
	}

	if comment != "" {
		if err := pkg.ValidateComment(comment); err != nil {
			return fmt.Errorf("invalid comment: %s", err.Message)
		}
	}

	return lock.WithQuickLock(c.hostsFile, func() error {
		store := hosts.NewStore(c.hostsFile, false)

		hostsFile, err := store.Load()
		if err != nil {
			return fmt.Errorf("failed to load hosts file: %w", err)
		}

		entry := hosts.Entry{
			IP:      pkg.NormalizeIP(ip),
			Names:   names,
			Comment: comment,
		}

		hostsFile.AddEntry(entry)

		if err := store.Save(hostsFile); err != nil {
			return fmt.Errorf("failed to save hosts file: %w", err)
		}

		fmt.Printf("Added entry: %s -> %s\n", entry.IP, strings.Join(entry.Names, ", "))
		return nil
	})
}

func (c *CLI) runRemove(id int, name string) error {
	if id == 0 && name == "" {
		return fmt.Errorf("either --id or --name must be specified")
	}

	return lock.WithQuickLock(c.hostsFile, func() error {
		store := hosts.NewStore(c.hostsFile, false)

		hostsFile, err := store.Load()
		if err != nil {
			return fmt.Errorf("failed to load hosts file: %w", err)
		}

		if id != 0 {
			if !hostsFile.RemoveEntry(id) {
				return fmt.Errorf("entry with ID %d not found", id)
			}
			fmt.Printf("Removed entry with ID %d\n", id)
		} else {
			entries := hostsFile.FindByName(name)
			if len(entries) == 0 {
				return fmt.Errorf("no entries found with hostname %s", name)
			}

			for _, entry := range entries {
				hostsFile.RemoveEntry(entry.ID)
				fmt.Printf("Removed entry with ID %d (%s)\n", entry.ID, name)
			}
		}

		return store.Save(hostsFile)
	})
}

func (c *CLI) runEnable(id int, name string) error {
	if id == 0 && name == "" {
		return fmt.Errorf("either --id or --name must be specified")
	}

	return lock.WithQuickLock(c.hostsFile, func() error {
		store := hosts.NewStore(c.hostsFile, false)

		hostsFile, err := store.Load()
		if err != nil {
			return fmt.Errorf("failed to load hosts file: %w", err)
		}

		if id != 0 {
			if !hostsFile.EnableEntry(id) {
				return fmt.Errorf("entry with ID %d not found", id)
			}
			fmt.Printf("Enabled entry with ID %d\n", id)
		} else {
			entries := hostsFile.FindByName(name)
			if len(entries) == 0 {
				return fmt.Errorf("no entries found with hostname %s", name)
			}

			for _, entry := range entries {
				hostsFile.EnableEntry(entry.ID)
				fmt.Printf("Enabled entry with ID %d (%s)\n", entry.ID, name)
			}
		}

		return store.Save(hostsFile)
	})
}

func (c *CLI) runDisable(id int, name string) error {
	if id == 0 && name == "" {
		return fmt.Errorf("either --id or --name must be specified")
	}

	return lock.WithQuickLock(c.hostsFile, func() error {
		store := hosts.NewStore(c.hostsFile, false)

		hostsFile, err := store.Load()
		if err != nil {
			return fmt.Errorf("failed to load hosts file: %w", err)
		}

		if id != 0 {
			if !hostsFile.DisableEntry(id) {
				return fmt.Errorf("entry with ID %d not found", id)
			}
			fmt.Printf("Disabled entry with ID %d\n", id)
		} else {
			entries := hostsFile.FindByName(name)
			if len(entries) == 0 {
				return fmt.Errorf("no entries found with hostname %s", name)
			}

			for _, entry := range entries {
				hostsFile.DisableEntry(entry.ID)
				fmt.Printf("Disabled entry with ID %d (%s)\n", entry.ID, name)
			}
		}

		return store.Save(hostsFile)
	})
}

func (c *CLI) runBackup(output string) error {
	store := hosts.NewStore(c.hostsFile, false)

	backup, err := store.Backup(output)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if c.jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(backup)
	}

	fmt.Printf("Backup created: %s\n", backup.Path)
	fmt.Printf("Size: %d bytes\n", backup.Size)
	fmt.Printf("Created: %s\n", backup.CreatedAt.Format(time.RFC3339))
	return nil
}

func (c *CLI) runRestore(file string) error {
	return lock.WithQuickLock(c.hostsFile, func() error {
		store := hosts.NewStore(c.hostsFile, false)

		if err := store.Restore(file); err != nil {
			return fmt.Errorf("failed to restore from backup: %w", err)
		}

		fmt.Printf("Restored hosts file from: %s\n", file)
		return nil
	})
}

func (c *CLI) runImport(file, format string) error {
	// Validate file path to prevent directory traversal
	if err := pkg.ValidateSecurePath(file); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	data, err := os.ReadFile(file) // #nosec G304 -- path validated above
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	var profile hosts.Profile
	switch format {
	case "json":
		err = json.Unmarshal(data, &profile)
	case "yaml":
		err = yaml.Unmarshal(data, &profile)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}

	return lock.WithQuickLock(c.hostsFile, func() error {
		store := hosts.NewStore(c.hostsFile, false)

		hostsFile, err := store.Load()
		if err != nil {
			return fmt.Errorf("failed to load hosts file: %w", err)
		}

		for _, entry := range profile.Entries {
			hostsFile.AddEntry(entry)
		}

		if err := store.Save(hostsFile); err != nil {
			return fmt.Errorf("failed to save hosts file: %w", err)
		}

		fmt.Printf("Imported %d entries from %s\n", len(profile.Entries), file)
		return nil
	})
}

func (c *CLI) runExport(file, format string) error {
	store := hosts.NewStore(c.hostsFile, false)

	hostsFile, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load hosts file: %w", err)
	}

	profile := hosts.Profile{
		Name:        "exported",
		Description: "Exported hosts entries",
		Entries:     hostsFile.Entries,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	var data []byte
	switch format {
	case "json":
		data, err = json.MarshalIndent(profile, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(profile)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}

	if err := os.WriteFile(file, data, 0600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	fmt.Printf("Exported %d entries to %s\n", len(profile.Entries), file)
	return nil
}

func (c *CLI) runVerify() error {
	store := hosts.NewStore(c.hostsFile, true)

	issues, err := store.Verify()
	if err != nil {
		return fmt.Errorf("failed to verify hosts file: %w", err)
	}

	if c.jsonOutput {
		result := map[string]interface{}{
			"valid":  len(issues) == 0,
			"issues": issues,
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	if len(issues) == 0 {
		fmt.Println("✓ Hosts file is valid")
		return nil
	}

	fmt.Printf("✗ Found %d issue(s):\n\n", len(issues))
	for i, issue := range issues {
		fmt.Printf("%d. %s\n", i+1, issue)
	}

	return fmt.Errorf("hosts file has validation issues")
}

func (c *CLI) printEntriesFiltered(entries []hosts.Entry, filters ListFilters) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tSTATUS\tIP\tHOSTNAMES\tCOMMENT")
	_, _ = fmt.Fprintln(w, "--\t------\t--\t---------\t-------")

	for _, entry := range entries {
		status := "enabled"
		if entry.Disabled {
			status = "disabled"
		}

		hostnames := strings.Join(entry.Names, ", ")
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", entry.ID, status, entry.IP, hostnames, entry.Comment)
	}

	_ = w.Flush()
}

// applyListFilters applies all specified filters to the entries list.
func (c *CLI) applyListFilters(entries []hosts.Entry, filters ListFilters) []hosts.Entry {
	var filtered []hosts.Entry

	for _, entry := range entries {
		// Filter by disabled status
		if !filters.ShowAll && entry.Disabled {
			continue
		}

		// Filter by status
		if filters.StatusFilter != "" {
			wantEnabled := strings.ToLower(filters.StatusFilter) == "enabled"
			if entry.Disabled == wantEnabled {
				continue
			}
		}

		// Filter by IP
		if filters.IPFilter != "" && !c.matchesPattern(entry.IP, filters.IPFilter) {
			continue
		}

		// Filter by comment
		if filters.CommentFilter != "" && !c.matchesPattern(entry.Comment, filters.CommentFilter) {
			continue
		}

		// Filter by hostname
		if filters.NameFilter != "" {
			nameMatches := false
			for _, name := range entry.Names {
				if c.matchesPattern(name, filters.NameFilter) {
					nameMatches = true
					break
				}
			}
			if !nameMatches {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

// matchesPattern checks if text matches a pattern (supports wildcards).
func (c *CLI) matchesPattern(text, pattern string) bool {
	if pattern == "" {
		return true
	}

	// Simple wildcard matching
	if strings.Contains(pattern, "*") {
		return c.matchWildcard(text, pattern)
	}

	// Simple substring match
	return strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
}

// matchWildcard performs wildcard pattern matching.
func (c *CLI) matchWildcard(text, pattern string) bool {
	// Convert pattern to regex
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	regexPattern = "^" + regexPattern + "$"

	regex, err := regexp.Compile("(?i)" + regexPattern)
	if err != nil {
		// Fall back to simple substring match if regex fails
		return strings.Contains(strings.ToLower(text), strings.ToLower(strings.ReplaceAll(pattern, "*", "")))
	}

	return regex.MatchString(text)
}
