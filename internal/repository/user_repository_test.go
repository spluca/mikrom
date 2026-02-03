package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/apardo/mikrom-go/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	return db, mock
}

func TestCreate_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	now := time.Now()

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.PasswordHash, user.Name).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, now, now))

	err := repo.Create(user)

	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Name:         "Test User",
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, user.PasswordHash, user.Name).
		WillReturnError(errors.New("database error"))

	err := repo.Create(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByEmail_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	now := time.Now()

	email := "test@example.com"

	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email`).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
			AddRow(1, email, "hashedpassword", "Test User", now, now))

	user, err := repo.FindByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "hashedpassword", user.PasswordHash)
	assert.Equal(t, "Test User", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByEmail_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	email := "notfound@example.com"

	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.FindByEmail(email)

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByEmail_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	email := "test@example.com"

	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email`).
		WithArgs(email).
		WillReturnError(errors.New("database error"))

	user, err := repo.FindByEmail(email)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "error finding user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	now := time.Now()

	userID := 1

	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "created_at", "updated_at"}).
			AddRow(1, "test@example.com", "hashedpassword", "Test User", now, now))

	user, err := repo.FindByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "hashedpassword", user.PasswordHash)
	assert.Equal(t, "Test User", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	userID := 999

	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.FindByID(userID)

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	userID := 1

	mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id`).
		WithArgs(userID).
		WillReturnError(errors.New("database error"))

	user, err := repo.FindByID(userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "error finding user")
	assert.NoError(t, mock.ExpectationsWereMet())
}
