# Phase 2: Git Worktree Integration - MVP Version

## Overview

Phase 2 MVP focuses on basic Git integration with validation and core worktree functionality. This simplified version provides essential Git operations without complex features.

**Timeline:** 1-2 weeks
**Goal:** Basic Git worktree integration with validation and core functionality

---

## Release 2.1: Git Infrastructure Foundation (Week 1)

### Backend Git Integration

#### TASK-2.1.1: Git Manager Service Foundation

**Priority:** High
**Estimated Time:** 2-3 days
**Description:** Implement basic Git CLI wrapper service with validation

**Steps:**

1. Create Git service structure in `internal/service/git/`:

   - `git_manager.go` - Main Git operations manager
   - `git_commands.go` - Basic Git CLI command wrappers
   - `git_validation.go` - Repository and branch validation
   - `git_errors.go` - Git-specific error types

2. Implement basic Git CLI wrapper functions:

   - Execute Git commands with proper error handling
   - Command timeout and cancellation support
   - Git output parsing and validation
   - Working directory management

3. Add Git repository validation:

   - Check if directory is a valid Git repository
   - Validate remote repository URLs
   - Check Git version compatibility
   - Validate user Git configuration

4. Add basic Git status and information retrieval:

   - Get current branch information
   - Check repository status (clean/dirty)
   - List existing branches
   - Get basic commit information

5. Implement Git error handling and logging:

   - Standardized Git error types
   - Error message parsing and translation
   - Basic logging for Git operations
   - Error recovery suggestions

6. Create Git service unit tests:
   - Mock Git CLI responses
   - Test error scenarios
   - Test authentication flows
   - Test validation functions

**Acceptance Criteria:**

- Git CLI wrapper executes commands correctly
- Authentication methods work properly
- Error handling provides clear feedback
- Service interface is clean and testable
- Unit tests achieve >70% coverage

---

#### TASK-2.1.2: Basic Branch Management System

**Priority:** High
**Estimated Time:** 1-2 days
**Description:** Implement basic branch naming and management

**Steps:**

1. Define simple branch naming conventions:

   ```go
   type BranchNamingConfig struct {
       Prefix       string // e.g., "task"
       IncludeID    bool   // Include task ID in name
       Separator    string // e.g., "-"
   }
   ```

2. Implement basic branch name generation:

   - Task ID to branch name conversion
   - Basic title slugification
   - Duplicate branch name handling
   - Branch name validation against Git rules

3. Create basic branch lifecycle management:

   - Branch creation from main/default branch
   - Branch checkout and switching
   - Basic branch deletion and cleanup

4. Add branch conflict detection:

   - Check for existing branches with same name
   - Detect branch conflicts before creation
   - Handle basic branch naming collisions

5. Create basic branch validation rules:
   - Branch name format validation
   - Branch name length limits
   - Special character restrictions

**Acceptance Criteria:**

- Branch naming follows consistent conventions
- Branch conflicts are detected and handled
- Basic branch lifecycle is properly managed
- Validation catches common errors

---

#### TASK-2.1.3: Worktree Base Directory Management

**Priority:** High
**Estimated Time:** 1-2 days
**Description:** Implement basic worktree directory structure and management

**Steps:**

1. Design simple worktree directory structure:

   ```
   /worktrees/
   ├── project-{id}/
   │   └── task-{id}/     # Task-specific worktrees
   ```

2. Implement basic worktree directory configuration:

   - Configurable base directory path
   - Directory permission and access validation
   - Basic storage space monitoring

3. Create basic worktree path management:

   - Generate unique worktree paths per task
   - Handle basic path conflicts
   - Validate path lengths and characters

4. Add basic worktree directory validation:

   - Check directory existence and permissions
   - Validate available disk space
   - Ensure directory is not in use

5. Implement basic worktree cleanup:
   - Remove worktree when task is completed/cancelled
   - Clean up branch from local repository
   - Remove worktree directory and files

**Acceptance Criteria:**

- Worktree directories are properly structured
- Path management handles basic edge cases
- Cleanup operations work reliably
- Basic validation prevents common errors

---

#### TASK-2.1.4: Enhanced Project Model for Git

**Priority:** High
**Estimated Time:** 1-2 days
**Description:** Extend project entity and database schema to support basic Git operations

**Steps:**

1. Update Project entity with basic Git fields:

   ```go
   type Project struct {
       // Existing fields...
       RepositoryURL      string    `json:"repository_url" gorm:"column:repository_url"`
       MainBranch         string    `json:"main_branch" gorm:"column:main_branch;default:main"`
       WorktreeBasePath   string    `json:"worktree_base_path" gorm:"column:worktree_base_path"`
       GitAuthMethod      string    `json:"git_auth_method" gorm:"column:git_auth_method"` // "ssh" or "https"
       GitEnabled         bool      `json:"git_enabled" gorm:"column:git_enabled;default:false"`
   }
   ```

