# Mikrom Go API - Complete Documentation

> [← Back to main README](../README.md) | [View Testing documentation](TESTING.md)

REST API built with Gin and PostgreSQL for user authentication.

## 📋 Features

- ✅ User registration
- ✅ JWT login
- ✅ Token-based authentication
- ✅ Bcrypt password hashing
- ✅ PostgreSQL persistence
- ✅ Clean architecture (handlers, services, repositories)

## 🚀 Endpoints

### Public

#### Register User
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

#### Login
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

### Protected (require authentication)

#### Get Profile
```bash
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

#### Health Check
```bash
GET /health
```

## 🛠️ Installation

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Docker and Docker Compose (optional)

### Configuration

**Quick setup with Makefile:**

```bash
# 1. Clone the repository
git clone <repo-url>
cd mikrom-go

# 2. Initial setup (installs deps and starts DB)
make setup

# 3. Configure environment variables
cp .env.example .env
# Edit .env with your configurations

# 4. Run the application
make run

# Or with hot-reload
make dev
```

**Manual configuration:**

1. **Clone the repository**
```bash
git clone <repo-url>
cd mikrom-go
```

2. **Configure environment variables**
```bash
cp .env.example .env
```

Edit `.env` with your configurations:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=mikrom
SERVER_PORT=8080
JWT_SECRET=your-secure-secret-key
```

3. **Start PostgreSQL with Docker (optional)**
```bash
make docker-up
# Or manually:
docker-compose up -d
```

Or configure your own PostgreSQL instance and create the database:
```sql
CREATE DATABASE mikrom;
```

4. **Install dependencies**
```bash
make install
# Or manually:
go mod download
```

5. **Run the application**
```bash
make run
# Or manually:
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

## 🔧 Available Commands (Makefile)

```bash
make help              # Show all available commands
make run               # Run the application
make dev               # Run with hot-reload (requires air)
make build             # Build for production
make test              # Run tests
make test-coverage     # Tests with coverage report
make lint              # Run linter
make fmt               # Format code
make check             # Run all checks
make docker-up         # Start PostgreSQL
make docker-down       # Stop PostgreSQL
make clean             # Clean generated files
make setup             # Complete initial setup
```

To see all available commands run:
```bash
make help
```

## 📁 Project Structure

```
mikrom-go/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── config/
│   └── config.go                # Configuration
├── internal/
│   ├── handlers/
│   │   └── auth_handler.go      # HTTP controllers
│   ├── middleware/
│   │   └── auth.go              # Authentication middleware
│   ├── models/
│   │   └── user.go              # Data models
│   ├── repository/
│   │   └── user_repository.go   # Data access layer
│   └── service/
│       └── auth_service.go      # Business logic
├── pkg/
│   ├── database/
│   │   └── database.go          # DB connection
│   └── utils/
│       ├── jwt.go               # JWT utilities
│       └── password.go          # Password utilities
├── .env.example
├── docker-compose.yml
├── go.mod
└── go.sum
```

## 🧪 Usage Example

### 1. Register a user
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "test@example.com",
    "name": "Test User",
    "created_at": "2024-02-03T00:00:00Z",
    "updated_at": "2024-02-03T00:00:00Z"
  }
}
```

### 3. Get profile (authenticated)
```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer <your-token-here>"
```

## 🧪 Testing

The project includes **53 unit tests** distributed across **6 files** with **1,239 lines of testing code**.

### Running Tests

```bash
# Run all tests
make test

# Run tests with detailed output
make test-verbose

# Run tests with coverage report
make test-coverage

# Generate HTML coverage report
make coverage-html
```

### Current Coverage

- **Handlers**: 88.2%
- **Middleware**: 100%
- **Repository**: 100%
- **Service**: 76.0%
- **Utils**: 87.5%
- **Total**: 63.6%

### Test Structure

```
mikrom-go/
├── internal/
│   ├── handlers/auth_handler_test.go
│   ├── middleware/auth_test.go
│   ├── repository/user_repository_test.go
│   └── service/auth_service_test.go
└── pkg/
    └── utils/
        ├── jwt_test.go
        └── password_test.go
```

## 🔒 Security

- Passwords are hashed with bcrypt before being stored
- JWT tokens expire in 24 hours
- Protected endpoints require a valid JWT token
- **IMPORTANT**: Change `JWT_SECRET` in production

## 🛠️ Technologies

- **Gin** - HTTP web framework
- **PostgreSQL** - Database
- **JWT** - Authentication
- **bcrypt** - Password hashing
- **godotenv** - Environment variables

## 📝 License

MIT
