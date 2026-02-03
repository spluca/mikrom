# AGENTS.md - Coding Agent Guidelines

This guide provides essential information for AI coding agents working in this repository.

## Project Overview

**Mikrom-Go** is a Firecracker VM management API built with Go 1.25+, Gin, GORM, PostgreSQL, Redis, and asynq for background workers. It features JWT authentication, IP pool management, and asynchronous VM operations.

## Build & Run Commands

```bash
# Install dependencies
make install                    # Download and tidy dependencies

# Development
make dev                        # Run with hot-reload (requires air)
make run                        # Run API server directly
go run cmd/worker/main.go       # Run background worker

# Build
make build                      # Build for current platform
make build-linux                # Build for Linux
make build-all                  # Build for multiple platforms

# Database & Infrastructure
make docker-up                  # Start PostgreSQL & Redis
make docker-down                # Stop containers
make db-shell                   # Open PostgreSQL shell
go run cmd/seed/main.go         # Seed database with IP pools

# Code Quality
make fmt                        # Format code with gofmt
make fmt-check                  # Verify formatting
make vet                        # Run go vet
make lint                       # Run golangci-lint (if installed)
make check                      # Run all checks (fmt, vet, lint, test)
```

## Testing Commands

```bash
# Run all tests
make test                       # Run with race detector
make test-verbose               # Run with verbose output
make test-short                 # Skip integration tests

# Run a single test
go test ./internal/handlers -run TestRegisterHandler_Success -v
go test ./pkg/utils -run TestHashPassword -v

# Run tests for specific package
go test ./internal/handlers -v
go test ./internal/service -v
go test ./pkg/utils -v

# Coverage
make test-coverage              # Generate coverage report
make coverage-html              # Generate HTML coverage report

# Benchmarks
make bench                      # Run benchmarks
```

## Code Style Guidelines

### Import Organization

Group imports in this order (separated by blank lines):
1. Standard library
2. External dependencies
3. Internal packages

```go
import (
    "errors"
    "fmt"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "gorm.io/gorm"

    "github.com/apardo/mikrom-go/config"
    "github.com/apardo/mikrom-go/internal/models"
    "github.com/apardo/mikrom-go/internal/repository"
)
```

### Naming Conventions

- **Files**: `snake_case.go` (e.g., `auth_service.go`, `user_repository.go`)
- **Packages**: Short, lowercase, singular (e.g., `models`, `service`, `handlers`)
- **Types/Structs**: `PascalCase` (e.g., `AuthService`, `VMRepository`)
- **Functions/Methods**: `PascalCase` for exported, `camelCase` for unexported
- **Variables**: `camelCase` (e.g., `userID`, `authHandler`)
- **Constants**: `PascalCase` for exported, `camelCase` for unexported
- **Acronyms**: Keep uppercase in type names (e.g., `VMID`, `IPAddress`)

### Type Definitions

```go
// Struct with JSON tags and GORM tags
type User struct {
    ID           int       `json:"id" gorm:"primaryKey"`
    Email        string    `json:"email" gorm:"uniqueIndex;not null"`
    PasswordHash string    `json:"-" gorm:"not null"` // Hide from JSON
    Name         string    `json:"name" gorm:"not null"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// Request/Response models with validation tags
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Name     string `json:"name" binding:"required"`
}

// Table name customization (if needed)
func (VM) TableName() string {
    return "vms"
}
```

### Error Handling

```go
// Return errors with context using fmt.Errorf
if err := r.db.Create(user).Error; err != nil {
    return fmt.Errorf("error creating user: %w", err)
}

// Check for specific errors
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, nil // Return nil for not found cases
}