2. Create database migration for Git fields:

   - Add new columns to projects table
   - Create basic indexes for Git-related queries
   - Handle existing project data migration

3. Implement basic Git authentication configuration:

   - Store authentication method preference (SSH/HTTPS)
   - Basic authentication status tracking
   - Let users configure Git credentials manually

4. Add repository URL validation:

   - URL format validation (HTTPS/SSH)
   - Basic repository accessibility testing
   - Clone permission verification

5. Create basic Git-enabled project validation:
   - Validate required Git fields when enabled
   - Basic repository clone testing during setup
   - Authentication validation workflow

**Acceptance Criteria:**

- Project model supports basic Git operations
- Database migration preserves existing data
- Git authentication works securely
- Basic validation catches common errors

---

### Database Schema Updates

#### TASK-2.1.5: Basic Worktree Database Schema

**Priority:** High
**Estimated Time:** 1-2 days
**Description:** Create basic database schema for Git worktree tracking

**Steps:**

1. Design and create `worktrees` table:

   ```sql
   CREATE TABLE worktrees (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
       project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
       branch_name VARCHAR(255) NOT NULL,
       worktree_path TEXT NOT NULL,
       status VARCHAR(50) NOT NULL DEFAULT 'creating',
       created_at TIMESTAMP DEFAULT NOW(),
       updated_at TIMESTAMP DEFAULT NOW(),
       deleted_at TIMESTAMP NULL,
       UNIQUE(task_id),
       UNIQUE(worktree_path)
   );
   ```

2. Update `tasks` table with basic Git fields:

   ```sql
   ALTER TABLE tasks ADD COLUMN branch_name VARCHAR(255);
   ALTER TABLE tasks ADD COLUMN worktree_path TEXT;
   ALTER TABLE tasks ADD COLUMN git_status VARCHAR(50) DEFAULT 'none';
   ```

3. Create basic indexes for performance:

   - Index on `worktrees.task_id` and `worktrees.project_id`
   - Index on `tasks.branch_name` and `tasks.git_status`

4. Add basic database constraints:

   - Check constraints for valid status values
   - Foreign key constraints with proper cascading
   - NOT NULL constraints for required fields

5. Create basic database migration scripts:
   - Up migration with all new structures
   - Down migration for rollback capability
   - Data migration for existing records

**Acceptance Criteria:**

- All tables and relationships work correctly
- Basic indexes provide adequate performance
- Constraints prevent invalid data
- Migrations handle basic scenarios

---

## Release 2.2: Basic Worktree Operations (Week 2)

### Core Worktree Operations

#### TASK-2.2.1: Basic Worktree Creation and Management

**Priority:** High
**Estimated Time:** 2-3 days
**Description:** Implement basic worktree lifecycle management

**Steps:**

1. Implement basic worktree creation workflow:

   - Validate task eligibility for worktree creation
   - Generate unique branch name using naming conventions
   - Create Git worktree from main branch
   - Initialize basic worktree directory structure
   - Update task and worktree database records

2. Create basic branch management within worktrees:

   - Create new branch for task from main/default branch
   - Switch to task branch in worktree
   - Basic branch configuration

3. Add basic worktree initialization:

   - Copy necessary configuration files
   - Create basic worktree metadata files

4. Implement basic worktree status tracking:

   - Track worktree creation progress
   - Basic worktree health monitoring
   - Update status in database throughout lifecycle

5. Create basic worktree validation:

   - Verify worktree is properly created
   - Validate Git repository state
   - Check basic branch configuration

6. Add basic worktree cleanup:

   - Remove worktree when task is completed/cancelled
   - Clean up branch from local repository
   - Remove worktree directory and files
   - Update database records (soft delete)

7. Implement basic error handling:
   - Handle partial worktree creation failures
   - Basic recovery from interrupted operations
   - Clean up failed worktree attempts

**Acceptance Criteria:**

- Worktrees are created reliably for eligible tasks
- Basic branch management works correctly
- Cleanup operations leave no artifacts
- Error handling provides basic feedback

---

#### TASK-2.2.2: Basic Task-Branch Integration

**Priority:** High
**Estimated Time:** 1-2 days
**Description:** Integrate task management with basic Git branch operations

**Steps:**

1. Implement automatic branch creation workflow:

   - Trigger branch creation when task moves to IMPLEMENTING status
   - Generate branch name based on task information
   - Create worktree and branch simultaneously
   - Update task record with Git information

2. Create basic task status to Git status mapping:

   ```go
   type TaskGitStatus string
   const (
       GitStatusNone       TaskGitStatus = "none"
       GitStatusCreating   TaskGitStatus = "creating"
       GitStatusActive     TaskGitStatus = "active"
       GitStatusCompleted  TaskGitStatus = "completed"
       GitStatusCleaning   TaskGitStatus = "cleaning"
   )
   ```

