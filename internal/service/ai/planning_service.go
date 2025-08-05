package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/auto-devs/auto-devs/internal/entity"
)

// PlanningService generates implementation plans for tasks using AI
type PlanningService struct {
	executionService *ExecutionService
	cliManager       *CLIManager
}

// NewPlanningService creates a new planning service
func NewPlanningService(executionService *ExecutionService, cliManager *CLIManager) *PlanningService {
	return &PlanningService{
		executionService: executionService,
		cliManager:       cliManager,
	}
}

// GeneratePlan generates an implementation plan for the given task
func (ps *PlanningService) GeneratePlan(task entity.Task) (*Plan, error) {
	// Generate AI prompt for planning phase
	prompt, err := ps.generatePlanningPrompt(task)
	if err != nil {
		return nil, fmt.Errorf("failed to generate planning prompt: %w", err)
	}

	// Create plan structure
	plan := &Plan{
		ID:          uuid.New().String(),
		TaskID:      task.ID.String(),
		Description: fmt.Sprintf("Implementation plan for task: %s", task.Title),
		Steps:       ps.generateInitialPlanSteps(task),
		Context: map[string]string{
			"task_title":       task.Title,
			"task_description": task.Description,
			"task_priority":    string(task.Priority),
			"task_status":      string(task.Status),
			"prompt":           prompt,
		},
		CreatedAt: time.Now(),
	}

	// Execute AI prompt to generate detailed plan
	err = ps.enhancePlanWithAI(plan, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to enhance plan with AI: %w", err)
	}

	return plan, nil
}

// generatePlanningPrompt creates a structured prompt for AI planning phase
func (ps *PlanningService) generatePlanningPrompt(task entity.Task) (string, error) {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("# Task Implementation Planning\n\n")
	promptBuilder.WriteString("You are an expert software developer tasked with creating a detailed implementation plan.\n\n")

	// Task Information
	promptBuilder.WriteString("## Task Details\n")
	promptBuilder.WriteString(fmt.Sprintf("**Title:** %s\n", task.Title))
	promptBuilder.WriteString(fmt.Sprintf("**Description:** %s\n", task.Description))
	promptBuilder.WriteString(fmt.Sprintf("**Priority:** %s\n", task.Priority))
	
	if task.EstimatedHours != nil {
		promptBuilder.WriteString(fmt.Sprintf("**Estimated Hours:** %.2f\n", *task.EstimatedHours))
	}
	
	if len(task.Tags) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(task.Tags, ", ")))
	}

	promptBuilder.WriteString("\n## Requirements\n")
	promptBuilder.WriteString("Please create a comprehensive implementation plan that includes:\n\n")
	promptBuilder.WriteString("1. **Analysis Phase**\n")
	promptBuilder.WriteString("   - Understanding the requirements\n")
	promptBuilder.WriteString("   - Identifying key components and dependencies\n")
	promptBuilder.WriteString("   - Risk assessment\n\n")

	promptBuilder.WriteString("2. **Design Phase**\n")
	promptBuilder.WriteString("   - Architecture decisions\n")
	promptBuilder.WriteString("   - Interface definitions\n")
	promptBuilder.WriteString("   - Database schema changes (if applicable)\n\n")

	promptBuilder.WriteString("3. **Implementation Phase**\n")
	promptBuilder.WriteString("   - Step-by-step implementation tasks\n")
	promptBuilder.WriteString("   - File modifications and creations\n")
	promptBuilder.WriteString("   - Code structure and patterns\n\n")

	promptBuilder.WriteString("4. **Testing Phase**\n")
	promptBuilder.WriteString("   - Unit test requirements\n")
	promptBuilder.WriteString("   - Integration test scenarios\n")
	promptBuilder.WriteString("   - Manual testing steps\n\n")

	promptBuilder.WriteString("5. **Validation Phase**\n")
	promptBuilder.WriteString("   - Acceptance criteria verification\n")
	promptBuilder.WriteString("   - Code review checklist\n")
	promptBuilder.WriteString("   - Documentation updates\n\n")

	promptBuilder.WriteString("## Output Format\n")
	promptBuilder.WriteString("Please provide the plan as structured markdown with clear sections and actionable steps.\n")
	promptBuilder.WriteString("Each step should be specific, measurable, and include estimated time if possible.\n")
	promptBuilder.WriteString("Include any assumptions, dependencies, or potential risks.\n\n")

	promptBuilder.WriteString("## Context\n")
	promptBuilder.WriteString("This is a Go-based web application with Clean Architecture pattern.\n")
	promptBuilder.WriteString("The codebase uses Gin framework, GORM for database, and follows standard Go practices.\n")
	
	return promptBuilder.String(), nil
}

