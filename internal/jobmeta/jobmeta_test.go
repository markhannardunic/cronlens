package jobmeta_test

import (
	"testing"

	"github.com/cronlens/internal/jobmeta"
)

func TestSet_And_Get(t *testing.T) {
	r := jobmeta.New()
	if err := r.Set("backup", "owner", "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := r.Get("backup")
	if m == nil {
		t.Fatal("expected metadata map, got nil")
	}
	if m["owner"] != "alice" {
		t.Errorf("expected owner=alice, got %q", m["owner"])
	}
}

func TestSet_EmptyJobName_ReturnsError(t *testing.T) {
	r := jobmeta.New()
	if err := r.Set("", "owner", "alice"); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestSet_EmptyKey_ReturnsError(t *testing.T) {
	r := jobmeta.New()
	if err := r.Set("backup", "", "value"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestGet_UnknownJob_ReturnsNil(t *testing.T) {
	r := jobmeta.New()
	if m := r.Get("nonexistent"); m != nil {
		t.Errorf("expected nil, got %v", m)
	}
}

func TestDelete_RemovesKey(t *testing.T) {
	r := jobmeta.New()
	_ = r.Set("sync", "team", "ops")
	_ = r.Set("sync", "owner", "bob")
	r.Delete("sync", "team")
	m := r.Get("sync")
	if _, ok := m["team"]; ok {
		t.Error("expected 'team' key to be deleted")
	}
	if m["owner"] != "bob" {
		t.Errorf("expected owner=bob after delete, got %q", m["owner"])
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	r := jobmeta.New()
	_ = r.Set("export", "desc", "nightly export")
	m := r.Get("export")
	m["desc"] = "mutated"
	original := r.Get("export")
	if original["desc"] != "nightly export" {
		t.Error("Get should return a copy, not a reference")
	}
}

func TestJobNames_ReturnsJobsWithEntries(t *testing.T) {
	r := jobmeta.New()
	_ = r.Set("jobA", "k", "v")
	_ = r.Set("jobB", "k", "v")
	names := r.JobNames()
	if len(names) != 2 {
		t.Errorf("expected 2 job names, got %d", len(names))
	}
}

func TestJobNames_ExcludesEmptyMaps(t *testing.T) {
	r := jobmeta.New()
	_ = r.Set("jobA", "k", "v")
	r.Delete("jobA", "k")
	names := r.JobNames()
	if len(names) != 0 {
		t.Errorf("expected 0 job names after delete, got %d", len(names))
	}
}
