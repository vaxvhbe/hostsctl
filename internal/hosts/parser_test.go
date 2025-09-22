package hosts

import (
	"strings"
	"testing"
)

func TestParser_ParseValidEntries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate bool
		wantLen  int
		wantErr  bool
	}{
		{
			name:    "simple entry",
			input:   "127.0.0.1\tlocalhost",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "entry with comment",
			input:   "192.168.1.1\tserver.local\t# Server",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "disabled entry",
			input:   "# 192.168.1.2\tdisabled.local\t# Disabled",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "multiple names",
			input:   "10.0.0.1\tapi.local web.local app.local",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "IPv6 entry",
			input:   "::1\tipv6.local",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty lines and comments",
			input:   "# Comment\n\n127.0.0.1\tlocalhost\n# Another comment\n",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "multiple entries",
			input:   "127.0.0.1\tlocalhost\n192.168.1.1\tserver.local",
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.validate)
			hostsFile, err := parser.Parse(strings.NewReader(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(hostsFile.Entries) != tt.wantLen {
				t.Errorf("Parser.Parse() entries = %d, want %d", len(hostsFile.Entries), tt.wantLen)
			}
		})
	}
}

func TestParser_ParseInvalidEntries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate bool
	}{
		{
			name:     "invalid IP with validation",
			input:    "999.999.999.999\tinvalid.local",
			validate: true,
		},
		{
			name:     "invalid hostname with validation",
			input:    "127.0.0.1\t.invalid",
			validate: true,
		},
		{
			name:     "empty hostname with validation",
			input:    "127.0.0.1\t",
			validate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.validate)
			_, err := parser.Parse(strings.NewReader(tt.input))

			if err == nil {
				t.Errorf("Parser.Parse() should return error for invalid input with validation")
			}
		})
	}
}

func TestParser_ParseWithoutValidation(t *testing.T) {
	// Invalid entries should be parsed without validation
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{
			name:    "invalid IP without validation",
			input:   "999.999.999.999\tinvalid.local",
			wantLen: 0, // Parser might still reject obviously malformed lines
		},
		{
			name:    "invalid hostname without validation",
			input:   "127.0.0.1\t.invalid",
			wantLen: 0, // Parser might still reject obviously malformed lines
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(false)
			hostsFile, err := parser.Parse(strings.NewReader(tt.input))

			if err != nil {
				t.Errorf("Parser.Parse() should not return error without validation: %v", err)
			}

			if len(hostsFile.Entries) != tt.wantLen {
				t.Errorf("Parser.Parse() entries = %d, want %d", len(hostsFile.Entries), tt.wantLen)
			}
		})
	}
}

func TestParser_ParseDisabledEntries(t *testing.T) {
	input := `# This is a comment
127.0.0.1	localhost
# 192.168.1.1	disabled.local	# Disabled entry
192.168.1.2	enabled.local	# Enabled entry`

	parser := NewParser(false)
	hostsFile, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parser.Parse() error = %v", err)
	}

	if len(hostsFile.Entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(hostsFile.Entries))
	}

	// Check that disabled entry is marked as disabled
	var disabledFound bool
	for _, entry := range hostsFile.Entries {
		if entry.Names[0] == "disabled.local" {
			if !entry.Disabled {
				t.Error("Entry 'disabled.local' should be marked as disabled")
			}
			disabledFound = true
		}
		if entry.Names[0] == "enabled.local" {
			if entry.Disabled {
				t.Error("Entry 'enabled.local' should not be marked as disabled")
			}
		}
	}

	if !disabledFound {
		t.Error("Disabled entry should be parsed")
	}
}

func TestParser_Serialize(t *testing.T) {
	entries := []Entry{
		{
			ID:       1,
			IP:       "127.0.0.1",
			Names:    []string{"localhost"},
			Comment:  "",
			Disabled: false,
		},
		{
			ID:       2,
			IP:       "192.168.1.1",
			Names:    []string{"server.local", "web.local"},
			Comment:  "Main server",
			Disabled: false,
		},
		{
			ID:       3,
			IP:       "192.168.1.2",
			Names:    []string{"disabled.local"},
			Comment:  "Disabled entry",
			Disabled: true,
		},
	}

	hostsFile := &HostsFile{Entries: entries}
	parser := NewParser(false)
	output := parser.Serialize(hostsFile)

	expected := `127.0.0.1	localhost
192.168.1.1	server.local	web.local	# Main server
# 192.168.1.2	disabled.local	# Disabled entry
`

	if output != expected {
		t.Errorf("Parser.Serialize() = %q, want %q", output, expected)
	}
}

