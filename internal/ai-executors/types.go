package aiexecutors

type PlanOutput struct {
	Type            string      `json:"type"`
	Message         PlanMessage `json:"message"`
	ParentToolUseID string      `json:"parent_tool_use_id"`
	SessionID       string      `json:"session_id"`
}

type PlanMessage struct {
	ID      string        `json:"id"`
	Type    string        `json:"type"`
	Role    string        `json:"role"`
	Model   string        `json:"model"`
	Content []PlanContent `json:"content"`
}

type PlanContent struct {
	Type  string           `json:"type"`
	ID    string           `json:"id"`
	Role  string           `json:"role"`
	Model string           `json:"model"`
	Input PlanContentInput `json:"input"`
}

type PlanContentInput struct {
	Plan string `json:"plan"`
}
