// Package utils provides utility functions for the monitoring service.
package utils

// ContainsString checks if a slice contains a string.
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
