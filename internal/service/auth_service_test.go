package service

import (
	"testing"

	"github.com/spluca/mikrom/internal/models"
	"github.com/spluca/mikrom/internal/repository"
	"github.com/spluca/mikrom/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	req := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	user, err := authService.Register(req)

	assert.NoError(t, err)
	require.NotNil(t, user)
	assert.NotZero(t, user.ID)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.Name, user.Name)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, req.Password, user.PasswordHash, "Password should be hashed")

	// Verify user was created in DB
	found, err := userRepo.FindByEmail(req.Email)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	// Create first user
	existingUser := &models.User{
		Email:        "existing@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Existing User",
	}
	db.Create(existingUser)

	// Try to register with same email
	req := &models.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	user, err := authService.Register(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email already exists", err.Error())
}

func TestRegister_FindByEmailError(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	// Close DB to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	req := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	user, err := authService.Register(req)

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestLogin_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	password := "password123"
	// Create user with hashed password
	hash, err := utils.HashPassword(password)
	require.NoError(t, err)

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: hash,
		Name:         "Test User",
	}
	db.Create(user)

	req := &models.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	response, err := authService.Login(req)

	assert.NoError(t, err)
	require.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, user.Email, response.User.Email)
	assert.Equal(t, user.Name, response.User.Name)
	assert.Equal(t, user.ID, response.User.ID)
}

func TestLogin_UserNotFound(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	req := &models.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	response, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "invalid credentials", err.Error())
}

func TestLogin_InvalidPassword(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	// Create user with hashed password
	correctPassword := "password123"
	hash, err := utils.HashPassword(correctPassword)
	require.NoError(t, err)

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: hash,
		Name:         "Test User",
	}
	db.Create(user)

	// Try to login with wrong password
	req := &models.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	response, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "invalid credentials", err.Error())
}

func TestLogin_DatabaseError(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	// Close DB to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	req := &models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	response, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetUserByID_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	// Create user
	expectedUser := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
	}
	db.Create(expectedUser)

	user, err := authService.GetUserByID(int(expectedUser.ID))

	assert.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Equal(t, expectedUser.Name, user.Name)
}

func TestGetUserByID_NotFound(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	user, err := authService.GetUserByID(999)

	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestGetUserByID_DatabaseError(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret")

	// Close DB to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	user, err := authService.GetUserByID(1)

	assert.Error(t, err)
	assert.Nil(t, user)
}
