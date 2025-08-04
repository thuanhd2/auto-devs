# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## [IMPORTANT] Tools use must to use if possible

- **serena**: Semantic code retrieval and editing tools and so many more
- **github**: Git repository operations and PR management

## Project Overview

This is the **Auto-Devs** project - an AI-powered developer task automation system that orchestrates AI CLI tools to automate software development workflows. The system manages the complete lifecycle from task planning to implementation and PR creation.

## Architecture

### Core Components

- **Task Management System**: REST API with WebSocket real-time updates for managing projects and tasks
- **AI Agent Controller**: Orchestrates external AI CLI tools (Claude Code, Gemini CLI, etc.) for automated planning and implementation
- **Git Integration**: Manages Git worktrees, branches, and Pull Request automation
- **Process Management**: Spawns, monitors, and controls AI CLI processes with comprehensive lifecycle management

### Technology Stack

- **Backend**: Go 1.24+ with Gin framework, Clean Architecture pattern
- **Frontend**: React 19 + TypeScript + ShadcnUI + TanStack Router/Query
- **Database**: PostgreSQL with GORM, golang-migrate for migrations
- **WebSocket**: Real-time updates with Gorilla WebSocket
- **AI Integration**: External CLI tools via process spawning
- **Testing**: Testify for Go, comprehensive test coverage
- **Documentation**: Swagger/OpenAPI with swaggo
- **DI**: Google Wire for dependency injection
- **Infrastructure**: Docker, GitHub Actions for CI/CD

## Development Phases

The project follows a 3-phase development approach:

### Phase 1: Task Management System (4-6 weeks)

- Core infrastructure with Go backend and React frontend
- Project and task CRUD operations
- Real-time WebSocket updates
- Manual task workflow without AI automation

### Phase 2: Git Worktree Integration (3-4 weeks)

- Git CLI integration for branch and worktree management
- Isolated development environments per task
- Manual implementation support with Git operations

### Phase 3: AI Executor (6-8 weeks)

- AI CLI integration for automated planning and implementation
- Process management for AI agent orchestration
- Automated Pull Request creation and merge detection

## Key Concepts

### Task Lifecycle

Tasks follow this state machine:

- `TODO` → `PLANNING` → `PLAN_REVIEWING` → `IMPLEMENTING` → `CODE_REVIEWING` → `DONE`
- Any state can transition to `CANCELLED`

### AI Integration Strategy

- System acts as orchestrator, delegating all AI work to external CLI tools
- Extensible plugin architecture for different AI CLIs
- MVP focuses on Claude Code CLI integration
- Process monitoring and lifecycle management for spawned AI processes

### Git Worktree Management

- Each task gets isolated Git worktree and branch
- Branch naming: `task-{task_id}-{slug}`
- Automatic cleanup after task completion

## Code Architecture

### Backend Structure (Clean Architecture)

- **Handler Layer** (`internal/handler/`): HTTP handlers, DTOs, middleware, WebSocket handlers
- **Usecase Layer** (`internal/usecase/`): Business logic, orchestration between repositories
- **Repository Layer** (`internal/repository/`): Data access with PostgreSQL implementation
- **Entity Layer** (`internal/entity/`): Domain models and business entities
- **Service Layer** (`internal/service/`): External integrations (Git operations, worktree management)
- **WebSocket** (`internal/websocket/`): Real-time communication hub and connection management
- **DI** (`internal/di/`): Wire-based dependency injection

### Frontend Structure

- **Components**: Reusable UI components with ShadcnUI
- **Features**: Feature-based organization (dashboard, projects, settings)
- **Hooks**: Custom React hooks for API calls and state management
- **Services**: API clients and WebSocket service
- **Types**: TypeScript type definitions
- **Routes**: TanStack Router for navigation

### Database Schema

- `projects`: Project configuration and Git repository settings
- `tasks`: Task details, status, and lifecycle tracking
- `worktrees`: Git worktree management for isolated development
- `audit_logs`: System audit trail and task history
- `project_settings`: Configuration and preferences per project

## Important Files & Directories

### Documentation

- `docs/prd.md`: Complete Product Requirements Document
- `docs/technical-design.md`: Detailed technical architecture and API specifications
- `docs/core-concepts.md`: Core system concepts and implementation guidelines
- `development-roadmap.md`: 3-phase development plan with detailed tasks

### Development Guidelines

- **Architecture**: Follow Clean Architecture with clear layer separation
- **Dependency Injection**: Use Google Wire for DI code generation
- **Testing**: Comprehensive testing with testify, pgtestdb for database tests
- **Documentation**: Maintain OpenAPI/Swagger docs for all endpoints
- **Git Integration**: Use worktree service for isolated branch development
- **WebSocket**: Real-time updates for task status and project changes
- **Error Handling**: Structured error responses with proper HTTP status codes
- **Validation**: Use go-playground/validator for request validation

## Development Commands

### Backend (Go)

- `make build` - Build the Go application
- `make run` - Run the Go server (http://localhost:8098)
- `make test` - Run all Go tests
- `go run cmd/server/main.go` - Run server directly
- `make deps` - Download and tidy Go dependencies
- `make clean` - Clean build artifacts

### Frontend (React + TypeScript)

- `cd frontend && npm run dev` - Start development server (http://localhost:5173)
- `cd frontend && npm run build` - Build for production
- `cd frontend && npm run lint` - Run ESLint
- `cd frontend && npm run format` - Format code with Prettier
- `cd frontend && npm run format:check` - Check code formatting

### Database (PostgreSQL)

- `make migrate-up` - Run all pending migrations
- `make migrate-down` - Rollback last migration
- `make migrate-reset` - Rollback all migrations
- `make migrate-create name=<name>` - Create new migration
- `make db-setup` - Setup database with migrations

### Documentation & Code Generation

- `make swagger` - Generate Swagger documentation
- `make swagger-serve` - Show Swagger UI URLs (http://localhost:8098/swagger/index.html)
- `make wire` - Generate Wire dependency injection code
- `make mocks` - Generate mocks for testing
- `make mocks-regen` - Regenerate all mocks

### Development Environment

- `./run-dev.sh` - Start both backend and frontend in development mode
- `make help` - Show all available Makefile commands

## Notes for Claude Code

- This is a complex multi-phase project currently in design/planning stage
- Focus on the technical architecture and API specifications in `docs/technical-design.md`
- The system design emphasizes AI CLI orchestration rather than implementing AI logic directly
- Pay attention to the process management patterns for spawning and monitoring external CLI tools
- Database schema supports comprehensive execution tracking and process monitoring
