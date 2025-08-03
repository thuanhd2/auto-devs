# Auto-Devs API Core

API Core implementation for the Auto-Devs project management system.

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL
- Make (optional)

### Installation

```bash
# Clone repository
git clone <repository-url>
cd vk-c373-api-core-i

# Install dependencies
go mod download
go mod tidy

# Build application
go build -o server cmd/server/main.go

# Run server
./server
```

## 📚 API Documentation

### Swagger UI

The API documentation is available through Swagger UI:

- **Swagger UI**: http://localhost:8098/swagger/index.html
- **Root redirect**: http://localhost:8098/ (redirects to Swagger UI)

### API Documentation Files

- **Swagger JSON**: http://localhost:8098/swagger.json
- **Swagger YAML**: http://localhost:8098/swagger.yaml

## 🔧 Development

### Generate Swagger Documentation

```bash
# Using script
./scripts/generate-swagger.sh

# Using Makefile
make swagger

# Using swag CLI directly
swag init -g cmd/server/main.go
```

### Available Make Commands

```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run the application
make test          # Run tests
make swagger       # Generate Swagger documentation
make swagger-serve # Show Swagger UI URLs
make deps          # Download dependencies
make clean         # Clean build artifacts
```

### Database Setup

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Reset database
make migrate-reset
```

## 📋 API Endpoints

### Health Check

- `GET /api/v1/health` - Check server and database status

### Projects

- `POST /api/v1/projects` - Create new project
- `GET /api/v1/projects` - List all projects
- `GET /api/v1/projects/{id}` - Get project by ID
- `PUT /api/v1/projects/{id}` - Update project
- `DELETE /api/v1/projects/{id}` - Delete project
- `GET /api/v1/projects/{id}/tasks` - Get project with tasks

### Tasks

- `POST /api/v1/tasks` - Create new task
- `GET /api/v1/tasks` - List tasks with filtering
- `GET /api/v1/tasks/{id}` - Get task by ID
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Delete task
- `PATCH /api/v1/tasks/{id}/status` - Update task status
- `GET /api/v1/tasks/{id}/project` - Get task with project

## 🏗️ Architecture

### Layers

1. **DTO Layer** (`internal/handler/dto/`) - Request/response models
2. **Handler Layer** (`internal/handler/`) - HTTP request handlers
3. **Usecase Layer** (`internal/usecase/`) - Business logic
4. **Repository Layer** (`internal/repository/`) - Data access

### Features

- ✅ RESTful API with comprehensive validation
- ✅ OpenAPI/Swagger documentation
- ✅ CORS configuration for frontend integration
- ✅ Rate limiting and security headers
- ✅ Request logging and error handling
- ✅ Database migrations with GORM
- ✅ Clean Architecture pattern

## 📁 Project Structure

```
├── cmd/server/           # Application entry point
├── config/              # Configuration management
├── docs/                # API documentation
├── internal/            # Internal application code
│   ├── di/             # Dependency injection
│   ├── entity/         # Domain entities
│   ├── handler/        # HTTP handlers
│   ├── repository/     # Data access layer
│   └── usecase/        # Business logic
├── migrations/          # Database migrations
├── pkg/                # Public packages
│   └── database/       # Database utilities
└── scripts/            # Utility scripts
```

## 🔗 Links

- [API Implementation Guide](docs/api-implementation.md)
- [Swagger Usage Guide](docs/swagger-usage.md)
- [Technical Design](docs/technical-design.md)
- [Development Roadmap](docs/development-roadmap.md)

## 📄 License

MIT License
