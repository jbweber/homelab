package testutil

import (
	"strings"
	"testing"
)

func TestNewTestDSN(t *testing.T) {
	dsn := NewTestDSN("TestName")
	if !strings.Contains(dsn, "file:TestName?mode=memory&cache=shared") {
		t.Errorf("NewTestDSN did not generate expected DSN, got: %s", dsn)
	}
}
