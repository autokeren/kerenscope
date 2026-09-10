package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/autokeren/kerenscope/internal/sectors"
)

func testClient(t *testing.T, handler http.HandlerFunc) *sectors.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sectors.New(sectors.Options{
		APIKey:   "k",
		BaseURL:  srv.URL,
		HomeDir:  t.TempDir(),
		CacheDir: t.TempDir() + "/cache",
		StateDir: t.TempDir() + "/state",
		NoCache:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestSectorsAPIToolHappyPath(t *testing.T) {
	var gotPath, gotQ string
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQ = r.URL.Query().Get("limit")
		w.Write([]byte(`[{"ok":true}]`))
	})
	tool := SectorsAPITool{Client: client}
	res := tool.Run(context.Background(), map[string]any{
		"path":   "/v2/filings/",
		"params": map[string]any{"limit": float64(5), "symbol": "BBRI"},
	})
	if !res.OK {
		t.Fatalf("expected ok, got error: %s", res.Error)
	}
	if gotPath != "/v2/filings/" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQ != "5" {
		t.Fatalf("expected limit=5, got %q", gotQ)
	}
}

func TestSectorsAPIToolRejectsBadInput(t *testing.T) {
	tool := SectorsAPITool{Client: testClient(t, func(w http.ResponseWriter, r *http.Request) {})}
	cases := []struct {
		name string
		args map[string]any
	}{
		{"full URL", map[string]any{"path": "https://evil.example.com/v2/x/"}},
		{"with query", map[string]any{"path": "/v2/x/?limit=1"}},
		{"not v2", map[string]any{"path": "/v1/companies/"}},
		{"empty", map[string]any{"path": ""}},
	}
	for _, c := range cases {
		res := tool.Run(context.Background(), c.args)
		if res.OK {
			t.Fatalf("%s: expected rejection", c.name)
		}
	}
}

func TestSectorsAPIToolCacheReusesResult(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write([]byte(`{"n":1}`))
	}))
	t.Cleanup(srv.Close)
	client, err := sectors.New(sectors.Options{
		APIKey:  "k",
		BaseURL: srv.URL,
		HomeDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	tool := &SectorsAPITool{Client: client}
	first := tool.Run(context.Background(), map[string]any{"path": "v2/subsectors/"})
	if !first.OK {
		t.Fatal(first.Error)
	}
	second := tool.Run(context.Background(), map[string]any{"path": "v2/subsectors/"})
	if !second.OK {
		t.Fatal(second.Error)
	}
	if calls != 1 {
		t.Fatalf("expected cache to prevent second HTTP call, got %d calls", calls)
	}
	if _, ok := second.Data.(json.RawMessage); !ok {
		t.Fatalf("expected raw JSON data, got %T", second.Data)
	}
}

func TestSectorsAPIGuardTranslatesDaysAndRejectsUnknown(t *testing.T) {
	var gotParams url.Values
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotParams = r.URL.Query()
		w.Write([]byte(`{"ok":true}`))
	})
	tool := SectorsAPITool{Client: client}

	res := tool.Run(context.Background(), map[string]any{
		"path":   "/v2/news/",
		"params": map[string]any{"symbols": "BBRI", "days": float64(30), "limit": float64(10)},
	})
	if !res.OK {
		t.Fatalf("expected ok, got: %s", res.Error)
	}
	if gotParams.Get("symbols") != "BBRI" {
		t.Fatalf("expected symbols param, got %v", gotParams)
	}
	if gotParams.Get("start") == "" || gotParams.Get("end") == "" {
		t.Fatalf("expected days translated to start/end, got %v", gotParams)
	}
	if gotParams.Get("days") != "" {
		t.Fatal("days must not be forwarded raw")
	}

	res = tool.Run(context.Background(), map[string]any{
		"path":   "/v2/news/",
		"params": map[string]any{"bogus_param": "x"},
	})
	if res.OK {
		t.Fatal("expected unknown param rejection")
	}
	if !strings.Contains(res.Error, "bogus_param") || !strings.Contains(res.Error, "allowed") {
		t.Fatalf("expected helpful rejection message, got: %s", res.Error)
	}

	res = tool.Run(context.Background(), map[string]any{"path": "/v2/not/a/real/endpoint/"})
	if res.OK {
		t.Fatal("expected unknown endpoint rejection")
	}
}

func TestSectorsAPIGuardCleanErrorMessage(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Unsupported query parameter(s): days, symbols"}`))
	})
	// 'where' is schema-valid for /v2/companies/, so the guard lets it through;
	// the mock server then rejects it — cleanAPIError must strip the raw JSON.
	tool := SectorsAPITool{Client: client}
	res := tool.Run(context.Background(), map[string]any{"path": "/v2/companies/", "params": map[string]any{"where": "sector = 'banks'"}})
	if res.OK {
		t.Fatal("expected error")
	}
	if strings.Contains(res.Error, `{"error"`) {
		t.Fatalf("error message should be cleaned, got: %s", res.Error)
	}
	if !strings.Contains(res.Error, "Unsupported query parameter(s)") {
		t.Fatalf("expected underlying message preserved, got: %s", res.Error)
	}
}
