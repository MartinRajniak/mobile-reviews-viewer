package storage

import (
    "fmt"
    "os"
    "path/filepath"
    "testing"
    "time"

    "backend/internal/models"
)

func TestFileStorage_PersistAndLoad(t *testing.T) {
    // Create temp directory
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "test_reviews.json")

    // Create storage
    storage, err := NewFileStorage(testFile)
    if err != nil {
        t.Fatalf("Failed to create storage: %v", err)
    }

    // Create test reviews
    reviews := []models.Review{
        {
            ID:          "review1",
            AppID:       "123",
            Author:      "Test User",
            Content:     "Test content",
            Rating:      5,
            SubmittedAt: time.Now(),
            FetchedAt:   time.Now(),
        },
    }

    // Save reviews
    if err := storage.SaveReviews(reviews); err != nil {
        t.Fatalf("Failed to save reviews: %v", err)
    }

    // Verify file exists
    if _, err := os.Stat(testFile); os.IsNotExist(err) {
        t.Fatal("File was not created")
    }

    // Create new storage instance and load
    storage2, _ := NewFileStorage(testFile)
    if err := storage2.LoadState(); err != nil {
        t.Fatalf("Failed to load state: %v", err)
    }

    // Verify loaded reviews
    allReviews, _ := storage2.GetAllReviews()
    if len(allReviews) != 1 {
        t.Errorf("Expected 1 review, got %d", len(allReviews))
    }

    if allReviews[0].ID != "review1" {
        t.Errorf("Expected review1, got %s", allReviews[0].ID)
    }
}

func TestFileStorage_AtomicWrite(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "test_reviews.json")

    storage, _ := NewFileStorage(testFile)

    // Verify temp file is cleaned up after successful write
    storage.SaveReviews([]models.Review{{ID: "test"}})

    tempFile := testFile + ".tmp"
    if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
        t.Error("Temp file should not exist after successful write")
    }
}

func TestFileStorage_LoadStateNonExistentFile(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "nonexistent.json")

    storage, err := NewFileStorage(testFile)
    if err != nil {
        t.Fatalf("Failed to create storage: %v", err)
    }

    // LoadState should not error when file doesn't exist
    err = storage.LoadState()
    if err != nil {
        t.Errorf("LoadState should not error for nonexistent file, got: %v", err)
    }

    // Should have no reviews
    reviews, err := storage.GetAllReviews()
    if err != nil {
        t.Fatalf("GetAllReviews failed: %v", err)
    }
    if len(reviews) != 0 {
        t.Errorf("Expected 0 reviews, got %d", len(reviews))
    }
}

func TestFileStorage_LoadStateCorruptedFile(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "corrupted.json")

    // Create corrupted JSON file
    err := os.WriteFile(testFile, []byte("invalid json content"), 0644)
    if err != nil {
        t.Fatalf("Failed to create corrupted file: %v", err)
    }

    storage, err := NewFileStorage(testFile)
    if err != nil {
        t.Fatalf("Failed to create storage: %v", err)
    }

    // LoadState should error with corrupted JSON
    err = storage.LoadState()
    if err == nil {
        t.Error("Expected error when loading corrupted JSON file")
    }
    if !contains(err.Error(), "failed to unmarshal reviews") {
        t.Errorf("Expected unmarshal error, got: %v", err)
    }
}

func TestFileStorage_EmptyReviews(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "empty.json")

    storage, _ := NewFileStorage(testFile)

    // Save empty slice
    err := storage.SaveReviews([]models.Review{})
    if err != nil {
        t.Fatalf("Failed to save empty reviews: %v", err)
    }

    // Load and verify
    storage2, _ := NewFileStorage(testFile)
    err = storage2.LoadState()
    if err != nil {
        t.Fatalf("Failed to load state: %v", err)
    }

    reviews, err := storage2.GetAllReviews()
    if err != nil {
        t.Fatalf("GetAllReviews failed: %v", err)
    }
    if len(reviews) != 0 {
        t.Errorf("Expected 0 reviews, got %d", len(reviews))
    }
}

