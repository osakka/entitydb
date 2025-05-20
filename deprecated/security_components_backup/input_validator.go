package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// InputValidator handles input validation for API endpoints
type InputValidator struct {
	// Validation patterns
	patterns map[string]*regexp.Regexp
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	validator := &InputValidator{
		patterns: make(map[string]*regexp.Regexp),
	}

	// Initialize common validation patterns
	validator.patterns["username"] = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	validator.patterns["password"] = regexp.MustCompile(`^.{8,}$`) // Minimum 8 chars
	validator.patterns["entityID"] = regexp.MustCompile(`^entity_[a-zA-Z0-9_]{1,64}$`)
	validator.patterns["relID"] = regexp.MustCompile(`^rel_[a-zA-Z0-9_]{1,64}$`)
	validator.patterns["status"] = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
	validator.patterns["type"] = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
	validator.patterns["title"] = regexp.MustCompile(`^.{1,256}$`) // Non-empty, max 256
	validator.patterns["tag"] = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
	validator.patterns["role"] = regexp.MustCompile(`^(admin|user|readonly)$`)

	return validator
}

// Validate validates input based on the specified rules
func (v *InputValidator) Validate(input map[string]interface{}, rules map[string]string) []ValidationError {
	var errors []ValidationError

	for field, rule := range rules {
		// Skip validation if field is not required and not present
		value, exists := input[field]
		if !exists {
			if strings.Contains(rule, "required") {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: "Field is required",
				})
			}
			continue
		}

		// Handle different validation rules
		if strings.Contains(rule, "required") {
			// Already checked above
		}

		// Type validations
		if strings.Contains(rule, "string") {
			if _, ok := value.(string); !ok {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: "Field must be a string",
				})
				continue
			}
		}

		if strings.Contains(rule, "array") {
			if _, ok := value.([]interface{}); !ok {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: "Field must be an array",
				})
				continue
			}
		}

		if strings.Contains(rule, "object") {
			if _, ok := value.(map[string]interface{}); !ok {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: "Field must be an object",
				})
				continue
			}
		}

		// Pattern validations
		for pattern, regex := range v.patterns {
			if strings.Contains(rule, pattern) {
				if strValue, ok := value.(string); ok {
					if !regex.MatchString(strValue) {
						errors = append(errors, ValidationError{
							Field:   field,
							Message: fmt.Sprintf("Field does not match %s pattern", pattern),
						})
					}
				}
			}
		}

		// Array item validations
		if strings.Contains(rule, "array:") {
			if arr, ok := value.([]interface{}); ok {
				// Extract array item validation
				parts := strings.Split(rule, "array:")
				if len(parts) > 1 {
					itemRules := parts[1]
					for i, item := range arr {
						if strings.Contains(itemRules, "string") {
							if _, ok := item.(string); !ok {
								errors = append(errors, ValidationError{
									Field:   fmt.Sprintf("%s[%d]", field, i),
									Message: "Array item must be a string",
								})
								continue
							}
						}

						for pattern, regex := range v.patterns {
							if strings.Contains(itemRules, pattern) {
								if strItem, ok := item.(string); ok {
									if !regex.MatchString(strItem) {
										errors = append(errors, ValidationError{
											Field:   fmt.Sprintf("%s[%d]", field, i),
											Message: fmt.Sprintf("Array item does not match %s pattern", pattern),
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return errors
}

// ValidateAndRespond validates input and writes errors to response if any
func (v *InputValidator) ValidateAndRespond(w http.ResponseWriter, input map[string]interface{}, rules map[string]string) bool {
	errors := v.Validate(input, rules)
	if len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Validation failed",
			"errors":  errors,
		})
		return false
	}
	return true
}

// ValidateEntityCreate validates entity creation input
func (v *InputValidator) ValidateEntityCreate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"type":        "required|string|type",
		"title":       "required|string|title",
		"description": "string",
		"status":      "string|status",
		"tags":        "array:string|tag",
		"properties":  "object",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateEntityUpdate validates entity update input
func (v *InputValidator) ValidateEntityUpdate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"title":       "string|title",
		"description": "string",
		"status":      "string|status",
		"tags":        "array:string|tag",
		"properties":  "object",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateRelationshipCreate validates relationship creation input
func (v *InputValidator) ValidateRelationshipCreate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"source_id":  "required|string|entityID",
		"target_id":  "required|string|entityID",
		"type":       "required|string|type",
		"properties": "object",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateRelationshipUpdate validates relationship update input
func (v *InputValidator) ValidateRelationshipUpdate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"properties": "required|object",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateLogin validates login input
func (v *InputValidator) ValidateLogin(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"username": "required|string|username",
		"password": "required|string|password",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateUserCreate validates user creation input
func (v *InputValidator) ValidateUserCreate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"type":        "required|string|type",
		"title":       "required|string|username",
		"description": "string",
		"status":      "string|status",
		"tags":        "array:string|tag",
		"properties": "required|object",
	}
	
	// Basic validation
	if !v.ValidateAndRespond(w, input, rules) {
		return false
	}
	
	// Special validation for user properties
	properties, ok := input["properties"].(map[string]interface{})
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Validation failed",
			"errors": []ValidationError{{
				Field:   "properties",
				Message: "Must be a valid properties object with username, roles and password_hash",
			}},
		})
		return false
	}
	
	// Validate properties
	propRules := map[string]string{
		"username":      "required|string|username",
		"roles":         "required|array:string|role",
		"password_hash": "required|string",
	}
	
	propErrors := v.Validate(properties, propRules)
	if len(propErrors) > 0 {
		// Prefix field names with "properties."
		for i := range propErrors {
			propErrors[i].Field = "properties." + propErrors[i].Field
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Validation failed",
			"errors":  propErrors,
		})
		return false
	}
	
	return true
}