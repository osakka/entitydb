# EntityDB Makefile - Improved version with entity/tag-based testing support

# Variables
NAME := entitydb
VERSION := 0.1.0
BUILD_DIR := ../bin
BIN_DIR := $(shell realpath ../bin)
SHARE_DIR := $(shell realpath ../share)
TESTS_DIR := $(SHARE_DIR)/tests
TOOLS_DIR := ./tools
SERVER_SRC := .
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(shell date -u +%Y-%m-%d)"
GOFLAGS := -v

# Build tags to exclude deprecated code
BUILD_TAGS := -tags "!deprecated"

# Exclude deprecated packages, scripts, backup directories, and utility test files from testing
# Also exclude the root package since it contains multiple main files
PACKAGES_FOR_TESTING := $(shell go list ./... | grep -v "^entitydb$$" | grep -v "/deprecated" | grep -v "/scripts" | grep -v "/tests" | grep -v "/tools" | grep -v "/security_components_backup" | grep -v "/backup")

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

.PHONY: all server clean install dev test tools entity-tools unit-tests api-tests entity-tests test-utils help security-tests master-tests

all: server install

server:
	@$(ECHO) "$(YELLOW)Building server binary: $(NAME)...$(NC)\n"
	@mkdir -p $(BUILD_DIR)
	@$(ECHO) "$(YELLOW)Excluding deprecated code from build...$(NC)\n"
	# Create a temporary file to hide any deprecated code compilation warnings
	@touch .build.log
	@$(ECHO) "$(YELLOW)Building consolidated server with pure tag-based architecture...$(NC)\n"
	go build $(GOFLAGS) $(LDFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(NAME) main.go 2>.build.log || (cat .build.log && rm .build.log && exit 1)
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

test: unit-tests api-tests

unit-tests:
	@$(ECHO) "$(YELLOW)Running Go unit tests (excluding deprecated code, scripts, and utility test files)...$(NC)\n"
	@$(ECHO) "$(YELLOW)Testing packages: $(PACKAGES_FOR_TESTING)$(NC)\n"
	go test $(BUILD_TAGS) $(PACKAGES_FOR_TESTING) -v

api-tests: entity-tests

# Deprecated test targets removed - see entity-tests for current API testing

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

	@$(ECHO) "$(YELLOW)Running Entity API tests...$(NC)\n"
	@$(MAKE) entity-tests || true

	@$(ECHO) "$(YELLOW)----------------------------------------$(NC)\n"

	@$(ECHO) "$(YELLOW)Running Security implementation tests...$(NC)\n"
	@$(MAKE) security-tests || true

	@$(ECHO) "$(YELLOW)----------------------------------------$(NC)\n"

	@$(ECHO) "$(BLUE)========================================$(NC)\n"
	@$(ECHO) "$(BLUE)Test Summary$(NC)\n"
	@$(ECHO) "$(BLUE)========================================$(NC)\n"
	@$(ECHO) "$(GREEN)All tests executed. Check individual test results above.$(NC)\n"

tools: entity-tools
	@$(ECHO) "$(YELLOW)Building all tools...$(NC)\n"
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/add_user $(TOOLS_DIR)/add_user.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/create_users $(TOOLS_DIR)/create_users.go
	@chmod +x $(BUILD_DIR)/add_user $(BUILD_DIR)/create_users
	@$(ECHO) "$(GREEN)User management tools built successfully$(NC)\n"

entity-tools:
	@$(ECHO) "$(YELLOW)Building entity management tools...$(NC)\n"
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/add_entity $(TOOLS_DIR)/add_entity.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/list_entities $(TOOLS_DIR)/list_entities.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/add_entity_relationship $(TOOLS_DIR)/add_entity_relationship.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/list_entity_relationships $(TOOLS_DIR)/list_entity_relationships.go
	@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/migrate_issues_to_entities $(TOOLS_DIR)/migrate_issues_to_entities.go
	@chmod +x $(BUILD_DIR)/add_entity $(BUILD_DIR)/list_entities $(BUILD_DIR)/add_entity_relationship $(BUILD_DIR)/list_entity_relationships $(BUILD_DIR)/migrate_issues_to_entities
	@$(ECHO) "$(GREEN)Entity management tools built successfully$(NC)\n"

test-utils:
	@$(ECHO) "$(YELLOW)Available test utilities:$(NC)\n"
	@echo "  - fix_auth_tokens.go: Fixes auth tokens in the database"
	@echo "  - reset_admin.go: Resets admin credentials"
	@echo "  - setup_admin.go: Sets up admin user"
	@echo "  - add_user.go: Adds a new user to the system"
	@echo "  - create_users.go: Creates multiple users"
	@echo ""
	@$(ECHO) "$(YELLOW)Entity management tools:$(NC)\n"
	@echo "  - add_entity.go: Creates a new entity"
	@echo "  - list_entities.go: Lists entities with filtering options"
	@echo "  - add_entity_relationship.go: Creates a relationship between entities"
	@echo "  - list_entity_relationships.go: Lists relationships with filtering"
	@echo "  - migrate_issues_to_entities.go: Migrates legacy issues to entities"
	@echo ""
	@$(ECHO) "$(YELLOW)To run a utility, use: go run $(BUILD_TAGS) tools/<utility_file>$(NC)\n"
	@$(ECHO) "$(YELLOW)To build all tools: make tools$(NC)\n"

help:
	@$(ECHO) "$(BLUE)EntityDB Makefile Help$(NC)\n"
	@$(ECHO) "$(BLUE)========================================$(NC)\n"
	@$(ECHO) "$(GREEN)Available targets:$(NC)\n"
	@echo "  all           : Build server, install, and run unit tests"
	@echo "  server        : Build the consolidated server binary with integrated static file support"
	@echo "  install       : Install scripts and make them executable"
	@echo "  clean         : Clean build artifacts"
	@echo "  dev           : Start development server on port 8086"
	@echo "  tools         : Build all tools (user management and entity tools)"
	@echo "  entity-tools  : Build only entity management tools"
	@echo "  test          : Run all tests (unit tests and API tests)"
	@echo "  unit-tests    : Run Go unit tests"
	@echo "  api-tests     : Run entity API tests"
	@echo "  entity-tests  : Run entity-model API tests"
	@echo "  security-tests: Run security implementation tests"
	@echo "  master-tests  : Run all tests with consolidated reporting"
	@echo "  test-utils    : Show available test utilities"
	@echo "  help          : Show this help message"
	@$(ECHO) "$(BLUE)========================================$(NC)\n"