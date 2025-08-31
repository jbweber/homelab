package testutil

import (
	"fmt"
)

// NewTestDSN generates a DSN for an in-memory SQLite database for testing purposes.
func NewTestDSN(testName string) string {
	return fmt.Sprintf("file:%s?mode=memory&cache=shared", testName)
}
