package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/go-github/v74/github"
)

// RateLimitInfo holds information about current rate limits
type RateLimitInfo struct {
	Limit     int
	Remaining int
	ResetAt   time.Time
}

// RateLimiter handles GitHub API rate limiting
type RateLimiter struct {
	mu        sync.RWMutex
	limitInfo RateLimitInfo
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limitInfo: RateLimitInfo{
			Limit:     5000, // Default GitHub rate limit for authenticated users
			Remaining: 5000,
			ResetAt:   time.Now().Add(time.Hour),
		},
		lastReset: time.Now(),
	}
}

// Wait blocks until the rate limit allows the request
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.RLock()
	info := rl.limitInfo
	rl.mu.RUnlock()

	// Check if we need to reset the counter
	now := time.Now()
	if now.After(info.ResetAt) {
		rl.mu.Lock()
		if now.After(rl.limitInfo.ResetAt) {
			rl.limitInfo.Remaining = rl.limitInfo.Limit
			rl.limitInfo.ResetAt = now.Add(time.Hour)
			rl.lastReset = now
		}
		rl.mu.Unlock()
		return nil
	}

	// If we have remaining requests, proceed
	if info.Remaining > 0 {
		rl.mu.Lock()
		if rl.limitInfo.Remaining > 0 {
			rl.limitInfo.Remaining--
		}
		rl.mu.Unlock()
		return nil
	}

	// We need to wait until reset
	waitTime := time.Until(info.ResetAt)
	if waitTime <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		// Reset the counter
		rl.mu.Lock()
		rl.limitInfo.Remaining = rl.limitInfo.Limit
		rl.limitInfo.ResetAt = time.Now().Add(time.Hour)
		rl.lastReset = time.Now()
		rl.mu.Unlock()
		return nil
	}
}

// UpdateFromResponse updates rate limit info from HTTP response headers
func (rl *RateLimiter) UpdateFromResponse(resp *http.Response) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Parse rate limit headers
	if limitHeader := resp.Header.Get("X-RateLimit-Limit"); limitHeader != "" {
		if limit, err := strconv.Atoi(limitHeader); err == nil {
			rl.limitInfo.Limit = limit
		}
	}

	if remainingHeader := resp.Header.Get("X-RateLimit-Remaining"); remainingHeader != "" {
		if remaining, err := strconv.Atoi(remainingHeader); err == nil {
			rl.limitInfo.Remaining = remaining
		}
	}

	if resetHeader := resp.Header.Get("X-RateLimit-Reset"); resetHeader != "" {
		if resetUnix, err := strconv.ParseInt(resetHeader, 10, 64); err == nil {
			rl.limitInfo.ResetAt = time.Unix(resetUnix, 0)
		}
	}
}

// UpdateFromGitHubResponse updates rate limit info from GitHub API response
func (rl *RateLimiter) UpdateFromGitHubResponse(resp *github.Response) {
	if resp == nil {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Parse rate limit headers from GitHub response
	if resp.Rate.Limit > 0 {
		rl.limitInfo.Limit = resp.Rate.Limit
	}
	if resp.Rate.Remaining >= 0 {
		rl.limitInfo.Remaining = resp.Rate.Remaining
	}
	if !resp.Rate.Reset.IsZero() {
		rl.limitInfo.ResetAt = resp.Rate.Reset.Time
	}
}

// GetInfo returns current rate limit information
func (rl *RateLimiter) GetInfo() RateLimitInfo {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.limitInfo
}

// IsRateLimited returns true if we're currently rate limited
func (rl *RateLimiter) IsRateLimited() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limitInfo.Remaining <= 0 && time.Now().Before(rl.limitInfo.ResetAt)
}

// TimeUntilReset returns the duration until the rate limit resets
func (rl *RateLimiter) TimeUntilReset() time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return time.Until(rl.limitInfo.ResetAt)
}

// String returns a string representation of the rate limiter state
func (rl *RateLimiter) String() string {
	info := rl.GetInfo()
	return fmt.Sprintf("RateLimit: %d/%d remaining, resets at %s",
		info.Remaining, info.Limit, info.ResetAt.Format(time.RFC3339))
}
