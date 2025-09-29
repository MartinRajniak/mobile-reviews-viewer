package poller

import (
	"log"
	"sync"
	"time"
)

type Poller struct {
	logger       *log.Logger
	appIDs       []string
	pollInterval time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	started      bool
}

func NewPoller(logger *log.Logger, appIDs []string, interval time.Duration) *Poller {
	if logger == nil {
		logger = log.Default()
	}
	return &Poller{
		logger:       logger,
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
    return nil
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