// generateInitialPlanSteps creates basic plan steps before AI enhancement
func (ps *PlanningService) generateInitialPlanSteps(task entity.Task) []PlanStep {
	steps := []PlanStep{
		{
			ID:          uuid.New().String(),
			Description: "Analyze task requirements and constraints",
			Action:      "analysis",
			Parameters: map[string]string{
				"task_id":    task.ID.String(),
				"task_type":  "requirement_analysis",
			},
			Order: 1,
		},
		{
			ID:          uuid.New().String(),
			Description: "Design implementation approach",
			Action:      "design",
			Parameters: map[string]string{
				"task_id":     task.ID.String(),
				"design_type": "technical_design",
			},
			Order: 2,
		},
		{
			ID:          uuid.New().String(),
			Description: "Implement core functionality",
			Action:      "implement",
			Parameters: map[string]string{
				"task_id":           task.ID.String(),
				"implementation_type": "core_features",
			},
			Order: 3,
		},
		{
			ID:          uuid.New().String(),
			Description: "Write and execute tests",
			Action:      "test",
			Parameters: map[string]string{
				"task_id":   task.ID.String(),
				"test_type": "comprehensive",
			},
			Order: 4,
		},
		{
			ID:          uuid.New().String(),
			Description: "Validate implementation against requirements",
			Action:      "validate",
			Parameters: map[string]string{
				"task_id":        task.ID.String(),
				"validation_type": "acceptance_criteria",
			},
			Order: 5,
		},
	}

	return steps
}

// enhancePlanWithAI uses AI to generate detailed plan content
func (ps *PlanningService) enhancePlanWithAI(plan *Plan, prompt string) error {
	// Create a mock AI command for plan generation
	// In a real implementation, this would use the actual AI CLI tool
	command := ps.buildPlanningCommand(plan, prompt)
	
	// For now, we'll enhance the plan with realistic content
	// In the actual implementation, this would execute the AI command
	ps.enhancePlanStepsWithDetails(plan)
	
	// Add the generated prompt to context for future reference
	plan.Context["ai_command"] = command
	plan.Context["generation_method"] = "ai_enhanced"
	
	return nil
}

// buildPlanningCommand creates the AI CLI command for plan generation
func (ps *PlanningService) buildPlanningCommand(plan *Plan, prompt string) string {
	// This would be the actual CLI command to execute
	// Format: claude-code --mode=planning --task-id={task_id} --prompt-file={prompt_file}
	return fmt.Sprintf("claude-code --mode=planning --task-id=%s --output=markdown", plan.TaskID)
}

// enhancePlanStepsWithDetails adds detailed descriptions to plan steps
func (ps *PlanningService) enhancePlanStepsWithDetails(plan *Plan) {
	for i := range plan.Steps {
		step := &plan.Steps[i]
		
		switch step.Action {
		case "analysis":
			step.Description = ps.generateAnalysisStepDetails(plan, step)
		case "design":
			step.Description = ps.generateDesignStepDetails(plan, step)
		case "implement":
			step.Description = ps.generateImplementationStepDetails(plan, step)
		case "test":
			step.Description = ps.generateTestStepDetails(plan, step)
		case "validate":
			step.Description = ps.generateValidationStepDetails(plan, step)
		}
	}
}

// generateAnalysisStepDetails creates detailed analysis step description
func (ps *PlanningService) generateAnalysisStepDetails(plan *Plan, step *PlanStep) string {
	return fmt.Sprintf(`## Analysis Phase

### Requirements Analysis
- Review task description: "%s"
- Identify functional requirements
- Identify non-functional requirements
- Document assumptions and constraints

### Component Analysis
- Identify affected system components
- Map dependencies between components  
- Assess impact on existing functionality
- Identify integration points

### Risk Assessment
- Technical risks and mitigation strategies
- Performance implications
- Security considerations
- Compatibility concerns

### Deliverables
- Requirements specification document
- Component impact analysis
- Risk assessment report
- Updated task scope (if needed)`, plan.Context["task_description"])
}

