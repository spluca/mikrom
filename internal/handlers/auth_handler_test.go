package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spluca/mikrom/internal/models"
	"github.com/spluca/mikrom/internal/repository"
	"github.com/spluca/mikrom/internal/service"
	"github.com/spluca/mikrom/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestHandler(t *testing.T) (*AuthHandler, *gorm.DB, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	db := repository.SetupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, "test-secret")
	authHandler := NewAuthHandler(authService)

	router := gin.New()

	return authHandler, db, router
}

func TestRegisterHandler_Success(t *testing.T) {
	handler, db, router := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	router.POST("/register", handler.Register)

	reqBody := models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User created successfully", response["message"])
	assert.NotNil(t, response["user"])

	// Verify user was created in DB
	var user models.User
	db.Where("email = ?", reqBody.Email).First(&user)
	assert.Equal(t, reqBody.Email, user.Email)
	assert.Equal(t, reqBody.Name, user.Name)
}

func TestRegisterHandler_EmailAlreadyExists(t *testing.T) {
	handler, db, router := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	router.POST("/register", handler.Register)

	// Create existing user
	existingUser := &models.User{
		Email:        "existing@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Existing User",
	}
	db.Create(existingUser)

	reqBody := models.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "email already exists", response.Error)
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	handler, db, router := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	router.POST("/register", handler.Register)

	invalidJSON := []byte(`{"email": "test@example.com"`) // Invalid JSON

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_ValidationError(t *testing.T) {
	handler, db, router := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	router.POST("/register", handler.Register)

	reqBody := models.RegisterRequest{
		Email:    "invalid-email", // Invalid email
		Password: "123",           // Password too short
		Name:     "",              // Empty name
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_Success(t *testing.T) {
	handler, db, router := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	router.POST("/login", handler.Login)

	password := "password123"
	hashedPassword, err := utils.HashPassword(password)
	require.NoError(t, err)

	// Create user with hashed password
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Name:         "Test User",
	}
	db.Create(user)

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LoginResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, reqBody.Email, response.User.Email)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	handler, db, router := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	router.POST("/login", handler.Login)

	reqBody := models.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid credentials", response.Error)
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	handler, db, router := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	router.POST("/login", handler.Login)

	correctPassword := "correctPassword"
	hashedPassword, err := utils.HashPassword(correctPassword)
	require.NoError(t, err)

	// Create user
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Name:         "Test User",
	}
	db.Create(user)

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongPassword",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid credentials", response.Error)
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	handler, db, router := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	router.POST("/login", handler.Login)

	invalidJSON := []byte(`{"email": "test@example.com"`) // Invalid JSON

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetProfileHandler_Success(t *testing.T) {
	handler, db, _ := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	// Create user
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
	}
	db.Create(user)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()

	// Simulate context with user_id from middleware
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", int(user.ID))

	handler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseUser models.User
	err := json.Unmarshal(w.Body.Bytes(), &responseUser)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, responseUser.ID)
	assert.Equal(t, user.Email, responseUser.Email)
}

func TestGetProfileHandler_Unauthorized(t *testing.T) {
	handler, db, _ := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id in context

	handler.GetProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Unauthorized", response.Error)
}

func TestGetProfileHandler_UserNotFound(t *testing.T) {
	handler, db, _ := setupTestHandler(t)
	defer repository.CleanupTestDB(t, db)

	userID := 999

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User not found", response.Error)
}
