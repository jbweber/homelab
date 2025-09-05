package api

import (
	"fmt"
	"net"
	"net/http"
)

// extractClientIP extracts the client IP from the request, preferring X-Forwarded-For header
// over RemoteAddr. Returns an error if the IP cannot be parsed.
func extractClientIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return "", fmt.Errorf("unable to parse remote address: %w", err)
		}
	}
	return ip, nil
}
