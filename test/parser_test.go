package test

import (
	"strings"
	"testing"

	"github.com/vaxvhbe/hostsctl/internal/hosts"
)

func TestParseValidHosts(t *testing.T) {
	input := `# This is a comment
127.0.0.1	localhost
192.168.1.100	server.local web.local	# Main server
# 192.168.1.200	disabled.local	# Disabled entry
::1	ipv6.local`

	parser := hosts.NewParser(false)
	hostsFile, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(hostsFile.Entries) != 4 {
		t.Fatalf("Expected 4 entries, got %d", len(hostsFile.Entries))
	}

	entry1 := hostsFile.Entries[0]
	if entry1.IP != "127.0.0.1" {
		t.Errorf("Expected IP '127.0.0.1', got '%s'", entry1.IP)
	}
	if len(entry1.Names) != 1 || entry1.Names[0] != "localhost" {
		t.Errorf("Expected names ['localhost'], got %v", entry1.Names)
	}

	entry2 := hostsFile.Entries[1]
	if entry2.IP != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got '%s'", entry2.IP)
	}
	if len(entry2.Names) != 2 || entry2.Names[0] != "server.local" || entry2.Names[1] != "web.local" {
		t.Errorf("Expected names ['server.local', 'web.local'], got %v", entry2.Names)
	}
	if entry2.Comment != "Main server" {
		t.Errorf("Expected comment 'Main server', got '%s'", entry2.Comment)
	}

	entry3 := hostsFile.Entries[2]
	if !entry3.Disabled {
		t.Errorf("Expected entry3 to be disabled, but it's enabled")
	}
	if entry3.IP != "192.168.1.200" {
		t.Errorf("Expected IP '192.168.1.200', got '%s'", entry3.IP)
	}
	if len(entry3.Names) != 1 || entry3.Names[0] != "disabled.local" {
		t.Errorf("Expected names ['disabled.local'], got %v", entry3.Names)
	}

	entry4 := hostsFile.Entries[3]
	if entry4.Disabled {
		t.Errorf("Expected entry4 to be enabled, but it's disabled")
	}
	if entry4.IP != "::1" {
		t.Errorf("Expected IP '::1', got '%s'", entry4.IP)
	}
}

func TestParseDisabledEntry(t *testing.T) {
	input := `# 192.168.1.200	disabled.local	# Disabled entry`

	parser := hosts.NewParser(false)
	hostsFile, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(hostsFile.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(hostsFile.Entries))
	}

	entry := hostsFile.Entries[0]
	if !entry.Disabled {
		t.Errorf("Expected entry to be disabled")
	}
	if entry.IP != "192.168.1.200" {
		t.Errorf("Expected IP '192.168.1.200', got '%s'", entry.IP)
	}
	if len(entry.Names) != 1 || entry.Names[0] != "disabled.local" {
		t.Errorf("Expected names ['disabled.local'], got %v", entry.Names)
	}
	if entry.Comment != "Disabled entry" {
		t.Errorf("Expected comment 'Disabled entry', got '%s'", entry.Comment)
	}
}

func TestParseInvalidIP(t *testing.T) {
	input := `999.999.999.999	invalid.local`

	parser := hosts.NewParser(true)
	_, err := parser.Parse(strings.NewReader(input))

	if err == nil {
		t.Fatal("Expected error for invalid IP, got none")
	}
}

func TestSerialize(t *testing.T) {
	hostsFile := &hosts.HostsFile{
		Entries: []hosts.Entry{
			{
				ID:       1,
				IP:       "127.0.0.1",
				Names:    []string{"localhost"},
				Comment:  "",
				Disabled: false,
			},
			{
				ID:       2,
				IP:       "192.168.1.100",
				Names:    []string{"server.local", "web.local"},
				Comment:  "Main server",
				Disabled: false,
			},
			{
				ID:       3,
				IP:       "192.168.1.200",
				Names:    []string{"disabled.local"},
				Comment:  "Disabled entry",
				Disabled: true,
			},
		},
	}

	parser := hosts.NewParser(false)
	output := parser.Serialize(hostsFile)

	expected := `127.0.0.1	localhost
192.168.1.100	server.local	web.local	# Main server
# 192.168.1.200	disabled.local	# Disabled entry
`

	if output != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, output)
	}
}
