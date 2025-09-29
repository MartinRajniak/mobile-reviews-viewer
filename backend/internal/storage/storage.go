package storage

import (
	"backend/internal/models"
)

// Storage defines the interface for review persistence
type Storage interface {
	SaveReviews(reviews []models.Review) error
	GetAllReviews() ([]models.Review, error)
	LoadState() error
	SaveState() error
}
