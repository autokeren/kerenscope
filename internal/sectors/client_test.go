package sectors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCachePutGetExpiry(t *testing.T) {
	c, err := newCache(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	body := json.RawMessage(`{"ok":true}`)
	if err := c.put("/v2/test/", body); err != nil {
		t.Fatal(err)
	}
	if got, ok := c.get("/v2/test/", time.Hour); !ok || string(got) != string(body) {
		t.Fatalf("expected cache hit with %s, ok=%v", string(got), ok)
	}
	if _, ok := c.get("/v2/test/", 0); ok {
		t.Fatal("ttl of zero must never hit cache")
	}
	if _, ok := c.get("/v2/other/", time.Hour); ok {
		t.Fatal("unknown key must miss")
	}
}

func TestLedgerRecordSummarize(t *testing.T) {
	l, err := newLedger(t.TempDir() + "/credits.log")
	if err != nil {
		t.Fatal(err)
	}
	l.record("/v2/a/", false)
	l.record("/v2/a/", false)
	l.record("/v2/a/", true)
	s := l.summarize()
	if s.Spent != 2 || s.CacheHits != 1 || s.Requests != 3 {
		t.Fatalf("unexpected summary: %+v", s)
	}
	if s.ByEndpoint["/v2/a/"] != 2 {
		t.Fatalf("expected 2 spent on endpoint, got %d", s.ByEndpoint["/v2/a/"])
	}
}

func TestClientSendsAuthAndFetches(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"symbol":"TEST.JK","company_name":"Test"}`))
	}))
	defer srv.Close()
	c, err := New(Options{APIKey: "secret", BaseURL: srv.URL, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := c.GetJSON(context.Background(), "/v2/company/report/TEST/", nil, time.Hour, &out); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "secret" {
		t.Fatalf("expected Authorization header to carry the raw key, got %q", gotAuth)
	}
	if out["symbol"] != "TEST.JK" {
		t.Fatalf("unexpected payload: %v", out)
	}
}

func TestClientRetriesOn429(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"details":"rate limited"}`))
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	c, err := New(Options{APIKey: "k", BaseURL: srv.URL, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := c.GetJSON(context.Background(), "/v2/x/", nil, time.Hour, &out); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls after one 429, got %d", calls)
	}
}
