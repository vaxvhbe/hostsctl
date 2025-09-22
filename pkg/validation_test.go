package pkg

import (
	"strings"
	"testing"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		ip      string
		wantErr bool
	}{
		{"127.0.0.1", false},
		{"192.168.1.1", false},
		{"0.0.0.0", false},
		{"255.255.255.255", false},
		{"::1", false},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"2001:db8:85a3::8a2e:370:7334", false},
		{"", true},
		{"256.1.1.1", true},
		{"127.0.0", true},
		{"127.0.0.1.1", true},
		{"invalid", true},
		{"192.168.1.", true},
		{"192.168.1.256", true},
		{"gggg::1", true},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			err := ValidateIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIP(%q) error = %v, wantErr %v", tt.ip, err, tt.wantErr)
			}
			if err != nil {
				if err.Field != "ip" {
					t.Errorf("ValidateIP(%q) Field = %v, want 'ip'", tt.ip, err.Field)
				}
				if err.Value != tt.ip {
					t.Errorf("ValidateIP(%q) Value = %v, want %v", tt.ip, err.Value, tt.ip)
				}
			}
		})
	}
}

// ValidateSecurePath tests
func TestValidateSecurePath(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
		desc    string
	}{
		{"", true, "empty path"},
		{"valid/path.txt", false, "valid relative path"},
		{"/valid/absolute/path.txt", false, "valid absolute path"},
		{"../../../etc/passwd", true, "directory traversal"},
		{"path/../../../sensitive", true, "path with traversal"},
		{"path/with/../elements", true, "path with .. elements"},
		{"path/file\x00.txt", true, "null byte injection"},
		{"normal/file.json", false, "normal file path"},
		{"/home/user/file.txt", false, "absolute user path"},
		{"./file.txt", false, "current directory path"},
	}

	for _, test := range tests {
		err := ValidateSecurePath(test.path)
		hasErr := err != nil
		if hasErr != test.wantErr {
			t.Errorf("ValidateSecurePath(%q) error = %v, wantErr %v (%s)",
				test.path, err, test.wantErr, test.desc)
		}
	}
}

func TestValidateIPv4(t *testing.T) {
	tests := []struct {
		ip      string
		wantErr bool
	}{
		{"127.0.0.1", false},
		{"192.168.1.1", false},
		{"0.0.0.0", false},
		{"255.255.255.255", false},
		{"", true},
		{"256.1.1.1", true},
		{"127.0.0", true},
		{"127.0.0.1.1", true},
		{"invalid", true},
		{"::1", true},
		{"192.168.1.", true},
		{"192.168.1.256", true},
		{"-1.0.0.1", true},
		{"192.168.001.1", false}, // Leading zeros are allowed by regex
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			err := ValidateIPv4(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPv4(%q) error = %v, wantErr %v", tt.ip, err, tt.wantErr)
			}
		})
	}
}

func TestValidateIPv6(t *testing.T) {
	tests := []struct {
		ip      string
		wantErr bool
	}{
		{"::1", false},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"2001:db8:85a3::8a2e:370:7334", false},
		{"fe80::1", false},
		{"", true},
		{"127.0.0.1", true}, // IPv4 should fail
		{"invalid", true},
		{"gggg::1", true},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334:extra", true},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			err := ValidateIPv6(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPv6(%q) error = %v, wantErr %v", tt.ip, err, tt.wantErr)
			}
		})
	}
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		hostname string
		wantErr  bool
	}{
		{"localhost", false},
		{"example.com", false},
		{"sub.example.com", false},
		{"test-server.local", false},
		{"api.v1.example.com", false},
		{"x.com", false},
		{"", true},
		{".invalid", true},
		{"invalid.", true},
		{"invalid..com", true},
		{"toolong" + strings.Repeat("a", 250), true},
		{strings.Repeat("a", 64) + ".com", true}, // Label too long
		{"-invalid.com", true},                   // Label starts with hyphen
		{"invalid-.com", true},                   // Label ends with hyphen
		{"123.456.789.012", false},               // Numeric labels are allowed
		{"test_underscore.com", true},            // Underscores not allowed
		{"test space.com", true},                 // Spaces not allowed
		{"test.com.", true},                      // Trailing dot not allowed
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			err := ValidateHostname(tt.hostname)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHostname(%q) error = %v, wantErr %v", tt.hostname, err, tt.wantErr)
			}
		})
	}
}

func TestValidateHostnames(t *testing.T) {
	tests := []struct {
		name      string
		hostnames []string
		wantErrs  int
	}{
		{"valid single", []string{"localhost"}, 0},
		{"valid multiple", []string{"example.com", "test.local"}, 0},
		{"empty list", []string{}, 1},
		{"duplicates", []string{"example.com", "example.com"}, 1},
		{"mixed valid/invalid", []string{"valid.com", ".invalid"}, 1},
		{"multiple duplicates", []string{"a.com", "b.com", "a.com", "c.com", "b.com"}, 2},
		{"all invalid", []string{".invalid1", ".invalid2"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateHostnames(tt.hostnames)
			if len(errs) != tt.wantErrs {
				t.Errorf("ValidateHostnames(%v) errors = %d, want %d", tt.hostnames, len(errs), tt.wantErrs)
				for i, err := range errs {
					t.Errorf("  Error %d: %v", i, err)
				}
			}
		})
	}
}