func TestFileStorage_ReviewDeduplication(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "dedup.json")

    storage, _ := NewFileStorage(testFile)

    // Save initial reviews
    reviews1 := []models.Review{
        {ID: "review1", AppID: "app1", Author: "User1", Content: "Original", Rating: 5},
        {ID: "review2", AppID: "app1", Author: "User2", Content: "Second", Rating: 4},
    }
    err := storage.SaveReviews(reviews1)
    if err != nil {
        t.Fatalf("Failed to save initial reviews: %v", err)
    }

    // Save updated review with same ID
    reviews2 := []models.Review{
        {ID: "review1", AppID: "app2", Author: "Updated User", Content: "Updated", Rating: 3},
        {ID: "review3", AppID: "app1", Author: "User3", Content: "Third", Rating: 2},
    }
    err = storage.SaveReviews(reviews2)
    if err != nil {
        t.Fatalf("Failed to save updated reviews: %v", err)
    }

    // Should have 3 reviews total (review1 updated, review2 unchanged, review3 new)
    allReviews, err := storage.GetAllReviews()
    if err != nil {
        t.Fatalf("GetAllReviews failed: %v", err)
    }
    if len(allReviews) != 3 {
        t.Errorf("Expected 3 reviews, got %d", len(allReviews))
    }

    // Verify review1 was updated
    var review1 *models.Review
    for _, r := range allReviews {
        if r.ID == "review1" {
            review1 = &r
            break
        }
    }
    if review1 == nil {
        t.Fatal("review1 not found")
    }
    if review1.Content != "Updated" {
        t.Errorf("Expected updated content 'Updated', got '%s'", review1.Content)
    }
    if review1.AppID != "app2" {
        t.Errorf("Expected updated AppID 'app2', got '%s'", review1.AppID)
    }
}

func TestFileStorage_SaveStateExplicit(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "explicit.json")

    storage, _ := NewFileStorage(testFile)

    // Add reviews to memory
    reviews := []models.Review{
        {ID: "review1", AppID: "app1", Author: "User1", Content: "Test", Rating: 5},
    }
    err := storage.SaveReviews(reviews)
    if err != nil {
        t.Fatalf("Failed to save reviews: %v", err)
    }

    // Explicitly call SaveState
    err = storage.SaveState()
    if err != nil {
        t.Fatalf("SaveState failed: %v", err)
    }

    // Verify file was written
    if _, err := os.Stat(testFile); os.IsNotExist(err) {
        t.Error("File should exist after SaveState")
    }

    // Verify content by loading with new instance
    storage2, _ := NewFileStorage(testFile)
    err = storage2.LoadState()
    if err != nil {
        t.Fatalf("Failed to load state: %v", err)
    }

    allReviews, _ := storage2.GetAllReviews()
    if len(allReviews) != 1 {
        t.Errorf("Expected 1 review, got %d", len(allReviews))
    }
}

func TestFileStorage_ConcurrentAccess(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "concurrent.json")

    storage, _ := NewFileStorage(testFile)

    // Create goroutines that write concurrently
    const numGoroutines = 10
    done := make(chan bool, numGoroutines)

    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            reviews := []models.Review{
                {ID: fmt.Sprintf("review%d", id), AppID: "app1", Author: fmt.Sprintf("User%d", id), Rating: 5},
            }
            storage.SaveReviews(reviews)
            done <- true
        }(i)
    }

    // Wait for all goroutines to complete
    for i := 0; i < numGoroutines; i++ {
        <-done
    }

    // Verify all reviews were saved
    allReviews, err := storage.GetAllReviews()
    if err != nil {
        t.Fatalf("GetAllReviews failed: %v", err)
    }
    if len(allReviews) != numGoroutines {
        t.Errorf("Expected %d reviews, got %d", numGoroutines, len(allReviews))
    }
}

func TestFileStorage_DirectoryCreation(t *testing.T) {
    tempDir := t.TempDir()
    // Create path with nested directories that don't exist
    testFile := filepath.Join(tempDir, "nested", "deep", "reviews.json")

    storage, err := NewFileStorage(testFile)
    if err != nil {
        t.Fatalf("Failed to create storage with nested path: %v", err)
    }

    // Save a review to trigger directory creation
    reviews := []models.Review{
        {ID: "test", AppID: "app1", Author: "User", Rating: 5},
    }
    err = storage.SaveReviews(reviews)
    if err != nil {
        t.Fatalf("Failed to save to nested path: %v", err)
    }

    // Verify directory was created
    dir := filepath.Dir(testFile)
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        t.Error("Directory should have been created")
    }

    // Verify file was created
    if _, err := os.Stat(testFile); os.IsNotExist(err) {
        t.Error("File should have been created")
    }
}

