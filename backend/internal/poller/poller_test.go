package poller

import (
	"backend/internal/testutil"
	"io"
	"log"
	"strings"
	"testing"
	"time"
)

func TestNewPoller(t *testing.T) {
	interval := 100 * time.Millisecond
	logger := log.New(io.Discard, "", 0)
	poller := NewPoller(logger, interval)

	if poller == nil {
		t.Fatal("NewPoller returned nil")
	}

	if poller.pollInterval != interval {
		t.Errorf("Expected poll interval %v, got %v", interval, poller.pollInterval)
	}

	if poller.stopChan == nil {
		t.Error("stopChan should be initialized")
	}
}

func TestPoller_StartAndStop(t *testing.T) {
	interval := 100 * time.Millisecond
	logger := log.New(io.Discard, "", 0)
	poller := NewPoller(logger, interval)

	poller.Start()
	time.Sleep(50 * time.Millisecond)
	
	done := make(chan struct{})
	go func() {
		poller.Stop()
		close(done)
	}()

	// Wait for Stop to complete with timeout
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not complete within timeout")
	}
}

func TestPoller_ImmediatePollOnStart(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	poller := NewPoller(logger, time.Second)

	poller.Start()
	// Give it a moment to execute the immediate poll
	time.Sleep(50 * time.Millisecond)
	poller.Stop()

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Polling all apps concurrently...") {
		t.Error("Expected immediate poll on start, but log message not found")
	}
}

func TestPoller_PeriodicPolling(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	interval := 100 * time.Millisecond
	poller := NewPoller(logger, interval)

	poller.Start()
	// Wait for multiple poll cycles
	time.Sleep(350 * time.Millisecond)
	poller.Stop()

	logOutput := buf.String()
	count := strings.Count(logOutput, "Polling all apps concurrently...")

	// Should have at least 3 polls: immediate + 2-3 periodic
	if count < 3 {
		t.Errorf("Expected at least 3 polls, got %d", count)
	}
}

func TestPoller_MultipleStartCallsSafe(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	poller := NewPoller(logger, 100*time.Millisecond)
	
	// Start multiple times - should only start once
	poller.Start()
	poller.Start()
	poller.Start()
	
	time.Sleep(250 * time.Millisecond)
	poller.Stop()
	
	logOutput := buf.String()
	count := strings.Count(logOutput, "Polling all apps concurrently...")
	
	// Should be ~3 polls (1 immediate + 2 periodic), not 9 (if 3 goroutines started)
	if count > 6 {
		t.Errorf("Multiple Start() calls should not create multiple goroutines. Got %d polls", count)
	}
}


func TestPoller_StopWithoutStart(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	poller := NewPoller(logger, 100*time.Millisecond)
	
	poller.Stop()
}

func TestPoller_MultipleStopCallsSafe(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	poller := NewPoller(logger, 100*time.Millisecond)
	
	poller.Start()
	time.Sleep(50 * time.Millisecond)
	poller.Stop()
	poller.Stop()
}