func TestValidateComment(t *testing.T) {
	tests := []struct {
		comment string
		wantErr bool
	}{
		{"", false},
		{"Valid comment", false},
		{"Comment with special chars !@#$%^&*()", false},
		{strings.Repeat("a", 255), false},
		{strings.Repeat("a", 256), true}, // Too long
		{"Comment with\nnewline", true},
		{"Comment with\rcarriage return", true},
		{"Comment with\r\nCRLF", true},
		{"Comment with tab\there", false}, // Tabs are allowed
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			err := ValidateComment(tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateComment(%q) error = %v, wantErr %v", tt.comment, err, tt.wantErr)
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Test IsValidIP
	if !IsValidIP("127.0.0.1") {
		t.Error("IsValidIP should return true for valid IP")
	}
	if IsValidIP("invalid") {
		t.Error("IsValidIP should return false for invalid IP")
	}

	// Test IsValidIPv4
	if !IsValidIPv4("192.168.1.1") {
		t.Error("IsValidIPv4 should return true for valid IPv4")
	}
	if IsValidIPv4("::1") {
		t.Error("IsValidIPv4 should return false for IPv6")
	}

	// Test IsValidIPv6
	if !IsValidIPv6("::1") {
		t.Error("IsValidIPv6 should return true for valid IPv6")
	}
	if IsValidIPv6("127.0.0.1") {
		t.Error("IsValidIPv6 should return false for IPv4")
	}

	// Test IsValidHostname
	if !IsValidHostname("example.com") {
		t.Error("IsValidHostname should return true for valid hostname")
	}
	if IsValidHostname(".invalid") {
		t.Error("IsValidHostname should return false for invalid hostname")
	}

	// Test IsValidComment
	if !IsValidComment("valid comment") {
		t.Error("IsValidComment should return true for valid comment")
	}
	if IsValidComment("invalid\ncomment") {
		t.Error("IsValidComment should return false for comment with newline")
	}
}

func TestNormalizeIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"127.0.0.1", "127.0.0.1"},
		{"::1", "::1"},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:db8:85a3::8a2e:370:7334"},
		{"invalid", "invalid"}, // Returns original if unparseable
		{"0000:0000:0000:0000:0000:0000:0000:0001", "::1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeIP(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeIP(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Example.COM", "example.com"},
		{"  TEST.local  ", "test.local"},
		{"localhost", "localhost"},
		{"", ""},
		{"  ", ""},
		{"UPPER.CASE.HOST", "upper.case.host"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeHostname(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeHostname(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test_field",
		Value:   "test_value",
		Message: "test message",
	}

	if err.Error() != "test message" {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), "test message")
	}
}

func TestEdgeCases(t *testing.T) {
	// Test very long hostname with valid labels (but respecting DNS label rules)
	longHostname := strings.Repeat("a", 50) + "." + strings.Repeat("b", 50) + "." + strings.Repeat("c", 50) + "." + strings.Repeat("d", 50) // Valid hostname under 253 chars
	if err := ValidateHostname(longHostname); err != nil {
		t.Errorf("ValidateHostname should accept valid long hostname: %v", err)
	}

	// Test hostname exactly at limit (but this is just one label, so it should fail because labels > 63 chars are invalid)
	exactLimit := strings.Repeat("a", 253)
	if err := ValidateHostname(exactLimit); err == nil {
		t.Error("ValidateHostname should reject single label with 253 characters (labels must be <= 63 chars)")
	}

	// Test comment at exact limit
	longComment := strings.Repeat("a", 255)
	if err := ValidateComment(longComment); err != nil {
		t.Errorf("ValidateComment should accept comment with 255 characters: %v", err)
	}

	// Test label at exact limit
	longLabel := strings.Repeat("a", 63) + ".com"
	if err := ValidateHostname(longLabel); err != nil {
		t.Errorf("ValidateHostname should accept label with 63 characters: %v", err)
	}
}

func TestIPv4Ranges(t *testing.T) {
	// Test boundary values for IPv4 octets
	tests := []struct {
		ip      string
		wantErr bool
	}{
		{"0.0.0.0", false},
		{"255.255.255.255", false},
		{"256.0.0.0", true},
		{"0.256.0.0", true},
		{"0.0.256.0", true},
		{"0.0.0.256", true},
		{"192.168.1.1", false},
		{"10.0.0.1", false},
		{"172.16.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			err := ValidateIPv4(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPv4(%q) error = %v, wantErr %v", tt.ip, err, tt.wantErr)
			}
		})
	}
}
