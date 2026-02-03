package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash, "Hash should not equal plain password")
	assert.True(t, len(hash) > 0, "Hash should have length > 0")
}

func TestHashPasswordEmpty(t *testing.T) {
	password := ""

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestCheckPassword_Valid(t *testing.T) {
	password := "mySecurePassword123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	result := CheckPassword(password, hash)

	assert.True(t, result, "Password should match hash")
}

func TestCheckPassword_Invalid(t *testing.T) {
	password := "mySecurePassword123"
	wrongPassword := "wrongPassword"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	result := CheckPassword(wrongPassword, hash)

	assert.False(t, result, "Wrong password should not match hash")
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	password := "mySecurePassword123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	result := CheckPassword("", hash)

	assert.False(t, result, "Empty password should not match")
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	password := "mySecurePassword123"
	invalidHash := "not-a-valid-hash"

	result := CheckPassword(password, invalidHash)

	assert.False(t, result, "Invalid hash should return false")
}

func TestHashPassword_SamePasswordDifferentHashes(t *testing.T) {
	password := "mySecurePassword123"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "Same password should generate different hashes (bcrypt uses salt)")

	// Both hashes should validate the same password
	assert.True(t, CheckPassword(password, hash1))
	assert.True(t, CheckPassword(password, hash2))
}
