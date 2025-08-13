# README.md Summary

## Project Overview
Auto-Devs API Core is a Go-based project management system API implementation that follows Clean Architecture principles.

## Key Features
- ✅ RESTful API with comprehensive validation
- ✅ OpenAPI/Swagger documentation with UI at `/swagger/index.html`
- ✅ CORS configuration for frontend integration
- ✅ Rate limiting and security headers
- ✅ Request logging and error handling
- ✅ Database migrations with GORM
- ✅ Clean Architecture pattern

## Prerequisites
- Go 1.24+
- PostgreSQL
- Make (optional)

## Quick Setup
```bash
git clone <repository-url>
cd vk-c373-api-core-i
go mod download && go mod tidy
go build -o server cmd/server/main.go
./server
```

## Development Commands
- `make build` - Build application
- `make run` - Run application
- `make test` - Run tests
- `make swagger` - Generate Swagger docs
- `make migrate-up/down/reset` - Database migrations

## API Structure
**Health Check:** `/api/v1/health`

**Projects:** CRUD operations at `/api/v1/projects`
- Create, list, get, update, delete projects
- Get project with tasks

**Tasks:** CRUD operations at `/api/v1/tasks`  
- Create, list, get, update, delete tasks
- Update task status, get task with project

## Architecture Layers
1. **DTO Layer** - Request/response models
2. **Handler Layer** - HTTP request handlers  
3. **Usecase Layer** - Business logic
4. **Repository Layer** - Data access

## Project Structure
- `cmd/server/` - Application entry point
- `internal/` - Core application code (di, entity, handler, repository, usecase)
- `migrations/` - Database migrations
- `docs/` - API documentation
- `scripts/` - Utility scripts

## Documentation
Swagger UI available at `http://localhost:8098/swagger/index.html` with JSON/YAML endpoints for API specs.