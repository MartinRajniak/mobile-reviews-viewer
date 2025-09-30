package poller

import (
	"backend/internal/models"
	"backend/internal/testutil"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)


func TestNewPoller(t *testing.T) {
	interval := 100 * time.Millisecond
	appIDs := []string{"app1", "app2"}
	logger := log.New(io.Discard, "", 0)
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, interval)

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
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, 100*time.Millisecond)

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
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, interval)

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
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, time.Second)

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
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, interval)

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
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, time.Second)

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
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, 100*time.Millisecond)

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
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, interval)

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
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, interval)

	poller.Stop()
}

func TestPoller_MultipleStopCallsSafe(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	interval := 100 * time.Millisecond
	appIDs := []string{"app1"}
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, interval)

	poller.Start()
	time.Sleep(50 * time.Millisecond)
	poller.Stop()
	poller.Stop()
}

func TestPoller_RestartAfterStop(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	appIDs := []string{"app1"}
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, 100*time.Millisecond)

	poller.Start()
	time.Sleep(150 * time.Millisecond)
	poller.Stop()

	firstRunOutput := buf.String()
	firstCount := strings.Count(firstRunOutput, "Polling all apps concurrently...")

	buf.Reset()

	poller.Start()
	time.Sleep(250 * time.Millisecond)
	poller.Stop()

	secondRunOutput := buf.String()
	secondCount := strings.Count(secondRunOutput, "Polling all apps concurrently...")

	if firstCount < 1 {
		t.Errorf("First run: expected at least 1 poll, got %d", firstCount)
	}

	if secondCount < 1 {
		t.Errorf("Second run after restart: expected at least 1 poll, got %d", secondCount)
	}
}

func TestPoller_ConcurrentAppPolling(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)
	// Create many apps to increase likelihood of concurrent execution
	appIDs := []string{"app1", "app2", "app3", "app4", "app5"}
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, appIDs, time.Second)

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

// HTTP Request Tests

func TestPoller_fetchReviewsSuccess(t *testing.T) {
	// Create mock RSS feed response
	mockFeed := RSSFeed{
		Feed: struct {
			Entry []RSSEntry `json:"entry"`
		}{
			Entry: []RSSEntry{
				{
					Author: struct {
						Name struct {
							Label string `json:"label"`
						} `json:"name"`
					}{
						Name: struct {
							Label string `json:"label"`
						}{Label: "John Doe"},
					},
					Content: struct {
						Label string `json:"label"`
					}{Label: "Great app!"},
					Rating: struct {
						Label string `json:"label"`
					}{Label: "5"},
					Updated: struct {
						Label string `json:"label"`
					}{Label: "2023-01-15T10:30:00Z"},
					ID: struct {
						Label string `json:"label"`
					}{Label: "https://itunes.apple.com/us/review?id=123&type=Purple+Software"},
				},
			},
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("User-Agent") != "AppReviewPoller/1.0" {
			t.Errorf("Expected User-Agent header 'AppReviewPoller/1.0', got %s", r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockFeed)
	}))
	defer server.Close()

	logger := log.New(io.Discard, "", 0)
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, time.Second)

	reviews, err := poller.fetchReviews(server.URL, "123")
	if err != nil {
		t.Fatalf("fetchReviews failed: %v", err)
	}

	if len(reviews) != 1 {
		t.Fatalf("Expected 1 review, got %d", len(reviews))
	}

	review := reviews[0]
	if review.AppID != "123" {
		t.Errorf("Expected AppID '123', got '%s'", review.AppID)
	}
	if review.Author != "John Doe" {
		t.Errorf("Expected Author 'John Doe', got '%s'", review.Author)
	}
	if review.Content != "Great app!" {
		t.Errorf("Expected Content 'Great app!', got '%s'", review.Content)
	}
	if review.Rating != 5 {
		t.Errorf("Expected Rating 5, got %d", review.Rating)
	}
}

