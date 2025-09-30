# Mobile Reviews Viewer

A service for monitoring iOS App Store reviews with concurrent polling, persistent storage, and HTTP API endpoints. Available in two backend implementations: Go and Kotlin (Ktor).

## Backend Implementations

This project includes two backend implementations:

- **`backend-go/`**: Go implementation with comprehensive test coverage (69 tests)
- **`backend-kotlin/`**: Kotlin/Ktor implementation for those preferring the JVM ecosystem

Both backends provide the same REST API for the React frontend.

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

## Features

### Backend Service
- **Concurrent App Polling**: Monitor multiple iOS apps simultaneously
- **iTunes RSS API Integration**: Fetches most recent reviews (50 per app)
- **Persistent Storage**: JSON-based file storage with atomic writes
- **Review Deduplication**: ID-based review management prevents duplicates
- **Thread-Safe Operations**: Concurrent access with proper synchronization
- **HTTP API Endpoints**: REST API for retrieving reviews with time filtering
- **Comprehensive Testing**: Full test coverage with mock interfaces
- **Error Handling**: Graceful handling of network, parsing, and storage errors

### Frontend Application
- **React + TypeScript + Vite**: Modern development stack with hot module replacement
- **App Selector**: Switch between multiple monitored iOS apps
- **Real-Time Updates**: Auto-refresh reviews every 5 minutes
- **Review Display**: Card-based layout with star ratings and timestamps
- **Responsive UI**: Clean interface for viewing recent reviews
- **Error Handling**: User-friendly error and loading states
- **API Service Tests**: Comprehensive test coverage for API client with Vitest

### Project Structure
```
backend-go/                        # Go implementation
├── main.go                     # Application entry point with HTTP server
├── config/
│   └── apps.json              # Application IDs to poll for
├── internal/
│   ├── models/
│   │   └── review.go          # Review data model
│   ├── poller/
│   │   ├── poller.go          # Core polling engine with HTTP client
│   │   ├── poller_test.go     # Comprehensive test suite (25 tests)
│   │   └── rss_types.go       # iTunes RSS feed data structures
│   ├── storage/
│   │   ├── storage.go         # Storage interface definition
│   │   ├── file_storage.go    # JSON file storage implementation
│   │   └── file_storage_test.go # Storage test suite (18 tests)
│   ├── handler/
│   │   ├── handler.go         # HTTP API endpoints
│   │   └── handler_test.go    # HTTP handler test suite (17 tests)
│   └── testutil/
│       ├── buffer.go          # Thread-safe buffer for log testing
│       ├── mock_storage.go    # Mock storage for testing
│       └── mock_storage_test.go # Mock storage validation (9 tests)

backend-kotlin/                 # Kotlin/Ktor implementation
├── src/
│   └── main/kotlin/com/example/
└── build.gradle.kts           # Gradle build configuration

frontend/
├── src/
│   ├── main.tsx               # React application entry point
│   ├── App.tsx                # Main app component with routing
│   ├── App.css                # Global application styles
│   ├── components/
│   │   ├── AppSelector.tsx    # App switcher dropdown component
│   │   ├── ReviewList.tsx     # Review listing with auto-refresh
│   │   └── ReviewCard.tsx     # Individual review card component
│   ├── services/
│   │   ├── api.ts             # API client for backend endpoints
│   │   └── api.test.ts        # API client test suite (5 tests)
│   └── types/
│       └── review.ts          # TypeScript type definitions
├── vite.config.ts             # Vite build & test configuration
└── package.json               # Dependencies and scripts
```

## Key Components

### Poller Engine (`internal/poller/`)
- **Concurrent Processing**: Polls multiple app RSS feeds simultaneously
- **HTTP Client**: 30-second timeout with proper User-Agent headers
- **Review Fetching**: Gets 50 most recent reviews per app from iTunes RSS API
- **Error Recovery**: Continues polling other apps if one fails
- **Review Parsing**: Converts iTunes RSS format to internal Review model

### Storage System (`internal/storage/`)
- **Interface-Based Design**: Pluggable storage backends
- **File Storage**: JSON persistence with atomic writes (temp file + rename)
- **Thread Safety**: Concurrent read/write operations with RWMutex
- **Data Integrity**: Review deduplication by ID
- **Time-Based Queries**: GetRecentReviews with configurable time window

