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