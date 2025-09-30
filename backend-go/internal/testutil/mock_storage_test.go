package testutil

import (
	"backend/internal/models"
	"errors"
	"testing"
	"time"
)

func TestMockStorage_BasicOperations(t *testing.T) {
	storage := NewMockStorage()

	// Test empty storage
	if storage.GetSavedReviewCount() != 0 {
		t.Errorf("Expected 0 reviews in new storage, got %d", storage.GetSavedReviewCount())
	}

	// Test saving reviews
	reviews := []models.Review{
		{
			ID:          "review1",
			AppID:       "app1",
			Author:      "User1",
			Content:     "Great app!",
			Rating:      5,
			SubmittedAt: time.Now(),
			FetchedAt:   time.Now(),
		},
		{
			ID:          "review2",
			AppID:       "app1",
			Author:      "User2",
			Content:     "Good app!",
			Rating:      4,
			SubmittedAt: time.Now(),
			FetchedAt:   time.Now(),
		},
	}

	err := storage.SaveReviews(reviews)
	if err != nil {
		t.Fatalf("SaveReviews failed: %v", err)
	}

	// Test count
	if storage.GetSavedReviewCount() != 2 {
		t.Errorf("Expected 2 reviews, got %d", storage.GetSavedReviewCount())
	}

	// Test retrieval
	allReviews, err := storage.GetAllReviews()
	if err != nil {
		t.Fatalf("GetAllReviews failed: %v", err)
	}

	if len(allReviews) != 2 {
		t.Errorf("Expected 2 reviews, got %d", len(allReviews))
	}

	// Test HasReview
	if !storage.HasReview("review1") {
		t.Error("Expected review1 to exist")
	}
	if !storage.HasReview("review2") {
		t.Error("Expected review2 to exist")
	}
	if storage.HasReview("nonexistent") {
		t.Error("Expected nonexistent review to not exist")
	}

	// Test GetReview
	review, exists := storage.GetReview("review1")
	if !exists {
		t.Error("Expected review1 to exist")
	}
	if review.Author != "User1" {
		t.Errorf("Expected author 'User1', got '%s'", review.Author)
	}

	_, exists = storage.GetReview("nonexistent")
	if exists {
		t.Error("Expected nonexistent review to not exist")
	}
}

func TestMockStorage_ErrorHandling(t *testing.T) {
	storage := NewMockStorage()

	// Test save error
	testErr := errors.New("save failed")
	storage.SetSaveError(testErr)

	reviews := []models.Review{{ID: "test", AppID: "app1"}}
	err := storage.SaveReviews(reviews)
	if err != testErr {
		t.Errorf("Expected save error, got %v", err)
	}

	// Test load error
	loadErr := errors.New("load failed")
	storage.SetLoadError(loadErr)

	err = storage.LoadState()
	if err != loadErr {
		t.Errorf("Expected load error, got %v", err)
	}
}

func TestMockStorage_Deduplication(t *testing.T) {
	storage := NewMockStorage()

	// Save initial review
	reviews1 := []models.Review{
		{ID: "review1", AppID: "app1", Author: "User1", Content: "Original", Rating: 5},
	}
	storage.SaveReviews(reviews1)

	// Save updated review with same ID
	reviews2 := []models.Review{
		{ID: "review1", AppID: "app2", Author: "User2", Content: "Updated", Rating: 3},
	}
	storage.SaveReviews(reviews2)

	// Should still have only 1 review
	if storage.GetSavedReviewCount() != 1 {
		t.Errorf("Expected 1 review after update, got %d", storage.GetSavedReviewCount())
	}

	// Should have updated content
	review, exists := storage.GetReview("review1")
	if !exists {
		t.Fatal("Review should exist")
	}

	if review.Content != "Updated" {
		t.Errorf("Expected updated content 'Updated', got '%s'", review.Content)
	}
	if review.AppID != "app2" {
		t.Errorf("Expected updated AppID 'app2', got '%s'", review.AppID)
	}
}

func TestMockStorage_Reset(t *testing.T) {
	storage := NewMockStorage()

	// Add some data and errors
	reviews := []models.Review{{ID: "test", AppID: "app1"}}
	storage.SaveReviews(reviews)
	storage.SetSaveError(errors.New("error"))
	storage.SetLoadError(errors.New("load error"))

	// Verify data exists
	if storage.GetSavedReviewCount() != 1 {
		t.Error("Expected review to be saved before reset")
	}

	// Reset
	storage.Reset()

	// Verify everything is cleared
	if storage.GetSavedReviewCount() != 0 {
		t.Errorf("Expected 0 reviews after reset, got %d", storage.GetSavedReviewCount())
	}

	// Verify errors are cleared
	err := storage.SaveReviews(reviews)
	if err != nil {
		t.Errorf("Expected no save error after reset, got %v", err)
	}

	err = storage.LoadState()
	if err != nil {
		t.Errorf("Expected no load error after reset, got %v", err)
	}
}

