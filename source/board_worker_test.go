//go:build windows

package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// A deliberately stalled child, not a camera worker. Test binaries only.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "--board-worker" && os.Getenv("LUMERIAN_TEST_STALLED_WORKER") == "1" {
		var req boardRequest
		_ = json.NewDecoder(os.Stdin).Decode(&req)
		d := newBoardDiagnostic(true)
		d.Policy = "s1-comprehensive-v1"
		d.ActiveStage = "usage"
		d.step("open_session", "ok", 0x2001, nil, time.Now())
		_ = json.NewEncoder(os.Stdout).Encode(boardReply{Kind: "suite_running", Diagnostic: d})
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func TestBoardWorkerCancellationPreservesEvidence(t *testing.T) {
	t.Setenv("LUMERIAN_TEST_STALLED_WORKER", "1")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	start := time.Now()
	out, e := boardRunWorkerContext(ctx, boardRequest{Suite: true}, func(boardReply) { cancel() })
	if e == nil || out.Kind != "complete" || out.Diagnostic.Outcome != "interrupted" || out.Diagnostic.ActiveStage != "usage" || time.Since(start) > 5*time.Second {
		t.Fatal("cancellation failed", out, e)
	}
	if len(out.Diagnostic.Steps) != 3 || out.Diagnostic.Steps[0].PTP != "0x2001" || out.Diagnostic.Steps[1].Code != "cancelled" || out.Diagnostic.Steps[2].Status != "unconfirmed" {
		t.Fatal("lost partial log", out)
	}
	if b, e := boardDiagnosticZIP(out.Diagnostic); e != nil || len(b) == 0 {
		t.Fatal("cancelled ZIP", e)
	}
}

func TestBoardWorkerTimeoutPreservesEvidence(t *testing.T) {
	t.Setenv("LUMERIAN_TEST_STALLED_WORKER", "1")
	start := time.Now()
	out, e := boardRunWorkerContext(context.Background(), boardRequest{Suite: true}, func(boardReply) {})
	if e == nil || time.Since(start) > 27*time.Second || time.Since(start) < 19*time.Second {
		t.Fatal("inactivity deadline", time.Since(start), e)
	}
	if out.Diagnostic.Outcome != "interrupted" || out.Diagnostic.ActiveStage != "usage" || len(out.Diagnostic.Steps) != 3 || out.Diagnostic.Steps[1].Code != "worker_incomplete" {
		t.Fatal("timeout evidence lost", out)
	}
}
func TestBoardSuitePostSessionAndCloseFailure(t *testing.T) {
	for _, closeOK := range []bool{true, false} {
		a, b, c := suiteCamera(), suiteCamera(), suiteCamera()
		c.usageRC = 0x2001
		if !closeOK {
			b.closeRC = 0x2002
		}
		devices := []boardTransport{a, b, c}
		opens := 0
		out := boardSuite(func(session bool) (boardTransport, error) { v := devices[opens]; opens++; return v, nil }, func(boardReply) {}, func(time.Duration) {})
		if closeOK {
			if !out.Valid || opens != 3 || c.closeCount != 1 || out.Diagnostic.Outcome != "completed" {
				t.Fatal(out, opens)
			}
		} else {
			if out.Valid || opens != 2 || out.Diagnostic.Outcome != "stopped" {
				t.Fatal("reopened after uncertain close", out, opens)
			}
		}
		count := 0
		for _, call := range b.calls {
			if call == "9103" {
				count++
			}
		}
		if count != 1 {
			t.Fatal("duplicate session close", b.calls)
		}
	}
}
