package handler

import (
	"backend/internal/models"
	"backend/internal/testutil"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewHandler(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}
	if handler.storage != storage {
		t.Error("Handler storage not properly set")
	}
}

func TestHandler_GetRecentReviews_Success(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	// Add test reviews with different timestamps
	now := time.Now()
	reviews := []models.Review{
		{
			ID:          "review1",
			AppID:       "123",
			Author:      "User1",
			Content:     "Recent review",
			Rating:      5,
			SubmittedAt: now.Add(-1 * time.Hour), // 1 hour ago
			FetchedAt:   now,
		},
		{
			ID:          "review2",
			AppID:       "123",
			Author:      "User2",
			Content:     "Older review",
			Rating:      4,
			SubmittedAt: now.Add(-24 * time.Hour), // 1 day ago (within 48h window)
			FetchedAt:   now,
		},
		{
			ID:          "review3",
			AppID:       "456",
			Author:      "User3",
			Content:     "Different app",
			Rating:      3,
			SubmittedAt: now.Add(-1 * time.Hour),
			FetchedAt:   now,
		},
	}
	storage.SaveReviews(reviews)

	// Create request
	req := httptest.NewRequest("GET", "/api/reviews?app_id=123", nil)
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetRecentReviews(rr, req)

	// Check response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", contentType)
	}

	// Parse response
	var responseReviews []models.Review
	err := json.NewDecoder(rr.Body).Decode(&responseReviews)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should return both review1 and review2 (app_id=123 and within default 48h)
	if len(responseReviews) != 2 {
		t.Errorf("Expected 2 reviews, got %d", len(responseReviews))
	}

	// Should be sorted by newest first, so review1 (1h ago) comes before review2 (24h ago)
	if len(responseReviews) >= 1 && responseReviews[0].ID != "review1" {
		t.Errorf("Expected review1 first, got %s", responseReviews[0].ID)
	}
	if len(responseReviews) >= 2 && responseReviews[1].ID != "review2" {
		t.Errorf("Expected review2 second, got %s", responseReviews[1].ID)
	}
}

