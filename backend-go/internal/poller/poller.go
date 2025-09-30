package poller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"backend/internal/models"
	"backend/internal/storage"
)

type Poller struct {
	storage      storage.Storage
	logger       *log.Logger
	client       *http.Client
	appIDs       []string
	pollInterval time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	started      bool
}

func NewPoller(storage storage.Storage, logger *log.Logger, appIDs []string, interval time.Duration) *Poller {
	if logger == nil {
		logger = log.Default()
	}
	return &Poller{
		storage: storage,
		logger:  logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		appIDs:       appIDs,
		pollInterval: interval,
		stopChan:     make(chan struct{}),
	}
}

func (p *Poller) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return // Already started, prevent multiple goroutines
	}

	p.started = true
	p.wg.Add(1)
	go p.run()
}

func (p *Poller) run() {
	defer p.wg.Done()

	// Poll immediately on startup (e.g. after restart)
	p.pollAllAppsConcurrently()

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.pollAllAppsConcurrently()
		case <-p.stopChan:
			return
		}
	}
}

// TODO: check if rate-limiting is needed
func (p *Poller) pollAllAppsConcurrently() {
	p.logger.Println("Polling all apps concurrently...")

	start := time.Now()

	var wg sync.WaitGroup

	for _, appID := range p.appIDs {
		wg.Add(1)

		// Launch goroutine for each app
		go func(id string) {
			defer wg.Done()

			if err := p.fetchAndStore(id); err != nil {
				p.logger.Printf("Error polling app %s: %v", id, err)
			} else {
				p.logger.Printf("Successfully polled app %s", id)
			}
		}(appID)
	}

	// Wait for all apps to complete
	wg.Wait()

	p.logger.Printf("Poll complete in %v", time.Since(start))
}

func (p *Poller) fetchAndStore(appID string) error {
	p.logger.Printf("Fetching reviews for app %s", appID)

	url := fmt.Sprintf(
		"https://itunes.apple.com/us/rss/customerreviews/id=%s/sortBy=mostRecent/page=1/json",
		appID,
	)

	reviews, err := p.fetchReviews(url, appID)
	if err != nil {
		return err
	}

	if len(reviews) == 0 {
		p.logger.Printf("No reviews found for app %s", appID)
		return nil
	}

	if err := p.storage.SaveReviews(reviews); err != nil {
		return err
	}

	p.logger.Printf("Stored %d reviews for app %s", len(reviews), appID)
	return nil
}

// TODO: add error handling and retry logic
func (p *Poller) fetchReviews(url, appID string) ([]models.Review, error) {
	// Send HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "AppReviewPoller/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reviews: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse RSS feed
	var feed RSSFeed
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("failed to decode RSS feed: %w", err)
	}

	// Convert to internal Review models
	reviews := make([]models.Review, 0, len(feed.Feed.Entry))
	now := time.Now()

	for _, entry := range feed.Feed.Entry {
		review, err := p.parseReviewEntry(entry, appID, now)
		if err != nil {
			p.logger.Printf("Warning: failed to parse review entry: %v", err)
			continue // Skip malformed entries
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (p *Poller) parseReviewEntry(entry RSSEntry, appID string, fetchedAt time.Time) (models.Review, error) {
	// Parse rating
	rating, err := strconv.Atoi(entry.Rating.Label)
	if err != nil {
		return models.Review{}, fmt.Errorf("invalid rating: %w", err)
	}

	// Parse submission timestamp
	submittedAt, err := time.Parse(time.RFC3339, entry.Updated.Label)
	if err != nil {
		return models.Review{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Generate a unique ID from the entry ID
	// The entry.ID.Label looks like: "https://itunes.apple.com/us/review?id=12345&type=..."
	// We'll use this as the unique identifier
	reviewID := entry.ID.Label

	return models.Review{
		ID:          reviewID,
		AppID:       appID,
		Author:      entry.Author.Name.Label,
		Content:     entry.Content.Label,
		Rating:      rating,
		SubmittedAt: submittedAt,
		FetchedAt:   fetchedAt,
	}, nil
}

func (p *Poller) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return // Not started, nothing to stop
	}

	select {
	case <-p.stopChan:
		// Already closed
	default:
		close(p.stopChan)
	}

	p.wg.Wait()
	p.started = false
	p.stopChan = make(chan struct{})
}
