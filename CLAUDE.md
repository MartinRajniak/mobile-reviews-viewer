# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mobile Reviews Viewer is a service for monitoring iOS App Store reviews with two backend implementations (Go and Kotlin/Ktor) and a React frontend. The backends poll iTunes RSS feeds, store reviews persistently, and expose REST APIs. The frontend displays reviews with auto-refresh.

## Architecture

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   RSS Feed  │ ───> │   Backend    │ <──> │  React UI   │
│ (App Store) │      │   (Poller)   │      │  (Viewer)   │
└─────────────┘      └──────────────┘      └─────────────┘
                            │
                            ▼
                     ┌─────────────┐
                     │ File Storage│
                     │ (JSON/Disk) │
                     └─────────────┘
```

**Key architectural concepts:**
- **Concurrent polling**: Both backends poll multiple apps simultaneously (Go uses goroutines, Kotlin uses coroutines)
- **Storage interface pattern**: Pluggable storage backends with interface-based design
- **Atomic writes**: JSON files written to temp file first, then renamed to prevent corruption
- **Review deduplication**: Reviews tracked by unique ID to prevent duplicates
- **Thread-safe operations**: Concurrent access with RWMutex (Go) or ConcurrentHashMap (Kotlin)
- **Time-based filtering**: API supports retrieving reviews within configurable time windows

## Repository Structure

- **`backend-go/`**: Go implementation (69 tests)
  - `main.go`: Entry point with HTTP server
  - `internal/models/`: Review data model
  - `internal/poller/`: RSS feed polling engine
  - `internal/storage/`: File storage with interface abstraction
  - `internal/handler/`: HTTP API endpoints
  - `internal/testutil/`: Mock implementations and test utilities
  - `config/apps.json`: App IDs to monitor

- **`backend-kotlin/`**: Kotlin/Ktor implementation (46 tests)
  - `src/main/kotlin/com/example/Application.kt`: Entry point, lifecycle management
  - `src/main/kotlin/com/example/Routing.kt`: HTTP routes
  - `src/main/kotlin/com/example/Polling.kt`: Coroutine-based polling service
  - `src/main/kotlin/com/example/reviews/`: Review domain logic
    - `Model.kt`: Data models and interfaces
    - `ReviewsRepository.kt`: Orchestrates fetching and storage
    - `itunes/ITunesReviewsFetcher.kt`: iTunes RSS client
    - `file_storage/ReviewsFileStorage.kt`: JSON persistence
  - `src/main/resources/application.yaml`: Server and polling config
  - `src/main/resources/config.json`: App IDs to monitor
  - `src/test/kotlin/com/example/TestFixtures.kt`: Shared test utilities (FakeReviewsStorage, FakeReviewsFetcher)

- **`frontend/`**: React + TypeScript + Vite (5 tests)
  - `src/components/`: UI components (AppSelector, ReviewList, ReviewCard)
  - `src/services/api.ts`: API client for backend
  - `src/types/review.ts`: TypeScript types

## Common Commands

### Go Backend

**Development:**
```bash
cd backend-go
go run main.go                    # Run server (http://localhost:8080)
go test ./...                      # Run all tests
go test -v ./internal/poller       # Run specific package tests
go test -cover ./...               # Test with coverage
go build -o mobile-reviews-poller main.go  # Build binary
```

**Code quality:**
```bash
go fmt ./...                       # Format code
go vet ./...                       # Vet code
```

### Kotlin Backend

**Development:**
```bash
cd backend-kotlin
./gradlew run                      # Run with Gradle (http://localhost:8080)
./run.sh                           # Run JAR directly (recommended for shutdown logs)
./gradlew test                     # Run all tests (46 tests)
./gradlew build                    # Build project
./gradlew buildFatJar              # Build fat JAR
```

**Testing:**
```bash
./gradlew test jacocoTestReport    # Test with coverage report
```

### Frontend

**Development:**
```bash
cd frontend
npm install                        # Install dependencies
npm run dev                        # Start dev server (http://localhost:5173)
npm test                           # Run tests
npm test -- --watch                # Run tests in watch mode
npm run build                      # Build for production
npm run lint                       # Run linter
```

## Configuration

### App IDs to Monitor

- **Go**: Edit `backend-go/config/apps.json`
- **Kotlin**: Edit `backend-kotlin/src/main/resources/config.json`

Format:
```json
{
  "apps": ["389801252", "447188370", "310633997"]
}
```

### Polling Interval (Kotlin only)

Edit `backend-kotlin/src/main/resources/application.yaml`:
```yaml
polling:
    intervalSeconds: 300  # 5 minutes (default)