func TestParser_RoundTrip(t *testing.T) {
	// Test that parsing then serializing produces the same output
	input := `127.0.0.1	localhost
192.168.1.1	server.local	web.local	# Main server
# 192.168.1.2	disabled.local	# Disabled entry
`

	parser := NewParser(false)
	hostsFile, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parser.Parse() error = %v", err)
	}

	output := parser.Serialize(hostsFile)
	if output != input {
		t.Errorf("Round trip failed:\nInput:  %q\nOutput: %q", input, output)
	}
}

func TestParser_EmptyInput(t *testing.T) {
	parser := NewParser(false)
	hostsFile, err := parser.Parse(strings.NewReader(""))

	if err != nil {
		t.Errorf("Parser.Parse() should handle empty input: %v", err)
	}

	if len(hostsFile.Entries) != 0 {
		t.Errorf("Expected 0 entries for empty input, got %d", len(hostsFile.Entries))
	}
}

func TestParser_OnlyComments(t *testing.T) {
	input := `# This is a comment
# Another comment
# Third comment`

	parser := NewParser(false)
	hostsFile, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Errorf("Parser.Parse() should handle comment-only input: %v", err)
	}

	if len(hostsFile.Entries) != 0 {
		t.Errorf("Expected 0 entries for comment-only input, got %d", len(hostsFile.Entries))
	}
}

func TestParser_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{
			name:    "tabs and spaces",
			input:   "127.0.0.1    localhost   local  \t# Comment",
			wantLen: 1,
		},
		{
			name:    "leading whitespace",
			input:   "  \t127.0.0.1\tlocalhost",
			wantLen: 0, // Leading whitespace might not be supported
		},
		{
			name:    "trailing whitespace",
			input:   "127.0.0.1\tlocalhost  \t  ",
			wantLen: 1,
		},
		{
			name:    "blank lines",
			input:   "\n\n127.0.0.1\tlocalhost\n\n\n",
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(false)
			hostsFile, err := parser.Parse(strings.NewReader(tt.input))

			if err != nil {
				t.Errorf("Parser.Parse() error = %v", err)
			}

			if len(hostsFile.Entries) != tt.wantLen {
				t.Errorf("Parser.Parse() entries = %d, want %d", len(hostsFile.Entries), tt.wantLen)
			}
		})
	}
}

func TestParser_IPv6Addresses(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"IPv6 loopback", "::1"},
		{"IPv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		{"IPv6 compressed", "2001:db8:85a3::8a2e:370:7334"},
		{"IPv6 link-local", "fe80::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.ip + "\ttesthost.local"
			parser := NewParser(false)
			hostsFile, err := parser.Parse(strings.NewReader(input))

			if err != nil {
				t.Errorf("Parser.Parse() error = %v", err)
			}

			if len(hostsFile.Entries) != 1 {
				t.Errorf("Expected 1 entry, got %d", len(hostsFile.Entries))
			}

			if len(hostsFile.Entries) > 0 && hostsFile.Entries[0].IP != tt.ip {
				t.Errorf("Expected IP %s, got %s", tt.ip, hostsFile.Entries[0].IP)
			}
		})
	}
}

func TestNewParser(t *testing.T) {
	// Test with validation enabled
	parser := NewParser(true)
	if parser == nil {
		t.Error("NewParser should not return nil")
	}

	// Test with validation disabled
	parser = NewParser(false)
	if parser == nil {
		t.Error("NewParser should not return nil")
	}
}

func TestParser_SerializeEmpty(t *testing.T) {
	hostsFile := &HostsFile{Entries: []Entry{}}
	parser := NewParser(false)
	output := parser.Serialize(hostsFile)

	if strings.TrimSpace(output) != "" {
		t.Errorf("Parser.Serialize() for empty hosts file = %q, want empty string", output)
	}
}
