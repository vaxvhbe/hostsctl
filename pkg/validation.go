// Package pkg provides validation utilities for IP addresses, hostnames, and comments.
// It ensures that hosts file entries conform to RFC standards and best practices.
package pkg

import (
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// hostnameRegex validates hostname format according to RFC standards
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	// ipv4Regex validates IPv4 address format with proper octet ranges
	ipv4Regex = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
)

// ValidationError represents a validation failure with context information.
type ValidationError struct {
	Field   string // The field that failed validation
	Value   string // The invalid value
	Message string // Human-readable error message
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	return e.Message
}

// ValidateIP validates an IP address (IPv4 or IPv6) using Go's net package.
// Returns a ValidationError if the IP address format is invalid.
func ValidateIP(ip string) *ValidationError {
	if ip == "" {
		return &ValidationError{
			Field:   "ip",
			Value:   ip,
			Message: "IP address cannot be empty",
		}
	}

	if net.ParseIP(ip) == nil {
		return &ValidationError{
			Field:   "ip",
			Value:   ip,
			Message: "invalid IP address format",
		}
	}

	return nil
}

// ValidateIPv4 specifically validates IPv4 address format.
// Uses regex to ensure proper octet ranges (0-255).
func ValidateIPv4(ip string) *ValidationError {
	if ip == "" {
		return &ValidationError{
			Field:   "ip",
			Value:   ip,
			Message: "IPv4 address cannot be empty",
		}
	}

	if !ipv4Regex.MatchString(ip) {
		return &ValidationError{
			Field:   "ip",
			Value:   ip,
			Message: "invalid IPv4 address format",
		}
	}

	return nil
}

// ValidateIPv6 specifically validates IPv6 address format.
// Ensures the address is valid IPv6 and not IPv4-mapped.
func ValidateIPv6(ip string) *ValidationError {
	if ip == "" {
		return &ValidationError{
			Field:   "ip",
			Value:   ip,
			Message: "IPv6 address cannot be empty",
		}
	}

	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.To4() != nil {
		return &ValidationError{
			Field:   "ip",
			Value:   ip,
			Message: "invalid IPv6 address format",
		}
	}

	return nil
}

// ValidateHostname validates a hostname according to RFC 1123 standards.
// Checks length limits, character restrictions, and label format rules.
func ValidateHostname(hostname string) *ValidationError {
	if hostname == "" {
		return &ValidationError{
			Field:   "hostname",
			Value:   hostname,
			Message: "hostname cannot be empty",
		}
	}

	if len(hostname) > 253 {
		return &ValidationError{
			Field:   "hostname",
			Value:   hostname,
			Message: "hostname too long (max 253 characters)",
		}
	}

	if hostname == "localhost" {
		return nil
	}

	if strings.HasPrefix(hostname, ".") || strings.HasSuffix(hostname, ".") {
		return &ValidationError{
			Field:   "hostname",
			Value:   hostname,
			Message: "hostname cannot start or end with a dot",
		}
	}

	if strings.Contains(hostname, "..") {
		return &ValidationError{
			Field:   "hostname",
			Value:   hostname,
			Message: "hostname cannot contain consecutive dots",
		}
	}

	if !hostnameRegex.MatchString(hostname) {
		return &ValidationError{
			Field:   "hostname",
			Value:   hostname,
			Message: "invalid hostname format",
		}
	}

	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) > 63 {
			return &ValidationError{
				Field:   "hostname",
				Value:   hostname,
				Message: "hostname label too long (max 63 characters per label)",
			}
		}

		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return &ValidationError{
				Field:   "hostname",
				Value:   hostname,
				Message: "hostname label cannot start or end with hyphen",
			}
		}
	}

	return nil
}

// ValidateHostnames validates a list of hostnames and checks for duplicates.
// Returns a slice of ValidationErrors for any invalid or duplicate hostnames.
func ValidateHostnames(hostnames []string) []*ValidationError {
	var errors []*ValidationError

	if len(hostnames) == 0 {
		errors = append(errors, &ValidationError{
			Field:   "hostnames",
			Value:   "",
			Message: "at least one hostname is required",
		})
		return errors
	}

	seen := make(map[string]bool)
	for i, hostname := range hostnames {
		if seen[hostname] {
			errors = append(errors, &ValidationError{
				Field:   "hostnames",
				Value:   hostname,
				Message: "duplicate hostname in list",
			})
			continue
		}
		seen[hostname] = true

		if err := ValidateHostname(hostname); err != nil {
			err.Field = "hostnames[" + string(rune(i)) + "]"
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidateComment validates a comment string for hosts file entries.
// Ensures the comment doesn't contain newlines and isn't too long.
func ValidateComment(comment string) *ValidationError {
	if len(comment) > 255 {
		return &ValidationError{
			Field:   "comment",
			Value:   comment,
			Message: "comment too long (max 255 characters)",
		}
	}

	if strings.Contains(comment, "\n") || strings.Contains(comment, "\r") {
		return &ValidationError{
			Field:   "comment",
			Value:   comment,
			Message: "comment cannot contain newlines",
		}
	}

	return nil
}

// IsValidIP is a convenience function that returns true if the IP is valid.
func IsValidIP(ip string) bool {
	return ValidateIP(ip) == nil
}

// IsValidIPv4 is a convenience function that returns true if the IPv4 is valid.
func IsValidIPv4(ip string) bool {
	return ValidateIPv4(ip) == nil
}

// IsValidIPv6 is a convenience function that returns true if the IPv6 is valid.
func IsValidIPv6(ip string) bool {
	return ValidateIPv6(ip) == nil
}

// IsValidHostname is a convenience function that returns true if the hostname is valid.
func IsValidHostname(hostname string) bool {
	return ValidateHostname(hostname) == nil
}

// IsValidComment is a convenience function that returns true if the comment is valid.
func IsValidComment(comment string) bool {
	return ValidateComment(comment) == nil
}

// NormalizeIP normalizes an IP address to its canonical string representation.
// Uses Go's net.ParseIP for consistent formatting.
func NormalizeIP(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed != nil {
		return parsed.String()
	}
	return ip
}

// NormalizeHostname normalizes a hostname to lowercase and trims whitespace.
// This ensures consistent comparison and storage.
func NormalizeHostname(hostname string) string {
	return strings.ToLower(strings.TrimSpace(hostname))
}

// ValidateSecurePath validates a file path to prevent directory traversal attacks (CWE-22).
// It checks for path traversal patterns and ensures the path is safe to use.
func ValidateSecurePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for null bytes which can be used to bypass security checks
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte: %s", path)
	}

	// Check for path traversal attempts before cleaning
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains directory traversal pattern: %s", path)
	}

	// Additional checks for encoded traversal attempts
	if strings.Contains(path, "%2e%2e") || strings.Contains(path, "%2E%2E") {
		return fmt.Errorf("path contains encoded traversal pattern: %s", path)
	}

	// Clean the path to resolve . elements and normalize
	cleanPath := filepath.Clean(path)

	// After cleaning, ensure no .. elements remain (shouldn't happen if we caught them above)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path resolves to traversal pattern: %s", path)
	}

	return nil
}
