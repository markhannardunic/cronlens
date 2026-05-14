package webhook_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronlens/internal/job"
	"github.com/cronlens/internal/webhook"
)

func makeRun(name string, success bool, errMsg string) job.Run {
	r := job.NewRun(name)
	r.Finish(success, errMsg)
	return r
}

func TestDispatcher_NoConfigsNoError(t *testing.T) {
	d := webhook.New(nil)
	r := makeRun("backup", false, "disk full")
	if err := d.Notify(r); err != nil {
		t.Fatalf("expected no error with no configs, got %v", err)
	}
}

func TestDispatcher_SendsOnFailureEvent(t *testing.T) {
	var received webhook.Payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := webhook.Config{URL: ts.URL, Events: []string{"failure"}}
	d := webhook.New([]webhook.Config{cfg})

	r := makeRun("nightly", false, "timeout")
	if err := d.Notify(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.JobName != "nightly" {
		t.Errorf("expected job_name=nightly, got %q", received.JobName)
	}
	if received.Status != "failure" {
		t.Errorf("expected status=failure, got %q", received.Status)
	}
}

func TestDispatcher_SkipsSuccessWhenEventIsFailure(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := webhook.Config{URL: ts.URL, Events: []string{"failure"}}
	d := webhook.New([]webhook.Config{cfg})

	r := makeRun("sync", true, "")
	if err := d.Notify(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("webhook should not have been called for a successful run")
	}
}

func TestDispatcher_AllEventFiresOnSuccess(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := webhook.Config{URL: ts.URL, Events: []string{"all"}}
	d := webhook.New([]webhook.Config{cfg})

	r := makeRun("report", true, "")
	if err := d.Notify(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("webhook should have been called for 'all' event")
	}
}

func TestDispatcher_CustomHeaders(t *testing.T) {
	var gotToken string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("X-Api-Token")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := webhook.Config{
		URL:     ts.URL,
		Events:  []string{"all"},
		Headers: map[string]string{"X-Api-Token": "secret123"},
	}
	d := webhook.New([]webhook.Config{cfg})

	r := makeRun("ping", true, "")
	_ = d.Notify(r)
	if gotToken != "secret123" {
		t.Errorf("expected token secret123, got %q", gotToken)
	}
}

func TestDispatcher_ErrorOnBadStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	cfg := webhook.Config{URL: ts.URL, Events: []string{"all"}}
	d := webhook.New([]webhook.Config{cfg})

	r := makeRun("job", false, "err")
	if err := d.Notify(r); err == nil {
		t.Error("expected error on 500 response")
	}
}

func TestDispatcher_PayloadDuration(t *testing.T) {
	var received webhook.Payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := webhook.Config{URL: ts.URL, Events: []string{"all"}}
	d := webhook.New([]webhook.Config{cfg})

	run := job.NewRun("timed")
	time.Sleep(2 * time.Millisecond)
	run.Finish(true, "")

	_ = d.Notify(run)
	if received.Duration <= 0 {
		t.Errorf("expected positive duration, got %v", received.Duration)
	}
}
