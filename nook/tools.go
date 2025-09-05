//go:build tools
// +build tools

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.

package nook

import (
	_ "golang.org/x/tools/cmd/goimports"
	// Add other tools here as needed:
	// _ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