```

### Frontend App Configuration

The app selector in `frontend/src/components/AppSelector.tsx` should match the backend configuration.

## API Endpoints

Both backends expose the same REST API:

**GET /api/reviews**
- Query params: `app_id` (required), `hours` (optional, default: 48)
- Returns: Array of reviews within time window

**GET /api/health**
- Returns: Service health status and total review count

## Testing Patterns

### Go Backend
- Mock storage interface defined in `internal/testutil/mock_storage.go`
- SafeBuffer for concurrent log testing in `internal/testutil/buffer.go`
- Table-driven tests for comprehensive coverage
- HTTP handler tests use `httptest.ResponseRecorder`

### Kotlin Backend
- Shared test fixtures in `TestFixtures.kt`: `FakeReviewsStorage`, `FakeReviewsFetcher`, `testJson`
- Ktor test host for HTTP endpoint testing
- Mock HTTP client for iTunes fetcher tests
- Coroutine test utilities (`runTest`, `TestScope`)
- SupervisorScope testing to ensure error isolation

### Frontend
- Vitest for unit tests
- API client tests mock global `fetch`

## Key Implementation Details

### Go Backend

**Concurrent polling** (`internal/poller/poller.go`):
- Uses `sync.WaitGroup` to coordinate goroutines
- 30-second HTTP timeout with User-Agent headers
- Continues polling other apps if one fails

**Storage** (`internal/storage/file_storage.go`):
- `sync.RWMutex` for thread-safe concurrent access
- Atomic writes: write to temp file → rename
- `GetRecentReviews()` filters by time window

### Kotlin Backend

**Coroutine-based polling** (`Polling.kt`):
- `SupervisorScope` isolates failures across apps
- Lifecycle tied to Ktor application events
- Proper cleanup with `cancelAndJoin()`

**Repository pattern** (`reviews/ReviewsRepository.kt`):
- Coordinates fetching from iTunes and saving to storage
- Launches parallel coroutines for each app
- Error handling per app with logging

**Dispatcher usage** (`reviews/file_storage/ReviewsFileStorage.kt`):
- Uses `Dispatchers.IO` for blocking file operations
- iTunes fetcher uses suspend functions (no explicit dispatcher needed)

**Lifecycle management** (`Application.kt`):
- Loads reviews from disk on startup
- Saves reviews to disk on shutdown via `ApplicationStopping` event
- Proper HTTP client cleanup

### Frontend

**Auto-refresh**: ReviewList polls backend every 5 minutes using `setInterval`

**API integration**: `services/api.ts` uses fetch with URL params for filtering

## Important Notes

- Both backends run on port 8080 by default
- Frontend dev server runs on port 5173
- Review data stored in `data/` directory (created automatically)
- iTunes RSS endpoint: `https://itunes.apple.com/us/rss/customerreviews/id={APP_ID}/sortBy=mostRecent/page=1/json`
- User-Agent header: `AppReviewPoller/1.0`
- Default time window for reviews: 48 hours (2 days), configurable via `hours` parameter

## Test Coverage Summary

- **Go Backend**: 69 tests (Poller: 25, Storage: 18, Handler: 17, TestUtil: 9)
- **Kotlin Backend**: 46 tests (Polling: 9, Repository: 6, Fetcher: 6, Storage: 6, Config: 6, Routing: 5, Application: 1)
- **Frontend**: 5 tests (API client)
