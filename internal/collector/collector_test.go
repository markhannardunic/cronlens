package collector_test

import (
	"errors"
	"testing"
	"time"

	"github.com/user/cronlens/internal/collector"
	"github.com/user/cronlens/internal/store"
)

func newCollector() *collector.Collector {
	s := store.New()
	return collector.New(s)
}

func TestCollector_RunFunc_Success(t *testing.T) {
	c := newCollector()
	before := time.Now()

	err := c.RunFunc("backup", func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	runs := c.Since(before)
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].JobName != "backup" {
		t.Errorf("expected job name 'backup', got %q", runs[0].JobName)
	}
	if !runs[0].Success {
		t.Errorf("expected run to be successful")
	}
}

func TestCollector_RunFunc_Failure(t *testing.T) {
	c := newCollector()
	before := time.Now()
	wantErr := errors.New("something went wrong")

	err := c.RunFunc("report", func() error {
		return wantErr
	})

	if err != wantErr {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}

	runs := c.Since(before)
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Success {
		t.Errorf("expected run to be a failure")
	}
	if runs[0].Err != wantErr {
		t.Errorf("expected error %v, got %v", wantErr, runs[0].Err)
	}
}

func TestCollector_RunCommand_Success(t *testing.T) {
	c := newCollector()
	before := time.Now()

	err := c.Run("echo-job", "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	runs := c.Since(before)
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if !runs[0].Success {
		t.Errorf("expected run to succeed")
	}
}
