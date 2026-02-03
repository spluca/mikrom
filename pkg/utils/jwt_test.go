package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT(t *testing.T) {
	userID := 123
	email := "test@example.com"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateJWT_Valid(t *testing.T) {
	userID := 123
	email := "test@example.com"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)
	assert.NoError(t, err)

	claims, err := ValidateJWT(token, secret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	secret := "test-secret-key"
	invalidToken := "invalid.token.here"

	claims, err := ValidateJWT(invalidToken, secret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := 123
	email := "test@example.com"
	secret := "test-secret-key"
	wrongSecret := "wrong-secret-key"

	token, err := GenerateJWT(userID, email, secret)
	assert.NoError(t, err)

	claims, err := ValidateJWT(token, wrongSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateJWT_EmptyToken(t *testing.T) {
	secret := "test-secret-key"

	claims, err := ValidateJWT("", secret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWT_ExpirationTime(t *testing.T) {
	userID := 123
	email := "test@example.com"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)
	assert.NoError(t, err)

	claims, err := ValidateJWT(token, secret)
	assert.NoError(t, err)

	// Verificar que el token expira en aproximadamente 24 horas
	now := time.Now()
	expiresAt := claims.ExpiresAt.Time

	duration := expiresAt.Sub(now)

	// Debe ser aproximadamente 24 horas (con un margen de 1 minuto)
	expectedDuration := 24 * time.Hour
	assert.InDelta(t, expectedDuration, duration, float64(time.Minute), "Token should expire in ~24 hours")
}

func TestJWT_IssuedAt(t *testing.T) {
	userID := 123
	email := "test@example.com"
	secret := "test-secret-key"

	beforeGeneration := time.Now().Add(-time.Second) // Margen de 1 segundo antes
	token, err := GenerateJWT(userID, email, secret)
	afterGeneration := time.Now().Add(time.Second) // Margen de 1 segundo después

	assert.NoError(t, err)

	claims, err := ValidateJWT(token, secret)
	assert.NoError(t, err)

	issuedAt := claims.IssuedAt.Time

	assert.True(t, issuedAt.After(beforeGeneration) || issuedAt.Equal(beforeGeneration))
	assert.True(t, issuedAt.Before(afterGeneration) || issuedAt.Equal(afterGeneration))
}

func TestGenerateJWT_DifferentTokensForSameUser(t *testing.T) {
	userID := 123
	email := "test@example.com"
	secret := "test-secret-key"

	token1, err1 := GenerateJWT(userID, email, secret)
	time.Sleep(2 * time.Second) // Pausa de 2 segundos para asegurar diferentes timestamps
	token2, err2 := GenerateJWT(userID, email, secret)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, token1, token2, "Tokens generated at different times should be different")

	// Ambos tokens deben ser válidos
	claims1, err := ValidateJWT(token1, secret)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims1.UserID)

	claims2, err := ValidateJWT(token2, secret)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims2.UserID)
}

func TestGenerateJWT_EmptyEmail(t *testing.T) {
	userID := 123
	email := ""
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateJWT(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, "", claims.Email)
}

func TestGenerateJWT_ZeroUserID(t *testing.T) {
	userID := 0
	email := "test@example.com"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateJWT(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, 0, claims.UserID)
}

func TestGenerateJWT_NegativeUserID(t *testing.T) {
	userID := -1
	email := "test@example.com"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateJWT(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, -1, claims.UserID)
}

func TestGenerateJWT_EmptySecret(t *testing.T) {
	userID := 123
	email := "test@example.com"
	secret := ""

	token, err := GenerateJWT(userID, email, secret)

	// Should still work with empty secret (not recommended in production)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateJWT_EmptySecret(t *testing.T) {
	userID := 123
	email := "test@example.com"
	secret := ""

	token, err := GenerateJWT(userID, email, secret)
	assert.NoError(t, err)

	claims, err := ValidateJWT(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestGenerateJWT_LongEmail(t *testing.T) {
	userID := 123
	email := "verylongemailaddress" + string(make([]byte, 200)) + "@example.com"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateJWT(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	secret := "test-secret-key"
	malformedTokens := []string{
		"not.a.jwt",
		"only-one-part",
		"two.parts",
		"",
		"Bearer token",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", // Only header
	}

	for _, token := range malformedTokens {
		claims, err := ValidateJWT(token, secret)
		assert.Error(t, err, "Should error on malformed token: %s", token)
		assert.Nil(t, claims)
	}
}

func TestGenerateJWT_SpecialCharactersInEmail(t *testing.T) {
	userID := 123
	email := "test+special@example.com"
	secret := "test-secret-key"

	token, err := GenerateJWT(userID, email, secret)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateJWT(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, email, claims.Email)
}
