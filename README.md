# Mikrom Go API

REST API built with Gin and PostgreSQL for user authentication.

## 🚀 Quick Start

```bash
# 1. Clone the repository
git clone <repo-url>
cd mikrom-go

# 2. Initial setup
make setup

# 3. Configure environment variables
cp .env.example .env

# 4. Run the application
make run
```

The API will be available at `http://localhost:8080`

## 📋 Features

- ✅ User registration
- ✅ JWT authentication
- ✅ Token-based authentication
- ✅ Bcrypt password hashing
- ✅ PostgreSQL persistence
- ✅ Clean architecture (handlers, services, repositories)
- ✅ 53 unit tests with 63.6% coverage

## 🚀 Endpoints

### Public
- `POST /api/v1/auth/register` - Register user
- `POST /api/v1/auth/login` - Login

### Protected (requires JWT)
- `GET /api/v1/auth/profile` - Get user profile

### Health Check
- `GET /health` - Server status

## 🔧 Main Commands

```bash
make help              # Show all available commands
make run               # Run the application
make dev               # Run with hot-reload
make test              # Run tests
make test-coverage     # Tests with coverage
make build             # Build for production
make docker-up         # Start PostgreSQL
```

## 📚 Documentation

- [**📑 Documentation Index**](docs/INDEX.md) - Complete documentation index
- [**📖 Full Documentation**](docs/README.md) - Complete installation, usage and examples guide
- [**🧪 Testing**](docs/TESTING.md) - Testing documentation, coverage and best practices

## 📁 Project Structure

```
mikrom-go/
├── cmd/api/              # Application entry point
├── config/               # Configuration
├── internal/
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # Middleware (auth)
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   └── service/          # Business logic
├── pkg/
│   ├── database/         # PostgreSQL connection
│   └── utils/            # JWT and password hashing
├── docs/                 # Documentation
├── Makefile              # Automated commands
└── docker-compose.yml    # PostgreSQL in Docker
```

## 🧪 Testing

```bash
# Run tests
make test

# Tests with coverage
make test-coverage

# HTML report
make coverage-html
```

**Coverage:** 63.6% total (53 tests)
- Middleware: 100%
- Repository: 100%
- Handlers: 88.2%
- Utils: 87.5%
- Service: 76.0%

## 🛠️ Technologies

- **Gin** - HTTP web framework
- **PostgreSQL** - Database
- **JWT** - Authentication
- **bcrypt** - Password hashing
- **testify** - Testing
- **sqlmock** - Database mocking

## 📝 License

MIT