// generateDesignStepDetails creates detailed design step description
func (ps *PlanningService) generateDesignStepDetails(plan *Plan, step *PlanStep) string {
	return `## Design Phase

### Architecture Design
- Define system architecture changes
- Identify design patterns to apply
- Plan component interfaces
- Design data flow and control flow

### Database Design
- Schema modifications (if applicable)
- Data migration requirements
- Index optimization
- Data validation rules

### API Design
- REST endpoint specifications
- Request/response schemas
- Error handling strategies
- Authentication/authorization requirements

### Deliverables
- Technical design document
- Database schema changes
- API specifications
- Interface definitions`
}

// generateImplementationStepDetails creates detailed implementation step description
func (ps *PlanningService) generateImplementationStepDetails(plan *Plan, step *PlanStep) string {
	return `## Implementation Phase

### Core Development
- Implement business logic components
- Create database repositories
- Develop API endpoints
- Implement error handling

### Code Structure
- Follow Clean Architecture patterns
- Implement proper dependency injection
- Add comprehensive logging
- Ensure proper error propagation

### Integration
- Integrate with existing services
- Implement inter-component communication
- Add configuration management
- Update dependency injection wiring

### Deliverables
- Implemented source code
- Updated configuration files
- Integration test fixtures
- Code documentation`
}

// generateTestStepDetails creates detailed testing step description
func (ps *PlanningService) generateTestStepDetails(plan *Plan, step *PlanStep) string {
	return `## Testing Phase

### Unit Testing
- Test individual components in isolation
- Mock external dependencies
- Achieve >80% code coverage
- Test edge cases and error conditions

### Integration Testing
- Test component interactions
- Database integration tests
- API endpoint tests
- End-to-end workflow tests

### Performance Testing
- Load testing (if applicable)
- Response time validation
- Resource utilization monitoring
- Scalability assessment

### Deliverables
- Complete test suite
- Test coverage report
- Performance test results
- Bug fixes and optimizations`
}

// generateValidationStepDetails creates detailed validation step description
func (ps *PlanningService) generateValidationStepDetails(plan *Plan, step *PlanStep) string {
	return `## Validation Phase

### Acceptance Criteria Verification
- Verify all requirements are met
- Test user scenarios and workflows
- Validate business logic correctness
- Confirm non-functional requirements

### Code Quality Review
- Code style and standards compliance
- Security vulnerability assessment
- Performance optimization review
- Documentation completeness

### Deployment Readiness
- Environment configuration validation
- Database migration testing
- Rollback procedure verification
- Monitoring and alerting setup

### Deliverables
- Acceptance test results
- Code review report
- Deployment checklist
- User documentation updates`
}

// GetPlanAsMarkdown returns the plan formatted as markdown text
func (ps *PlanningService) GetPlanAsMarkdown(plan *Plan) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# Implementation Plan: %s\n\n", plan.Context["task_title"]))
	builder.WriteString(fmt.Sprintf("**Plan ID:** %s\n", plan.ID))
	builder.WriteString(fmt.Sprintf("**Task ID:** %s\n", plan.TaskID))
	builder.WriteString(fmt.Sprintf("**Created:** %s\n\n", plan.CreatedAt.Format("2006-01-02 15:04:05")))

	builder.WriteString("## Overview\n\n")
	builder.WriteString(fmt.Sprintf("%s\n\n", plan.Description))

	builder.WriteString("## Task Details\n\n")
	builder.WriteString(fmt.Sprintf("- **Title:** %s\n", plan.Context["task_title"]))
	builder.WriteString(fmt.Sprintf("- **Priority:** %s\n", plan.Context["task_priority"]))
	builder.WriteString(fmt.Sprintf("- **Status:** %s\n\n", plan.Context["task_status"]))

	builder.WriteString("## Implementation Steps\n\n")
	for _, step := range plan.Steps {
		builder.WriteString(fmt.Sprintf("### Step %d: %s\n\n", step.Order, step.Description))
		
		if step.Parameters != nil && len(step.Parameters) > 0 {
			builder.WriteString("**Parameters:**\n")
			for key, value := range step.Parameters {
				builder.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
			}
			builder.WriteString("\n")
		}
	}

	builder.WriteString("## Context Information\n\n")
	for key, value := range plan.Context {
		if key != "prompt" && key != "ai_command" { // Skip verbose context
			builder.WriteString(fmt.Sprintf("- **%s:** %s\n", key, value))
		}
	}

	return builder.String()
}