#!/bin/bash
set -e

# EntityDB Development Environment Setup
# This script gets a developer productive in under 10 minutes

ENTITYDB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

check_command() {
    if command -v "$1" >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

install_go() {
    if check_command go; then
        local go_version=$(go version | cut -d' ' -f3 | sed 's/go//')
        log_success "Go $go_version is already installed"
        return 0
    fi
    
    log_info "Installing Go..."
    
    # Detect platform
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $arch in
        x86_64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) log_error "Unsupported architecture: $arch"; exit 1 ;;
    esac
    
    # Download and install Go
    local go_version="1.21.5"
    local go_url="https://golang.org/dl/go${go_version}.${os}-${arch}.tar.gz"
    
    if [[ "$os" == "darwin" ]]; then
        log_info "On macOS, please install Go using Homebrew:"
        echo "  brew install go"
        echo "Or download from: https://golang.org/dl/"
        exit 1
    elif [[ "$os" == "linux" ]]; then
        wget -O /tmp/go.tar.gz "$go_url"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf /tmp/go.tar.gz
        rm /tmp/go.tar.gz
        
        # Add to PATH if not already there
        if ! echo "$PATH" | grep -q "/usr/local/go/bin"; then
            echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
            export PATH="/usr/local/go/bin:$PATH"
        fi
    else
        log_error "Unsupported operating system: $os"
        exit 1
    fi
    
    log_success "Go installed successfully"
}

install_development_tools() {
    log_info "Installing development tools..."
    
    # Air for hot reloading
    if ! check_command air; then
        go install github.com/cosmtrek/air@latest
        log_success "Installed Air (hot reload)"
    fi
    
    # golangci-lint for code quality
    if ! check_command golangci-lint; then
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
        log_success "Installed golangci-lint"
    fi
    
    # gosec for security scanning
    if ! which gosec >/dev/null 2>&1; then
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        log_success "Installed gosec"
    fi
    
    # goimports for import management
    if ! which goimports >/dev/null 2>&1; then
        go install golang.org/x/tools/cmd/goimports@latest
        log_success "Installed goimports"
    fi
    
    # Swagger tools for API documentation
    if ! which swag >/dev/null 2>&1; then
        go install github.com/swaggo/swag/cmd/swag@latest
        log_success "Installed Swagger tools"
    fi
}

setup_git_hooks() {
    log_info "Setting up Git pre-commit hooks..."
    
    # Create pre-commit hook directory
    mkdir -p .git/hooks
    
    # Create pre-commit hook
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# EntityDB pre-commit hook

set -e

echo "🔍 Running pre-commit checks..."

# Change to src directory
cd src

# Format code
echo "  📝 Formatting code..."
gofmt -w .
goimports -w .

# Lint code
echo "  🔧 Running linter..."
golangci-lint run --timeout=5m

# Security scan
echo "  🔒 Security scan..."
gosec -quiet ./...

# Run tests
echo "  🧪 Running tests..."
go test -short ./...

echo "✅ All pre-commit checks passed!"
EOF

    chmod +x .git/hooks/pre-commit
    log_success "Git pre-commit hooks configured"
}

create_development_config() {
    log_info "Creating development configuration..."
    
    # Create development environment file
    cat > src/var/entitydb.dev.env << 'EOF'
# EntityDB Development Configuration
ENTITYDB_LOG_LEVEL=debug
ENTITYDB_PORT=8085
ENTITYDB_USE_SSL=false
ENTITYDB_DATA_PATH=./var
ENTITYDB_STATIC_DIR=../share/htdocs
ENTITYDB_TRACE_SUBSYSTEMS=auth,storage,temporal
EOF

    # Create Air configuration for hot reload
    cat > .air.toml << 'EOF'
root = "src"
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["--entitydb-log-level", "debug", "--entitydb-port", "8085"]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "docs", "../docs", "../tests"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
EOF

    log_success "Development configuration created"
}