func TestFileStorage_NewFileStorageWithInvalidPath(t *testing.T) {
    // Try to create storage with an invalid path (permission denied scenario)
    // This test might not work on all systems, so we'll skip if we can't create the condition

    // Test with root directory path that we can't write to (on most systems)
    invalidPath := "/root/cannot_create/test.json"

    _, err := NewFileStorage(invalidPath)
    // This should fail on most systems due to permission issues
    // If it doesn't fail, it means the test environment has unusual permissions
    if err == nil {
        t.Skip("Skipping test - unable to create permission denied scenario")
    }

    if !contains(err.Error(), "failed to create directory") {
        t.Errorf("Expected directory creation error, got: %v", err)
    }
}

func TestFileStorage_LargeDataSet(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "large.json")

    storage, _ := NewFileStorage(testFile)

    // Create a large number of reviews
    const numReviews = 1000
    reviews := make([]models.Review, numReviews)
    for i := 0; i < numReviews; i++ {
        reviews[i] = models.Review{
            ID:          fmt.Sprintf("review%d", i),
            AppID:       fmt.Sprintf("app%d", i%10), // 10 different apps
            Author:      fmt.Sprintf("User%d", i),
            Content:     fmt.Sprintf("This is review content for review %d", i),
            Rating:      (i % 5) + 1, // Ratings 1-5
            SubmittedAt: time.Now(),
            FetchedAt:   time.Now(),
        }
    }

    // Save large dataset
    err := storage.SaveReviews(reviews)
    if err != nil {
        t.Fatalf("Failed to save large dataset: %v", err)
    }

    // Load with new instance
    storage2, _ := NewFileStorage(testFile)
    err = storage2.LoadState()
    if err != nil {
        t.Fatalf("Failed to load large dataset: %v", err)
    }

    // Verify all reviews were loaded
    allReviews, err := storage2.GetAllReviews()
    if err != nil {
        t.Fatalf("GetAllReviews failed: %v", err)
    }
    if len(allReviews) != numReviews {
        t.Errorf("Expected %d reviews, got %d", numReviews, len(allReviews))
    }
}

func TestFileStorage_GetRecentReviews_BasicFiltering(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "recent_reviews.json")

    storage, _ := NewFileStorage(testFile)

    // Create test reviews with different timestamps and apps
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
        {
            ID:          "review4",
            AppID:       "app1",
            Author:      "User4",
            Content:     "Very recent review",
            Rating:      5,
            SubmittedAt: now.Add(-30 * time.Minute), // 30 minutes ago
            FetchedAt:   now,
        },
    }

    err := storage.SaveReviews(reviews)
    if err != nil {
        t.Fatalf("Failed to save reviews: %v", err)
    }

    // Test: Get reviews for app1 within 48 hours
    recentReviews, err := storage.GetRecentReviews("app1", 48*time.Hour)
    if err != nil {
        t.Fatalf("GetRecentReviews failed: %v", err)
    }

    // Should return review1 and review4 (both app1 and within 48h)
    if len(recentReviews) != 2 {
        t.Errorf("Expected 2 reviews for app1 within 48h, got %d", len(recentReviews))
    }

    // Verify correct reviews were returned
    reviewIDs := make([]string, len(recentReviews))
    for i, review := range recentReviews {
        reviewIDs[i] = review.ID
    }

    expectedIDs := []string{"review1", "review4"}
    for _, expectedID := range expectedIDs {
        found := false
        for _, id := range reviewIDs {
            if id == expectedID {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("Expected to find review %s in results", expectedID)
        }
    }
}

