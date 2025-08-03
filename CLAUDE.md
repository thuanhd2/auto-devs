# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **Auto-Devs** project - an AI-powered developer task automation system that orchestrates AI CLI tools to automate software development workflows. The system manages the complete lifecycle from task planning to implementation and PR creation.

## Architecture

### Core Components

- **Task Management System**: REST API with WebSocket real-time updates for managing projects and tasks
- **AI Agent Controller**: Orchestrates external AI CLI tools (Claude Code, Gemini CLI, etc.) for automated planning and implementation
- **Git Integration**: Manages Git worktrees, branches, and Pull Request automation
- **Process Management**: Spawns, monitors, and controls AI CLI processes with comprehensive lifecycle management

### Technology Stack

- **Backend**: Go with Gin framework, Clean Architecture pattern
- **Frontend**: React + TypeScript + ShadcnUI (cloned from shadcn-admin template)
- **Database**: PostgreSQL with Redis for caching
- **AI Integration**: External CLI tools via process spawning
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

## Database Schema

### Core Tables

- `projects`: Project configuration and Git repository settings
- `tasks`: Task details, status, and associated branches/PRs
- `executions`: AI CLI execution tracking with process management
- `processes`: Individual process monitoring (setup, cli_agent, monitor, cleanup)

## Important Files & Directories

### Documentation

- `docs/prd.md`: Complete Product Requirements Document
- `docs/technical-design.md`: Detailed technical architecture and API specifications
- `docs/core-concepts.md`: Core system concepts and implementation guidelines
- `development-roadmap.md`: 3-phase development plan with detailed tasks

### Development Guidelines

- Follow Clean Architecture pattern with layers (handler, usecase, repository)
- Use Wire for dependency injection
- Implement comprehensive testing (unit, integration, e2e)
- Maintain OpenAPI/Swagger documentation for all APIs

## MCP Server Requirements

This project requires these MCP servers:

- **serena**: For task and project management capabilities
- **github**: For Git repository operations and PR management

## Development Commands

Implementation will begin with Phase 1 backend setup using:

- Go with Gin framework setup
- PostgreSQL database with golang-migrate
- React frontend cloned from shadcn-admin template
- Docker containerization for development and production
- Run build: `make build`

## Notes for Claude Code

- This is a complex multi-phase project currently in design/planning stage
- Focus on the technical architecture and API specifications in `docs/technical-design.md`
- The system design emphasizes AI CLI orchestration rather than implementing AI logic directly
- Pay attention to the process management patterns for spawning and monitoring external CLI tools
- Database schema supports comprehensive execution tracking and process monitoring
