package kanban

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auto-devs/auto-devs/config"
)

// Client talks to the Hermes Kanban dashboard plugin API.
type Client interface {
	// CommentTask posts a markdown comment on the kanban card.
	CommentTask(ctx context.Context, kanbanTaskID string, body string) error
	// UnblockTask moves the card back to "ready" so the Hermes dispatcher
	// spawns a worker for it.
	UnblockTask(ctx context.Context, kanbanTaskID string) error
	// Enabled reports whether the feature is configured.
	Enabled() bool
}

const (
	requestTimeout = 15 * time.Second
	commentAuthor  = "auto-devs"
)

type httpClient struct {
	enabled    bool
	baseURL    string
	token      string
	board      string
	httpClient *http.Client
}

// NewClient builds a Client from config. When the feature is disabled (or
// misconfigured) every method is a no-op returning nil.
func NewClient(cfg *config.HermesKanbanConfig) Client {
	enabled := cfg.Enabled && cfg.BaseURL != ""
	return &httpClient{
		enabled: enabled,
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		token:   cfg.Token,
		board:   cfg.Board,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// boardQuery returns the ?board= suffix pinning requests to the configured
// board. Without it the dashboard falls back to the user's "current board",
// which silently breaks callbacks when they switch boards.
func (c *httpClient) boardQuery() string {
	if c.board == "" {
		return ""
	}
	return "?board=" + url.QueryEscape(c.board)
}

func (c *httpClient) Enabled() bool {
	return c.enabled
}

func (c *httpClient) CommentTask(ctx context.Context, kanbanTaskID string, body string) error {
	if !c.enabled {
		return nil
	}

	endpoint := fmt.Sprintf("%s/api/plugins/kanban/tasks/%s/comments%s", c.baseURL, url.PathEscape(kanbanTaskID), c.boardQuery())
	payload := map[string]string{
		"body":   body,
		"author": commentAuthor,
	}
	return c.doJSON(ctx, http.MethodPost, endpoint, payload)
}

func (c *httpClient) UnblockTask(ctx context.Context, kanbanTaskID string) error {
	if !c.enabled {
		return nil
	}

	endpoint := fmt.Sprintf("%s/api/plugins/kanban/tasks/%s%s", c.baseURL, url.PathEscape(kanbanTaskID), c.boardQuery())
	payload := map[string]string{
		"status": "ready",
	}
	return c.doJSON(ctx, http.MethodPatch, endpoint, payload)
}

func (c *httpClient) doJSON(ctx context.Context, method, endpoint string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal kanban payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create kanban request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("kanban request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("kanban API %s %s returned %d: %s", method, endpoint, resp.StatusCode, string(respBody))
	}

	return nil
}