### HTTP API (`internal/handler/`)
- **REST Endpoints**: JSON API for accessing stored reviews
- **Time Filtering**: Configurable hours parameter (default: 30 days)
- **Health Monitoring**: Health check endpoint with storage stats
- **CORS Support**: Cross-origin requests enabled
- **Error Handling**: Proper HTTP status codes and error responses

### Review Model (`internal/models/`)
```go
type Review struct {
    ID          string    `json:"id"`           // Unique iTunes review ID
    AppID       string    `json:"app_id"`       // iTunes app ID
    Author      string    `json:"author"`       // Review author name
    Content     string    `json:"content"`      // Review text content
    Rating      int       `json:"rating"`       // Star rating (1-5)
    SubmittedAt time.Time `json:"submitted_at"` // When user submitted
    FetchedAt   time.Time `json:"fetched_at"`   // When we fetched it
}
```

### Test Infrastructure (`internal/testutil/`)
- **MockStorage**: Full storage interface implementation for testing
- **SafeBuffer**: Thread-safe buffer for concurrent log testing
- **Test Utilities**: Reusable components for comprehensive test coverage

## Usage

### Backend Quick Start

**Go Backend:**
```bash
cd backend
go run main.go
```

**Kotlin/Ktor Backend:**
```bash
cd backend-kotlin
./gradlew run
```

Both backend servers run on `http://localhost:8080` by default.

### Frontend Quick Start
```bash
cd frontend
npm install
npm run dev
```

The frontend dev server runs on `http://localhost:5173` by default.

### Configuration

**Backend**: Edit `backend/config/apps.json` to specify which iOS apps to monitor:
```json
{
  "apps": ["595068606", "447188370", "310633997"]
}
```

**Frontend**: The app selector in `src/components/AppSelector.tsx` should match the apps configured in the backend.

### Testing & Development

**Go Backend**:
```bash
cd backend

# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./internal/poller
go test -v ./internal/storage
go test -v ./internal/testutil

# Test with coverage
go test -cover ./...

# Format code
go fmt ./...

# Vet code
go vet ./...

# Build binary
go build -o mobile-reviews-poller main.go
```

**Kotlin/Ktor Backend**:
```bash
cd backend-kotlin

# Run tests
./gradlew test

# Build project
./gradlew build

# Build fat JAR
./gradlew buildFatJar

# Build docker image
./gradlew buildImage

# Run with docker
./gradlew runDocker
```

**Frontend**:
```bash
cd frontend

# Run tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run linter
npm run lint

# Build for production
npm run build

# Preview production build
npm run preview
```

## API Integration

The poller integrates with iTunes Store RSS feeds:
- **Endpoint**: `https://itunes.apple.com/us/rss/customerreviews/id={APP_ID}/sortBy=mostRecent/page=1/json`
- **User Agent**: `AppReviewPoller/1.0`
- **Format**: JSON RSS feed with nested review entries

## HTTP API Endpoints

### GET /api/reviews
Returns recent reviews for a specific app with time filtering.

**Parameters:**
- `app_id` (required): iTunes app ID
- `hours` (optional): Hours to look back (default: 720 - 30 days)

**Example:**
```bash
curl "http://localhost:8080/api/reviews?app_id=595068606&hours=48"
```

**Response:**
```json
[
  {
    "id": "12345678",
    "app_id": "595068606",
    "author": "John Doe",
    "content": "Great app!",
    "rating": 5,
    "submitted_at": "2025-09-29T10:30:00Z",
    "fetched_at": "2025-09-29T11:00:00Z"
  }
]
```

### GET /api/health
Returns service health status and review statistics.

**Example:**
```bash
curl "http://localhost:8080/api/health"
```

**Response:**
```json
{
  "status": "healthy",
  "total_reviews": 1234
}
```

## Test Coverage

### Backend
- **69 total tests** across all packages
- **Poller**: 25 tests covering HTTP requests, parsing, and error handling
- **Storage**: 18 tests covering file I/O, concurrency, and edge cases
- **Handler**: 17 tests covering HTTP endpoints and API functionality
- **Test Utils**: 9 tests validating mock implementations
- **Coverage Areas**: Happy path, error conditions, edge cases, concurrency

### Frontend
- **5 tests** for API service
- **API Client**: Tests for fetch operations, error handling, and parameter validation
- **Coverage Areas**: Default/custom parameters, HTTP errors, network failures, empty responses