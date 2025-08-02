# API Core Implementation

This document describes the completed API Core implementation for the Auto-Devs project.

## Overview

The API Core provides RESTful endpoints for managing projects and tasks with comprehensive validation, error handling, and documentation.

## Architecture

### Layers

1. **DTO Layer** (`internal/handler/dto/`)
   - Request/response models for API endpoints
   - Validation tags and example values
   - Conversion helpers between entities and DTOs

2. **Handler Layer** (`internal/handler/`)
   - HTTP request handlers for projects and tasks
   - Input validation and error handling
   - Swagger documentation annotations

3. **Usecase Layer** (`internal/usecase/`)
   - Business logic implementation
   - Extended interfaces for additional functionality
   - Clean separation between HTTP and business logic

4. **Middleware** (`internal/handler/middleware.go`)
   - CORS configuration for frontend integration
   - Request logging for monitoring
   - Error handling and recovery
   - Input validation with user-friendly messages
   - Rate limiting (100 requests/minute)
   - Security headers

## API Endpoints

### Projects

- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects` - List all projects
- `GET /api/v1/projects/{id}` - Get project by ID
- `PUT /api/v1/projects/{id}` - Update project
- `DELETE /api/v1/projects/{id}` - Delete project
- `GET /api/v1/projects/{id}/tasks` - Get project with tasks

### Tasks

- `POST /api/v1/tasks` - Create task
- `GET /api/v1/tasks` - List tasks with filtering
- `GET /api/v1/tasks/{id}` - Get task by ID
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Delete task
- `PATCH /api/v1/tasks/{id}/status` - Update task status
- `GET /api/v1/projects/{project_id}/tasks` - List tasks by project

### Health

- `GET /api/v1/health` - Health check with database status

## Features Implemented

### ✅ Request/Response Models
- Comprehensive DTO models with validation
- Proper JSON tags and examples
- Conversion helpers between entities and DTOs

### ✅ Validation Middleware
- Input validation using struct tags
- User-friendly error messages
- Validation error details in responses

### ✅ CORS Configuration
- Frontend development server support
- Configurable allowed origins, methods, headers
- Credential support for authenticated requests

### ✅ Error Handling
- Consistent error response format
- Proper HTTP status codes
- Panic recovery middleware

### ✅ Request Logging
- Structured request/response logging
- Performance monitoring capabilities

### ✅ Rate Limiting
- Basic rate limiting (100 req/min)
- Configurable limits per endpoint

### ✅ Security Headers
- XSS protection
- Content-type sniffing prevention
- Frame options and HSTS

### ✅ OpenAPI Documentation
- Complete Swagger 3.0 specification
- Request/response schemas
- Example values and descriptions

## Usage

### Starting the Server

```bash
go run cmd/server/main.go
```

The server starts on port 8098 by default.

### API Documentation

The complete API documentation is available in `docs/swagger.yaml` and can be viewed using any OpenAPI-compatible tool.

### Example Requests

#### Create Project
```bash
curl -X POST http://localhost:8098/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Project",
    "description": "A sample project",
    "repo_url": "https://github.com/user/repo"
  }'
```

#### Create Task
```bash
curl -X POST http://localhost:8098/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Implement feature",
    "description": "Add new functionality"
  }'
```

#### Update Task Status
```bash
curl -X PATCH http://localhost:8098/api/v1/tasks/{id}/status \
  -H "Content-Type: application/json" \
  -d '{"status": "IMPLEMENTING"}'
```

## Configuration

### Environment Variables

- `SERVER_PORT` - Server port (default: 8098)
- Database configuration (see config package)

### CORS Origins

Currently configured for development:
- `http://localhost:3000` (React dev server)
- `http://localhost:5173` (Vite dev server)

Update `internal/handler/middleware.go` for production origins.

## Next Steps

The API Core implementation is complete and ready for integration with:

1. **Database Layer** - Repository implementations with GORM
2. **Frontend Integration** - React application API calls
3. **Authentication** - JWT middleware and user management
4. **WebSocket Support** - Real-time task updates
5. **Testing** - Unit and integration tests

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/handler/...
```

## Dependencies Added

- `github.com/gin-contrib/cors` - CORS middleware
- `golang.org/x/time/rate` - Rate limiting
- `github.com/go-playground/validator/v10` - Input validation (already included)

All dependencies have been added to `go.mod` and the application builds successfully.