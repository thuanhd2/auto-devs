package aiexecutors

import (
	"strings"
	"testing"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/service/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReasonixExecutor_GetPlanningCommand(t *testing.T) {
	e := NewReasonixExecutor()
	command, prompt, envVars, err := e.GetPlanningCommand(t.Context(), &entity.Task{
		Title:       "Add auth",
		Description: "Implement login",
	})
	require.NoError(t, err)
	assert.Contains(t, command, "reasonix run -p")
	assert.Contains(t, command, "--output-format=stream-json")
	assert.Contains(t, prompt, "Add auth")
	assert.Empty(t, envVars)
}

func TestReasonixExecutor_GetImplementationCommand(t *testing.T) {
	e := NewReasonixExecutor()
	command, prompt, _, err := e.GetImplementationCommand(t.Context(), &entity.Task{
		Title:       "Add auth",
		Description: "Implement login",
		Plans:       []entity.Plan{{Content: "Step 1: ..."}},
	})
	require.NoError(t, err)
	assert.Contains(t, command, "reasonix run -p")
	assert.Contains(t, command, "--permission-mode=bypassPermissions")
	assert.Contains(t, command, "--output-format=stream-json")
	assert.Contains(t, prompt, "Step 1: ...")
}

func TestReasonixExecutor_ParseOutputToPlan(t *testing.T) {
	e := NewReasonixExecutor()
	output := strings.Join([]string{
		`{"type":"event","kind":"run_start"}`,
		`{"type":"result","subtype":"success","is_error":false,"num_turns":2,"result":"# Plan\n\n1. Add auth module\n2. Wire it up"}`,
	}, "\n")

	plan, err := e.ParseOutputToPlan(output)
	require.NoError(t, err)
	assert.Equal(t, "# Plan\n\n1. Add auth module\n2. Wire it up", plan)
}

func TestReasonixExecutor_ParseOutputToPlan_NoResult(t *testing.T) {
	e := NewReasonixExecutor()
	output := strings.Join([]string{
		`{"type":"event","kind":"run_start"}`,
		`{"type":"event","kind":"run_done"}`,
	}, "\n")

	_, err := e.ParseOutputToPlan(output)
	require.Error(t, err)
}

func TestReasonixExecutor_ParseOutputToLogs(t *testing.T) {
	e := NewReasonixExecutor()
	output := strings.Join([]string{
		`{"type":"system","subtype":"init","session_id":"s1"}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_1","name":"Read"}]}}`,
		`plain text line`,
		`{"type":"result","subtype":"success","result":"done"}`,
	}, "\n")

	logs := e.ParseOutputToLogs(output)
	require.Len(t, logs, 4)

	assert.Equal(t, "system", logs[0].LogType)
	assert.Equal(t, "stdout", logs[0].Source)

	assert.Equal(t, "assistant", logs[1].LogType)
	assert.Equal(t, "toolu_1", logs[1].ToolUseID)
	assert.Equal(t, "Read", logs[1].ToolName)

	assert.Equal(t, "", logs[2].LogType)
	assert.Equal(t, "plain text line", logs[2].Message)

	assert.Equal(t, "result", logs[3].LogType)
}

func TestReasonixExecutor_ImplementsAiCodingCli(t *testing.T) {
	// compile-time check: *ReasonixExecutor must satisfy the AiCodingCli interface
	var _ ai.AiCodingCli = NewReasonixExecutor()
}
