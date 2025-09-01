package repository

import "errors"

// Common repository errors that can be checked with errors.Is()
var (
	// ErrNotFound is returned when an entity is not found
	ErrNotFound = errors.New("entity not found")

	// ErrDuplicate is returned when attempting to create an entity that already exists
	ErrDuplicate = errors.New("entity already exists")

	// ErrInvalidEntity is returned when an entity fails validation
	ErrInvalidEntity = errors.New("invalid entity")

	// ErrOperationNotSupported is returned when an operation is not supported
	ErrOperationNotSupported = errors.New("operation not supported")
)
