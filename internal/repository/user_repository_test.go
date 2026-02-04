package repository

import (
	"testing"

	"github.com/spluca/mikrom/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCreate_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
	}

	err := repo.Create(user)

	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)

	// Verify user was actually created
	var found models.User
	db.First(&found, user.ID)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Name, found.Name)
}

func TestCreate_Error(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Close the database to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
	}
	err := repo.Create(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating user")
}

func TestFindByEmail_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create a user first
	expected := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
	}
	db.Create(expected)

	// Find the user
	user, err := repo.FindByEmail("test@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expected.ID, user.ID)
	assert.Equal(t, expected.Email, user.Email)
	assert.Equal(t, expected.PasswordHash, user.PasswordHash)
	assert.Equal(t, expected.Name, user.Name)
}

func TestFindByEmail_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	user, err := repo.FindByEmail("notfound@example.com")

	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestFindByEmail_Error(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Close the database to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	user, err := repo.FindByEmail("test@example.com")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "error finding user")
}

func TestFindByID_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create a user first
	expected := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
	}
	db.Create(expected)

	// Find the user
	user, err := repo.FindByID(int(expected.ID))

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expected.ID, user.ID)
	assert.Equal(t, expected.Email, user.Email)
	assert.Equal(t, expected.PasswordHash, user.PasswordHash)
	assert.Equal(t, expected.Name, user.Name)
}

func TestFindByID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	user, err := repo.FindByID(999)

	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestFindByID_Error(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Close the database to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	user, err := repo.FindByID(1)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "error finding user")
}

// Additional test: Multiple users
func TestFindByEmail_MultipleUsers(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create multiple users
	users := []*models.User{
		{Email: "user1@example.com", PasswordHash: "hash1", Name: "User 1"},
		{Email: "user2@example.com", PasswordHash: "hash2", Name: "User 2"},
		{Email: "user3@example.com", PasswordHash: "hash3", Name: "User 3"},
	}

	for _, u := range users {
		db.Create(u)
	}

	// Find each user
	for _, expected := range users {
		found, err := repo.FindByEmail(expected.Email)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, expected.Email, found.Email)
		assert.Equal(t, expected.Name, found.Name)
	}
}

// Additional test: Create with validation
func TestCreate_EmptyEmail(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	user := &models.User{
		Email:        "", // Empty email
		PasswordHash: "hashedpassword",
		Name:         "Test User",
	}

	err := repo.Create(user)

	// GORM should allow empty email unless we add validation
	// This test documents current behavior
	assert.NoError(t, err)
}
