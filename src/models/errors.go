package models

import (
	"errors"
)

// Standard repository errors
var (
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrDuplicate is returned when attempting to create a resource that already exists
	ErrDuplicate = errors.New("resource already exists")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when user lacks required permissions
	ErrForbidden = errors.New("forbidden")

	// ErrInternal is returned for internal server errors
	ErrInternal = errors.New("internal error")

	// ErrNoDatabaseConnection is returned when database connection fails
	ErrNoDatabaseConnection = errors.New("no database connection")

	// ErrInvalidQuery is returned when a query is malformed
	ErrInvalidQuery = errors.New("invalid query")

	// ErrDatabaseError is returned for general database errors
	ErrDatabaseError = errors.New("database error")
	
	// ErrFactoryNotRegistered is returned when a factory function is not registered
	ErrFactoryNotRegistered = errors.New("factory not registered")
)