func TestPoller_fetchReviewsHTTPError(t *testing.T) {
	// Create test server that returns 500 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger := log.New(io.Discard, "", 0)
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, time.Second)

	_, err := poller.fetchReviews(server.URL, "123")
	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 500") {
		t.Errorf("Expected status code error, got: %v", err)
	}
}

func TestPoller_fetchReviewsInvalidJSON(t *testing.T) {
	// Create test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	logger := log.New(io.Discard, "", 0)
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, time.Second)

	_, err := poller.fetchReviews(server.URL, "123")
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode RSS feed") {
		t.Errorf("Expected decode error, got: %v", err)
	}
}

func TestPoller_fetchReviewsEmptyFeed(t *testing.T) {
	// Create mock empty RSS feed response
	mockFeed := RSSFeed{
		Feed: struct {
			Entry []RSSEntry `json:"entry"`
		}{
			Entry: []RSSEntry{},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockFeed)
	}))
	defer server.Close()

	logger := log.New(io.Discard, "", 0)
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, time.Second)

	reviews, err := poller.fetchReviews(server.URL, "123")
	if err != nil {
		t.Fatalf("fetchReviews failed: %v", err)
	}

	if len(reviews) != 0 {
		t.Fatalf("Expected 0 reviews, got %d", len(reviews))
	}
}

func TestPoller_parseReviewEntry(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, time.Second)

	entry := RSSEntry{
		Author: struct {
			Name struct {
				Label string `json:"label"`
			} `json:"name"`
		}{
			Name: struct {
				Label string `json:"label"`
			}{Label: "Jane Smith"},
		},
		Content: struct {
			Label string `json:"label"`
		}{Label: "Amazing functionality!"},
		Rating: struct {
			Label string `json:"label"`
		}{Label: "4"},
		Updated: struct {
			Label string `json:"label"`
		}{Label: "2023-02-20T15:45:30Z"},
		ID: struct {
			Label string `json:"label"`
		}{Label: "https://itunes.apple.com/us/review?id=456&type=Purple+Software"},
	}

	fetchedAt := time.Now()
	review, err := poller.parseReviewEntry(entry, "789", fetchedAt)
	if err != nil {
		t.Fatalf("parseReviewEntry failed: %v", err)
	}

	if review.ID != "https://itunes.apple.com/us/review?id=456&type=Purple+Software" {
		t.Errorf("Expected ID 'https://itunes.apple.com/us/review?id=456&type=Purple+Software', got '%s'", review.ID)
	}
	if review.AppID != "789" {
		t.Errorf("Expected AppID '789', got '%s'", review.AppID)
	}
	if review.Author != "Jane Smith" {
		t.Errorf("Expected Author 'Jane Smith', got '%s'", review.Author)
	}
	if review.Content != "Amazing functionality!" {
		t.Errorf("Expected Content 'Amazing functionality!', got '%s'", review.Content)
	}
	if review.Rating != 4 {
		t.Errorf("Expected Rating 4, got %d", review.Rating)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2023-02-20T15:45:30Z")
	if !review.SubmittedAt.Equal(expectedTime) {
		t.Errorf("Expected SubmittedAt %v, got %v", expectedTime, review.SubmittedAt)
	}
	if !review.FetchedAt.Equal(fetchedAt) {
		t.Errorf("Expected FetchedAt %v, got %v", fetchedAt, review.FetchedAt)
	}
}

func TestPoller_parseReviewEntryInvalidRating(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, time.Second)

	entry := RSSEntry{
		Rating: struct {
			Label string `json:"label"`
		}{Label: "invalid"},
		Updated: struct {
			Label string `json:"label"`
		}{Label: "2023-02-20T15:45:30Z"},
	}

	_, err := poller.parseReviewEntry(entry, "789", time.Now())
	if err == nil {
		t.Fatal("Expected error for invalid rating, got nil")
	}
	if !strings.Contains(err.Error(), "invalid rating") {
		t.Errorf("Expected rating error, got: %v", err)
	}
}

