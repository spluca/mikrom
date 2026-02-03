package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateVMID(t *testing.T) {
	vmID := GenerateVMID()

	// Check format: srv-xxxxxxxx
	assert.True(t, strings.HasPrefix(vmID, "srv-"), "VM ID should start with 'srv-'")
	assert.Equal(t, 12, len(vmID), "VM ID should be 12 characters long (srv- + 8 chars)")

	// Check that the ID part is alphanumeric
	idPart := strings.TrimPrefix(vmID, "srv-")
	assert.Equal(t, 8, len(idPart), "ID part should be 8 characters")
	for _, c := range idPart {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'), "ID should be hexadecimal")
	}
}

func TestGenerateVMID_Uniqueness(t *testing.T) {
	// Generate multiple VM IDs and check they're all unique
	ids := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		vmID := GenerateVMID()
		assert.False(t, ids[vmID], "Generated duplicate VM ID: %s", vmID)
		ids[vmID] = true
	}

	assert.Equal(t, iterations, len(ids), "Should have generated %d unique IDs", iterations)
}

func TestGenerateVMID_Format(t *testing.T) {
	vmID := GenerateVMID()

	// Test the exact format
	parts := strings.Split(vmID, "-")
	assert.Equal(t, 2, len(parts), "VM ID should have exactly one dash")
	assert.Equal(t, "srv", parts[0], "First part should be 'srv'")
	assert.Equal(t, 8, len(parts[1]), "Second part should be 8 characters")
}

func TestGenerateVMID_NoDashes(t *testing.T) {
	vmID := GenerateVMID()

	// Extract the ID part (after srv-)
	idPart := strings.TrimPrefix(vmID, "srv-")

	// Verify no dashes in the ID part
	assert.NotContains(t, idPart, "-", "ID part should not contain dashes")
}

func TestGenerateVMID_MultipleGenerations(t *testing.T) {
	// Test that multiple generations produce valid IDs
	for i := 0; i < 10; i++ {
		vmID := GenerateVMID()

		assert.True(t, strings.HasPrefix(vmID, "srv-"))
		assert.Equal(t, 12, len(vmID))

		// Verify format matches expected pattern
		idPart := strings.TrimPrefix(vmID, "srv-")
		assert.Regexp(t, "^[0-9a-f]{8}$", idPart, "ID part should be 8 hex characters")
	}
}
