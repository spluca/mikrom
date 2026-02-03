package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/apardo/mikrom-go/internal/models"
	"github.com/apardo/mikrom-go/internal/repository"
	"github.com/apardo/mikrom-go/internal/service"
	"github.com/apardo/mikrom-go/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler(t *testing.T) (*AuthHandler, sqlmock.Sqlmock, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, "test-secret")
	authHandler := NewAuthHandler(authService)

	router := gin.New()

	return authHandler, mock, router
}

func TestRegisterHandler_Success(t *testing.T) {
	handler, mock, router := setupTestHandler(t)

	router.POST("/register", handler.Register)

	reqBody := models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	// Mock FindByEmail (no existe)
	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email`).
		WithArgs(reqBody.Email).
		WillReturnError(sql.ErrNoRows)

	// Mock Create
	now := time.Now()
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, now, now))

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
}

func TestRegisterHandler_EmailAlreadyExists(t *testing.T) {
	handler, mock, router := setupTestHandler(t)

	router.POST("/register", handler.Register)

	reqBody := models.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	now := time.Now()
	// Mock FindByEmail (usuario ya existe)
	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email`).
		WithArgs(reqBody.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
			AddRow(1, reqBody.Email, "hashedpassword", "Existing User", now, now))

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
	handler, _, router := setupTestHandler(t)

	router.POST("/register", handler.Register)

	invalidJSON := []byte(`{"email": "test@example.com"`) // JSON inválido

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_ValidationError(t *testing.T) {
	handler, _, router := setupTestHandler(t)

	router.POST("/register", handler.Register)

	reqBody := models.RegisterRequest{
		Email:    "invalid-email", // Email inválido
		Password: "123",           // Password muy corto
		Name:     "",              // Nombre vacío
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_Success(t *testing.T) {
	handler, mock, router := setupTestHandler(t)

	router.POST("/login", handler.Login)

	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	now := time.Now()
	// Mock FindByEmail
	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email`).
		WithArgs(reqBody.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
			AddRow(1, reqBody.Email, hashedPassword, "Test User", now, now))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, reqBody.Email, response.User.Email)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	handler, mock, router := setupTestHandler(t)

	router.POST("/login", handler.Login)

	reqBody := models.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	// Mock FindByEmail (no existe)
	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email`).
		WithArgs(reqBody.Email).
		WillReturnError(sql.ErrNoRows)

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
	handler, mock, router := setupTestHandler(t)

	router.POST("/login", handler.Login)

	correctPassword := "correctPassword"
	hashedPassword, _ := utils.HashPassword(correctPassword)

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongPassword",
	}

	now := time.Now()
	// Mock FindByEmail
	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email`).
		WithArgs(reqBody.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
			AddRow(1, reqBody.Email, hashedPassword, "Test User", now, now))

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

func TestLoginHandler_InvalidJSON(t *testing.T) {
	handler, _, router := setupTestHandler(t)

	router.POST("/login", handler.Login)

	invalidJSON := []byte(`{"email": "test@example.com"`) // JSON inválido

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetProfileHandler_Success(t *testing.T) {
	handler, mock, router := setupTestHandler(t)

	router.GET("/profile", handler.GetProfile)

	userID := 1
	now := time.Now()

	// Mock FindByID
	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
			AddRow(userID, "test@example.com", "hashedpassword", "Test User", now, now))

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()

	// Simular el contexto con user_id del middleware
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var user models.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestGetProfileHandler_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// No establecer user_id en el contexto

	handler.GetProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Unauthorized", response.Error)
}

func TestGetProfileHandler_UserNotFound(t *testing.T) {
	handler, mock, _ := setupTestHandler(t)

	userID := 999

	// Mock FindByID (no existe)
	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

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
