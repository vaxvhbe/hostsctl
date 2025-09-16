{
  description = "hostsctl - A CLI manager for /etc/hosts files";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # Simple version extraction - Nix will handle this properly
        version = "dev";

      in {
        packages = {
          hostsctl = pkgs.buildGoModule {
            pname = "hostsctl";
            version = version;

            src = ./.;

            # Vendor hash for Go dependencies
            vendorHash = "sha256-VfNA2DxJJl7wRdSa7jSJNgwxvyFMszUpRGNsyNVHiN0=";

            # Build configuration
            env.CGO_ENABLED = "0";

            # Custom build flags
            ldflags = [ "-X main.version=${version}" ];

            # Main package to build
            subPackages = [ "cmd/hostsctl" ];

            # Include example configs in the output
            postInstall = ''
              mkdir -p $out/share/hostsctl
              cp -r configs/ $out/share/hostsctl/examples
            '';

            meta = with pkgs.lib; {
              description = "A command-line tool for safely managing entries in /etc/hosts files";
              longDescription = ''
                hostsctl is a modern, safe, and efficient command-line tool for managing /etc/hosts files.
                It provides atomic writes, automatic backups, file locking, and RFC-compliant validation.

                Features:
                - Safe operations with atomic writes and automatic backups
                - File locking to prevent corruption from concurrent access
                - IPv4/IPv6 and hostname validation according to RFC standards
                - Import/export profiles in JSON and YAML formats
                - Enable/disable entries without removing them
                - Custom hosts file support for development and testing
              '';
              homepage = "https://github.com/vaxvhbe/hostsctl";
              license = licenses.mit;
              maintainers = [ ];
              platforms = platforms.unix;
              mainProgram = "hostsctl";
            };
          };

          # Default package for nix profile install
          default = self.packages.${system}.hostsctl;
        };

        # Development shell with Go and tools
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            golangci-lint
            git
            gnumake
          ];

          shellHook = ''
            echo "🚀 hostsctl development environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Available commands:"
            echo "  make build    - Build the binary"
            echo "  make test     - Run tests"
            echo "  make fmt      - Format code"
            echo "  make lint     - Run linter"
            echo "  nix build     - Build with Nix"
            echo ""
          '';
        };

        # Apps for nix run
        apps = {
          hostsctl = flake-utils.lib.mkApp {
            drv = self.packages.${system}.hostsctl;
            name = "hostsctl";
          };
          default = self.apps.${system}.hostsctl;
        };
      });
}