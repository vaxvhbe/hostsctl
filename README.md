# hostsctl — CLI Manager for /etc/hosts

[![Build Status](https://github.com/vaxvhbe/hostsctl/workflows/CI/badge.svg)](https://github.com/vaxvhbe/hostsctl/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vaxvhbe/hostsctl)](https://goreportcard.com/report/github.com/vaxvhbe/hostsctl)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A modern, safe, and efficient command-line tool for managing `/etc/hosts` files. Built in Go with safety features like atomic writes, automatic backups, and file locking.

## Features

- 🔒 **Safe Operations**: Atomic writes with automatic backups
- 🔐 **File Locking**: Prevents corruption from concurrent access
- ✅ **Validation**: RFC-compliant IP address and hostname validation
- 📦 **Import/Export**: JSON and YAML support for profile management
- 🎯 **Flexible Operations**: Add, remove, enable/disable entries
- 🧪 **Testing-Friendly**: Custom hosts file support for development
- 🐳 **Docker Ready**: Containerized execution support

## Installation

### From Source

```bash
git clone https://github.com/vaxvhbe/hostsctl.git
cd hostsctl
make build
sudo make install
```

### Using Go

```bash
go install github.com/vaxvhbe/hostsctl/cmd/hostsctl@latest
```

### Using Nix

hostsctl is available as a Nix flake for easy installation and reproducible builds:

```bash
# Install from GitHub
nix profile install github:vaxvhbe/hostsctl

# Install from local directory
nix profile install .

# Run without installing
nix run github:vaxvhbe/hostsctl -- --help

# Development environment
nix develop
```

### Docker

```bash
docker build -t hostsctl .
docker run --rm -v /etc/hosts:/etc/hosts hostsctl list
```

## Quick Start

```bash
# List current entries
sudo hostsctl list

# Add a new entry
sudo hostsctl add --ip 127.0.0.1 --name myapp.local --comment "Local development"

# Disable an entry
sudo hostsctl disable --name myapp.local

# Create a backup
sudo hostsctl backup

# Verify hosts file integrity
sudo hostsctl verify
```

## Usage

### Commands

#### `list` - List hosts entries

```bash
# Show active entries only
sudo hostsctl list

# Show all entries including disabled ones
sudo hostsctl list --all

# JSON output for scripting
sudo hostsctl list --json
```

#### `add` - Add new entries

```bash
# Add single hostname
sudo hostsctl add --ip 192.168.1.100 --name server.local

# Add multiple hostnames
sudo hostsctl add --ip 192.168.1.100 --name server.local --name api.local

# Add with comment
sudo hostsctl add --ip 127.0.0.1 --name app.local --comment "Development server"
```

#### `rm` - Remove entries

```bash
# Remove by hostname
sudo hostsctl rm --name server.local

# Remove by ID
sudo hostsctl rm --id 5
```

#### `enable/disable` - Toggle entries

```bash
# Disable entry (comments it out)
sudo hostsctl disable --name server.local

# Enable entry (uncomments it)
sudo hostsctl enable --name server.local

# Use ID instead of name
sudo hostsctl enable --id 3
```

#### `backup/restore` - Backup management

```bash
# Create timestamped backup
sudo hostsctl backup

# Create backup to specific location
sudo hostsctl backup --out /tmp/hosts.backup

# Restore from backup
sudo hostsctl restore --file /etc/hosts.hostsctl.20240101-120000.bak
```

#### `import/export` - Profile management

```bash
# Export current entries to JSON
sudo hostsctl export --file development.json --format json

# Import entries from YAML
sudo hostsctl import --file production.yaml --format yaml

# Export to YAML
sudo hostsctl export --file current.yaml --format yaml
```

#### `verify` - Validate hosts file

```bash
# Check for syntax errors and duplicates
sudo hostsctl verify

# JSON output for automation
sudo hostsctl verify --json
```

### Global Options

- `--hosts-file PATH`: Use custom hosts file (default: `/etc/hosts`)
- `--json`: Output results in JSON format
- `--no-color`: Disable colored output

### Examples with Custom Hosts File

Perfect for development and testing:

```bash
# Create a test hosts file
echo "127.0.0.1 localhost" > test_hosts

# Use hostsctl with custom file (no sudo needed)
hostsctl --hosts-file test_hosts list
hostsctl --hosts-file test_hosts add --ip 192.168.1.1 --name test.local
hostsctl --hosts-file test_hosts verify
```

## Profile Management

hostsctl supports importing and exporting groups of hosts entries as profiles:

### Example JSON Profile

```json
{
  "name": "development",
  "description": "Development environment hosts",
  "entries": [
    {
      "ip": "127.0.0.1",
      "names": ["api.local", "app.local"],
      "comment": "Local services",
      "disabled": false
    }
  ]
}
```

### Example YAML Profile

```yaml
name: production
description: Production environment hosts
entries:
  - ip: "10.0.1.50"
    names:
      - api.prod.company.com
      - api-v1.prod.company.com
    comment: "Production API"
    disabled: false
```

## Safety Features

### Automatic Backups

Before any modification, hostsctl automatically creates a timestamped backup:

```
/etc/hosts.hostsctl.20240101-120000.bak
```

### Atomic Operations

All file writes use atomic operations (write to temp file + rename) to prevent corruption.

### File Locking

Concurrent access is prevented using file locking mechanisms.

### Validation

- IPv4 and IPv6 address validation
- RFC-compliant hostname validation
- Duplicate detection
- Syntax verification

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for build automation)
- Nix (optional, for reproducible builds)

### Building

#### Traditional Go Build

```bash
# Clone repository
git clone https://github.com/vaxvhbe/hostsctl.git
cd hostsctl

# Install dependencies
make deps

# Build binary
make build

# Run tests
make test

# Run all checks (format, vet, lint, test)
make check
```

#### Nix Build (Recommended)

For reproducible builds and development environments:

```bash
# Clone repository
git clone https://github.com/vaxvhbe/hostsctl.git
cd hostsctl

# Enter development shell with all tools
nix develop

# Build with Nix
make nix-build
# or directly: nix build

# Run tests
make test

# Install locally
make nix-install
```

The Nix flake provides:
- **Reproducible builds** across different systems
- **Development shell** with Go, golangci-lint, and all tools
- **Cross-platform** support (Linux, macOS, x86_64, ARM64)
- **Isolated environment** with pinned dependencies
```

### Running Tests

```bash
# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Create test hosts file for development
make test-hosts
```

### Project Structure

```
hostsctl/
├── cmd/hostsctl/           # CLI entrypoint
├── internal/
│   ├── hosts/              # Core hosts file operations
│   ├── cli/                # CLI command implementations
│   └── lock/               # File locking utilities
├── pkg/                    # Public utilities (validation)
├── configs/                # Example profiles
├── test/                   # Unit tests
├── Makefile               # Build automation
├── Dockerfile             # Container support
└── README.md
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Security

### Reporting Security Issues

Please report security vulnerabilities to [security@example.com](mailto:security@example.com).

### Security Best Practices

- Always run with minimum required privileges
- Use `--hosts-file` for testing instead of modifying system files
- Regularly backup your hosts file
- Verify imports from untrusted sources

## FAQ

### Q: Why do I need sudo?

hostsctl requires root privileges only when modifying `/etc/hosts`. You can use custom hosts files without sudo for development.

### Q: Can I use this in scripts?

Yes! Use `--json` output and check exit codes:

```bash
if hostsctl verify --json; then
    echo "Hosts file is valid"
else
    echo "Hosts file has issues"
fi
```

### Q: How do I restore if something goes wrong?

hostsctl automatically creates backups before any modification. Find them with:

```bash
ls /etc/hosts.hostsctl.*.bak
sudo hostsctl restore --file /etc/hosts.hostsctl.TIMESTAMP.bak
```

### Q: Can I manage multiple hosts files?

Yes, use the `--hosts-file` flag:

```bash
hostsctl --hosts-file /path/to/custom/hosts list
```

## Roadmap

- [ ] IPv6 support improvements
- [ ] Remote profile management (Git-backed)
- [ ] Daemon mode with REST API
- [ ] Integration with common development tools
- [ ] Advanced filtering and search capabilities

---

Made with ❤️ by the hostsctl team