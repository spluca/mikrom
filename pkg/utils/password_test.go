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

func TestHashPassword_LongPassword(t *testing.T) {
	// Bcrypt has a max length of 72 bytes - test that it returns an error
	password := string(make([]byte, 100)) + "test"

	hash, err := HashPassword(password)

	// Should return an error for passwords > 72 bytes
	assert.Error(t, err, "Should error on password > 72 bytes")
	assert.Empty(t, hash)
	assert.Contains(t, err.Error(), "password length exceeds 72 bytes")
}

func TestHashPassword_ExactlyMaxLength(t *testing.T) {
	// Test with exactly 72 bytes (the max)
	password := string(make([]byte, 72))

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, CheckPassword(password, hash))
}

func TestHashPassword_SpecialCharacters(t *testing.T) {
	passwords := []string{
		"p@ssw0rd!",
		"пароль", // Cyrillic
		"密码",     // Chinese
		"contraseña",
		"!@#$%^&*()",
		"pass word with spaces",
	}

	for _, password := range passwords {
		hash, err := HashPassword(password)
		assert.NoError(t, err, "Should hash password: %s", password)
		assert.NotEmpty(t, hash)
		assert.True(t, CheckPassword(password, hash), "Should validate password: %s", password)
	}
}

func TestCheckPassword_EmptyHash(t *testing.T) {
	password := "mySecurePassword123"
	emptyHash := ""

	result := CheckPassword(password, emptyHash)

	assert.False(t, result, "Empty hash should return false")
}

func TestCheckPassword_BothEmpty(t *testing.T) {
	result := CheckPassword("", "")

	assert.False(t, result, "Both empty should return false")
}

func TestHashPassword_VeryShortPassword(t *testing.T) {
	password := "a"

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, CheckPassword(password, hash))
}

func TestCheckPassword_CaseSensitive(t *testing.T) {
	password := "MySecurePassword"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	// Password should be case-sensitive
	assert.True(t, CheckPassword("MySecurePassword", hash))
	assert.False(t, CheckPassword("mysecurepassword", hash))
	assert.False(t, CheckPassword("MYSECUREPASSWORD", hash))
}