setup_ide_configuration() {
    log_info "Setting up IDE configuration..."
    
    # VS Code settings
    mkdir -p .vscode
    
    cat > .vscode/settings.json << 'EOF'
{
    "go.buildOnSave": "package",
    "go.lintOnSave": "package",
    "go.formatTool": "goimports",
    "go.useLanguageServer": true,
    "go.testFlags": ["-v", "-race"],
    "go.testTimeout": "30s",
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    },
    "files.exclude": {
        "**/tmp": true,
        "**/var/*.ebf": true,
        "**/var/*.wal": true,
        "**/var/*.log": true
    }
}
EOF

    cat > .vscode/launch.json << 'EOF'
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch EntityDB",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/src",
            "args": [
                "--entitydb-log-level", "debug",
                "--entitydb-port", "8085"
            ],
            "cwd": "${workspaceFolder}/src",
            "env": {
                "ENTITYDB_LOG_LEVEL": "debug"
            }
        },
        {
            "name": "Debug Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/src",
            "args": ["-test.v"]
        }
    ]
}
EOF

    cat > .vscode/tasks.json << 'EOF'
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "build",
            "type": "shell",
            "command": "make",
            "args": ["build"],
            "options": {
                "cwd": "${workspaceFolder}/src"
            },
            "group": {
                "kind": "build",
                "isDefault": true
            }
        },
        {
            "label": "test",
            "type": "shell",
            "command": "make",
            "args": ["test"],
            "options": {
                "cwd": "${workspaceFolder}/src"
            },
            "group": {
                "kind": "test",
                "isDefault": true
            }
        },
        {
            "label": "dev",
            "type": "shell",
            "command": "air",
            "options": {
                "cwd": "${workspaceFolder}"
            },
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "new"
            }
        }
    ]
}
EOF

    # Recommended extensions
    cat > .vscode/extensions.json << 'EOF'
{
    "recommendations": [
        "golang.go",
        "ms-vscode.vscode-json",
        "redhat.vscode-yaml",
        "ms-vscode-remote.remote-containers",
        "github.vscode-pull-request-github"
    ]
}
EOF

    log_success "IDE configuration created"
}

verify_setup() {
    log_info "Verifying setup..."
    
    cd src
    
    # Check Go dependencies
    if ! go mod download; then
        log_error "Failed to download Go dependencies"
        exit 1
    fi
    
    # Build project
    if ! make build >/dev/null 2>&1; then
        log_error "Failed to build project"
        exit 1
    fi
    
    # Run quick test
    if ! go test -short ./... >/dev/null 2>&1; then
        log_warning "Some tests failed, but setup is complete"
    else
        log_success "All tests passed"
    fi
    
    cd ..
}

print_usage_instructions() {
    echo
    echo "🎉 EntityDB development environment is ready!"
    echo
    echo "Quick start commands:"
    echo "  ${GREEN}make dev${NC}           # Start development server with hot reload"
    echo "  ${GREEN}make test${NC}          # Run tests"
    echo "  ${GREEN}make lint${NC}          # Run code quality checks"
    echo "  ${GREEN}make build${NC}         # Build binary"
    echo
    echo "Development workflow:"
    echo "  1. Start the dev server: ${BLUE}make dev${NC}"
    echo "  2. Edit code - changes auto-reload in ~2 seconds"
    echo "  3. Tests run automatically on commit"
    echo "  4. Code is formatted automatically on save (in VS Code)"
    echo
    echo "Useful URLs:"
    echo "  • Server: http://localhost:8085"
    echo "  • API Docs: http://localhost:8085/swagger/"
    echo "  • Health: http://localhost:8085/health"
    echo
    echo "Need help? Check docs/development/ or run: ${BLUE}make help${NC}"
    echo
}

main() {
    echo "🚀 Setting up EntityDB development environment..."
    echo
    
    # Check if we're in the right directory
    if [[ ! -f "src/main.go" ]]; then
        log_error "Please run this script from the EntityDB project root"
        exit 1
    fi
    
    # Install dependencies
    install_go
    install_development_tools
    
    # Setup development environment
    setup_git_hooks
    create_development_config
    setup_ide_configuration
    
    # Verify everything works
    verify_setup
    
    # Show usage instructions
    print_usage_instructions
}

# Allow sourcing this script for individual functions
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi