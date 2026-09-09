package sectors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const BaseURL = "https://api.sectors.app"

type Options struct {
	APIKey   string
	BaseURL  string
	NoCache  bool
	CacheDir string
	StateDir string
	HomeDir  string
}

type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
	cache   *cache
	ledger  *ledger
	nocache bool
}

func New(opts Options) (*Client, error) {
	key := opts.APIKey
	if key == "" {
		key = os.Getenv("SECTORS_API_KEY")
	}
	if key == "" {
		return nil, errors.New("SECTORS_API_KEY is not set — create one at sectors.app/api (API Key Management)")
	}
	home := opts.HomeDir
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot resolve home directory: %w", err)
		}
	}
	base := opts.BaseURL
	if base == "" {
		base = BaseURL
	}
	cacheDir := opts.CacheDir
	if cacheDir == "" {
		cacheDir = filepath.Join(home, ".cache", "kerenscope")
	}
	stateDir := opts.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(home, ".local", "state", "kerenscope")
	}
	c, err := newCache(cacheDir)
	if err != nil {
		return nil, err
	}
	l, err := newLedger(filepath.Join(stateDir, "credits.log"))
	if err != nil {
		return nil, err
	}
	return &Client{
		apiKey:  key,
		baseURL: strings.TrimRight(base, "/"),
		http:    &http.Client{Timeout: 90 * time.Second},
		cache:   c,
		ledger:  l,
		nocache: opts.NoCache,
	}, nil
}

func (c *Client) GetJSON(ctx context.Context, path string, params url.Values, ttl time.Duration, out any) error {
	endpoint := strings.TrimRight(path, "/")
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	if !c.nocache {
		if body, ok := c.cache.get(endpoint, ttl); ok {
			c.ledger.record(endpoint, true)
			return json.Unmarshal(body, out)
		}
	}
	full := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(params) > 0 {
		full += "?" + params.Encode()
	}
	body, err := c.fetch(ctx, full)
	if err != nil {
		return err
	}
	c.ledger.record(endpoint, false)
	if !c.nocache {
		if err := c.cache.put(endpoint, body); err != nil {
			fmt.Fprintf(os.Stderr, "warn: cache write failed: %v\n", err)
		}
	}
	return json.Unmarshal(body, out)
}

func (c *Client) LedgerSummary() LedgerSummary {
	return c.ledger.summarize()
}

func (c *Client) fetch(ctx context.Context, target string) (json.RawMessage, error) {
	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, backoffDelay(attempt, lastErr)); err != nil {
				return nil, err
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", c.apiKey)
		req.Header.Set("Accept", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = &Error{Kind: ErrNetwork, Cause: err}
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if readErr != nil {
				return nil, readErr
			}
			return json.RawMessage(body), nil
		}
		apiErr := &Error{Kind: ErrHTTP, Status: resp.StatusCode, Endpoint: target}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			apiErr.Detail = fmt.Sprintf("API key rejected (%d) — check SECTORS_API_KEY", resp.StatusCode)
			return nil, apiErr
		}
		if resp.StatusCode == http.StatusNotFound {
			apiErr.Detail = "endpoint or symbol not found"
			return nil, apiErr
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			apiErr.RetryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
			apiErr.Detail = strings.TrimSpace(string(body))
			lastErr = apiErr
			continue
		}
		apiErr.Detail = strings.TrimSpace(string(body))
		return nil, apiErr
	}
	return nil, lastErr
}

func backoffDelay(attempt int, err error) time.Duration {
	delay := time.Duration(1<<uint(attempt)) * time.Second
	if e, ok := err.(*Error); ok && e.RetryAfter > 0 {
		delay = e.RetryAfter
	}
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	return delay
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return 0
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type ErrorKind int

const (
	ErrNetwork ErrorKind = iota
	ErrHTTP
)

type Error struct {
	Kind       ErrorKind
	Status     int
	Endpoint   string
	Detail     string
	RetryAfter time.Duration
	Cause      error
}

func (e *Error) Error() string {
	switch e.Kind {
	case ErrNetwork:
		return fmt.Sprintf("sectors: network error: %v", e.Cause)
	default:
		if e.Detail != "" {
			return fmt.Sprintf("sectors: HTTP %d: %s", e.Status, e.Detail)
		}
		return fmt.Sprintf("sectors: HTTP %d (%s)", e.Status, e.Endpoint)
	}
}
