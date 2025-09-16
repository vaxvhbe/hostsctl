package hosts

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	// commentRegex matches lines that are pure comments or empty lines
	commentRegex = regexp.MustCompile(`^(\s*#.*|^\s*)$`)
	// entryRegex matches hosts file entries, capturing disabled prefix, IP, hostnames, and comment
	entryRegex = regexp.MustCompile(`^(\s*#\s*)?(\S+)\s+(.+?)(?:\s*#\s*(.*))?$`)
	// hostnameRegex validates hostname format according to RFC standards
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
)

// ParseError represents an error that occurred during hosts file parsing.
type ParseError struct {
	Line    int    // Line number where the error occurred
	Content string // Content of the problematic line
	Reason  string // Human-readable description of the error
}

// Error implements the error interface for ParseError.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d: %s (content: %q)", e.Line, e.Reason, e.Content)
}

// Parser handles parsing and serialization of hosts files.
type Parser struct {
	strict bool // Whether to fail on parse errors or continue with warnings
}

// NewParser creates a new parser instance.
// If strict is true, parsing will fail on any error.
// If strict is false, invalid lines will be skipped with warnings.
func NewParser(strict bool) *Parser {
	return &Parser{strict: strict}
}

// Parse reads a hosts file from the provided reader and returns a parsed HostsFile.
// It processes each line, extracting entries while handling comments and disabled entries.
func (p *Parser) Parse(reader io.Reader) (*HostsFile, error) {
	hostsFile := &HostsFile{
		Entries: []Entry{},
	}

	scanner := bufio.NewScanner(reader)
	lineNum := 0
	entryID := 1

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || (strings.HasPrefix(trimmed, "#") && !p.isDisabledEntry(line)) {
			continue
		}

		entry, err := p.parseLine(line, lineNum, entryID)
		if err != nil {
			if p.strict {
				return nil, err
			}
			continue
		}

		if entry != nil {
			hostsFile.Entries = append(hostsFile.Entries, *entry)
			entryID++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return hostsFile, nil
}

// isDisabledEntry checks if a line represents a disabled (commented) hosts entry.
// Returns true if the line matches the entry format but is commented out.
func (p *Parser) isDisabledEntry(line string) bool {
	matches := entryRegex.FindStringSubmatch(line)
	return len(matches) >= 4 && strings.TrimSpace(matches[1]) != ""
}

// parseLine parses a single line from a hosts file into an Entry.
// Returns the parsed entry or an error if the line is invalid.
func (p *Parser) parseLine(line string, lineNum, entryID int) (*Entry, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil, nil
	}

	matches := entryRegex.FindStringSubmatch(line)
	if len(matches) < 4 {
		return nil, &ParseError{
			Line:    lineNum,
			Content: line,
			Reason:  "invalid line format",
		}
	}

	disabled := strings.TrimSpace(matches[1]) != ""
	ip := strings.TrimSpace(matches[2])
	hostnames := strings.Fields(strings.TrimSpace(matches[3]))
	comment := ""
	if len(matches) > 4 {
		comment = strings.TrimSpace(matches[4])
	}

	if !p.isValidIP(ip) {
		return nil, &ParseError{
			Line:    lineNum,
			Content: line,
			Reason:  fmt.Sprintf("invalid IP address: %s", ip),
		}
	}

	for _, hostname := range hostnames {
		if !p.isValidHostname(hostname) {
			return nil, &ParseError{
				Line:    lineNum,
				Content: line,
				Reason:  fmt.Sprintf("invalid hostname: %s", hostname),
			}
		}
	}

	return &Entry{
		ID:       entryID,
		IP:       ip,
		Names:    hostnames,
		Comment:  comment,
		Disabled: disabled,
		Raw:      line,
	}, nil
}

// isValidIP validates an IP address (IPv4 or IPv6).
// Returns true if the IP address is properly formatted.
func (p *Parser) isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return p.isValidIPv6(ip)
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}

		num := 0
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
			num = num*10 + int(char-'0')
		}

		if num > 255 {
			return false
		}

		if len(part) > 1 && part[0] == '0' {
			return false
		}
	}

	return true
}

// isValidIPv6 validates an IPv6 address format.
// Supports both full and compressed IPv6 notation.
func (p *Parser) isValidIPv6(ip string) bool {
	if strings.Contains(ip, "::") {
		parts := strings.Split(ip, "::")
		if len(parts) != 2 {
			return false
		}

		leftParts := []string{}
		rightParts := []string{}

		if parts[0] != "" {
			leftParts = strings.Split(parts[0], ":")
		}
		if parts[1] != "" {
			rightParts = strings.Split(parts[1], ":")
		}

		totalParts := len(leftParts) + len(rightParts)
		if totalParts >= 8 {
			return false
		}

		allParts := append(leftParts, rightParts...)
		for _, part := range allParts {
			if !p.isValidIPv6Part(part) {
				return false
			}
		}

		return true
	}

	parts := strings.Split(ip, ":")
	if len(parts) != 8 {
		return false
	}

	for _, part := range parts {
		if !p.isValidIPv6Part(part) {
			return false
		}
	}

	return true
}

// isValidIPv6Part validates a single part (segment) of an IPv6 address.
// Each part should be 1-4 hexadecimal characters.
func (p *Parser) isValidIPv6Part(part string) bool {
	if len(part) == 0 || len(part) > 4 {
		return false
	}

	for _, char := range part {
		if !((char >= '0' && char <= '9') ||
			(char >= 'a' && char <= 'f') ||
			(char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}

// isValidHostname validates a hostname according to RFC standards.
// Returns true if the hostname format is valid.
func (p *Parser) isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	if hostname == "localhost" {
		return true
	}

	return hostnameRegex.MatchString(hostname)
}

// Serialize converts a HostsFile back to its string representation.
// Each entry is converted to a line using the Entry.String() method.
func (p *Parser) Serialize(hostsFile *HostsFile) string {
	var lines []string

	for _, entry := range hostsFile.Entries {
		lines = append(lines, entry.String())
	}

	return strings.Join(lines, "\n") + "\n"
}

// ParseFile is a convenience function to parse a hosts file with default settings.
// It creates a new parser and parses the content from the provided reader.
func ParseFile(reader io.Reader, strict bool) (*HostsFile, error) {
	parser := NewParser(strict)
	return parser.Parse(reader)
}

// SerializeFile is a convenience function to serialize a HostsFile to string.
// It uses a non-strict parser for serialization.
func SerializeFile(hostsFile *HostsFile) string {
	parser := NewParser(false)
	return parser.Serialize(hostsFile)
}