func TestHandler_GetRecentReviews_CustomHours(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	// Add test reviews
	now := time.Now()
	reviews := []models.Review{
		{
			ID:          "review1",
			AppID:       "123",
			Author:      "User1",
			Content:     "Recent review",
			Rating:      5,
			SubmittedAt: now.Add(-1 * time.Hour), // 1 hour ago
			FetchedAt:   now,
		},
		{
			ID:          "review2",
			AppID:       "123",
			Author:      "User2",
			Content:     "Older review",
			Rating:      4,
			SubmittedAt: now.Add(-25 * time.Hour), // 25 hours ago
			FetchedAt:   now,
		},
	}
	storage.SaveReviews(reviews)

	// Request with custom hours parameter
	req := httptest.NewRequest("GET", "/api/reviews?app_id=123&hours=24", nil)
	rr := httptest.NewRecorder()

	handler.GetRecentReviews(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var responseReviews []models.Review
	json.NewDecoder(rr.Body).Decode(&responseReviews)

	// Should only return review1 (within 24h)
	if len(responseReviews) != 1 {
		t.Errorf("Expected 1 review, got %d", len(responseReviews))
	}
}

func TestHandler_GetRecentReviews_SortedByNewest(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	// Add reviews with different timestamps
	now := time.Now()
	reviews := []models.Review{
		{
			ID:          "review1",
			AppID:       "123",
			Author:      "User1",
			Content:     "Older review",
			Rating:      5,
			SubmittedAt: now.Add(-10 * time.Hour), // 10 hours ago
			FetchedAt:   now,
		},
		{
			ID:          "review2",
			AppID:       "123",
			Author:      "User2",
			Content:     "Newer review",
			Rating:      4,
			SubmittedAt: now.Add(-2 * time.Hour), // 2 hours ago
			FetchedAt:   now,
		},
	}
	storage.SaveReviews(reviews)

	req := httptest.NewRequest("GET", "/api/reviews?app_id=123", nil)
	rr := httptest.NewRecorder()

	handler.GetRecentReviews(rr, req)

	var responseReviews []models.Review
	json.NewDecoder(rr.Body).Decode(&responseReviews)

	// Should be sorted by newest first
	if len(responseReviews) != 2 {
		t.Fatalf("Expected 2 reviews, got %d", len(responseReviews))
	}

	if responseReviews[0].ID != "review2" {
		t.Errorf("Expected review2 (newer) first, got %s", responseReviews[0].ID)
	}
	if responseReviews[1].ID != "review1" {
		t.Errorf("Expected review1 (older) second, got %s", responseReviews[1].ID)
	}
}

func TestHandler_GetRecentReviews_MissingAppID(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	req := httptest.NewRequest("GET", "/api/reviews", nil)
	rr := httptest.NewRecorder()

	handler.GetRecentReviews(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	responseBody := strings.TrimSpace(rr.Body.String())
	expectedMessage := "app_id query parameter is required"
	if !strings.Contains(responseBody, expectedMessage) {
		t.Errorf("Expected error message '%s', got '%s'", expectedMessage, responseBody)
	}
}

func TestHandler_GetRecentReviews_InvalidMethod(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	req := httptest.NewRequest("POST", "/api/reviews?app_id=123", nil)
	rr := httptest.NewRecorder()

	handler.GetRecentReviews(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandler_GetRecentReviews_InvalidHours(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	testCases := []struct {
		name  string
		hours string
	}{
		{"negative hours", "hours=-5"},
		{"zero hours", "hours=0"},
		{"non-numeric", "hours=abc"},
		{"float", "hours=24.5"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/reviews?app_id=123&"+tc.hours, nil)
			rr := httptest.NewRecorder()

			handler.GetRecentReviews(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
			}
		})
	}
}

func TestHandler_GetRecentReviews_StorageError(t *testing.T) {
	storage := testutil.NewMockStorage()
	storage.SetGetRecentReviewsError(errors.New("storage error"))
	handler := NewHandler(storage)

	req := httptest.NewRequest("GET", "/api/reviews?app_id=123", nil)
	rr := httptest.NewRecorder()

	handler.GetRecentReviews(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestHandler_GetRecentReviews_EmptyResult(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	req := httptest.NewRequest("GET", "/api/reviews?app_id=999", nil)
	rr := httptest.NewRecorder()

	handler.GetRecentReviews(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var responseReviews []models.Review
	json.NewDecoder(rr.Body).Decode(&responseReviews)

	if len(responseReviews) != 0 {
		t.Errorf("Expected 0 reviews, got %d", len(responseReviews))
	}
}

// HealthCheck Tests

func TestHandler_HealthCheck_Success(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	// Add some test reviews
	reviews := []models.Review{
		{ID: "review1", AppID: "123", Author: "User1", Rating: 5},
		{ID: "review2", AppID: "456", Author: "User2", Rating: 4},
	}
	storage.SaveReviews(reviews)

	req := httptest.NewRequest("GET", "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", contentType)
	}

	// Parse response
	var response map[string]any
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check response fields
	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}

	if response["total_reviews"] != float64(2) { // JSON numbers are float64
		t.Errorf("Expected total_reviews 2, got %v", response["total_reviews"])
	}

	// Check timestamp is present and valid
	timestamp, ok := response["timestamp"].(string)
	if !ok {
		t.Error("Expected timestamp to be a string")
	} else {
		_, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			t.Errorf("Invalid timestamp format: %v", err)
		}
	}
}

func TestHandler_HealthCheck_InvalidMethod(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	req := httptest.NewRequest("POST", "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandler_HealthCheck_StorageError(t *testing.T) {
	storage := testutil.NewMockStorage()
	storage.SetGetAllReviewsError(errors.New("storage error"))
	handler := NewHandler(storage)

	req := httptest.NewRequest("GET", "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestHandler_HealthCheck_EmptyStorage(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	req := httptest.NewRequest("GET", "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]any
	json.NewDecoder(rr.Body).Decode(&response)

	if response["total_reviews"] != float64(0) {
		t.Errorf("Expected total_reviews 0, got %v", response["total_reviews"])
	}
}

// GetAverageRating Tests

func TestHandler_GetAverageRating_Success(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	// Add test reviews with different ratings
	now := time.Now()
	reviews := []models.Review{
		{
			ID:          "review1",
			AppID:       "123",
			Author:      "User1",
			Content:     "Great app",
			Rating:      5,
			SubmittedAt: now.Add(-1 * time.Hour),
			FetchedAt:   now,
		},
		{
			ID:          "review2",
			AppID:       "123",
			Author:      "User2",
			Content:     "Good app",
			Rating:      4,
			SubmittedAt: now.Add(-2 * time.Hour),
			FetchedAt:   now,
		},
		{
			ID:          "review3",
			AppID:       "123",
			Author:      "User3",
			Content:     "Okay app",
			Rating:      3,
			SubmittedAt: now.Add(-3 * time.Hour),
			FetchedAt:   now,
		},
	}
	storage.SaveReviews(reviews)

	req := httptest.NewRequest("GET", "/api/average-rating?app_id=123", nil)
	rr := httptest.NewRecorder()

	handler.GetAverageRating(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", contentType)
	}

	// Parse response
	var response map[string]any
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check response fields
	if response["app_id"] != "123" {
		t.Errorf("Expected app_id '123', got '%v'", response["app_id"])
	}

	// Average of 5, 4, 3 = 12/3 = 4.0
	expectedAverage := 4.0
	if response["average_rating"] != expectedAverage {
		t.Errorf("Expected average_rating %.1f, got %v", expectedAverage, response["average_rating"])
	}

	if response["review_count"] != float64(3) {
		t.Errorf("Expected review_count 3, got %v", response["review_count"])
	}

	if response["hours"] != float64(48) {
		t.Errorf("Expected hours 48, got %v", response["hours"])
	}
}

func TestHandler_GetAverageRating_CustomHours(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	// Add test reviews
	now := time.Now()
	reviews := []models.Review{
		{
			ID:          "review1",
			AppID:       "123",
			Author:      "User1",
			Content:     "Recent",
			Rating:      5,
			SubmittedAt: now.Add(-1 * time.Hour), // Within 24h
			FetchedAt:   now,
		},
		{
			ID:          "review2",
			AppID:       "123",
			Author:      "User2",
			Content:     "Old",
			Rating:      1,
			SubmittedAt: now.Add(-30 * time.Hour), // Outside 24h
			FetchedAt:   now,
		},
	}
	storage.SaveReviews(reviews)

	req := httptest.NewRequest("GET", "/api/average-rating?app_id=123&hours=24", nil)
	rr := httptest.NewRecorder()

	handler.GetAverageRating(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]any
	json.NewDecoder(rr.Body).Decode(&response)

	// Should only consider review1 (rating 5)
	if response["average_rating"] != float64(5) {
		t.Errorf("Expected average_rating 5.0, got %v", response["average_rating"])
	}

	if response["review_count"] != float64(1) {
		t.Errorf("Expected review_count 1, got %v", response["review_count"])
	}

	if response["hours"] != float64(24) {
		t.Errorf("Expected hours 24, got %v", response["hours"])
	}
}

func TestHandler_GetAverageRating_NoReviews(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	req := httptest.NewRequest("GET", "/api/average-rating?app_id=999", nil)
	rr := httptest.NewRecorder()

	handler.GetAverageRating(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]any
	json.NewDecoder(rr.Body).Decode(&response)

	// With no reviews, average should be 0
	if response["average_rating"] != float64(0) {
		t.Errorf("Expected average_rating 0, got %v", response["average_rating"])
	}

	if response["review_count"] != float64(0) {
		t.Errorf("Expected review_count 0, got %v", response["review_count"])
	}
}

func TestHandler_GetAverageRating_MissingAppID(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	req := httptest.NewRequest("GET", "/api/average-rating", nil)
	rr := httptest.NewRecorder()

	handler.GetAverageRating(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	responseBody := strings.TrimSpace(rr.Body.String())
	expectedMessage := "app_id query parameter is required"
	if !strings.Contains(responseBody, expectedMessage) {
		t.Errorf("Expected error message '%s', got '%s'", expectedMessage, responseBody)
	}
}

func TestHandler_GetAverageRating_InvalidMethod(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	req := httptest.NewRequest("POST", "/api/average-rating?app_id=123", nil)
	rr := httptest.NewRecorder()

	handler.GetAverageRating(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandler_GetAverageRating_InvalidHours(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	testCases := []struct {
		name  string
		hours string
	}{
		{"negative hours", "hours=-5"},
		{"zero hours", "hours=0"},
		{"non-numeric", "hours=abc"},
		{"float", "hours=24.5"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/average-rating?app_id=123&"+tc.hours, nil)
			rr := httptest.NewRecorder()

			handler.GetAverageRating(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
			}
		})
	}
}

func TestHandler_GetAverageRating_StorageError(t *testing.T) {
	storage := testutil.NewMockStorage()
	storage.SetGetRecentReviewsError(errors.New("storage error"))
	handler := NewHandler(storage)

	req := httptest.NewRequest("GET", "/api/average-rating?app_id=123", nil)
	rr := httptest.NewRecorder()

	handler.GetAverageRating(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestHandler_GetAverageRating_RoundingPrecision(t *testing.T) {
	storage := testutil.NewMockStorage()
	handler := NewHandler(storage)

	// Add reviews that result in a non-round average
	now := time.Now()
	reviews := []models.Review{
		{ID: "r1", AppID: "123", Rating: 5, SubmittedAt: now.Add(-1 * time.Hour), FetchedAt: now},
		{ID: "r2", AppID: "123", Rating: 4, SubmittedAt: now.Add(-2 * time.Hour), FetchedAt: now},
		{ID: "r3", AppID: "123", Rating: 4, SubmittedAt: now.Add(-3 * time.Hour), FetchedAt: now},
	}
	storage.SaveReviews(reviews)

	req := httptest.NewRequest("GET", "/api/average-rating?app_id=123", nil)
	rr := httptest.NewRecorder()

	handler.GetAverageRating(rr, req)

	var response map[string]any
	json.NewDecoder(rr.Body).Decode(&response)

	// Average of 5, 4, 4 = 13/3 = 4.333...
	// Should be rounded to 1 decimal place = 4.3
	expectedAverage := 4.3
	actualAverage := response["average_rating"].(float64)

	// Allow small floating point tolerance
	if actualAverage < expectedAverage-0.01 || actualAverage > expectedAverage+0.01 {
		t.Errorf("Expected average_rating approximately %.1f, got %.2f", expectedAverage, actualAverage)
	}
}