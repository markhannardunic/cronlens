package jobqueue

import (
	"testing"
)

func TestEnqueue_EmptyName_ReturnsError(t *testing.T) {
	q := New()
	if err := q.Enqueue("", 1); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestDequeue_EmptyQueue_ReturnsError(t *testing.T) {
	q := New()
	if _, err := q.Dequeue(); err == nil {
		t.Fatal("expected error when dequeuing from empty queue")
	}
}

func TestEnqueue_And_Len(t *testing.T) {
	q := New()
	_ = q.Enqueue("job-a", 1)
	_ = q.Enqueue("job-b", 2)
	if got := q.Len(); got != 2 {
		t.Fatalf("expected len 2, got %d", got)
	}
}

func TestDequeue_HighestPriorityFirst(t *testing.T) {
	q := New()
	_ = q.Enqueue("low", 1)
	_ = q.Enqueue("high", 10)
	_ = q.Enqueue("mid", 5)

	e, err := q.Dequeue()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.JobName != "high" {
		t.Fatalf("expected 'high', got %q", e.JobName)
	}

	e, _ = q.Dequeue()
	if e.JobName != "mid" {
		t.Fatalf("expected 'mid', got %q", e.JobName)
	}
}

func TestDequeue_SamePriority_FIFOOrder(t *testing.T) {
	q := New()
	_ = q.Enqueue("first", 3)
	_ = q.Enqueue("second", 3)

	e, _ := q.Dequeue()
	if e.JobName != "first" {
		t.Fatalf("expected 'first' (FIFO), got %q", e.JobName)
	}
}

func TestPeek_DoesNotRemoveEntries(t *testing.T) {
	q := New()
	_ = q.Enqueue("job-x", 1)
	_ = q.Enqueue("job-y", 2)

	entries := q.Peek()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries from Peek, got %d", len(entries))
	}
	if q.Len() != 2 {
		t.Fatalf("Peek should not remove entries; len=%d", q.Len())
	}
}

func TestDequeue_DrainsFully(t *testing.T) {
	q := New()
	_ = q.Enqueue("a", 1)
	_ = q.Enqueue("b", 1)
	_, _ = q.Dequeue()
	_, _ = q.Dequeue()
	if _, err := q.Dequeue(); err == nil {
		t.Fatal("expected error on empty queue after drain")
	}
}
