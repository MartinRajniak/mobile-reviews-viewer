# Recent iOS App store reviews viewer

## Architecture
```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   RSS Feed  │ ───> │  Go Backend  │ <──> │  React UI   │
│ (App Store) │      │   (Poller)   │      │  (Viewer)   │
└─────────────┘      └──────────────┘      └─────────────┘
                            │
                            ▼
                     ┌─────────────┐
                     │   Storage   │
                     │   (JSON)    │
                     └─────────────┘
```

## Backend
A backend service/app that polls an iOS app’s App Store Connect RSS feed to fetch and store App Store reviews for a specific iOS app.

### Project Structure
```
backend/
├── main.go
├── config/
│   ├── apps.json            # Application IDs to poll for
├── internal/
│   ├── poller/
│   │   ├── poller.go        # RSS polling logic
│   │   └── poller_test.go
│   └── testutil/
│       └── buffer.go        # Thread-safe buffer for log testing
```

### How to start
```bash
cd backend
go run main.go
```

### How to test
```bash
cd backend
go test -v ./...
```