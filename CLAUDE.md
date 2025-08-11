# CLAUDE.md

## IMPORTANT use tools

- serena
- github

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Auto-Devs API Core

This is a Go-based API server for the Auto-Devs project management system with a React/TypeScript frontend.

### Backend Development Commands

**Build and Run:**

```bash
# Build the application
make build
# or
go build -o bin/autodevs cmd/server/main.go

# Run the application
make run
# or
go run cmd/server/main.go

# Run tests
make test
# or
go test ./... -v
```

**Database Management:**

```bash
# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Reset database (rollback all)
make migrate-reset

# Create new migration
make migrate-create name=migration_name
```

**Code Generation:**

```bash
# Generate Swagger documentation
make swagger
# or
./scripts/generate-swagger.sh

# Generate Wire dependency injection code
make wire
# or
go generate ./internal/di

# Generate mocks
make mocks
```

**Worker Management:**

```bash
# Build worker binary
make build-worker

# Run job worker
make run-worker

# Run worker with verbose logging
make run-worker-verbose

# Run worker with custom name
make run-worker-named name=worker-1
```

### Frontend Development Commands

```bash
# Navigate to frontend directory first
cd frontend

# Development server
npm run dev

# Build for production
npm run build

# Linting and formatting
npm run lint
npm run format
npm run format:check

# Dependency analysis
npm run knip
```

## Architecture Overview

### Backend Architecture (Clean Architecture Pattern)

The backend follows a layered clean architecture:

1. **Entity Layer** (`internal/entity/`) - Core domain models
2. **Repository Layer** (`internal/repository/`) - Data access interfaces and implementations
3. **Usecase Layer** (`internal/usecase/`) - Business logic orchestration
4. **Handler Layer** (`internal/handler/`) - HTTP request handlers and DTOs
5. **Service Layer** (`internal/service/`) - External service integrations

### Key Backend Components

**Dependency Injection:**

- Uses Google Wire for compile-time dependency injection
- Configuration in `internal/di/wire.go`
- Generated code in `internal/di/wire_gen.go`

**Database:**

- PostgreSQL with GORM
- Migration management with golang-migrate
- Repository pattern for data access

**Job Processing:**

- Redis-based job queue using Hibiken Asynq
- Background workers for task execution
- AI-powered planning and execution services

**WebSocket Support:**

- Real-time communication using Centrifuge
- Notification system for task updates

**AI Integration:**

- CLI manager for Claude Code integration
- Process management for AI task execution
- Planning and execution services

**Git Integration:**

- Worktree management for isolated development
- GitHub integration for PR creation
- Branch and commit management

### Frontend Architecture

**Tech Stack:**

- React 19 with TypeScript
- Vite for build tooling
- TailwindCSS for styling
- Radix UI components
- TanStack Router for routing
- TanStack Query for data fetching
- Zustand for state management

**Key Frontend Features:**

- Drag-and-drop task management
- Real-time updates via WebSocket
- Form handling with React Hook Form and Zod validation
- Data visualization with Recharts

### Development Workflow

1. **Backend changes:** Modify code → Generate mocks/wire → Run tests → Update Swagger docs
2. **Database changes:** Create migration → Run migration → Update entities/repositories
3. **API changes:** Update handlers/DTOs → Generate Swagger → Test endpoints
4. **Frontend changes:** Develop components → Lint/format → Build → Test integration

### Testing Strategy

- Unit tests for business logic in usecase layer
- Integration tests using pgtestdb for database operations
- API endpoint testing via HTTP requests
- Mock generation for external dependencies

### Key Configuration Files

- `go.mod` - Go dependencies and module definition
- `Makefile` - Build and development commands
- `.env.example` - Environment configuration template
- `frontend/package.json` - Frontend dependencies and scripts
- `migrations/` - Database schema changes
- `docs/` - API documentation and technical guides