func TestFileStorage_GetRecentReviews_TimeFiltering(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "time_filtering.json")

    storage, _ := NewFileStorage(testFile)

    now := time.Now()
    reviews := []models.Review{
        {
            ID:          "review1",
            AppID:       "app1",
            Author:      "User1",
            Content:     "Very recent",
            Rating:      5,
            SubmittedAt: now.Add(-30 * time.Minute), // 30 minutes ago
            FetchedAt:   now,
        },
        {
            ID:          "review2",
            AppID:       "app1",
            Author:      "User2",
            Content:     "Recent",
            Rating:      4,
            SubmittedAt: now.Add(-2 * time.Hour), // 2 hours ago
            FetchedAt:   now,
        },
        {
            ID:          "review3",
            AppID:       "app1",
            Author:      "User3",
            Content:     "Old",
            Rating:      3,
            SubmittedAt: now.Add(-25 * time.Hour), // 25 hours ago
            FetchedAt:   now,
        },
    }

    storage.SaveReviews(reviews)

    // Test 1: Within 1 hour - should get only review1
    recentReviews, err := storage.GetRecentReviews("app1", 1*time.Hour)
    if err != nil {
        t.Fatalf("GetRecentReviews failed: %v", err)
    }
    if len(recentReviews) != 1 {
        t.Errorf("Expected 1 review within 1h, got %d", len(recentReviews))
    }
    if len(recentReviews) > 0 && recentReviews[0].ID != "review1" {
        t.Errorf("Expected review1, got %s", recentReviews[0].ID)
    }

    // Test 2: Within 3 hours - should get review1 and review2
    recentReviews, err = storage.GetRecentReviews("app1", 3*time.Hour)
    if err != nil {
        t.Fatalf("GetRecentReviews failed: %v", err)
    }
    if len(recentReviews) != 2 {
        t.Errorf("Expected 2 reviews within 3h, got %d", len(recentReviews))
    }

    // Test 3: Within 48 hours - should get all 3 reviews
    recentReviews, err = storage.GetRecentReviews("app1", 48*time.Hour)
    if err != nil {
        t.Fatalf("GetRecentReviews failed: %v", err)
    }
    if len(recentReviews) != 3 {
        t.Errorf("Expected 3 reviews within 48h, got %d", len(recentReviews))
    }
}

func TestFileStorage_GetRecentReviews_AppFiltering(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "app_filtering.json")

    storage, _ := NewFileStorage(testFile)

    now := time.Now()
    reviews := []models.Review{
        {
            ID:          "review1",
            AppID:       "app1",
            Author:      "User1",
            Content:     "App1 review",
            Rating:      5,
            SubmittedAt: now.Add(-1 * time.Hour),
            FetchedAt:   now,
        },
        {
            ID:          "review2",
            AppID:       "app2",
            Author:      "User2",
            Content:     "App2 review",
            Rating:      4,
            SubmittedAt: now.Add(-1 * time.Hour),
            FetchedAt:   now,
        },
        {
            ID:          "review3",
            AppID:       "app3",
            Author:      "User3",
            Content:     "App3 review",
            Rating:      3,
            SubmittedAt: now.Add(-1 * time.Hour),
            FetchedAt:   now,
        },
    }

    storage.SaveReviews(reviews)

    // Test app1 filtering
    app1Reviews, err := storage.GetRecentReviews("app1", 24*time.Hour)
    if err != nil {
        t.Fatalf("GetRecentReviews failed: %v", err)
    }
    if len(app1Reviews) != 1 {
        t.Errorf("Expected 1 review for app1, got %d", len(app1Reviews))
    }
    if len(app1Reviews) > 0 && app1Reviews[0].ID != "review1" {
        t.Errorf("Expected review1 for app1, got %s", app1Reviews[0].ID)
    }

    // Test app2 filtering
    app2Reviews, err := storage.GetRecentReviews("app2", 24*time.Hour)
    if err != nil {
        t.Fatalf("GetRecentReviews failed: %v", err)
    }
    if len(app2Reviews) != 1 {
        t.Errorf("Expected 1 review for app2, got %d", len(app2Reviews))
    }
    if len(app2Reviews) > 0 && app2Reviews[0].ID != "review2" {
        t.Errorf("Expected review2 for app2, got %s", app2Reviews[0].ID)
    }

    // Test non-existent app
    nonExistentReviews, err := storage.GetRecentReviews("nonexistent", 24*time.Hour)
    if err != nil {
        t.Fatalf("GetRecentReviews failed: %v", err)
    }
    if len(nonExistentReviews) != 0 {
        t.Errorf("Expected 0 reviews for nonexistent app, got %d", len(nonExistentReviews))
    }
}

func TestFileStorage_GetRecentReviews_EmptyStorage(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "empty.json")

    storage, _ := NewFileStorage(testFile)

    // Test on empty storage
    recentReviews, err := storage.GetRecentReviews("app1", 24*time.Hour)
    if err != nil {
        t.Fatalf("GetRecentReviews on empty storage failed: %v", err)
    }
    if len(recentReviews) != 0 {
        t.Errorf("Expected 0 reviews from empty storage, got %d", len(recentReviews))
    }
}

