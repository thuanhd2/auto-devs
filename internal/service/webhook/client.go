package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/auto-devs/auto-devs/config"
)

type Client struct {
	url        string
	httpClient *http.Client
}

func NewClient(cfg *config.WebhookConfig) *Client {
	timeout := 10 * time.Second
	if cfg != nil && cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Second
	}

	url := ""
	if cfg != nil {
		url = cfg.URL
	}

	return &Client{
		url: url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Send(ctx context.Context, payload any) error {
	if c.url == "" {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Printf("webhook returned non-2xx status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
