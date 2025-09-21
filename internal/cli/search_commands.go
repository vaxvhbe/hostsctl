package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vaxvhbe/hostsctl/internal/hosts"
)

// SearchOptions contains options for searching hosts entries.
type SearchOptions struct {
	Pattern         string // Search pattern
	UseRegex        bool   // Whether to treat pattern as regex
	IgnoreCase      bool   // Case-insensitive search
	SearchIP        bool   // Search in IP addresses
	SearchNames     bool   // Search in hostnames
	SearchComments  bool   // Search in comments
	IncludeDisabled bool   // Include disabled entries in results
}

// SearchResult represents a search result with match information.
type SearchResult struct {
	Entry     hosts.Entry `json:"entry"`
	MatchType string      `json:"match_type"` // "ip", "hostname", "comment"
	MatchText string      `json:"match_text"` // The actual text that matched
}

// buildSearchCommand creates the search command.
func (c *CLI) buildSearchCommand() *cobra.Command {
	var options SearchOptions

	cmd := &cobra.Command{
		Use:   "search [pattern]",
		Short: "Search hosts entries",
		Long: `Search for hosts entries matching the given pattern.

The search can use regular expressions or glob patterns, and can search across
IP addresses, hostnames, and comments.

Examples:
  hostsctl search "local"          # Simple text search
  hostsctl search "*.dev" --glob   # Glob pattern
  hostsctl search "^192\.168"      # Regex pattern
  hostsctl search "test" --ip      # Search only IP addresses
  hostsctl search "app" --comment  # Search only comments`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.Pattern = args[0]
			return c.runSearch(options)
		},
	}

	cmd.Flags().BoolVar(&options.UseRegex, "regex", false, "Treat pattern as regular expression")
	cmd.Flags().BoolVar(&options.IgnoreCase, "ignore-case", false, "Case-insensitive search")
	cmd.Flags().BoolVar(&options.SearchIP, "ip", false, "Search only IP addresses")
	cmd.Flags().BoolVar(&options.SearchNames, "names", false, "Search only hostnames")
	cmd.Flags().BoolVar(&options.SearchComments, "comments", false, "Search only comments")
	cmd.Flags().BoolVar(&options.IncludeDisabled, "include-disabled", false, "Include disabled entries")

	// Add aliases for common flags
	cmd.Flags().BoolP("case-insensitive", "i", false, "Case-insensitive search (alias for --ignore-case)")
	cmd.Flags().BoolP("glob", "g", false, "Treat pattern as glob pattern")

	return cmd
}

// runSearch executes the search command.
func (c *CLI) runSearch(options SearchOptions) error {

	// Default to searching all fields if none specified
	if !options.SearchIP && !options.SearchNames && !options.SearchComments {
		options.SearchIP = true
		options.SearchNames = true
		options.SearchComments = true
	}

	store := hosts.NewStore(c.hostsFile, false)
	hostsFile, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load hosts file: %w", err)
	}

	results, err := c.searchEntries(hostsFile.Entries, options)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if c.jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(results)
	}

	c.printSearchResults(results, options)
	return nil
}

// searchEntries searches through hosts entries based on the given options.
func (c *CLI) searchEntries(entries []hosts.Entry, options SearchOptions) ([]SearchResult, error) {
	var results []SearchResult
	var matcher func(string) bool
	var err error

	// Prepare the matcher function
	if options.UseRegex {
		var re *regexp.Regexp
		pattern := options.Pattern

		if options.IgnoreCase {
			pattern = "(?i)" + pattern
		}

		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regular expression: %w", err)
		}

		matcher = re.MatchString
	} else {
		// Simple substring search
		pattern := options.Pattern
		if options.IgnoreCase {
			pattern = strings.ToLower(pattern)
		}

		matcher = func(text string) bool {
			if options.IgnoreCase {
				text = strings.ToLower(text)
			}
			return strings.Contains(text, pattern)
		}
	}

	// Search through entries
	for _, entry := range entries {
		if !options.IncludeDisabled && entry.Disabled {
			continue
		}

		// Search IP address
		if options.SearchIP && matcher(entry.IP) {
			results = append(results, SearchResult{
				Entry:     entry,
				MatchType: "ip",
				MatchText: entry.IP,
			})
			continue
		}

		// Search hostnames
		if options.SearchNames {
			for _, name := range entry.Names {
				if matcher(name) {
					results = append(results, SearchResult{
						Entry:     entry,
						MatchType: "hostname",
						MatchText: name,
					})
					break // Don't add the same entry multiple times
				}
			}
		}

		// Search comments (only if not already matched)
		if options.SearchComments && entry.Comment != "" {
			// Check if this entry was already added
			alreadyAdded := false
			for _, result := range results {
				if result.Entry.ID == entry.ID {
					alreadyAdded = true
					break
				}
			}

			if !alreadyAdded && matcher(entry.Comment) {
				results = append(results, SearchResult{
					Entry:     entry,
					MatchType: "comment",
					MatchText: entry.Comment,
				})
			}
		}
	}

	return results, nil
}

// printSearchResults prints search results in a human-readable format.
func (c *CLI) printSearchResults(results []SearchResult, options SearchOptions) {
	if len(results) == 0 {
		fmt.Printf("No entries found matching pattern: %s\n", options.Pattern)
		return
	}

	fmt.Printf("Found %d entries matching pattern: %s\n\n", len(results), options.Pattern)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tSTATUS\tIP\tHOSTNAMES\tCOMMENT\tMATCH")
	_, _ = fmt.Fprintln(w, "--\t------\t--\t---------\t-------\t-----")

	for _, result := range results {
		status := "enabled"
		if result.Entry.Disabled {
			status = "disabled"
		}

		hostnames := strings.Join(result.Entry.Names, ", ")
		match := fmt.Sprintf("%s: %s", result.MatchType, result.MatchText)

		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			result.Entry.ID, status, result.Entry.IP, hostnames, result.Entry.Comment, match)
	}

	_ = w.Flush()
}
