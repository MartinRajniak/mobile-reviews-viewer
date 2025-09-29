package testutil

import (
	"sync"

	"backend/internal/models"
	"backend/internal/storage"
)

// MockStorage implements storage.Storage interface for testing
type MockStorage struct {
	mu      sync.RWMutex
	reviews map[string]models.Review
	saveErr error
	loadErr error
}

// Verify interface implementation at compile time
var _ storage.Storage = (*MockStorage)(nil)

func NewMockStorage() *MockStorage {
	return &MockStorage{
		reviews: make(map[string]models.Review),
	}
}

func (m *MockStorage) SaveReviews(reviews []models.Review) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, review := range reviews {
		m.reviews[review.ID] = review
	}
	return nil
}

func (m *MockStorage) GetAllReviews() ([]models.Review, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]models.Review, 0, len(m.reviews))
	for _, review := range m.reviews {
		result = append(result, review)
	}
	return result, nil
}

func (m *MockStorage) LoadState() error {
	return m.loadErr
}

func (m *MockStorage) SaveState() error {
	return nil
}

func (m *MockStorage) SetSaveError(err error) {
	m.saveErr = err
}

func (m *MockStorage) SetLoadError(err error) {
	m.loadErr = err
}

func (m *MockStorage) GetSavedReviewCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.reviews)
}

// Reset clears all stored reviews and errors
func (m *MockStorage) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reviews = make(map[string]models.Review)
	m.saveErr = nil
	m.loadErr = nil
}

// HasReview checks if a review with the given ID exists
func (m *MockStorage) HasReview(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.reviews[id]
	return exists
}

// GetReview retrieves a specific review by ID
func (m *MockStorage) GetReview(id string) (models.Review, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	review, exists := m.reviews[id]
	return review, exists
}