// Package platform is the plumbing every Go service shares.
package platform

import (
	"fmt"
	"os"
)

// Getenv returns the variable, or def when it is unset or empty.
func Getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// MustGetenv returns the variable or an error naming it.
func MustGetenv(key string) (string, error) {
	if v := os.Getenv(key); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("environment variable %s is required", key)
}
