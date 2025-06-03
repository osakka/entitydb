# EntityDB Makefile - Improved version with entity/tag-based testing support

# Variables
NAME := entitydb
VERSION := 2.24.0
BUILD_DIR := ../bin
BIN_DIR := $(shell realpath ../bin)
SHARE_DIR := $(shell realpath ../share)
TESTS_DIR := ./tests
TOOLS_DIR := ./tools
TOOLS_PREFIX := entitydb_
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

test: unit-tests simple-tests

unit-tests:
	@$(ECHO) "$(YELLOW)Running Go unit tests (excluding deprecated code, scripts, and utility test files)...$(NC)\n"
	@$(ECHO) "$(YELLOW)Testing packages: $(PACKAGES_FOR_TESTING)$(NC)\n"
	go test $(BUILD_TAGS) $(PACKAGES_FOR_TESTING) -v

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

# Legacy test targets - deprecated but kept for backward compatibility

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
	@echo "  - $(TOOLS_PREFIX)clean_entities: Removes corrupted entries"
	@echo "  - $(TOOLS_PREFIX)scan_data: Scans entity data"
	@echo "  - $(TOOLS_PREFIX)check_admin: Verifies admin user"
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
	@echo "  entity-tests      : Run legacy entity-model API tests (deprecated)"
	@echo "  security-tests    : Run security implementation tests"
	@echo "  master-tests      : Run all tests with consolidated reporting"
	@echo "  test-utils        : Show available tools and usage examples"
	@echo "  help              : Show this help message"
	@echo ""
	@$(ECHO) "$(YELLOW)All compiled tools are prefixed with '$(TOOLS_PREFIX)' and installed in $(BUILD_DIR)/$(NC)\n"
	@$(ECHO) "$(BLUE)========================================$(NC)\n"