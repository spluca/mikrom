# CLAUDE.md - Project Knowledge Base

> **Purpose**: This file contains comprehensive knowledge about the Mikrom Go API project for AI assistants (Claude) to understand the project context, architecture, decisions made, and development history.

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture & Design Decisions](#architecture--design-decisions)
3. [Project Structure](#project-structure)
4. [Technology Stack](#technology-stack)
5. [Core Components](#core-components)
6. [Testing Strategy](#testing-strategy)
7. [Development Workflow](#development-workflow)
8. [API Specification](#api-specification)
9. [Database Schema](#database-schema)
10. [Configuration & Environment](#configuration--environment)
11. [Development History](#development-history)
12. [Known Limitations & Future Work](#known-limitations--future-work)
13. [Common Tasks](#common-tasks)

---

## Project Overview

**Mikrom Go API** is a production-ready REST API built with Go, implementing user authentication with JWT tokens. The project follows clean architecture principles and includes comprehensive testing.

### Key Features
- User registration and login
- JWT-based authentication (24-hour token expiration)
- Password hashing with bcrypt
- PostgreSQL database with raw SQL queries
- Protected endpoints with middleware
- Comprehensive test suite (53 tests, 63.6% coverage)
- Docker Compose for local development
- Hot-reload development with Air
- Extensive Makefile automation (35+ commands)

### Project Goals
1. Build a secure, scalable authentication API
2. Follow Go best practices and clean architecture
3. Maintain high test coverage
4. Provide excellent developer experience
5. Keep documentation in English for international collaboration

---

## Architecture & Design Decisions

### Clean Architecture
The project follows clean architecture principles with clear separation of concerns:

```
Handlers (HTTP Layer)
    ↓
Services (Business Logic)
    ↓
Repositories (Data Access)
    ↓
Database (PostgreSQL)
```

**Key Principles**:
- **Dependency Inversion**: Higher layers depend on interfaces, not implementations
- **Single Responsibility**: Each component has one clear purpose
- **Testability**: All layers can be tested independently with mocks

### Why Raw SQL Instead of ORM?
**Decision**: Use `database/sql` with raw SQL queries instead of an ORM like GORM.

**Reasons**:
1. **Performance**: Direct SQL queries are faster and more predictable
2. **Transparency**: Exact control over what queries are executed
3. **Simplicity**: No need to learn ORM-specific syntax
4. **Flexibility**: Easy to optimize complex queries
5. **Testing**: Works seamlessly with `sqlmock` for unit tests

### Why Gin Framework?
**Decision**: Use Gin instead of standard library or other frameworks.

**Reasons**:
1. **Performance**: One of the fastest Go web frameworks
2. **Popularity**: Large community, extensive documentation
3. **Features**: Built-in validation, middleware support, JSON handling
4. **Simplicity**: Easy to learn, minimal boilerplate
5. **Production-Ready**: Used by many companies in production

### JWT Token Strategy
**Decision**: Use JWT tokens with 24-hour expiration, stored in Authorization header.

**Reasons**:
1. **Stateless**: No server-side session storage needed
2. **Scalable**: Works across multiple server instances
3. **Standard**: Industry-standard authentication method
4. **Secure**: Signed with HS256 algorithm
5. **Simple**: Easy to implement and validate

**Format**: `Authorization: Bearer <token>`

### Password Security
**Decision**: Use bcrypt with default cost (10) for password hashing.

**Reasons**:
1. **Industry Standard**: Widely recognized secure hashing algorithm
2. **Adaptive**: Cost factor can be increased as hardware improves
3. **Salt Built-in**: Automatic salting prevents rainbow table attacks
4. **Go Support**: Excellent library support with `golang.org/x/crypto/bcrypt`

---

## Project Structure

```
mikrom-go/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Configuration loader
├── internal/
│   ├── handlers/                # HTTP request handlers
│   │   ├── auth_handler.go
│   │   └── auth_handler_test.go
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go
│   │   └── auth_test.go
│   ├── models/
│   │   └── user.go              # Data models & DTOs
│   ├── repository/              # Data access layer
│   │   ├── user_repository.go
│   │   └── user_repository_test.go
│   └── service/                 # Business logic layer
│       ├── auth_service.go
│       └── auth_service_test.go
├── pkg/
│   ├── database/
│   │   └── database.go          # PostgreSQL connection
│   └── utils/
│       ├── jwt.go               # JWT utilities
│       ├── jwt_test.go
│       ├── password.go          # Password utilities
│       └── password_test.go
├── docs/
│   ├── INDEX.md                 # Documentation index
│   ├── README.md                # Full documentation
│   └── TESTING.md               # Testing guide
├── .air.toml                    # Hot-reload configuration
├── .env.example                 # Environment variables template
├── .gitignore                   # Git ignore rules
├── docker-compose.yml           # PostgreSQL setup
├── go.mod                       # Go module definition
├── go.sum                       # Go dependencies checksums
├── Makefile                     # Automation commands
├── README.md                    # Quick start guide
└── CLAUDE.md                    # This file
```

### Directory Purposes

#### `cmd/api/`
- Application entry points
- Main function, server initialization
- Dependency injection and wiring

#### `config/`
- Configuration management
- Environment variable loading
- Application settings

#### `internal/`
- Private application code (cannot be imported by other projects)
- Core business logic and HTTP handling

#### `internal/handlers/`
- HTTP request handlers
- Request validation
- Response formatting
- HTTP status codes

#### `internal/middleware/`
- HTTP middleware components
- Authentication, logging, CORS, etc.

#### `internal/models/`
- Data structures (User, Product, etc.)
- Request/Response DTOs
- Validation rules

#### `internal/repository/`
- Database operations
- SQL queries
- Data persistence layer

#### `internal/service/`
- Business logic
- Use cases implementation
- Orchestrates repositories

#### `pkg/`
- Public, reusable packages
- Can be imported by other projects
- Utilities, helpers, shared code

#### `docs/`
- All project documentation
- API guides, testing guides
- Architecture documentation

---

## Technology Stack

### Core Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/gin-gonic/gin` | v1.10.0 | Web framework |
| `github.com/lib/pq` | v1.10.9 | PostgreSQL driver |
| `github.com/golang-jwt/jwt/v5` | v5.2.1 | JWT token handling |
| `github.com/joho/godotenv` | v1.5.1 | Environment variables |
| `golang.org/x/crypto` | v0.31.0 | Password hashing (bcrypt) |

### Testing Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/stretchr/testify` | v1.10.0 | Testing assertions |
| `github.com/DATA-DOG/go-sqlmock` | v1.5.2 | Database mocking |

### Development Tools

| Tool | Purpose |
|------|---------|
| Air | Hot-reload development server |
| Docker Compose | PostgreSQL local instance |
| Make | Build automation |
| Go 1.23+ | Programming language |

---

## Core Components

### 1. Main Application (`cmd/api/main.go`)

**Purpose**: Application entry point, server initialization.

**Responsibilities**:
- Load configuration from environment variables
- Initialize database connection
- Create dependencies (repositories, services, handlers)
- Set up routes and middleware
- Start HTTP server

**Key Code**:
```go
func main() {
    cfg := config.Load()
    db := database.Connect(cfg.DatabaseURL)
    defer db.Close()
    
    // Initialize layers
    userRepo := repository.NewUserRepository(db)
    authService := service.NewAuthService(userRepo, cfg.JWTSecret)
    authHandler := handlers.NewAuthHandler(authService)
    
    // Setup routes
    router := gin.Default()
    setupRoutes(router, authHandler, cfg)
    
    router.Run(":" + cfg.ServerPort)
}
```

### 2. Configuration (`config/config.go`)

**Purpose**: Load and validate configuration from environment variables.

**Environment Variables**:
- `SERVER_PORT`: HTTP server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT signing (required)

**Key Features**:
- Loads from `.env` file in development
- Validates required variables
- Provides sensible defaults
- Fails fast if critical config missing

### 3. Database Connection (`pkg/database/database.go`)

**Purpose**: PostgreSQL connection management and table creation.

**Features**:
- Connection pooling with `database/sql`
- Automatic table creation (users table)
- Connection validation with Ping()
- Graceful error handling

**Users Table Schema**:
```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
```

### 4. User Repository (`internal/repository/user_repository.go`)

**Purpose**: Data access layer for user operations.

**Interface**:
```go
type UserRepository interface {
    Create(user *models.User) error
    FindByEmail(email string) (*models.User, error)
    FindByID(id int) (*models.User, error)
}
```

**Methods**:
- `Create`: Insert new user with hashed password
- `FindByEmail`: Query user by email (for login)
- `FindByID`: Query user by ID (for profile)

**Testing**: Uses `sqlmock` to mock database interactions (100% coverage).

### 5. Auth Service (`internal/service/auth_service.go`)

**Purpose**: Business logic for authentication operations.

**Methods**:
- `Register(req)`: Create new user account
  - Validates email format
  - Checks email uniqueness
  - Hashes password
  - Creates user record
- `Login(req)`: Authenticate user and generate JWT
  - Finds user by email
  - Validates password
  - Generates JWT token
- `GetUserProfile(userID)`: Retrieve user profile data

**Business Rules**:
- Email must be valid format
- Email must be unique
- Password must meet complexity requirements (future)
- JWT tokens expire after 24 hours

**Testing**: Uses repository mocks (76% coverage).

### 6. Auth Handler (`internal/handlers/auth_handler.go`)

**Purpose**: HTTP request/response handling for authentication endpoints.

**Endpoints**:
- `POST /api/v1/auth/register`: User registration
- `POST /api/v1/auth/login`: User login
- `GET /api/v1/auth/profile`: Get user profile (protected)

**Responsibilities**:
- Parse and validate request JSON
- Call service layer methods
- Format response JSON
- Set appropriate HTTP status codes
- Handle errors gracefully

**Error Responses**:
```json
{
  "error": "error message here"
}
```

**Testing**: Uses service mocks (88.2% coverage).

### 7. Auth Middleware (`internal/middleware/auth.go`)

**Purpose**: Protect routes by validating JWT tokens.

**Flow**:
1. Extract token from `Authorization: Bearer <token>` header
2. Validate token signature and expiration
3. Extract user ID from token claims
4. Set `userID` in Gin context
5. Allow request to proceed OR return 401 Unauthorized

**Usage**:
```go
protected := router.Group("/api/v1/auth")
protected.Use(middleware.AuthMiddleware(jwtSecret))
protected.GET("/profile", handler.GetProfile)
```

**Testing**: Comprehensive tests for all scenarios (100% coverage).

### 8. JWT Utils (`pkg/utils/jwt.go`)

**Purpose**: JWT token generation and validation.

**Functions**:
- `GenerateToken(userID, secret)`: Create signed JWT token
- `ValidateToken(token, secret)`: Verify and parse JWT token

**Token Structure**:
```json
{
  "user_id": 123,
  "exp": 1234567890
}
```

**Algorithm**: HS256 (HMAC with SHA-256)  
**Expiration**: 24 hours from generation

**Testing**: Tests generation, validation, expiration (87.5% coverage).

### 9. Password Utils (`pkg/utils/password.go`)

**Purpose**: Password hashing and verification.

**Functions**:
- `HashPassword(password)`: Hash password with bcrypt
- `CheckPassword(password, hash)`: Verify password against hash

**Security**:
- Uses bcrypt default cost (10)
- Automatic salting
- Resistant to timing attacks

**Testing**: Tests hashing, verification, edge cases (87.5% coverage).

### 10. Models (`internal/models/user.go`)

**Purpose**: Data structures and DTOs.

**Types**:
```go
type User struct {
    ID           int       `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`
    Name         string    `json:"name"`
    CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}
```

**Validation Rules**:
- Email: required, valid email format
- Password: required, minimum 6 characters
- Name: required

---

## Testing Strategy

### Overview
- **Total Tests**: 53 unit tests
- **Total Lines**: 1,239 lines of test code
- **Overall Coverage**: 63.6%
- **Testing Framework**: Go's built-in `testing` package
- **Assertion Library**: `testify/assert`
- **Mocking**: `sqlmock` for database, custom mocks for services

### Coverage by Component

| Component | Coverage | Tests | Status |
|-----------|----------|-------|--------|
| Middleware | 100% | 9 tests | ✅ Excellent |
| Repository | 100% | 9 tests | ✅ Excellent |
| Handlers | 88.2% | 12 tests | ✅ Good |
| Utils (JWT) | 87.5% | 7 tests | ✅ Good |
| Utils (Password) | 87.5% | 4 tests | ✅ Good |
| Service | 76.0% | 12 tests | ⚠️ Could improve |
| **Overall** | **63.6%** | **53 tests** | ✅ Good |

### Testing Approach

#### Unit Tests
- Test each component in isolation
- Use mocks to avoid external dependencies
- Cover happy paths and error cases
- Test edge cases and boundary conditions

#### Database Mocking
Uses `go-sqlmock` to mock database queries:
```go
db, mock, _ := sqlmock.New()
mock.ExpectQuery("SELECT (.+) FROM users WHERE email = ?").
    WithArgs("test@example.com").
    WillReturnRows(sqlmock.NewRows(columns).AddRow(values...))
```

#### HTTP Handler Testing
Uses `httptest` to test HTTP handlers:
```go
w := httptest.NewRecorder()
c, _ := gin.CreateTestContext(w)
c.Request = httptest.NewRequest("POST", "/register", body)
handler.Register(c)
assert.Equal(t, http.StatusCreated, w.Code)
```

#### Test Organization
Each component has its own `_test.go` file:
- `auth_handler_test.go`: Handler tests
- `auth_service_test.go`: Service tests
- `user_repository_test.go`: Repository tests
- `auth_test.go`: Middleware tests
- `jwt_test.go`: JWT utility tests
- `password_test.go`: Password utility tests

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with verbose output
make test-verbose

# Run specific package tests
go test ./internal/handlers/... -v
```

### Test Quality Metrics
- ✅ All tests pass
- ✅ Race detector enabled (`-race` flag)
- ✅ Tests run in parallel where possible
- ✅ No flaky tests
- ✅ Fast execution (< 2 seconds total)

---

## Development Workflow

### Initial Setup

```bash
# 1. Clone repository
git clone <repo-url>
cd mikrom-go

# 2. Install dependencies
make deps

# 3. Start PostgreSQL
make docker-up

# 4. Configure environment
cp .env.example .env
# Edit .env with your settings

# 5. Run application
make run
```

### Daily Development

```bash
# Start development with hot-reload
make dev

# Run tests (recommended before committing)
make test

# Check coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Build for production
make build
```

### Common Workflows

#### Adding a New Endpoint

1. **Define Model** (if needed) in `internal/models/`
2. **Add Repository Method** in `internal/repository/`
3. **Write Repository Tests** in `internal/repository/*_test.go`
4. **Add Service Method** in `internal/service/`
5. **Write Service Tests** in `internal/service/*_test.go`
6. **Add Handler Method** in `internal/handlers/`
7. **Write Handler Tests** in `internal/handlers/*_test.go`
8. **Register Route** in `cmd/api/main.go`
9. **Test Manually** with curl or Postman
10. **Update Documentation** in `docs/`

#### Adding Middleware

1. **Create Middleware** in `internal/middleware/`
2. **Write Tests** in `internal/middleware/*_test.go`
3. **Apply to Routes** in `cmd/api/main.go`
4. **Document Behavior** in `docs/`

#### Making Database Changes

1. **Update Schema** in `pkg/database/database.go`
2. **Update Model** in `internal/models/`
3. **Update Repository** in `internal/repository/`
4. **Update Tests** with new schema
5. **Run Migration** (currently manual, future: migrate tool)

### Git Workflow

```bash
# Check status
git status

# Create feature branch
git checkout -b feature/new-feature

# Make changes and test
make test

# Commit changes
git add .
git commit -m "feat: add new feature"

# Push to remote
git push origin feature/new-feature

# Create pull request (use GitHub/GitLab UI)
```

### Code Quality Checklist

Before committing:
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No linter errors (`make lint`)
- [ ] Coverage maintained or improved
- [ ] Documentation updated
- [ ] No sensitive data in code
- [ ] Environment variables in `.env.example`

---

## API Specification

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
Protected endpoints require JWT token in Authorization header:
```
Authorization: Bearer <jwt-token>
```

### Endpoints

#### 1. Register User

**POST** `/api/v1/auth/register`

Create a new user account.

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe"
}
```

**Validation**:
- `email`: required, valid email format
- `password`: required, minimum 6 characters
- `name`: required

**Success Response** (201 Created):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid input or email already exists
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123",
    "name": "John Doe"
  }'
```

#### 2. Login

**POST** `/api/v1/auth/login`

Authenticate user and receive JWT token.

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Validation**:
- `email`: required, valid email format
- `password`: required

**Success Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Invalid credentials
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

#### 3. Get Profile (Protected)

**GET** `/api/v1/auth/profile`

Get current user's profile information.

**Headers**:
```
Authorization: Bearer <jwt-token>
```

**Success Response** (200 OK):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses**:
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: User not found
- `500 Internal Server Error`: Server error

**Example**:
```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 4. Health Check

**GET** `/health`

Check if API is running.

**Success Response** (200 OK):
```json
{
  "status": "ok"
}
```

**Example**:
```bash
curl http://localhost:8080/health
```

### Error Response Format

All errors follow this format:
```json
{
  "error": "descriptive error message"
}
```

### HTTP Status Codes

- `200 OK`: Request succeeded
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid input or validation error
- `401 Unauthorized`: Authentication failed or missing
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

---

## Database Schema

### Users Table

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Columns**:
- `id`: Auto-incrementing primary key
- `email`: Unique user email (used for login)
- `password_hash`: bcrypt-hashed password (never stored in plain text)
- `name`: User's display name
- `created_at`: Account creation timestamp

**Indexes**:
- Primary key on `id`
- Unique index on `email`

**Constraints**:
- `email` must be unique
- All fields are required (NOT NULL)

### Future Tables (Planned)

#### Refresh Tokens
```sql
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### Password Reset Tokens
```sql
CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Configuration & Environment

### Environment Variables

Required variables (must be set):
```bash
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
DATABASE_URL=postgres://postgres:postgres@localhost:5432/mikrom?sslmode=disable
```

Optional variables (have defaults):
```bash
SERVER_PORT=8080
```

### `.env.example`

Template file for environment variables:
```bash
# Server Configuration
SERVER_PORT=8080

# Database Configuration
DATABASE_URL=postgres://postgres:postgres@localhost:5432/mikrom?sslmode=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

### Docker Compose Configuration

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    container_name: mikrom-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: mikrom
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
volumes:
  postgres_data:
```

### Air Configuration (`.air.toml`)

Hot-reload configuration for development:
```toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/api"
  bin = "tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  delay = 1000
```

---

## Development History

### Initial Development Session

**Date**: 2024 (Initial commit: 0d0466e)

**Goals**:
1. Build a secure authentication API
2. Follow Go best practices
3. Achieve high test coverage
4. Create comprehensive documentation

**What Was Built**:

1. **Project Structure** ✅
   - Set up clean architecture with handlers, services, repositories
   - Created proper Go module structure
   - Organized code into `internal/` and `pkg/` directories

2. **Authentication System** ✅
   - User registration with email/password
   - User login with JWT tokens
   - Protected endpoints with middleware
   - bcrypt password hashing

3. **Database Layer** ✅
   - PostgreSQL connection with `database/sql`
   - User repository with CRUD operations
   - Raw SQL queries (no ORM)
   - Automatic table creation

4. **Testing Suite** ✅
   - 53 unit tests across all components
   - 63.6% overall coverage
   - Database mocking with `sqlmock`
   - HTTP handler testing with `httptest`

5. **Developer Tools** ✅
   - Makefile with 35+ commands
   - Docker Compose for PostgreSQL
   - Hot-reload with Air
   - Environment variable management

6. **Documentation** ✅
   - README.md (quick start)
   - docs/INDEX.md (documentation index)
   - docs/README.md (full documentation)
   - docs/TESTING.md (testing guide)
   - All documentation in English

### Key Decisions Made

1. **Go 1.23+**: Modern Go version with latest features
2. **Gin Framework**: Fast, popular, feature-rich
3. **Raw SQL**: Better performance and control than ORM
4. **JWT Tokens**: Stateless authentication, 24h expiration
5. **bcrypt**: Industry-standard password hashing
6. **English Docs**: International collaboration
7. **Clean Architecture**: Maintainable, testable code
8. **High Test Coverage**: 60%+ coverage target

### Files Created

**Total**: 27 files, 3,113 lines of code

**Configuration**:
- `.air.toml` - Hot-reload config
- `.env.example` - Environment template
- `.gitignore` - Git ignore rules
- `docker-compose.yml` - PostgreSQL setup
- `Makefile` - Build automation
- `go.mod`, `go.sum` - Go dependencies

**Application Code**:
- `cmd/api/main.go` - Entry point
- `config/config.go` - Configuration
- `pkg/database/database.go` - Database connection
- `pkg/utils/jwt.go`, `password.go` - Utilities
- `internal/models/user.go` - Data models
- `internal/repository/user_repository.go` - Data access
- `internal/service/auth_service.go` - Business logic
- `internal/handlers/auth_handler.go` - HTTP handlers
- `internal/middleware/auth.go` - JWT middleware

**Tests**:
- `internal/handlers/auth_handler_test.go`
- `internal/middleware/auth_test.go`
- `internal/repository/user_repository_test.go`
- `internal/service/auth_service_test.go`
- `pkg/utils/jwt_test.go`
- `pkg/utils/password_test.go`

**Documentation**:
- `README.md` - Quick start
- `docs/INDEX.md` - Doc index
- `docs/README.md` - Full docs
- `docs/TESTING.md` - Testing guide
- `CLAUDE.md` - This file

---

## Known Limitations & Future Work

### Current Limitations

1. **No Refresh Tokens**: JWT tokens expire after 24h, no way to refresh without re-login
2. **No Password Reset**: Users cannot reset forgotten passwords
3. **No Email Verification**: Email addresses are not verified
4. **No Rate Limiting**: API is vulnerable to brute-force attacks
5. **No RBAC**: All authenticated users have same permissions
6. **No Pagination**: Future list endpoints will need pagination
7. **No Logging**: No structured logging (only Gin's default logs)
8. **No Metrics**: No Prometheus/monitoring integration
9. **Manual Migrations**: Database schema changes are manual
10. **No CI/CD**: No automated testing/deployment pipeline

### Planned Features

#### High Priority
1. **Refresh Tokens**
   - Add refresh token table
   - Implement token refresh endpoint
   - Rotate refresh tokens on use

2. **Password Reset Flow**
   - Generate reset tokens
   - Send reset emails
   - Verify tokens and update passwords

3. **Email Verification**
   - Send verification emails on registration
   - Verify email with token
   - Require verified email for sensitive operations

4. **Rate Limiting**
   - Implement rate limiting middleware
   - Different limits for different endpoints
   - Redis for distributed rate limiting

5. **Logging & Monitoring**
   - Structured logging with zerolog or zap
   - Request ID tracking
   - Error tracking (Sentry)
   - Metrics (Prometheus)

#### Medium Priority
6. **Role-Based Access Control (RBAC)**
   - Add roles table (admin, user, etc.)
   - Add permissions system
   - Middleware for role checking

7. **Enhanced User Management**
   - Update profile endpoint
   - Change password endpoint
   - Delete account endpoint
   - Upload profile picture

8. **API Documentation**
   - Generate OpenAPI/Swagger spec
   - Interactive API documentation
   - Postman collection

9. **CI/CD Pipeline**
   - GitHub Actions for testing
   - Automated coverage reports
   - Docker image building
   - Deployment automation

10. **Database Migrations**
    - Use golang-migrate or similar
    - Version-controlled schema changes
    - Rollback support

#### Low Priority
11. **OAuth 2.0 Integration**
    - Login with Google
    - Login with GitHub
    - Social account linking

12. **Two-Factor Authentication (2FA)**
    - TOTP support
    - Backup codes
    - SMS verification

13. **Audit Logging**
    - Log all user actions
    - Admin access logs
    - Security event tracking

14. **Advanced Features**
    - User sessions management
    - Device tracking
    - Location-based security
    - Suspicious activity detection

### Technical Debt

1. **Increase Service Coverage**: Currently 76%, target 80%+
2. **Integration Tests**: Add tests with real database
3. **E2E Tests**: Full API workflow tests
4. **Linter Configuration**: Add golangci-lint config
5. **Pre-commit Hooks**: Automated checks before commit
6. **Security Scanning**: Gosec or similar tools
7. **Dependency Updates**: Dependabot or Renovate
8. **Performance Testing**: Load tests, benchmarks

---

## Common Tasks

### Adding a New Model

```go
// internal/models/product.go
package models

import "time"

type Product struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    CreatedAt   time.Time `json:"created_at"`
}

type CreateProductRequest struct {
    Name        string  `json:"name" binding:"required"`
    Description string  `json:"description"`
    Price       float64 `json:"price" binding:"required,gt=0"`
}
```

### Creating a Repository Interface

```go
// internal/repository/product_repository.go
package repository

import (
    "database/sql"
    "mikrom-go/internal/models"
)

type ProductRepository interface {
    Create(product *models.Product) error
    FindByID(id int) (*models.Product, error)
    FindAll() ([]*models.Product, error)
    Update(product *models.Product) error
    Delete(id int) error
}

type productRepository struct {
    db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) Create(product *models.Product) error {
    query := `
        INSERT INTO products (name, description, price)
        VALUES ($1, $2, $3)
        RETURNING id, created_at
    `
    return r.db.QueryRow(query, product.Name, product.Description, product.Price).
        Scan(&product.ID, &product.CreatedAt)
}
```

### Writing Repository Tests

```go
// internal/repository/product_repository_test.go
package repository

import (
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
    "mikrom-go/internal/models"
)

func TestCreateProduct(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    repo := NewProductRepository(db)
    
    product := &models.Product{
        Name:        "Test Product",
        Description: "Description",
        Price:       99.99,
    }

    rows := sqlmock.NewRows([]string{"id", "created_at"}).
        AddRow(1, time.Now())
    
    mock.ExpectQuery("INSERT INTO products").
        WithArgs(product.Name, product.Description, product.Price).
        WillReturnRows(rows)

    err = repo.Create(product)
    assert.NoError(t, err)
    assert.Equal(t, 1, product.ID)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

### Creating a Service

```go
// internal/service/product_service.go
package service

import (
    "mikrom-go/internal/models"
    "mikrom-go/internal/repository"
)

type ProductService interface {
    CreateProduct(req *models.CreateProductRequest) (*models.Product, error)
    GetProduct(id int) (*models.Product, error)
    ListProducts() ([]*models.Product, error)
}

type productService struct {
    repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
    return &productService{repo: repo}
}

func (s *productService) CreateProduct(req *models.CreateProductRequest) (*models.Product, error) {
    product := &models.Product{
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
    }
    
    if err := s.repo.Create(product); err != nil {
        return nil, err
    }
    
    return product, nil
}
```

### Creating a Handler

```go
// internal/handlers/product_handler.go
package handlers

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "mikrom-go/internal/models"
    "mikrom-go/internal/service"
)

type ProductHandler struct {
    service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
    return &ProductHandler{service: service}
}

func (h *ProductHandler) Create(c *gin.Context) {
    var req models.CreateProductRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    product, err := h.service.CreateProduct(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
        return
    }
    
    product, err := h.service.GetProduct(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
        return
    }
    
    c.JSON(http.StatusOK, product)
}
```

### Registering Routes

```go
// cmd/api/main.go
func setupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, productHandler *handlers.ProductHandler, cfg *config.Config) {
    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // API v1
    v1 := router.Group("/api/v1")
    
    // Auth routes (public)
    auth := v1.Group("/auth")
    {
        auth.POST("/register", authHandler.Register)
        auth.POST("/login", authHandler.Login)
        
        // Protected auth routes
        authProtected := auth.Group("")
        authProtected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
        authProtected.GET("/profile", authHandler.GetProfile)
    }
    
    // Product routes (protected)
    products := v1.Group("/products")
    products.Use(middleware.AuthMiddleware(cfg.JWTSecret))
    {
        products.POST("", productHandler.Create)
        products.GET("/:id", productHandler.GetByID)
        products.GET("", productHandler.List)
        products.PUT("/:id", productHandler.Update)
        products.DELETE("/:id", productHandler.Delete)
    }
}
```

### Testing with curl

```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'

# Login
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

# Get profile (protected)
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"

# Create product (protected)
curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Product 1","description":"Test","price":99.99}'
```

---

## Quick Reference

### Important Commands

```bash
# Development
make dev              # Run with hot-reload
make run              # Run normally
make test             # Run tests
make test-coverage    # Coverage report

# Database
make docker-up        # Start PostgreSQL
make docker-down      # Stop PostgreSQL
make docker-logs      # View logs

# Building
make build            # Build binary
make clean            # Clean artifacts

# Code Quality
make fmt              # Format code
make lint             # Run linter
make vet              # Run go vet

# Help
make help             # Show all commands
```

### Important Files

- `.env` - Environment configuration (create from `.env.example`)
- `Makefile` - All automation commands
- `cmd/api/main.go` - Application entry point
- `docs/README.md` - Full documentation

### Important URLs

- API Base: `http://localhost:8080/api/v1`
- Health Check: `http://localhost:8080/health`
- PostgreSQL: `localhost:5432`

### Code Style Guidelines

1. **Naming**:
   - Use camelCase for variables/functions
   - Use PascalCase for types/interfaces
   - Use UPPER_CASE for constants

2. **Error Handling**:
   - Always check errors
   - Return errors, don't panic
   - Wrap errors with context

3. **Comments**:
   - Public functions must have comments
   - Explain "why", not "what"
   - Keep comments up-to-date

4. **Testing**:
   - Test file name: `*_test.go`
   - Test function: `TestFunctionName`
   - Use table-driven tests when appropriate

5. **Project Structure**:
   - `internal/` for private code
   - `pkg/` for public libraries
   - `cmd/` for executables

---

## Conclusion

This document serves as a comprehensive knowledge base for the Mikrom Go API project. It should be updated whenever:

1. New features are added
2. Architecture decisions are made
3. Major refactoring occurs
4. Dependencies are updated
5. Development processes change

For questions or improvements, refer to the issue tracker or contact the development team.

**Last Updated**: 2024 (Initial creation)  
**Version**: 1.0.0  
**Status**: Active Development