3. Add basic Git status transition validation:

   - Define allowed Git status transitions
   - Validate transitions in business logic
   - Prevent invalid Git status changes

4. Implement basic task-branch relationship management:

   - Link tasks to their corresponding branches
   - Track basic branch metadata in task records
   - Maintain branch-task relationship integrity

5. Create basic Git operation triggers for task events:
   - Task status change triggers
   - Task deletion cleanup triggers
   - Basic task completion Git operations

**Acceptance Criteria:**

- Task status changes properly trigger Git operations
- Basic branch information is tracked and displayed
- Git status transitions follow business rules
- Integration maintains basic data consistency

---

### UI Enhancements for Git Integration

#### TASK-2.2.3: Basic Git Status Display

**Priority:** High
**Estimated Time:** 1-2 days
**Description:** Update task management UI to display basic Git information

**Steps:**

1. Update task card components to show basic Git information:

   - Add branch name display in task cards
   - Show basic Git status indicators
   - Display worktree path information

2. Create basic Git status indicator components:

   ```tsx
   // GitStatusBadge component
   interface GitStatusBadgeProps {
     status: TaskGitStatus;
     branchName?: string;
     className?: string;
   }
   ```

   - Color-coded status badges for different Git states
   - Basic tooltip information for status meanings
   - Loading states for Git operations in progress

3. Add basic Git information to task detail view:

   - Dedicated Git information section
   - Branch name with copy-to-clipboard functionality
   - Worktree path with open-in-file-manager link

4. Implement basic Git operation controls:

   - "Create Worktree" button for eligible tasks
   - "Open Worktree" button to open file manager
   - Basic "Cleanup Worktree" button

5. Add basic Git status filtering:
   - Filter tasks by Git status
   - Basic search tasks by branch name

**Acceptance Criteria:**

- Git information is clearly displayed in relevant views
- Basic Git operations can be triggered from the UI
- Status updates appear in real-time
- UI remains responsive during Git operations

---

#### TASK-2.2.4: Basic Project Git Configuration Interface

**Priority:** High
**Estimated Time:** 1-2 days
**Description:** Create basic UI for Git repository configuration

**Steps:**

1. Create basic project Git settings page:

   - Repository URL configuration form
   - Main branch selection dropdown
   - Basic authentication method selection

2. Implement basic Git authentication configuration:

   - Authentication method selection (SSH/HTTPS)
   - Basic authentication status display
   - Link to Git credential setup documentation

3. Add basic repository validation interface:

   - Repository URL validation
   - Basic repository accessibility testing
   - Clone permission verification

4. Create basic Git configuration testing:

   - "Test Connection" button with progress indicator
   - Basic "Test Clone" functionality
   - Authentication verification

5. Add basic Git configuration help:
   - Inline help for configuration options
   - Basic troubleshooting guides

**Acceptance Criteria:**

- Git configuration is intuitive and user-friendly
- Authentication setup works for supported methods
- Repository validation provides clear feedback
- Configuration changes take effect immediately

---

## Success Metrics for Phase 2 MVP

### Technical Metrics

- **Worktree Creation Time:** <60 seconds for standard repositories
- **Git Operation Success Rate:** >90% for basic Git operations
- **Worktree Cleanup Completion:** >95% successful cleanup rate
- **Git Status Accuracy:** >95% accurate status reflection

### User Experience Metrics

- **Implementation Setup Time:** <3 minutes from task to worktree
- **Git Configuration Completion:** <10 minutes for new projects
- **User Adoption Rate:** >60% of eligible tasks use Git integration
- **Error Resolution Time:** <15 minutes for common Git issues

### Quality Metrics

- **Git Integration Test Coverage:** >70% code coverage
- **Git Operation Error Rate:** <10% for basic operations
- **Data Consistency:** >95% consistency between Git and database state

---

## Phase 2 MVP Completion Criteria

1. **Functional Requirements:**

   - Basic Git repository integration
   - Automated worktree and branch management
   - Basic implementation workflow support
   - Git status monitoring and display
   - Basic project Git configuration interface

2. **Technical Requirements:**

   - Secure Git authentication handling
   - Reliable worktree lifecycle management
   - Basic Git operation error handling
   - Cross-platform Git integration
   - Basic Git operation testing

3. **Quality Requirements:**

   - Basic Git integration documentation
   - Security measures properly implemented
   - Basic user experience testing
   - Cross-platform compatibility verified

4. **User Experience Requirements:**
   - Intuitive Git configuration interface
   - Basic implementation workflow
   - Clear Git status visualization
   - Effective error handling and guidance

Upon completion of Phase 2 MVP, users will have a functional Git integration system that provides basic isolated development environments for each task, enabling manual implementation work with proper version control and branch management.