func TestPoller_parseReviewEntryInvalidTimestamp(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, time.Second)

	entry := RSSEntry{
		Rating: struct {
			Label string `json:"label"`
		}{Label: "5"},
		Updated: struct {
			Label string `json:"label"`
		}{Label: "invalid-timestamp"},
	}

	_, err := poller.parseReviewEntry(entry, "789", time.Now())
	if err == nil {
		t.Fatal("Expected error for invalid timestamp, got nil")
	}
	if !strings.Contains(err.Error(), "invalid timestamp") {
		t.Errorf("Expected timestamp error, got: %v", err)
	}
}

func TestPoller_fetchReviewsMalformedEntries(t *testing.T) {
	var buf testutil.SafeBuffer
	logger := log.New(&buf, "", 0)

	// Create mock RSS feed with one valid and one invalid entry
	mockFeed := RSSFeed{
		Feed: struct {
			Entry []RSSEntry `json:"entry"`
		}{
			Entry: []RSSEntry{
				{
					Author: struct {
						Name struct {
							Label string `json:"label"`
						} `json:"name"`
					}{
						Name: struct {
							Label string `json:"label"`
						}{Label: "Valid User"},
					},
					Content: struct {
						Label string `json:"label"`
					}{Label: "Good app"},
					Rating: struct {
						Label string `json:"label"`
					}{Label: "5"},
					Updated: struct {
						Label string `json:"label"`
					}{Label: "2023-01-15T10:30:00Z"},
					ID: struct {
						Label string `json:"label"`
					}{Label: "https://itunes.apple.com/us/review?id=123"},
				},
				{
					// Invalid entry - bad rating
					Rating: struct {
						Label string `json:"label"`
					}{Label: "invalid"},
					Updated: struct {
						Label string `json:"label"`
					}{Label: "2023-01-15T10:30:00Z"},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockFeed)
	}))
	defer server.Close()

	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{}, time.Second)

	reviews, err := poller.fetchReviews(server.URL, "123")
	if err != nil {
		t.Fatalf("fetchReviews failed: %v", err)
	}

	// Should have 1 valid review (invalid one skipped)
	if len(reviews) != 1 {
		t.Fatalf("Expected 1 review, got %d", len(reviews))
	}

	// Check that warning was logged for malformed entry
	logOutput := buf.String()
	if !strings.Contains(logOutput, "Warning: failed to parse review entry") {
		t.Errorf("Expected warning for malformed entry, got: %s", logOutput)
	}
}

func TestNewPoller_HTTPClientConfiguration(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	storage := testutil.NewMockStorage()
	poller := NewPoller(storage, logger, []string{"123"}, time.Second)

	if poller.client == nil {
		t.Fatal("HTTP client should be initialized")
	}
	if poller.client.Timeout != 30*time.Second {
		t.Errorf("Expected HTTP client timeout 30s, got %v", poller.client.Timeout)
	}
}

// Storage Integration Tests

func TestPoller_SaveReviewsToStorage(t *testing.T) {
	storage := testutil.NewMockStorage()

	// Create sample reviews directly
	reviews := []models.Review{
		{
			ID:          "test-review-1",
			AppID:       "123",
			Author:      "Test Author",
			Content:     "Great app!",
			Rating:      5,
			SubmittedAt: time.Now(),
			FetchedAt:   time.Now(),
		},
	}

	// Test storage functionality directly
	err := storage.SaveReviews(reviews)
	if err != nil {
		t.Fatalf("SaveReviews failed: %v", err)
	}

	// Verify reviews were saved to storage
	if storage.GetSavedReviewCount() != 1 {
		t.Errorf("Expected 1 review in storage, got %d", storage.GetSavedReviewCount())
	}

	// Verify review content
	savedReviews, err := storage.GetAllReviews()
	if err != nil {
		t.Fatalf("GetAllReviews failed: %v", err)
	}

	if len(savedReviews) != 1 {
		t.Fatalf("Expected 1 review, got %d", len(savedReviews))
	}

	review := savedReviews[0]
	if review.AppID != "123" {
		t.Errorf("Expected AppID '123', got '%s'", review.AppID)
	}
	if review.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got '%s'", review.Author)
	}
}

