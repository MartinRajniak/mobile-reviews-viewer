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
	appIDs := []string{"app1", "app2"}
	logger := log.New(io.Discard, "", 0)
	poller := NewPoller(logger, appIDs, interval)

	if poller == nil {
		t.Fatal("NewPoller returned nil")
	}
	if poller.pollInterval != interval {
		t.Errorf("Expected poll interval %v, got %v", interval, poller.pollInterval)
	}
	if len(poller.appIDs) != len(appIDs) {
		t.Errorf("Expected %d appIDs, got %d", len(appIDs), len(poller.appIDs))
	}
	if poller.stopChan == nil {
		t.Error("stopChan should be initialized")
	}
}

func TestNewPoller_EmptyAppIDs(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	poller := NewPoller(logger, []string{}, 100*time.Millisecond)

	if poller == nil {
		t.Fatal("NewPoller returned nil")
	}
	if len(poller.appIDs) != 0 {
		t.Errorf("Expected 0 appIDs, got %d", len(poller.appIDs))
	}
}

func TestPoller_StartAndStop(t *testing.T) {
	interval := 100 * time.Millisecond
	logger := log.New(io.Discard, "", 0)
	appIDs := []string{"app1"}
	poller := NewPoller(logger, appIDs, interval)

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
	appIDs := []string{"app1", "app2"}
	poller := NewPoller(logger, appIDs, time.Second)

	poller.Start()
	time.Sleep(200 * time.Millisecond)
	poller.Stop()

	logOutput := buf.String()
	if !strings.Contains(logOutput, "Polling all apps concurrently...") {
		t.Error("Expected immediate poll on start, but log message not found")
	}
	// Verify each app was polled
	if !strings.Contains(logOutput, "Fetching reviews for app app1") {
		t.Error("Expected app1 to be polled")
	}
	if !strings.Contains(logOutput, "Fetching reviews for app app2") {
		t.Error("Expected app2 to be polled")
	}
}

func TestPoller_PeriodicPolling(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	interval := 100 * time.Millisecond
	appIDs := []string{"app1"}
	poller := NewPoller(logger, appIDs, interval)

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

func TestPoller_PollsAllApps(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	appIDs := []string{"app1", "app2", "app3"}
	poller := NewPoller(logger, appIDs, time.Second)

	poller.Start()
	time.Sleep(100 * time.Millisecond)
	poller.Stop()

	logOutput := buf.String()

	// Verify all apps were fetched
	for _, appID := range appIDs {
		expectedMsg := "Fetching reviews for app " + appID
		if !strings.Contains(logOutput, expectedMsg) {
			t.Errorf("Expected to find '%s' in log output", expectedMsg)
		}

		successMsg := "Successfully polled app " + appID
		if !strings.Contains(logOutput, successMsg) {
			t.Errorf("Expected to find '%s' in log output", successMsg)
		}
	}

	// Verify completion message
	if !strings.Contains(logOutput, "Poll complete in") {
		t.Error("Expected poll completion message")
	}
}

func TestPoller_EmptyAppIDsNoCrash(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	poller := NewPoller(logger, []string{}, 100*time.Millisecond)

	poller.Start()
	time.Sleep(150 * time.Millisecond)
	poller.Stop()

	logOutput := buf.String()
	// Should still log the polling message even with no apps
	if !strings.Contains(logOutput, "Polling all apps concurrently...") {
		t.Error("Expected polling message even with no apps")
	}
}

func TestPoller_MultipleStartCallsSafe(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	interval := 100 * time.Millisecond
	appIDs := []string{"app1"}
	poller := NewPoller(logger, appIDs, interval)

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
	interval := 100 * time.Millisecond
	appIDs := []string{"app1"}
	poller := NewPoller(logger, appIDs, interval)

	poller.Stop()
}

func TestPoller_MultipleStopCallsSafe(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	interval := 100 * time.Millisecond
	appIDs := []string{"app1"}
	poller := NewPoller(logger, appIDs, interval)

	poller.Start()
	time.Sleep(50 * time.Millisecond)
	poller.Stop()
	poller.Stop()
}

func TestPoller_RestartAfterStop(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	appIDs := []string{"app1"}
	poller := NewPoller(logger, appIDs, 100*time.Millisecond)

	poller.Start()
	time.Sleep(150 * time.Millisecond)
	poller.Stop()

	firstRunOutput := buf.String()
	firstCount := strings.Count(firstRunOutput, "Polling all apps concurrently...")

	buf.Reset()

	poller.Start()
	time.Sleep(150 * time.Millisecond)
	poller.Stop()

	secondRunOutput := buf.String()
	secondCount := strings.Count(secondRunOutput, "Polling all apps concurrently...")

	if firstCount < 2 {
		t.Errorf("First run: expected at least 2 polls, got %d", firstCount)
	}

	if secondCount < 2 {
		t.Errorf("Second run after restart: expected at least 2 polls, got %d", secondCount)
	}
}

func TestPoller_ConcurrentAppPolling(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	// Create many apps to increase likelihood of concurrent execution
	appIDs := []string{"app1", "app2", "app3", "app4", "app5"}
	poller := NewPoller(logger, appIDs, time.Second)

	start := time.Now()
	poller.Start()
	time.Sleep(200 * time.Millisecond) // Give time for concurrent execution
	poller.Stop()
	elapsed := time.Since(start)

	logOutput := buf.String()
	
	// All apps should be polled
	for _, appID := range appIDs {
		if !strings.Contains(logOutput, "Fetching reviews for app "+appID) {
			t.Errorf("App %s was not polled", appID)
		}
	}
	
	// If truly concurrent, should complete much faster than sequential
	// (5 apps sequentially would take 5x the time of one app)
	// This is a soft check - concurrent execution should be reasonably fast
	if elapsed > 1*time.Second {
		t.Logf("Warning: Polling took %v, may not be running concurrently", elapsed)
	}
}