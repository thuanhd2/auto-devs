package kanban

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auto-devs/auto-devs/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(baseURL string) Client {
	return NewClient(&config.HermesKanbanConfig{
		Enabled: true,
		BaseURL: baseURL,
		Token:   "test-token",
		Board:   "autodevs",
	})
}

func TestCommentTask_Success(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]string

	var gotBoard string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotBoard = r.URL.Query().Get("board")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	err := client.CommentTask(context.Background(), "card-123", "[auto-devs] status=DONE")

	require.NoError(t, err)
	assert.Equal(t, "/api/plugins/kanban/tasks/card-123/comments", gotPath)
	assert.Equal(t, "Bearer test-token", gotAuth)
	assert.Equal(t, "autodevs", gotBoard)
	assert.Equal(t, "[auto-devs] status=DONE", gotBody["body"])
	assert.Equal(t, "auto-devs", gotBody["author"])
}

func TestUnblockTask_Success(t *testing.T) {
	var gotPath, gotMethod, gotBoard string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBoard = r.URL.Query().Get("board")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"task": {}}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	err := client.UnblockTask(context.Background(), "card-123")

	require.NoError(t, err)
	assert.Equal(t, "/api/plugins/kanban/tasks/card-123", gotPath)
	assert.Equal(t, http.MethodPatch, gotMethod)
	assert.Equal(t, "autodevs", gotBoard)
	assert.Equal(t, "ready", gotBody["status"])
}

func TestClient_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"detail": "invalid token"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	err := client.CommentTask(context.Background(), "card-123", "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")

	err = client.UnblockTask(context.Background(), "card-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestClient_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	err := client.CommentTask(context.Background(), "card-123", "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestClient_Disabled(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
	}))
	defer server.Close()

	client := NewClient(&config.HermesKanbanConfig{
		Enabled: false,
		BaseURL: server.URL,
		Token:   "test-token",
	})

	assert.False(t, client.Enabled())
	assert.NoError(t, client.CommentTask(context.Background(), "card-123", "hello"))
	assert.NoError(t, client.UnblockTask(context.Background(), "card-123"))
	assert.Equal(t, 0, requestCount)
}

func TestClient_EnabledWithoutBaseURL(t *testing.T) {
	client := NewClient(&config.HermesKanbanConfig{
		Enabled: true,
		BaseURL: "",
	})
	assert.False(t, client.Enabled())
	assert.NoError(t, client.CommentTask(context.Background(), "card-123", "hello"))
}
