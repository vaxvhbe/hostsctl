# hostsctl — CLI Manager for /etc/hosts

> This file specifies and guides the development of **hostsctl**, a **Go** CLI application to manage entries in `/etc/hosts` safely and efficiently.

---

## Project Goal

Provide a small command-line application to:

* List existing `/etc/hosts` entries,
* Add / remove entries (by IP + hostname or comment/label),
* Enable / disable entries (comment/uncomment them),
* Backup and restore the hosts file,
* Manage groups of entries (profiles), and import/export them as JSON/YAML,
* Edit the file atomically with file locking to prevent corruption.

`hostsctl` should be simple, testable, and script-friendly (usable in CI).

---

## Key Features

* `hostsctl list` — show current entries (`--all` to include commented ones)
* `hostsctl add --ip 1.2.3.4 --name example.local [--comment "my-tag"]`
* `hostsctl rm --name example.local` or `hostsctl rm --id 123`
* `hostsctl enable|disable --name example.local`
* `hostsctl backup [--out PATH]` — create a timestamped backup
* `hostsctl restore --file PATH`
* `hostsctl import|export --file PATH [--format json|yaml]`
* `hostsctl apply --profile prod` — apply a predefined profile
* `hostsctl verify` — syntax validation and duplicate detection

Global option: `--hosts-file` to target a custom file (useful for testing).

---

## Constraints & Safety

* **Permissions**: Writing to `/etc/hosts` requires root/sudo. The binary should warn the user and refuse or request elevation.
* **Backup**: Always create a timestamped backup (`/etc/hosts.hostsctl.YYYYMMDD-HHMMSS.bak`) before writing.
* **Atomic writes**: Use a temporary file + `os.Rename` to replace the target file and use flock-style locking.
* **Validation**: Validate IPv4/IPv6 and hostnames against RFC rules.

---

## Project Structure

```
hostsctl/
├─ cmd/hostsctl/main.go        # CLI entrypoint (cobra/urfave/cli)
├─ internal/
│  ├─ hosts/                   # parsing, manipulation, serialization
│  │   ├─ parser.go
│  │   ├─ store.go              # atomic read/write + backup
│  │   └─ model.go              # types: Entry, Profile
│  ├─ cli/                      # CLI commands implementations
│  └─ lock/                     # cross-platform flock utility
├─ pkg/                         # reusable utils (validation)
├─ configs/                     # example profiles (json/yaml)
├─ test/                        # fixtures and unit tests
├─ Dockerfile
├─ Makefile
└─ README.md
```

---

## Code Examples

### Atomic write and backup

```go
// store.go (excerpt)
func WriteHosts(path string, content []byte) error {
    // 1) backup
    bak := fmt.Sprintf("%s.hostsctl.%s.bak", path, time.Now().Format("20060102-150405"))
    if err := os.WriteFile(bak, contentOld, 0644); err != nil { return err }

    // 2) write to temp file
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, content, 0644); err != nil { return err }

    // 3) atomic rename
    return os.Rename(tmp, path)
}
```

### Simple parser for hosts lines

```go
// model.go
type Entry struct {
  ID int `json:"id"`
  IP string `json:"ip"`
  Names []string `json:"names"`
  Comment string `json:"comment"`
  Disabled bool `json:"disabled"`
}

// parser.go: function reads lines and builds []Entry
```

---

## CLI & UX

* Use `spf13/cobra` or `urfave/cli` for command-line interface.
* Default output: human-readable table; `--json` for machine-friendly output.
* Optional colored output (disable with `--no-color`).

---

## Tests & CI

* Unit tests for parser, IP/host validation, and write operations using a temp `--hosts-file`.
* GitHub Actions workflow: `go test ./...`, `golangci-lint`, and `go vet`.

---

## Packaging

* Build static binaries (`CGO_ENABLED=0`) for distribution.
* Provide a DEB package or installation script that installs to `/usr/local/bin/hostsctl`.

---

## Usage Examples

```bash
# list entries
sudo hostsctl list --hosts-file /etc/hosts
# add entry
sudo hostsctl add --ip 127.0.0.1 --name mylocal.test --comment dev
# backup
sudo hostsctl backup
```

---

## Future Improvements

* Advanced remote profiles (Git-backed)
* Optional daemon mode + local REST API
* Fine-grained ACLs and audit logging

---

## License

MIT

---

*End of document.*
