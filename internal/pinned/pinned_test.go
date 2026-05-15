package pinned_test

import (
	"testing"

	"github.com/cronlens/internal/pinned"
)

func TestPin_And_IsPinned(t *testing.T) {
	r := pinned.New()
	if err := r.Pin("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.IsPinned("backup") {
		t.Error("expected backup to be pinned")
	}
}

func TestPin_Duplicate_ReturnsError(t *testing.T) {
	r := pinned.New()
	_ = r.Pin("backup")
	if err := r.Pin("backup"); err == nil {
		t.Error("expected error for duplicate pin")
	}
}

func TestUnpin_ClearsPinnedState(t *testing.T) {
	r := pinned.New()
	_ = r.Pin("cleanup")
	if err := r.Unpin("cleanup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.IsPinned("cleanup") {
		t.Error("expected cleanup to be unpinned")
	}
}

func TestUnpin_NotPinned_ReturnsError(t *testing.T) {
	r := pinned.New()
	if err := r.Unpin("ghost"); err == nil {
		t.Error("expected error when unpinning unknown job")
	}
}

func TestList_ReturnsAllPinnedEntries(t *testing.T) {
	r := pinned.New()
	_ = r.Pin("alpha")
	_ = r.Pin("beta")
	list := r.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
}

func TestIsPinned_UnknownJob_ReturnsFalse(t *testing.T) {
	r := pinned.New()
	if r.IsPinned("unknown") {
		t.Error("expected false for unknown job")
	}
}