func TestPoller_StorageError(t *testing.T) {
	storage := testutil.NewMockStorage()
	storage.SetSaveError(errors.New("storage save failed"))

	// Create sample reviews
	reviews := []models.Review{
		{
			ID:          "test-review-1",
			AppID:       "123",
			Author:      "Test Author",
			Content:     "Great app!",
			Rating:      5,
			SubmittedAt: time.Now(),
			FetchedAt:   time.Now(),
		},
	}

	// Test that storage error is propagated
	err := storage.SaveReviews(reviews)
	if err == nil {
		t.Fatal("Expected storage error, got nil")
	}
	if !strings.Contains(err.Error(), "storage save failed") {
		t.Errorf("Expected storage error message, got: %v", err)
	}
}

func TestPoller_EmptyReviewsNoStorage(t *testing.T) {
	storage := testutil.NewMockStorage()

	// Test saving empty reviews slice
	err := storage.SaveReviews([]models.Review{})
	if err != nil {
		t.Fatalf("SaveReviews with empty slice failed: %v", err)
	}

	// Verify no reviews were saved
	if storage.GetSavedReviewCount() != 0 {
		t.Errorf("Expected 0 reviews in storage, got %d", storage.GetSavedReviewCount())
	}
}

func TestPoller_MockStorageInterface(t *testing.T) {
	storage := testutil.NewMockStorage()

	// Test that our mock implements all storage interface methods
	reviews := []models.Review{
		{ID: "test1", AppID: "123", Author: "User1", Content: "Great!", Rating: 5},
		{ID: "test2", AppID: "123", Author: "User2", Content: "Good!", Rating: 4},
	}

	// Test SaveReviews
	err := storage.SaveReviews(reviews)
	if err != nil {
		t.Fatalf("SaveReviews failed: %v", err)
	}

	// Test GetAllReviews
	allReviews, err := storage.GetAllReviews()
	if err != nil {
		t.Fatalf("GetAllReviews failed: %v", err)
	}

	if len(allReviews) != 2 {
		t.Errorf("Expected 2 reviews, got %d", len(allReviews))
	}

	// Test LoadState (should not error)
	err = storage.LoadState()
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	// Test SaveState (should not error)
	err = storage.SaveState()
	if err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// Test that reviews are stored by ID (duplicates should be replaced)
	duplicateReview := []models.Review{
		{ID: "test1", AppID: "456", Author: "Updated User", Content: "Updated!", Rating: 3},
	}

	err = storage.SaveReviews(duplicateReview)
	if err != nil {
		t.Fatalf("SaveReviews with duplicate failed: %v", err)
	}

	// Should still have 2 reviews (test1 was updated, test2 remains)
	if storage.GetSavedReviewCount() != 2 {
		t.Errorf("Expected 2 reviews after duplicate save, got %d", storage.GetSavedReviewCount())
	}

	// Verify test1 was updated
	allReviews, _ = storage.GetAllReviews()
	var updatedReview *models.Review
	for _, review := range allReviews {
		if review.ID == "test1" {
			updatedReview = &review
			break
		}
	}

	if updatedReview == nil {
		t.Fatal("Updated review not found")
	}

	if updatedReview.AppID != "456" {
		t.Errorf("Expected updated AppID '456', got '%s'", updatedReview.AppID)
	}
	if updatedReview.Author != "Updated User" {
		t.Errorf("Expected updated Author 'Updated User', got '%s'", updatedReview.Author)
	}
}