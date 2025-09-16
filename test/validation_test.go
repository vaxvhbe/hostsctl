package test

import (
	"strings"
	"testing"

	"github.com/vaxvhbe/hostsctl/pkg"
)

func TestValidateIPv4(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"127.0.0.1", true},
		{"192.168.1.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"", false},
		{"256.1.1.1", false},
		{"127.0.0", false},
		{"127.0.0.1.1", false},
		{"invalid", false},
	}

	for _, test := range tests {
		err := pkg.ValidateIPv4(test.ip)
		if test.valid && err != nil {
			t.Errorf("Expected IP '%s' to be valid, got error: %v", test.ip, err)
		}
		if !test.valid && err == nil {
			t.Errorf("Expected IP '%s' to be invalid, got no error", test.ip)
		}
	}
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		hostname string
		valid    bool
	}{
		{"localhost", true},
		{"example.com", true},
		{"sub.example.com", true},
		{"test-server.local", true},
		{"", false},
		{".invalid", false},
		{"invalid.", false},
		{"invalid..com", false},
		{"toolongtobevalid" + strings.Repeat("a", 250), false},
	}

	for _, test := range tests {
		err := pkg.ValidateHostname(test.hostname)
		if test.valid && err != nil {
			t.Errorf("Expected hostname '%s' to be valid, got error: %v", test.hostname, err)
		}
		if !test.valid && err == nil {
			t.Errorf("Expected hostname '%s' to be invalid, got no error", test.hostname)
		}
	}
}

func TestValidateHostnames(t *testing.T) {
	tests := []struct {
		hostnames []string
		valid     bool
	}{
		{[]string{"localhost"}, true},
		{[]string{"example.com", "test.local"}, true},
		{[]string{}, false},
		{[]string{"example.com", "example.com"}, false},
		{[]string{"valid.com", ".invalid"}, false},
	}

	for _, test := range tests {
		errs := pkg.ValidateHostnames(test.hostnames)
		if test.valid && len(errs) > 0 {
			t.Errorf("Expected hostnames %v to be valid, got errors: %v", test.hostnames, errs)
		}
		if !test.valid && len(errs) == 0 {
			t.Errorf("Expected hostnames %v to be invalid, got no errors", test.hostnames)
		}
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
	}

	for _, test := range tests {
		result := pkg.NormalizeIP(test.input)
		if result != test.expected {
			t.Errorf("Expected normalized IP '%s', got '%s'", test.expected, result)
		}
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
	}

	for _, test := range tests {
		result := pkg.NormalizeHostname(test.input)
		if result != test.expected {
			t.Errorf("Expected normalized hostname '%s', got '%s'", test.expected, result)
		}
	}
}
