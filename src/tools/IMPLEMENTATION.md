# EntityDB Tools Implementation Guide

## Overview

This document provides guidelines for implementing new tools for the EntityDB platform.

## Tool Structure

All tools should:

1. Be written in Go
2. Be located in the appropriate directory:
   - `users/`: User management tools
   - `entities/`: Entity management tools
   - `maintenance/`: System maintenance tools
3. Use a consistent command-line interface
4. Include proper error handling and logging
5. Be compiled with the `entitydb_` prefix

## Implementation Template

Here's a template for creating a new tool:

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Define command-line flags
	var (
		param1 string
		param2 bool
		param3 int
	)

	flag.StringVar(&param1, "param1", "", "Description of param1")
	flag.BoolVar(&param2, "param2", false, "Description of param2")
	flag.IntVar(&param3, "param3", 0, "Description of param3")
	flag.Parse()

	// Validate required parameters
	if param1 == "" {
		fmt.Println("Error: param1 is required")
		flag.Usage()
		os.Exit(1)
	}

	// Implement tool logic
	result, err := processData(param1, param2, param3)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Output results
	fmt.Printf("Operation completed successfully: %v\n", result)
}

func processData(param1 string, param2 bool, param3 int) (string, error) {
	// Implement your tool's functionality here
	return "result", nil
}
```

## Building Tools

After implementing your tool, add it to the appropriate section in the Makefile:

```makefile
# For a user tool:
@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)my_tool $(TOOLS_DIR)/users/my_tool.go

# For an entity tool:
@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)my_tool $(TOOLS_DIR)/entities/my_tool.go

# For a maintenance tool:
@go build $(GOFLAGS) $(BUILD_TAGS) -o $(BUILD_DIR)/$(TOOLS_PREFIX)my_tool $(TOOLS_DIR)/maintenance/my_tool.go
```

## Testing

Always test your tool thoroughly before committing. At minimum:

1. Test with valid parameters
2. Test with missing required parameters
3. Test with invalid parameters
4. Test error handling

## Documentation

Update the following files when adding a new tool:

1. Add tool description to `/opt/entitydb/src/tools/README.md`
2. Add usage example to the Makefile's `test-utils` target