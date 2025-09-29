package storage

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sync"

	"backend/internal/models"
)

type FileStorage struct {
	filepath string
	mu       sync.RWMutex
	reviews  map[string]models.Review
}

// Verify that FileStorage implements Storage interface at compile time
var _ Storage = (*FileStorage)(nil)

func NewFileStorage(filePath string) (*FileStorage, error) {
    // Ensure directory exists
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create directory: %w", err)
    }

    return &FileStorage{
        filepath: filePath,
        reviews:  make(map[string]models.Review),
    }, nil
}

// SaveReviews adds new reviews to storage and persists to disk
func (fs *FileStorage) SaveReviews(reviews []models.Review) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, review := range reviews {
		fs.reviews[review.ID] = review
	}

	return fs.persist()
}

func (fs *FileStorage) persist() error {
    // Convert map to slice for JSON serialization
    reviewSlice := make([]models.Review, 0, len(fs.reviews))
    for _, review := range fs.reviews {
        reviewSlice = append(reviewSlice, review)
    }

    // Marshal to JSON with indentation for readability
    data, err := json.MarshalIndent(reviewSlice, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal reviews: %w", err)
    }

    // Atomic write: write to temp file, then rename
    // This prevents corruption if the process crashes mid-write
    tempFile := fs.filepath + ".tmp"
    
    if err := os.WriteFile(tempFile, data, 0644); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }

    // Atomic rename - this is atomic on most filesystems
    if err := os.Rename(tempFile, fs.filepath); err != nil {
        return fmt.Errorf("failed to rename temp file: %w", err)
    }

    return nil
}

// LoadState loads reviews from disk into memory
func (fs *FileStorage) LoadState() error {
    fs.mu.Lock()
    defer fs.mu.Unlock()

    // Check if file exists
    if _, err := os.Stat(fs.filepath); os.IsNotExist(err) {
        // File doesn't exist yet - this is fine on first run
        return nil
    }

    // Read file
    data, err := os.ReadFile(fs.filepath)
    if err != nil {
        return fmt.Errorf("failed to read file: %w", err)
    }

    // Unmarshal JSON
    var reviewSlice []models.Review
    if err := json.Unmarshal(data, &reviewSlice); err != nil {
        return fmt.Errorf("failed to unmarshal reviews: %w", err)
    }

    // Load into map
    fs.reviews = make(map[string]models.Review, len(reviewSlice))
    for _, review := range reviewSlice {
        fs.reviews[review.ID] = review
    }

    return nil
}

// SaveState explicitly persists current state (called on shutdown)
func (fs *FileStorage) SaveState() error {
    fs.mu.Lock()
    defer fs.mu.Unlock()

    return fs.persist()
}

// GetAllReviews returns all stored reviews
func (fs *FileStorage) GetAllReviews() ([]models.Review, error) {
    fs.mu.RLock()
    defer fs.mu.RUnlock()

    result := make([]models.Review, 0, len(fs.reviews))
    for _, review := range fs.reviews {
        result = append(result, review)
    }

    return result, nil
}