func TestFileStorage_GetRecentReviews_PersistenceAndLoad(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "persistence.json")

    // Create first storage instance and save reviews
    storage1, _ := NewFileStorage(testFile)

    now := time.Now()
    reviews := []models.Review{
        {
            ID:          "review1",
            AppID:       "app1",
            Author:      "User1",
            Content:     "Persistent review",
            Rating:      5,
            SubmittedAt: now.Add(-1 * time.Hour),
            FetchedAt:   now,
        },
    }

    err := storage1.SaveReviews(reviews)
    if err != nil {
        t.Fatalf("Failed to save reviews: %v", err)
    }

    // Create second storage instance and load from disk
    storage2, _ := NewFileStorage(testFile)
    err = storage2.LoadState()
    if err != nil {
        t.Fatalf("Failed to load state: %v", err)
    }

    // Test GetRecentReviews on loaded storage
    recentReviews, err := storage2.GetRecentReviews("app1", 24*time.Hour)
    if err != nil {
        t.Fatalf("GetRecentReviews after load failed: %v", err)
    }

    if len(recentReviews) != 1 {
        t.Errorf("Expected 1 review after load, got %d", len(recentReviews))
    }

    if len(recentReviews) > 0 && recentReviews[0].ID != "review1" {
        t.Errorf("Expected review1 after load, got %s", recentReviews[0].ID)
    }
}

func TestFileStorage_GetRecentReviews_ThreadSafety(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "concurrent.json")

    storage, _ := NewFileStorage(testFile)

    // Add initial reviews
    now := time.Now()
    reviews := []models.Review{
        {
            ID:          "review1",
            AppID:       "app1",
            Author:      "User1",
            Content:     "Concurrent review",
            Rating:      5,
            SubmittedAt: now.Add(-1 * time.Hour),
            FetchedAt:   now,
        },
    }
    storage.SaveReviews(reviews)

    // Test concurrent reads
    const numGoroutines = 10
    done := make(chan error, numGoroutines)

    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            recentReviews, err := storage.GetRecentReviews("app1", 24*time.Hour)
            if err != nil {
                done <- fmt.Errorf("goroutine %d: GetRecentReviews failed: %v", id, err)
                return
            }
            if len(recentReviews) != 1 {
                done <- fmt.Errorf("goroutine %d: expected 1 review, got %d", id, len(recentReviews))
                return
            }
            done <- nil
        }(i)
    }

    // Wait for all goroutines and check for errors
    for i := 0; i < numGoroutines; i++ {
        if err := <-done; err != nil {
            t.Errorf("Concurrent access error: %v", err)
        }
    }
}

func TestFileStorage_GetRecentReviews_EdgeCaseTimes(t *testing.T) {
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "edge_times.json")

    storage, _ := NewFileStorage(testFile)

    now := time.Now()
    reviews := []models.Review{
        {
            ID:          "review1",
            AppID:       "app1",
            Author:      "User1",
            Content:     "Edge case review",
            Rating:      5,
            SubmittedAt: now.Add(-1 * time.Second), // Just 1 second ago
            FetchedAt:   now,
        },
    }

    storage.SaveReviews(reviews)

    // Test very small time window (should still include the review)
    recentReviews, err := storage.GetRecentReviews("app1", 2*time.Second)
    if err != nil {
        t.Fatalf("GetRecentReviews failed: %v", err)
    }
    if len(recentReviews) != 1 {
        t.Errorf("Expected 1 review within 2 seconds, got %d", len(recentReviews))
    }

    // Test very small time window that excludes the review
    recentReviews, err = storage.GetRecentReviews("app1", 500*time.Millisecond)
    if err != nil {
        t.Fatalf("GetRecentReviews failed: %v", err)
    }
    if len(recentReviews) != 0 {
        t.Errorf("Expected 0 reviews within 500ms, got %d", len(recentReviews))
    }
}

// Helper function for error message checking
func contains(str, substr string) bool {
    return len(str) >= len(substr) && (str == substr || len(substr) == 0 ||
           (len(str) > len(substr) &&
            (str[:len(substr)] == substr ||
             str[len(str)-len(substr):] == substr ||
             containsHelper(str, substr))))
}

func containsHelper(str, substr string) bool {
    for i := 0; i <= len(str)-len(substr); i++ {
        if str[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}