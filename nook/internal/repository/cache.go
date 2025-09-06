package repository

import (
	"database/sql"
	"sync"
)

// PreparedStatementCache caches prepared statements for better performance
type PreparedStatementCache struct {
	mu         sync.RWMutex
	statements map[string]*sql.Stmt
	db         *sql.DB
}

// NewPreparedStatementCache creates a new prepared statement cache
func NewPreparedStatementCache(db *sql.DB) *PreparedStatementCache {
	return &PreparedStatementCache{
		statements: make(map[string]*sql.Stmt),
		db:         db,
	}
}

// Get retrieves or creates a prepared statement
func (c *PreparedStatementCache) Get(query string) (*sql.Stmt, error) {
	// First try to get from cache (read lock)
	c.mu.RLock()
	if stmt, ok := c.statements[query]; ok {
		c.mu.RUnlock()
		return stmt, nil
	}
	c.mu.RUnlock()

	// Not in cache, prepare statement (write lock)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if stmt, ok := c.statements[query]; ok {
		return stmt, nil
	}

	// Prepare the statement
	stmt, err := c.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	c.statements[query] = stmt
	return stmt, nil
}

// Close closes all prepared statements and clears the cache
func (c *PreparedStatementCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var lastErr error
	for _, stmt := range c.statements {
		if err := stmt.Close(); err != nil {
			lastErr = err
		}
	}

	c.statements = make(map[string]*sql.Stmt)
	return lastErr
}

// Clear removes a specific prepared statement from cache
func (c *PreparedStatementCache) Clear(query string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if stmt, ok := c.statements[query]; ok {
		delete(c.statements, query)
		return stmt.Close()
	}

	return nil
}

// Size returns the number of cached prepared statements
func (c *PreparedStatementCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.statements)
}
