package database

import (
	"fmt"
	"time"
	"strings"
)

// generateUserID generates a unique user ID
func generateUserID() string {
	// In a real implementation, you would use UUID or another proper ID generation
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate key") || 
		   strings.Contains(errStr, "unique constraint") ||
		   strings.Contains(errStr, "23505")
}