func TestMockStorage_StateOperations(t *testing.T) {
	storage := NewMockStorage()

	// SaveState should never error
	err := storage.SaveState()
	if err != nil {
		t.Errorf("SaveState should not error, got %v", err)
	}

	// LoadState should not error by default
	err = storage.LoadState()
	if err != nil {
		t.Errorf("LoadState should not error by default, got %v", err)
	}
}

func TestMockStorage_GetRecentReviews(t *testing.T) {
	storage := NewMockStorage()

	// Add test reviews with different timestamps and apps
	now := time.Now()
	reviews := []models.Review{
		{
			ID:          "review1",
			AppID:       "app1",
			Author:      "User1",
			Content:     "Recent review",
			Rating:      5,
			SubmittedAt: now.Add(-1 * time.Hour), // 1 hour ago
			FetchedAt:   now,
		},
		{
			ID:          "review2",
			AppID:       "app1",
			Author:      "User2",
			Content:     "Old review",
			Rating:      4,
			SubmittedAt: now.Add(-72 * time.Hour), // 3 days ago
			FetchedAt:   now,
		},
		{
			ID:          "review3",
			AppID:       "app2",
			Author:      "User3",
			Content:     "Different app review",
			Rating:      3,
			SubmittedAt: now.Add(-1 * time.Hour), // 1 hour ago
			FetchedAt:   now,
		},
	}

	err := storage.SaveReviews(reviews)
	if err != nil {
		t.Fatalf("Failed to save reviews: %v", err)
	}

	// Test getting recent reviews for app1 within 48 hours
	recentReviews, err := storage.GetRecentReviews("app1", 48*time.Hour)
	if err != nil {
		t.Fatalf("GetRecentReviews failed: %v", err)
	}

	// Should only return review1 (app1 and within 48h)
	if len(recentReviews) != 1 {
		t.Errorf("Expected 1 recent review for app1, got %d", len(recentReviews))
	}

	if len(recentReviews) > 0 && recentReviews[0].ID != "review1" {
		t.Errorf("Expected review1, got %s", recentReviews[0].ID)
	}

	// Test with shorter time window
	recentReviews, err = storage.GetRecentReviews("app1", 30*time.Minute)
	if err != nil {
		t.Fatalf("GetRecentReviews failed: %v", err)
	}

	// Should return no reviews (none within 30 minutes)
	if len(recentReviews) != 0 {
		t.Errorf("Expected 0 recent reviews within 30 minutes, got %d", len(recentReviews))
	}

	// Test with different app
	recentReviews, err = storage.GetRecentReviews("app2", 48*time.Hour)
	if err != nil {
		t.Fatalf("GetRecentReviews failed: %v", err)
	}

	// Should return review3
	if len(recentReviews) != 1 {
		t.Errorf("Expected 1 recent review for app2, got %d", len(recentReviews))
	}

	if len(recentReviews) > 0 && recentReviews[0].ID != "review3" {
		t.Errorf("Expected review3, got %s", recentReviews[0].ID)
	}
}

func TestMockStorage_GetRecentReviewsError(t *testing.T) {
	storage := NewMockStorage()
	testErr := errors.New("get recent reviews failed")
	storage.SetGetRecentReviewsError(testErr)

	_, err := storage.GetRecentReviews("app1", 24*time.Hour)
	if err != testErr {
		t.Errorf("Expected GetRecentReviews error, got %v", err)
	}
}

func TestMockStorage_GetAllReviewsError(t *testing.T) {
	storage := NewMockStorage()
	testErr := errors.New("get all reviews failed")
	storage.SetGetAllReviewsError(testErr)

	_, err := storage.GetAllReviews()
	if err != testErr {
		t.Errorf("Expected GetAllReviews error, got %v", err)
	}
}

func TestMockStorage_ErrorReset(t *testing.T) {
	storage := NewMockStorage()

	// Set various errors
	storage.SetSaveError(errors.New("save error"))
	storage.SetLoadError(errors.New("load error"))
	storage.SetGetRecentReviewsError(errors.New("get recent error"))
	storage.SetGetAllReviewsError(errors.New("get all error"))

	// Add some data
	reviews := []models.Review{{ID: "test", AppID: "app1"}}
	// This should fail due to save error
	err := storage.SaveReviews(reviews)
	if err == nil {
		t.Error("Expected save error before reset")
	}

	// Reset should clear all errors
	storage.Reset()

	// Now operations should work
	err = storage.SaveReviews(reviews)
	if err != nil {
		t.Errorf("Expected no save error after reset, got %v", err)
	}

	err = storage.LoadState()
	if err != nil {
		t.Errorf("Expected no load error after reset, got %v", err)
	}

	_, err = storage.GetRecentReviews("app1", 24*time.Hour)
	if err != nil {
		t.Errorf("Expected no GetRecentReviews error after reset, got %v", err)
	}

	_, err = storage.GetAllReviews()
	if err != nil {
		t.Errorf("Expected no GetAllReviews error after reset, got %v", err)
	}
}