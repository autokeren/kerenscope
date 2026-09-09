package sectors

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LedgerEntry struct {
	Time     time.Time `json:"time"`
	Endpoint string    `json:"endpoint"`
	Cached   bool      `json:"cached"`
}

type LedgerSummary struct {
	Requests   int            `json:"requests"`
	Spent      int            `json:"spent"`
	CacheHits  int            `json:"cache_hits"`
	ByEndpoint map[string]int `json:"by_endpoint"`
}

type ledger struct {
	mu   sync.Mutex
	path string
}

func newLedger(path string) (*ledger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	return &ledger{path: path}, nil
}

func (l *ledger) record(endpoint string, cached bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	entry := LedgerEntry{Time: time.Now().UTC(), Endpoint: endpoint, Cached: cached}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	f.Write(append(data, '\n'))
}

func (l *ledger) summarize() LedgerSummary {
	l.mu.Lock()
	defer l.mu.Unlock()
	summary := LedgerSummary{ByEndpoint: map[string]int{}}
	f, err := os.Open(l.path)
	if err != nil {
		return summary
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var entry LedgerEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		summary.Requests++
		if entry.Cached {
			summary.CacheHits++
		} else {
			summary.Spent++
			summary.ByEndpoint[entry.Endpoint]++
		}
	}
	return summary
}

func ReadLedger(path string) LedgerSummary {
	l, err := newLedger(path)
	if err != nil {
		return LedgerSummary{ByEndpoint: map[string]int{}}
	}
	return l.summarize()
}

func DefaultLedgerPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "credits.log"
	}
	return filepath.Join(home, ".local", "state", "kerenscope", "credits.log")
}
