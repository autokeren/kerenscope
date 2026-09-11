package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	if err := Save(map[string]string{"SECTORS_API_KEY": "abc123", "KERENSCOPE_LLM_MODEL": "test-model"}); err != nil {
		t.Fatal(err)
	}
	loaded := Load()
	if loaded["SECTORS_API_KEY"] != "abc123" {
		t.Fatalf("expected key round-trip, got %v", loaded)
	}
	if loaded["KERENSCOPE_LLM_MODEL"] != "test-model" {
		t.Fatalf("expected model round-trip, got %v", loaded)
	}
	if v, ok := Get("SECTORS_API_KEY"); !ok || v != "abc123" {
		t.Fatalf("Get failed: %v %v", v, ok)
	}
	if _, ok := Get("MISSING"); ok {
		t.Fatal("missing key must not be found")
	}
	info, err := os.Stat(filepath.Join(dir, ".kerenscope", "config"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatal("config file must not be group/world readable")
	}
}

func TestSaveMergesExisting(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	Save(map[string]string{"A": "1"})
	Save(map[string]string{"B": "2"})
	loaded := Load()
	if loaded["A"] != "1" || loaded["B"] != "2" {
		t.Fatalf("merge failed: %v", loaded)
	}
}
