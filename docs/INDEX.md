# Documentation Index

> [← Back to main README](../README.md)

## 📚 Available Documentation

### 📖 Main Guides

- **[README.md](README.md)** - Complete project documentation
  - Installation and configuration
  - API endpoints
  - Usage examples
  - Project structure
  - Technologies used

- **[TESTING.md](TESTING.md)** - Testing documentation
  - Test coverage
  - Tests by package
  - How to run tests
  - Best practices
  - Quality metrics

## 🔍 Quick Navigation

### By Topic

#### Quick Start
1. [Installation](README.md#installation)
2. [Configuration](README.md#configuration)
3. [First use](README.md#usage-example)

#### API
1. [Available endpoints](README.md#endpoints)
2. [User registration](README.md#register-user)
3. [Login](README.md#login)
4. [Protected profile](README.md#get-profile)

#### Testing
1. [Run tests](TESTING.md#running-tests)
2. [Coverage](TESTING.md#current-coverage)
3. [Implemented tests](TESTING.md#implemented-tests)
4. [Adding new tests](TESTING.md#adding-new-tests)

#### Development
1. [Project structure](README.md#project-structure)
2. [Make commands](README.md#available-commands-makefile)
3. [Hot-reload](README.md#configuration)

## 📊 Project Statistics

- **Endpoints**: 4 (3 auth + 1 health)
- **Tests**: 53 unit tests
- **Coverage**: 63.6%
- **Test lines**: 1,239
- **Code files**: 16
- **Make commands**: 35+

## 🛠️ Useful Commands

```bash
# Show help
make help

# Development
make dev              # Run with hot-reload
make run              # Run application

# Testing
make test             # Run tests
make test-coverage    # Tests with coverage

# Docker
make docker-up        # Start PostgreSQL
make docker-down      # Stop PostgreSQL

# Utilities
make clean            # Clean generated files
make build            # Build for production
```

## 🤝 Contributing

To contribute to the project:

1. Read the [complete documentation](README.md)
2. Review the [testing guide](TESTING.md)
3. Make sure `make check` passes
4. Maintain test coverage > 60%

## 📞 Support

- Issues: [GitHub Issues](../../issues)
- Documentation: This directory
- Examples: [README.md](README.md#usage-example)
