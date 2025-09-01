package repository

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/jbweber/homelab/nook/internal/datastore"
)

// DatastoreRepository provides a generic implementation of Repository
// that works with our existing datastore.Datastore
type DatastoreRepository[T any, ID comparable] struct {
	ds     *datastore.Datastore
	entity reflect.Type
}

// NewDatastoreRepository creates a new generic repository
func NewDatastoreRepository[T any, ID comparable](ds *datastore.Datastore) *DatastoreRepository[T, ID] {
	var zero T
	entityType := reflect.TypeOf(zero)

	return &DatastoreRepository[T, ID]{
		ds:     ds,
		entity: entityType,
	}
}

// Save creates or updates an entity
// For now, this is a placeholder - we'll implement specific save logic
// in the concrete repository implementations
func (r *DatastoreRepository[T, ID]) Save(ctx context.Context, entity T) (T, error) {
	return entity, fmt.Errorf("Save not implemented for %s: %w", r.entity.Name(), ErrOperationNotSupported)
}

// FindByID retrieves an entity by its ID
// For now, this is a placeholder - we'll implement specific find logic
// in the concrete repository implementations
func (r *DatastoreRepository[T, ID]) FindByID(ctx context.Context, id ID) (T, error) {
	var zero T
	return zero, fmt.Errorf("FindByID not implemented for %s: %w", r.entity.Name(), ErrOperationNotSupported)
}

// FindAll retrieves all entities
// For now, this is a placeholder - we'll implement specific find logic
// in the concrete repository implementations
func (r *DatastoreRepository[T, ID]) FindAll(ctx context.Context) ([]T, error) {
	return nil, fmt.Errorf("FindAll not implemented for %s: %w", r.entity.Name(), ErrOperationNotSupported)
}

// DeleteByID deletes an entity by its ID
// For now, this is a placeholder - we'll implement specific delete logic
// in the concrete repository implementations
func (r *DatastoreRepository[T, ID]) DeleteByID(ctx context.Context, id ID) error {
	return fmt.Errorf("DeleteByID not implemented for %s: %w", r.entity.Name(), ErrOperationNotSupported)
}

// ExistsByID checks if an entity exists by its ID
// For now, this is a placeholder - we'll implement specific exists logic
// in the concrete repository implementations
func (r *DatastoreRepository[T, ID]) ExistsByID(ctx context.Context, id ID) (bool, error) {
	return false, fmt.Errorf("ExistsByID not implemented for %s: %w", r.entity.Name(), ErrOperationNotSupported)
}

// GetDatastore returns the underlying datastore (useful for specific implementations)
func (r *DatastoreRepository[T, ID]) GetDatastore() *datastore.Datastore {
	return r.ds
}

// Helper function to check if an error is a "not found" error from the database
func isNotFoundError(err error) bool {
	return err == sql.ErrNoRows
}
