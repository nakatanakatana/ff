## Gemini
This file outlines the project's conventions and configurations to guide Gemini in making accurate and effective contributions.

### Project Overview
`ff` is a Go-based RSS feed filtering and modification tool. It acts as a web service that fetches RSS feeds, applies a series of filters and modifiers based on query parameters, and serves the modified feed. The main application logic is in `cmd/ff/main.go`, which sets up an HTTP server.

### Key Technologies
- **Language:** Go (version 1.16)
- **Key Libraries:**
    - `github.com/mmcdole/gofeed`: For parsing RSS/Atom feeds.
    - `github.com/gorilla/feeds`: For generating RSS/Atom feeds.
- **CI/CD:** GitHub Actions (`.github/workflows/ci.yaml`)
- **Linting:** `golangci-lint` (configuration in `.golangci.yaml`)

### Development Workflow
- **Building:** To build the project, run:
  ```sh
  go build -o dist/ ./cmd/...
  ```
- **Testing:** To run tests, use:
  ```sh
  go test ./... -v -coverprofile=coverage.txt -covermode=atomic
  ```
- **Linting:** To run the linter, use:
  ```sh
  go tool golangci-lint run
  ```
  The configuration for `golangci-lint` is located in `.golangci.yaml`. Please adhere to the rules defined in this file.

### Code Style and Conventions
- **Filtering and Modification:** The core logic revolves around `FilterFunc` and `ModifierFunc`. These functions are dynamically created and applied based on URL query parameters.
  - `filter.go` and `filterFn.go` contain the filtering logic.
  - `modifier.go` and `modifierFn.go` contain the modification logic.
- **Configuration:** The application can be configured via environment variables (e.g., `MUTE_AUTHORS`, `MUTE_URLS`, `LATEST_ONLY`).
- **Caching:** The application uses a cache middleware (`cache.go`) to improve performance.

### How to Contribute
When making changes, please ensure that:
1.  All new code is accompanied by corresponding tests.
2.  The code is formatted according to Go standards (`go fmt`).
3.  The code passes all linting checks (`go tool golangci-lint run`).
4.  The changes are documented where necessary.
5.  After making code changes, confirm that all checks (lint, format, build, test) pass.