// Return domain-specific errors
if existingUser != nil {
    return nil, errors.New("email already exists")
}
```

### Handler Pattern

```go
func (h *AuthHandler) Register(c *gin.Context) {
    var req models.RegisterRequest

    // 1. Bind and validate input
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{
            Error: err.Error(),
        })
        return
    }

    // 2. Call service layer
    user, err := h.authService.Register(&req)
    if err != nil {
        // 3. Handle specific errors with appropriate status codes
        if err.Error() == "email already exists" {
            c.JSON(http.StatusConflict, models.ErrorResponse{
                Error: err.Error(),
            })
            return
        }
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{
            Error: "Internal server error",
        })
        return
    }

    // 4. Return success response
    c.JSON(http.StatusCreated, gin.H{
        "message": "User created successfully",
        "user":    user,
    })
}
```

### Repository Pattern

```go
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil // Return nil for not found
        }
        return nil, fmt.Errorf("error finding user: %w", err)
    }

    return &user, nil
}
```

### Testing Patterns

**Test File Naming**: `*_test.go` in the same package

**Test Function Naming**: `TestFunctionName_Scenario`

**Structure**: Use Arrange-Act-Assert pattern

```go
func TestRegister_Success(t *testing.T) {
    // Arrange
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    userRepo := repository.NewUserRepository(db)
    authService := service.NewAuthService(userRepo, "test-secret")

    req := &models.RegisterRequest{
        Email:    "test@example.com",
        Password: "password123",
        Name:     "Test User",
    }

    // Mock expectations
    mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrNoRows)
    mock.ExpectQuery(`INSERT`).WillReturnRows(
        sqlmock.NewRows([]string{"id"}).AddRow(1),
    )

    // Act
    user, err := authService.Register(req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, req.Email, user.Email)
}
```

### Comments & Documentation

- Add comments to exported functions, types, and packages
- Use Spanish for inline comments (as seen in codebase)
- Keep comments concise and meaningful
- Document complex logic and business rules

```go
// NewAuthService creates a new authentication service instance
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
    return &AuthService{
        userRepo:  userRepo,
        jwtSecret: jwtSecret,
    }
}

// Register maneja el registro de nuevos usuarios
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
    // Verificar si el email ya existe
    existingUser, err := s.userRepo.FindByEmail(req.Email)
    // ...
}
```

## Important Patterns & Conventions

1. **No ORM Query Builder in Tests**: Use `sqlmock` for database mocking
2. **Pointer Receivers**: Use pointer receivers for methods that modify state
3. **Constructor Pattern**: Use `New*` functions for initialization
4. **Dependency Injection**: Pass dependencies through constructors
5. **Context Propagation**: Use `c.Set()` and `c.Get()` for request-scoped data
6. **Status Constants**: Use `http.StatusX` constants for HTTP status codes
7. **Validation**: Use Gin's `binding` tags for input validation
8. **Password Security**: Always hash passwords with bcrypt

## Common Tasks

**Add a new endpoint**:
1. Define request/response models in `internal/models/`
2. Add repository method in `internal/repository/`
3. Implement business logic in `internal/service/`
4. Create handler in `internal/handlers/`
5. Register route in `cmd/api/main.go`
6. Write tests for each layer

**Run the full stack locally**:
```bash
make docker-up              # Start PostgreSQL & Redis
go run cmd/seed/main.go     # Seed database
make dev                    # Start API with hot-reload
go run cmd/worker/main.go   # Start worker (separate terminal)
```

## Project Structure

```
mikrom-go/
├── cmd/
│   ├── api/           # API server entrypoint
│   ├── worker/        # Background worker entrypoint
│   └── seed/          # Database seeder
├── config/            # Configuration loading
├── internal/          # Private application code
│   ├── handlers/      # HTTP handlers (controllers)
│   ├── middleware/    # HTTP middleware (auth, etc.)
│   ├── models/        # Domain models & DTOs
│   ├── repository/    # Data access layer
│   └── service/       # Business logic
└── pkg/               # Public libraries
    ├── database/      # Database connection
    ├── firecracker/   # Firecracker/Ansible client
    ├── utils/         # Utilities (JWT, password, etc.)
    └── worker/        # Task queue (asynq)
```

## References

- Main README: `README.md`
- Testing guide: `docs/TESTING.md`
- Project knowledge: `CLAUDE.md`
