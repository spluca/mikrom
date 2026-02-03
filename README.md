# Mikrom-Go - Firecracker VM Management API

A high-performance REST API for managing Firecracker microVMs built with Go, featuring asynchronous task processing, IP pool management, and Ansible-based VM provisioning.

## 🚀 Features

- **User Authentication** - JWT-based authentication
- **VM Lifecycle Management** - Create, start, stop, restart, delete VMs
- **Asynchronous Operations** - Background workers with Redis/asynq
- **IP Pool Management** - Automatic IP allocation from configurable pools
- **Ansible Integration** - VM provisioning via Ansible playbooks
- **RESTful API** - Clean, well-documented API endpoints
- **Database Migrations** - GORM-based auto-migrations
- **Docker Support** - Complete docker-compose setup

## 📋 Prerequisites

- Go 1.21 or later
- PostgreSQL 15+
- Redis 7+
- Ansible 2.10+ (for VM provisioning)
- Docker & Docker Compose (optional)

## 🛠️ Installation

### 1. Clone the repository

```bash
git clone https://github.com/apardo/mikrom-go.git
cd mikrom-go
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Configure environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 4. Start dependencies with Docker

```bash
docker compose up -d postgres redis
```

### 5. Initialize the database

Run the seeder to create the default IP pool:

```bash
go run cmd/seed/main.go
```

Or use the shell script:

```bash
./scripts/init-ippool.sh
```

## 🎯 Running the Application

### Start the API server

```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

### Start the worker (in separate terminal)

```bash
go run cmd/worker/main.go
```

The worker processes background tasks for VM operations.

## 📡 API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login and get JWT token
- `GET /api/v1/auth/profile` - Get user profile (authenticated)

### Virtual Machines

All VM endpoints require authentication (Bearer token).

- `POST /api/v1/vms` - Create a new VM
- `GET /api/v1/vms` - List all VMs (with pagination)
- `GET /api/v1/vms/:vm_id` - Get VM details
- `PATCH /api/v1/vms/:vm_id` - Update VM metadata
- `DELETE /api/v1/vms/:vm_id` - Delete VM
- `POST /api/v1/vms/:vm_id/start` - Start VM
- `POST /api/v1/vms/:vm_id/stop` - Stop VM
- `POST /api/v1/vms/:vm_id/restart` - Restart VM

### IP Pools

All IP pool endpoints require authentication.

- `POST /api/v1/ippools` - Create a new IP pool
- `GET /api/v1/ippools` - List all IP pools
- `GET /api/v1/ippools/:id` - Get IP pool details
- `PATCH /api/v1/ippools/:id` - Update IP pool
- `DELETE /api/v1/ippools/:id` - Delete IP pool
- `GET /api/v1/ippools/stats` - Get all pools statistics
- `GET /api/v1/ippools/:id/stats` - Get pool statistics
- `POST /api/v1/ippools/suggest-range` - Suggest IP range for CIDR

### Health Check

- `GET /health` - API health check

## 📝 Usage Examples

### 1. Register a user

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword",
    "name": "John Doe"
  }'
```

### 2. Login

```bash
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }' | jq -r '.token')
```

### 3. Create a VM

```bash
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-vm",
    "vcpu_count": 2,
    "memory_mb": 1024,
    "description": "My test VM",
    "kernel_path": "/path/to/kernel",
    "rootfs_path": "/path/to/rootfs"
  }'
```

### 4. List VMs

```bash
curl -X GET "http://localhost:8080/api/v1/vms?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
```

### 5. Create an IP Pool

```bash
curl -X POST http://localhost:8080/api/v1/ippools \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production",
    "network": "192.168.1.0",
    "cidr": "192.168.1.0/24",
    "gateway": "192.168.1.1",
    "start_ip": "192.168.1.10",
    "end_ip": "192.168.1.254"
  }'
```

### 6. Get Pool Statistics

```bash
curl -X GET http://localhost:8080/api/v1/ippools/stats \
  -H "Authorization: Bearer $TOKEN"
```

## 🏗️ Architecture

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│   Client    │ ──────> │  Gin API    │ ──────> │ PostgreSQL  │
│   (HTTP)    │ <────── │   (GORM)    │ <────── │   + GORM    │
└─────────────┘         └─────────────┘         └─────────────┘
                              │
                              │ Enqueue Task
                              ▼
                        ┌─────────────┐         ┌─────────────┐
                        │   Redis     │ ──────> │   asynq     │
                        │   Queue     │ <────── │  Workers    │
                        └─────────────┘         └─────────────┘
                                                      │
                                                      │ Execute
                                                      ▼
                                                ┌─────────────┐
                                                │  Ansible    │
                                                │  Playbooks  │
                                                └─────────────┘
                                                      │
                                                      ▼
                                                ┌─────────────┐
                                                │ Firecracker │
                                                │    VMs      │
                                                └─────────────┘
```

## 🔧 Configuration

Environment variables (`.env`):

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=mikrom

# Server
SERVER_PORT=8080

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Firecracker
FIRECRACKER_DEPLOY_PATH=/path/to/firecracker-deploy
FIRECRACKER_DEFAULT_HOST=your-firecracker-host

# Worker
WORKER_CONCURRENCY=10
```

## 📁 Project Structure

```
mikrom-go/
├── cmd/
│   ├── api/           # API server entry point
│   ├── worker/        # Background worker entry point
│   └── seed/          # Database seeder
├── config/            # Configuration management
├── internal/
│   ├── handlers/      # HTTP request handlers
│   ├── middleware/    # HTTP middleware
│   ├── models/        # Data models
│   ├── repository/    # Database repositories
│   └── service/       # Business logic
├── pkg/
│   ├── database/      # Database connection
│   ├── firecracker/   # Firecracker/Ansible client
│   ├── utils/         # Utility functions
│   └── worker/        # Task queue (asynq)
└── scripts/           # Helper scripts
```

## 🧪 Testing

Run tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## 🐳 Docker Deployment

Start all services:

```bash
docker compose up -d
```

Stop all services:

```bash
docker compose down
```

View logs:

```bash
docker compose logs -f
```

## 📚 Development

### Build

```bash
go build -o bin/api cmd/api/main.go
go build -o bin/worker cmd/worker/main.go
go build -o bin/seed cmd/seed/main.go
```

### Run

```bash
./bin/api
./bin/worker
./bin/seed
```

## 🔐 Security

- JWT tokens for authentication
- Password hashing with bcrypt
- Input validation on all endpoints
- SQL injection prevention via GORM
- CORS middleware ready

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 👥 Authors

- Antonio Pardo - [@apardo](https://github.com/apardo)

## 🙏 Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [GORM](https://gorm.io/) - ORM library
- [asynq](https://github.com/hibiken/asynq) - Distributed task queue
- [Firecracker](https://firecracker-microvm.github.io/) - MicroVM technology
