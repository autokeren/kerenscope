package sectors

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type cache struct {
	dir string
}

type cacheEntry struct {
	FetchedAt time.Time       `json:"fetched_at"`
	Body      json.RawMessage `json:"body"`
}

func newCache(dir string) (*cache, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &cache{dir: dir}, nil
}

func (c *cache) key(endpoint string) string {
	sum := sha256.Sum256([]byte(endpoint))
	return hex.EncodeToString(sum[:])
}

func (c *cache) get(endpoint string, ttl time.Duration) (json.RawMessage, bool) {
	if ttl <= 0 {
		return nil, false
	}
	data, err := os.ReadFile(filepath.Join(c.dir, c.key(endpoint)+".json"))
	if err != nil {
		return nil, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	if time.Since(entry.FetchedAt) > ttl {
		return nil, false
	}
	return entry.Body, true
}

func (c *cache) put(endpoint string, body json.RawMessage) error {
	entry := cacheEntry{FetchedAt: time.Now().UTC(), Body: body}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dir, c.key(endpoint)+".json"), data, 0o600)
}
