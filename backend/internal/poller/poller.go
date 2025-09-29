package poller

import (
	"log"
	"sync"
	"time"
)

type Poller struct {
	logger       *log.Logger
	pollInterval time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	started      bool
}

func NewPoller(logger *log.Logger, interval time.Duration) *Poller {
	if logger == nil {
		logger = log.Default()
	}
	return &Poller{
		logger:       logger,
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

	// TODO: fetch and store reviews for apps
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
