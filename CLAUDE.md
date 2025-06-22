# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build
```sh
go build -o dist/ ./cmd/...
```

### Lint
```sh
go tool golangci-lint run
```

### Format
```sh
go fmt ./...
```

### Test
```sh
go test ./... -v -coverprofile=coverage.txt -covermode=atomic
```

## Development Workflow

**IMPORTANT**: After making any code changes, always run the following commands in order to ensure code quality:

1. **Format code**: `go fmt ./...`
2. **Run tests**: `go test ./... -v`
3. **Run linter**: `go tool golangci-lint run`
4. **Build project**: `go build -o dist/ ./cmd/...`

This workflow ensures that all code changes are properly formatted, tested, and free of linting errors before committing.

## Architecture

This is a Go-based RSS feed filtering and modification service that:

1. **Fetches RSS feeds** and parses them using `github.com/mmcdole/gofeed`
2. **Applies filters** to RSS items based on query parameters (title, description, link, author matching)
3. **Applies modifiers** to transform RSS items (remove description/content)
4. **Converts and serves** the filtered feed using `github.com/gorilla/feeds`

### Core Components

- **Main HTTP server** (`cmd/ff/main.go`) - Listens on port 8080, handles environment variables for muting authors/URLs
- **Filter system** (`filter.go`, `filterFn.go`) - Provides filtering functions for RSS items based on various criteria
- **Modifier system** (`modifier.go`, `modifierFn.go`) - Provides modification functions to transform RSS items
- **Converter** (`converter.go`) - Converts between gofeed and gorilla/feeds formats
- **Query parser** (`func.go`) - Parses URL query parameters into filter and modifier functions

### Workspace Structure

- Uses Go workspaces with two modules: root module and `tools/` module
- The `tools/` module contains golangci-lint for linting
- Go version: 1.24.3 (workspace), 1.16 (main module)

### Environment Variables

- `MUTE_AUTHORS` - Comma-separated list of authors to filter out
- `MUTE_URLS` - Comma-separated list of URLs to filter out  
- `LATEST_ONLY` - When set, applies latest-only filters for published_at and updated_at