package labeler_test

import (
	"sort"
	"testing"

	"github.com/cronlens/internal/labeler"
)

func TestSet_And_Get(t *testing.T) {
	r := labeler.New()
	if err := r.Set("backup", "env", "prod"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lbls := r.Get("backup")
	if lbls["env"] != "prod" {
		t.Errorf("expected prod, got %s", lbls["env"])
	}
}

func TestSet_EmptyKey_ReturnsError(t *testing.T) {
	r := labeler.New()
	if err := r.Set("backup", "", "value"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestGet_UnknownJob_ReturnsNil(t *testing.T) {
	r := labeler.New()
	if lbls := r.Get("nonexistent"); lbls != nil {
		t.Errorf("expected nil, got %v", lbls)
	}
}

func TestDelete_RemovesKey(t *testing.T) {
	r := labeler.New()
	_ = r.Set("job", "team", "ops")
	r.Delete("job", "team")
	lbls := r.Get("job")
	if _, ok := lbls["team"]; ok {
		t.Error("expected key to be deleted")
	}
}

func TestJobsWithLabel_ReturnsMatching(t *testing.T) {
	r := labeler.New()
	_ = r.Set("job-a", "env", "prod")
	_ = r.Set("job-b", "env", "prod")
	_ = r.Set("job-c", "env", "staging")

	jobs := r.JobsWithLabel("env", "prod")
	sort.Strings(jobs)
	if len(jobs) != 2 || jobs[0] != "job-a" || jobs[1] != "job-b" {
		t.Errorf("unexpected jobs: %v", jobs)
	}
}

func TestJobsWithLabel_NoMatches(t *testing.T) {
	r := labeler.New()
	_ = r.Set("job-a", "env", "staging")
	if jobs := r.JobsWithLabel("env", "prod"); len(jobs) != 0 {
		t.Errorf("expected empty, got %v", jobs)
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	r := labeler.New()
	_ = r.Set("job", "k", "v")
	lbls := r.Get("job")
	lbls["k"] = "mutated"
	if r.Get("job")["k"] != "v" {
		t.Error("Get should return a copy, not a reference")
	}
}
