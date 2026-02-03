package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apardo/mikrom-go/internal/models"
	"github.com/apardo/mikrom-go/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestMiddleware() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := setupTestMiddleware()
	secret := "test-secret"

	// Generar un token válido
	token, err := utils.GenerateJWT(1, "test@example.com", secret)
	assert.NoError(t, err)

	router.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		email, _ := c.Get("email")

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email":   email,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	router := setupTestMiddleware()
	secret := "test-secret"

	router.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// No se establece el header Authorization
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Authorization header is required", response.Error)
}

func TestAuthMiddleware_InvalidHeaderFormat_NoBearer(t *testing.T) {
	router := setupTestMiddleware()
	secret := "test-secret"

	token, err := utils.GenerateJWT(1, "test@example.com", secret)
	assert.NoError(t, err)

	router.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", token) // Sin "Bearer "
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidHeaderFormat_WrongPrefix(t *testing.T) {
	router := setupTestMiddleware()
	secret := "test-secret"

	token, err := utils.GenerateJWT(1, "test@example.com", secret)
	assert.NoError(t, err)

	router.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic "+token) // Debe ser "Bearer"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := setupTestMiddleware()
	secret := "test-secret"

	router.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	router := setupTestMiddleware()
	secret := "test-secret"
	wrongSecret := "wrong-secret"

	// Generar token con un secreto
	token, err := utils.GenerateJWT(1, "test@example.com", secret)
	assert.NoError(t, err)

	// Usar middleware con otro secreto
	router.GET("/protected", AuthMiddleware(wrongSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ContextValues(t *testing.T) {
	router := setupTestMiddleware()
	secret := "test-secret"
	userID := 123
	email := "test@example.com"

	token, err := utils.GenerateJWT(userID, email, secret)
	assert.NoError(t, err)

	router.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
		extractedUserID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, userID, extractedUserID)

		extractedEmail, exists := c.Get("email")
		assert.True(t, exists)
		assert.Equal(t, email, extractedEmail)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	router := setupTestMiddleware()
	secret := "test-secret"

	router.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MultipleSpacesInHeader(t *testing.T) {
	router := setupTestMiddleware()
	secret := "test-secret"

	token, err := utils.GenerateJWT(1, "test@example.com", secret)
	assert.NoError(t, err)

	router.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer  "+token+" extra") // Espacios extra
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Debería fallar porque el formato es incorrecto
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
