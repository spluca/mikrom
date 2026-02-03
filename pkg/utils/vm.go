package utils

import (
	"fmt"

	"github.com/google/uuid"
)

// GenerateVMID generates a unique VM ID in the format srv-xxxxxxxx
func GenerateVMID() string {
	// Generate a UUID and use the first 8 characters
	id := uuid.New().String()
	// Take first 8 chars without dashes
	shortID := ""
	count := 0
	for _, c := range id {
		if c != '-' {
			shortID += string(c)
			count++
			if count == 8 {
				break
			}
		}
	}
	return fmt.Sprintf("srv-%s", shortID)
}
