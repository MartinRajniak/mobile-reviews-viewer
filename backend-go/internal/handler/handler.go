package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"backend/internal/storage"
)

type Handler struct {
	storage storage.Storage
}

func NewHandler(storage storage.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// GetRecentReviews handles GET /api/reviews
// Query parameters:
//   - app_id: (required) The iTunes app ID
//   - hours: (optional) Number of hours to look back (default: 720 - 30 days)
func (h *Handler) GetRecentReviews(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get app_id from query params
	appID := r.URL.Query().Get("app_id")
	if appID == "" {
		http.Error(w, "app_id query parameter is required", http.StatusBadRequest)
		return
	}

	// Get hours parameter (default to 48 - 2 days)
	hoursStr := r.URL.Query().Get("hours")
	hours := 48
	if hoursStr != "" {
		parsedHours, err := strconv.Atoi(hoursStr)
		if err != nil || parsedHours <= 0 {
			http.Error(w, "hours must be a positive integer", http.StatusBadRequest)
			return
		}
		hours = parsedHours
	}

	// Calculate time window
	since := time.Duration(hours) * time.Hour

	// Fetch reviews from storage
	reviews, err := h.storage.GetRecentReviews(appID, since)
	if err != nil {
		log.Printf("Error fetching reviews: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Sort by newest first (submitted_at descending)
	sort.Slice(reviews, func(i, j int) bool {
		return reviews[i].SubmittedAt.After(reviews[j].SubmittedAt)
	})

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(reviews); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// HealthCheck handles GET /api/health
// Returns a simple health check response
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get total review count
	allReviews, err := h.storage.GetAllReviews()
	if err != nil {
		log.Printf("Error getting all reviews: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"status":        "healthy",
		"timestamp":     time.Now().Format(time.RFC3339),
		"total_reviews": len(allReviews),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAverageRating handles GET /api/average-rating
// Query parameters:
//   - app_id: (required) The iTunes app ID
//   - hours: (optional) Number of hours to look back (default: 48)
func (h *Handler) GetAverageRating(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get app_id from query params
	appID := r.URL.Query().Get("app_id")
	if appID == "" {
		http.Error(w, "app_id query parameter is required", http.StatusBadRequest)
		return
	}

	// Get hours parameter (default to 48)
	hoursStr := r.URL.Query().Get("hours")
	hours := 48
	if hoursStr != "" {
		parsedHours, err := strconv.Atoi(hoursStr)
		if err != nil || parsedHours <= 0 {
			http.Error(w, "hours must be a positive integer", http.StatusBadRequest)
			return
		}
		hours = parsedHours
	}

	// Calculate time window
	since := time.Duration(hours) * time.Hour

	// Fetch reviews from storage
	reviews, err := h.storage.GetRecentReviews(appID, since)
	if err != nil {
		log.Printf("Error fetching reviews: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Calculate average rating
	var totalRating int
	reviewCount := len(reviews)

	for _, review := range reviews {
		totalRating += review.Rating
	}

	var averageRating float64
	if reviewCount > 0 {
		averageRating = float64(totalRating) / float64(reviewCount)
		// Round to 1 decimal place
		averageRating = float64(int(averageRating*10+0.5)) / 10
	}

	response := map[string]any{
		"app_id":         appID,
		"average_rating": averageRating,
		"review_count":   reviewCount,
		"hours":          hours,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
