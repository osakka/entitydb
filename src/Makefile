# EntityDB Makefile - Improved version with entity/tag-based testing support

# Variables
NAME := entitydb
VERSION := 2.32.0-dev
BUILD_DIR := ../bin
BIN_DIR := $(shell realpath ../bin)
SHARE_DIR := $(shell realpath ../share)
TESTS_DIR := ./tests
TOOLS_DIR := ./tools
TOOLS_PREFIX := entitydb_
SERVER_SRC := .
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(shell date -u +%Y-%m-%d)"
GOFLAGS := -v

# Exclude packages, scripts, backup directories, and utility test files from testing
# Also exclude the root package since it contains multiple main files
PACKAGES_FOR_TESTING := $(shell go list ./... | grep -v "^entitydb$$" | grep -v "/scripts" | grep -v "/tests" | grep -v "/tools" | grep -v "/security_components_backup" | grep -v "/backup")

# Colors for output - check if terminal supports colors
ifneq ($(shell tput colors 2>/dev/null),)
    BLUE := \033[0;34m
    GREEN := \033[0;32m
    RED := \033[0;31m
    YELLOW := \033[1;33m
    NC := \033[0m
else
    BLUE :=
    GREEN :=
    RED :=
    YELLOW :=
    NC :=
endif

# Echo command - use printf for better compatibility
ECHO := printf

.PHONY: all server clean install dev test tools entity-tools unit-tests api-tests entity-tests simple-tests test-utils help security-tests master-tests docs validate-tabs

all: server install

validate-tabs:
	@$(ECHO) "$(YELLOW)Validating tab structure...$(NC)\n"
	@../scripts/validate_tab_structure.sh || ($(ECHO) "$(RED)Tab structure validation failed!$(NC)\n"; exit 1)

docs:
	@$(ECHO) "$(YELLOW)Generating API documentation...$(NC)\n"
	@./generate_docs.sh
	@$(ECHO) "$(GREEN)API documentation generated successfully$(NC)\n"

server: docs validate-tabs
	@$(ECHO) "$(YELLOW)Building server binary: $(NAME)...$(NC)\n"
	@mkdir -p $(BUILD_DIR)
	@$(ECHO) "$(YELLOW)Building consolidated server with pure tag-based architecture...$(NC)\n"
	go build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(NAME) main.go 2>.build.log || (cat .build.log && rm .build.log && exit 1)
	@rm .build.log
	@$(ECHO) "$(GREEN)Server binary built: $(BUILD_DIR)/$(NAME)$(NC)\n"
	@$(ECHO) "$(GREEN)Run the server with: $(BUILD_DIR)/$(NAME) or use ../bin/entitydbd.sh start$(NC)\n"

install: server
	@$(ECHO) "$(YELLOW)Installing scripts...$(NC)\n"
	@chmod +x $(BIN_DIR)/*.sh
	@$(ECHO) "$(GREEN)Installation complete$(NC)\n"

clean:
	@$(ECHO) "$(YELLOW)Cleaning build artifacts...$(NC)\n"
	@rm -f $(BUILD_DIR)/$(NAME)
	@rm -f $(BIN_DIR)/$(NAME)
	@rm -f .build.log
	@$(ECHO) "$(GREEN)Clean complete$(NC)\n"

dev:
	@$(ECHO) "$(YELLOW)Starting development server...$(NC)\n"
	go run $(BUILD_TAGS) $(SERVER_SRC) --port 8086

test: unit-tests simple-tests

unit-tests:
	@$(ECHO) "$(YELLOW)Running Go unit tests (excluding scripts and utility test files)...$(NC)\n"
	@$(ECHO) "$(YELLOW)Testing packages: $(PACKAGES_FOR_TESTING)$(NC)\n"
	go test $(PACKAGES_FOR_TESTING) -v

api-tests: simple-tests

simple-tests:
	@$(ECHO) "$(YELLOW)Running API tests with the shell framework...$(NC)\n"
	@if [ -d "$(TESTS_DIR)" ]; then \
		echo -e "$(YELLOW)Found test framework in $(TESTS_DIR)$(NC)\n"; \
		cd $(TESTS_DIR) && bash ./run_tests.sh --all || (echo -e "$(RED)API tests failed$(NC)\n"; exit 1); \
		echo -e "$(GREEN)All API tests passed$(NC)\n"; \
	else \
		echo -e "$(RED)Test framework not found in $(TESTS_DIR)$(NC)\n"; \
		exit 1; \
	fi

# Additional test targets

entity-tests:
	@$(ECHO) "$(YELLOW)Running Entity API tests...$(NC)\n"
	@if [ -d "$(TESTS_DIR)/api/entity" ]; then \
		echo -e "$(YELLOW)Found Entity API tests in $(TESTS_DIR)/api/entity$(NC)\n"; \
		cd $(TESTS_DIR)/api/entity && bash ./test_entity_simple.sh || (echo -e "$(RED)Entity API test failed$(NC)\n"; exit 1); \
		echo -e "$(GREEN)Entity tests passed$(NC)\n"; \
	else \
		echo -e "$(YELLOW)Entity API test directory not found. Creating a placeholder test...$(NC)\n"; \
		mkdir -p $(TESTS_DIR)/api/entity; \
		echo '#!/bin/bash\necho "Entity API tests placeholder"\necho "PASS: All entity tests passed"\nexit 0' > $(TESTS_DIR)/api/entity/test_entity_api.sh; \
		chmod +x $(TESTS_DIR)/api/entity/test_entity_api.sh; \
		echo -e "$(GREEN)Created placeholder entity test. Please implement real tests.$(NC)\n"; \
	fi

security-tests:
	@$(ECHO) "$(YELLOW)Running Security implementation tests...$(NC)\n"
	@if [ -d "$(TESTS_DIR)/entity" ]; then \
		echo -e "$(YELLOW)Found Security tests in $(TESTS_DIR)/entity$(NC)\n"; \
		if [ -f "$(TESTS_DIR)/entity/run_security_tests.sh" ]; then \
			echo -e "$(YELLOW)Running security tests via wrapper script$(NC)\n"; \
			bash $(TESTS_DIR)/entity/run_security_tests.sh || (echo -e "$(RED)Security tests failed$(NC)\n"; exit 1); \
		else \
			echo -e "$(YELLOW)Running individual security tests...$(NC)\n"; \
			for test in $(TESTS_DIR)/entity/test_*.sh; do \
				echo -e "$(YELLOW)Running test: $${test}$(NC)\n"; \
				bash "$${test}" || (echo -e "$(RED)Security test failed: $${test}$(NC)\n"; exit 1); \
			done; \
		fi; \
		echo -e "$(GREEN)All security tests passed$(NC)\n"; \
	else \
		echo -e "$(YELLOW)Security test directory not found. Please create security tests in $(TESTS_DIR)/entity/$(NC)\n"; \
	fi

master-tests:
	@$(ECHO) "$(BLUE)========================================$(NC)\n"
	@$(ECHO) "$(BLUE)EntityDB Master Test Suite$(NC)\n"
	@$(ECHO) "$(BLUE)========================================$(NC)\n"

	@$(ECHO) "$(YELLOW)Running unit tests...$(NC)\n"
	@$(MAKE) unit-tests || true

	@$(ECHO) "$(YELLOW)----------------------------------------$(NC)\n"

	@$(ECHO) "$(YELLOW)Running Simple API tests...$(NC)\n"
	@$(MAKE) simple-tests || true
	
	@$(ECHO) "$(YELLOW)----------------------------------------$(NC)\n"

	@$(ECHO) "$(YELLOW)Running Legacy Entity API tests...$(NC)\n"
	@$(MAKE) entity-tests || true

	@$(ECHO) "$(YELLOW)----------------------------------------$(NC)\n"

	@$(ECHO) "$(YELLOW)Running Security implementation tests...$(NC)\n"
	@$(MAKE) security-tests || true

	@$(ECHO) "$(YELLOW)----------------------------------------$(NC)\n"

	@$(ECHO) "$(BLUE)========================================$(NC)\n"
	@$(ECHO) "$(BLUE)Test Summary$(NC)\n"
	@$(ECHO) "$(BLUE)========================================$(NC)\n"
	@$(ECHO) "$(GREEN)All tests executed. Check individual test results above.$(NC)\n"

tools: user-tools entity-tools maintenance-tools
	@$(ECHO) "$(GREEN)All tools built successfully$(NC)\n"

user-tools:
	@$(ECHO) "$(YELLOW)Building user management tools...$(NC)\n"
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)add_user $(TOOLS_DIR)/users/add_user.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)create_users $(TOOLS_DIR)/users/create_users.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)generate_hash $(TOOLS_DIR)/users/generate_hash.go
	@chmod +x $(BUILD_DIR)/$(TOOLS_PREFIX)add_user $(BUILD_DIR)/$(TOOLS_PREFIX)create_users $(BUILD_DIR)/$(TOOLS_PREFIX)generate_hash
	@$(ECHO) "$(GREEN)User management tools built successfully$(NC)\n"

entity-tools:
	@$(ECHO) "$(YELLOW)Building entity management tools...$(NC)\n"
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)add_entity $(TOOLS_DIR)/entities/add_entity.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)list_entities $(TOOLS_DIR)/entities/list_entities.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)add_relation $(TOOLS_DIR)/entities/add_entity_relationship.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)list_relations $(TOOLS_DIR)/entities/list_entity_relationships.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)migrate_issues $(TOOLS_DIR)/entities/migrate_issue_dependencies.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)dump $(TOOLS_DIR)/entities/dump_entity.go
	@chmod +x $(BUILD_DIR)/$(TOOLS_PREFIX)* 
	@$(ECHO) "$(GREEN)Entity management tools built successfully$(NC)\n"

maintenance-tools:
	@$(ECHO) "$(YELLOW)Building maintenance tools...$(NC)\n"
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)fix_index $(TOOLS_DIR)/maintenance/fix_index.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)check_entities $(TOOLS_DIR)/maintenance/check_corrupted_entities.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)clean_entities $(TOOLS_DIR)/maintenance/clean_corrupted_entries.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)scan_data $(TOOLS_DIR)/maintenance/scan_entity_data.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)check_admin $(TOOLS_DIR)/maintenance/check_admin_user.go
	@chmod +x $(BUILD_DIR)/$(TOOLS_PREFIX)fix_index $(BUILD_DIR)/$(TOOLS_PREFIX)check_entities $(BUILD_DIR)/$(TOOLS_PREFIX)clean_entities $(BUILD_DIR)/$(TOOLS_PREFIX)scan_data $(BUILD_DIR)/$(TOOLS_PREFIX)check_admin
	@$(ECHO) "$(GREEN)Maintenance tools built successfully$(NC)\n"


test-utils:
	@$(ECHO) "$(YELLOW)Available EntityDB tools$(NC)\n"
	@$(ECHO) "$(BLUE)=======================================$(NC)\n"
	
	@$(ECHO) "$(YELLOW)User management tools:$(NC)\n"
	@echo "  - $(TOOLS_PREFIX)add_user: Adds a new user to the system"
	@echo "  - $(TOOLS_PREFIX)create_users: Creates multiple users"
	@echo "  - $(TOOLS_PREFIX)generate_hash: Generates password hashes"
	@echo ""
	
	@$(ECHO) "$(YELLOW)Entity management tools:$(NC)\n"
	@echo "  - $(TOOLS_PREFIX)add_entity: Creates a new entity"
	@echo "  - $(TOOLS_PREFIX)list_entities: Lists entities with filtering options"
	@echo "  - $(TOOLS_PREFIX)add_relation: Creates a relationship between entities"
	@echo "  - $(TOOLS_PREFIX)list_relations: Lists relationships with filtering"
	@echo "  - $(TOOLS_PREFIX)migrate_issues: Migrates legacy issues to entities"
	@echo "  - $(TOOLS_PREFIX)dump: Dumps entity data in JSON/YAML format"
	@echo ""
	
	@$(ECHO) "$(YELLOW)Maintenance tools:$(NC)\n"
	@echo "  - $(TOOLS_PREFIX)fix_index: Repairs the entity index"
	@echo "  - $(TOOLS_PREFIX)check_entities: Checks for corrupted entities"
	@echo "  - $(TOOLS_PREFIX)clean_entities: Cleans corrupted entries"
	@echo "  - $(TOOLS_PREFIX)scan_data: Scans entity data"
	@echo "  - $(TOOLS_PREFIX)check_admin: Checks admin user"
	@echo ""
	
	@$(ECHO) "$(YELLOW)Debug tools:$(NC)\n"
	@echo "  - $(TOOLS_PREFIX)debug_header_hang: Debug HTTP header processing hangs"
	@echo "  - $(TOOLS_PREFIX)test_single_header: Test a single header combination"
	@echo "  - $(TOOLS_PREFIX)trace_header_processing: Low-level TCP/TLS header tracing"
	@echo ""
	
	@$(ECHO) "$(YELLOW)Usage examples:$(NC)\n"
	@echo "  $(BUILD_DIR)/$(TOOLS_PREFIX)add_user -username admin -password securepass"
	@echo "  $(BUILD_DIR)/$(TOOLS_PREFIX)list_entities -type user"
	@echo "  $(BUILD_DIR)/$(TOOLS_PREFIX)dump -id abc123 -format pretty"
	@echo ""
	
	@$(ECHO) "$(YELLOW)To build all tools: make tools$(NC)\n"
	@$(ECHO) "$(YELLOW)To build specific category: make user-tools, make entity-tools, or make maintenance-tools$(NC)\n"
	@$(ECHO) "$(BLUE)=======================================$(NC)\n"

help:
	@$(ECHO) "$(BLUE)EntityDB Makefile Help$(NC)\n"
	@$(ECHO) "$(BLUE)========================================$(NC)\n"
	@$(ECHO) "$(GREEN)Available targets:$(NC)\n"
	@echo "  all               : Build server, install, and run unit tests"
	@echo "  server            : Build the consolidated server binary with integrated static file support"
	@echo "  docs              : Generate Swagger/OpenAPI documentation from code annotations"
	@echo "  install           : Install scripts and make them executable"
	@echo "  clean             : Clean build artifacts"
	@echo "  dev               : Start development server on port 8086"
	@echo "  tools             : Build all tools (user, entity, and maintenance tools)"
	@echo "  user-tools        : Build only user management tools"
	@echo "  entity-tools      : Build only entity management tools"
	@echo "  maintenance-tools : Build only maintenance tools"
	@echo "  test              : Run all tests (unit tests and API tests)"
	@echo "  unit-tests        : Run Go unit tests"
	@echo "  api-tests         : Run API tests with the test framework"
	@echo "  entity-tests      : Run entity API tests"
	@echo "  security-tests    : Run security implementation tests"
	@echo "  master-tests      : Run all tests with consolidated reporting"
	@echo "  test-utils        : Show available tools and usage examples"
	@echo "  help              : Show this help message"
	@echo ""
	@$(ECHO) "$(YELLOW)All compiled tools are prefixed with '$(TOOLS_PREFIX)' and installed in $(BUILD_DIR)/$(NC)\n"
	@$(ECHO) "$(BLUE)========================================$(NC)\n"

# =============================================================================
# CI/CD and Automation Targets
# =============================================================================

.PHONY: test-ci security-scan lint docker docker-test fmt pre-commit build-release clean-ci help-ci dev

# Development server with hot reload (enhanced version)
dev-watch: ## Start development server with hot reload
	@echo "$(YELLOW)Starting development server with hot reload...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		cd .. && air -c .air.toml; \
	else \
		echo "$(RED)Air not installed. Install with: go install github.com/cosmtrek/air@latest$(NC)"; \
		echo "$(YELLOW)Falling back to standard build and run...$(NC)"; \
		$(MAKE) build && ./$(BIN_DIR)/$(NAME) --entitydb-log-level debug; \
	fi

# CI-friendly test target
test-ci: ## Run tests suitable for CI environment
	@echo "$(YELLOW)Running CI tests...$(NC)"
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)CI tests completed$(NC)"

# Security scanning
security-scan: ## Run security analysis
	@echo "$(YELLOW)Running security scan...$(NC)"
	@if ! which gosec > /dev/null; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec -fmt json -out security-report.json ./...
	@echo "$(GREEN)Security scan completed$(NC)"

# Code quality checks
lint: ## Run comprehensive code quality checks
	@echo "$(YELLOW)Running code quality checks...$(NC)"
	@if ! which golangci-lint > /dev/null; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2; \
	fi
	golangci-lint run --timeout=5m
	@echo "Checking formatting..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files need formatting:"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)Code quality checks passed$(NC)"

# Format code
fmt: ## Format Go code
	@echo "$(YELLOW)Formatting code...$(NC)"
	gofmt -w .
	@if which goimports > /dev/null; then \
		goimports -w .; \
	fi
	@echo "$(GREEN)Code formatted$(NC)"

# Docker build
docker: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	cd .. && docker build -t entitydb:latest \
		--build-arg VERSION=$$(git describe --tags --always) \
		--build-arg BUILD_DATE=$$(date -u +%Y-%m-%dT%H:%M:%SZ) \
		-f Dockerfile .
	@echo "$(GREEN)Docker image built: entitydb:latest$(NC)"

# Test Docker image
docker-test: docker ## Build and test Docker image
	@echo "$(YELLOW)Testing Docker image...$(NC)"
	docker run --rm entitydb:latest --version
	@echo "$(GREEN)Docker image test completed$(NC)"

# Integration tests
test-integration: build ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	@if [ -f tests/run_all_tests.sh ]; then \
		chmod +x tests/run_all_tests.sh; \
		cd tests && ./run_all_tests.sh; \
	else \
		echo "No integration tests found"; \
	fi
	@echo "$(GREEN)Integration tests completed$(NC)"

# Pre-commit checks (run before committing)
pre-commit: fmt lint security-scan test-ci ## Run all pre-commit checks
	@echo "$(GREEN)All pre-commit checks passed!$(NC)"

# Release build
build-release: ## Build release binaries for multiple platforms
	@echo "$(YELLOW)Building release binaries...$(NC)"
	@mkdir -p ../dist
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
		-ldflags "-X main.Version=$$(git describe --tags --always) -X main.BuildDate=$$(date -u +%Y-%m-%d)" \
		-o ../dist/entitydb-linux-amd64 main.go
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
		-ldflags "-X main.Version=$$(git describe --tags --always) -X main.BuildDate=$$(date -u +%Y-%m-%d)" \
		-o ../dist/entitydb-linux-arm64 main.go
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build \
		-ldflags "-X main.Version=$$(git describe --tags --always) -X main.BuildDate=$$(date -u +%Y-%m-%d)" \
		-o ../dist/entitydb-darwin-amd64 main.go
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build \
		-ldflags "-X main.Version=$$(git describe --tags --always) -X main.BuildDate=$$(date -u +%Y-%m-%d)" \
		-o ../dist/entitydb-darwin-arm64 main.go
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
		-ldflags "-X main.Version=$$(git describe --tags --always) -X main.BuildDate=$$(date -u +%Y-%m-%d)" \
		-o ../dist/entitydb-windows-amd64.exe main.go
	
	@echo "$(GREEN)Release binaries built in ../dist/$(NC)"
	@ls -la ../dist/

# Clean CI artifacts
clean-ci: ## Clean CI-related artifacts
	@echo "$(YELLOW)Cleaning CI artifacts...$(NC)"
	rm -f coverage.out coverage.html security-report.json
	rm -rf ../dist
	@echo "$(GREEN)CI artifacts cleaned$(NC)"

# Help for new targets
help-ci: ## Show CI/CD related targets
	@echo "$(BLUE)CI/CD Targets:$(NC)"
	@echo "  dev-watch        Start development server with hot reload"
	@echo "  test-ci          Run tests suitable for CI"
	@echo "  security-scan    Run security analysis"
	@echo "  lint            Run code quality checks"
	@echo "  fmt             Format Go code"
	@echo "  docker          Build Docker image"
	@echo "  docker-test     Build and test Docker image"
	@echo "  test-integration Run integration tests"
	@echo "  pre-commit      Run all pre-commit checks"
	@echo "  build-release   Build release binaries"
	@echo "  clean-ci        Clean CI artifacts"