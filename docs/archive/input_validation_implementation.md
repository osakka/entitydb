# Input Validation Implementation

## Overview

This document describes the input validation system implemented for the EntityDB platform. The input validation system ensures that all data received by the API endpoints meets the required format and content constraints, enhancing security and reliability.

## Validation System Design

The input validation system is implemented as a separate module in `/opt/entitydb/src/input_validator.go`. It provides a structured, rule-based approach to validating input data across all API endpoints.

### Core Features

1. **Rule-based validation**: Define validation rules for each field
2. **Pattern validation**: Regular expression-based pattern validation
3. **Type validation**: Validate field data types (string, array, object)
4. **Required field validation**: Ensure required fields are present
5. **Structured error reporting**: Return clear, actionable validation errors
6. **Endpoint-specific validation**: Specialized validation for different API operations

## Validation Rules

Validation rules are defined using a simple string-based syntax:

```
"field": "required|string|pattern"
```

### Available Rule Types

| Rule Type | Description | Example |
|-----------|-------------|---------|
| required | Field must be present | "required" |
| string | Field must be a string | "string" |
| array | Field must be an array | "array" |
| object | Field must be an object | "object" |
| {pattern} | Field must match a predefined pattern | "username" |
| array:{rules} | Array items must match specified rules | "array:string" |

### Predefined Patterns

| Pattern | Description | Regular Expression |
|---------|-------------|-------------------|
| username | Username format | `^[a-zA-Z0-9_]{3,32}$` |
| password | Password format | `^.{8,}$` (8+ chars) |
| entityID | Entity ID format | `^entity_[a-zA-Z0-9_]{1,64}$` |
| relID | Relationship ID format | `^rel_[a-zA-Z0-9_]{1,64}$` |
| status | Status format | `^[a-zA-Z0-9_-]{1,32}$` |
| type | Type format | `^[a-zA-Z0-9_-]{1,32}$` |
| title | Title format | `^.{1,256}$` |
| tag | Tag format | `^[a-zA-Z0-9_-]{1,64}$` |
| role | Role format | `^(admin|user|readonly)$` |

## Implementation Details

### Validation Process

1. For each field in the input data, check if it exists and is required
2. Validate the field's data type (string, array, object)
3. If applicable, validate the field against a predefined pattern
4. For arrays, validate each item according to specified rules
5. Collect all validation errors
6. If any errors are found, return them to the client as a structured response

### Example Validation

```go
// Initialize validator
validator := NewInputValidator()

// Define validation rules
rules := map[string]string{
    "type":        "required|string|type",
    "title":       "required|string|title",
    "description": "string",
    "status":      "string|status",
    "tags":        "array:string|tag",
    "properties":  "object",
}

// Validate input
errors := validator.Validate(input, rules)

// Check for errors
if len(errors) > 0 {
    // Handle validation errors
}
```

### Endpoint-Specific Validation

The system provides specialized validation methods for each API endpoint:

```go
// Validate entity creation
validator.ValidateEntityCreate(w, input)

// Validate entity update
validator.ValidateEntityUpdate(w, input)

// Validate relationship creation
validator.ValidateRelationshipCreate(w, input)

// Validate relationship update
validator.ValidateRelationshipUpdate(w, input)

// Validate login
validator.ValidateLogin(w, input)

// Validate user creation
validator.ValidateUserCreate(w, input)
```

## Error Responses

When validation fails, the API returns a structured error response:

```json
{
  "status": "error",
  "message": "Validation failed",
  "errors": [
    {
      "field": "username",
      "message": "Field does not match username pattern"
    },
    {
      "field": "tags[1]",
      "message": "Array item does not match tag pattern"
    }
  ]
}
```

This format allows clients to clearly understand what validation issues need to be addressed.

## Integration Points

The input validation system is integrated at the following points in the codebase:

1. **Entity API**
   - Entity creation
   - Entity update

2. **Relationship API**
   - Relationship creation
   - Relationship update

3. **Authentication**
   - Login endpoint
   - User creation

4. **Administrative Operations**
   - All administrative endpoints

## Security Benefits

1. **Prevented Injection Attacks**: Validate input format to prevent injection attacks
2. **Reduced Invalid Data**: Ensure data consistency and prevent invalid data
3. **Improved Error Handling**: Clear error messages for client developers
4. **Enhanced API Robustness**: Reduce unexpected behavior from invalid input
5. **Enforcement of Business Rules**: Ensure data conforms to business requirements

## Testing

A test script is available at `/opt/entitydb/share/tests/entity/test_input_validation.sh` to verify the input validation implementation. The script tests:

1. Valid entity creation
2. Invalid entity creation (missing required fields)
3. Invalid entity creation (invalid field values)
4. Valid relationship creation
5. Invalid relationship creation (missing required fields)
6. Invalid relationship creation (invalid field values)
7. Invalid login (missing fields)
8. Invalid login (invalid username format)
9. Invalid user creation (invalid roles)
10. Valid user creation

Run the test script to verify that the input validation implementation is working correctly:

```bash
/opt/entitydb/share/tests/entity/test_input_validation.sh
```

## Future Enhancements

1. **Custom Validation Rules**: Allow defining custom validation rules at runtime
2. **Cross-field Validation**: Validate fields based on values of other fields
3. **Conditional Validation**: Apply validation rules conditionally
4. **Internationalization**: Support localized error messages
5. **Data Sanitization**: Extend validation to include data sanitization