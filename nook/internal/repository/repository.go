package repository

import "context"

// Repository defines the basic CRUD operations for any entity type.
// This follows a similar pattern to Spring Data's Repository interface.
type Repository[T any, ID comparable] interface {
	// Save creates or updates an entity
	Save(ctx context.Context, entity T) (T, error)

	// FindByID retrieves an entity by its ID
	// Returns ErrNotFound if the entity doesn't exist
	FindByID(ctx context.Context, id ID) (T, error)

	// FindAll retrieves all entities
	FindAll(ctx context.Context) ([]T, error)

	// DeleteByID deletes an entity by its ID
	// Returns ErrNotFound if the entity doesn't exist
	DeleteByID(ctx context.Context, id ID) error

	// ExistsByID checks if an entity exists by its ID
	ExistsByID(ctx context.Context, id ID) (bool, error)
}
