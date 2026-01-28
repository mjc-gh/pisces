# AGENTS.md

This file provides guidance for AI coding agents working in the Pisces codebase.

## Project Overview

Pisces is a Go-based tool for analyzing phishing attack sites. It uses the Chrome
DevTools Protocol via `chromedp` to automate browser interactions for security
analysis. The project produces two binaries: `pisces` (CLI) and `pisces-web` (REST API).

**Go Version**: 1.24.4  
**Module**: `github.com/mjc-gh/pisces`

## Build Commands

```bash
make build.cli      # Build CLI to build/pisces
make build.web      # Build web server to build/pisces-web
make go.get         # Download dependencies
make go.tidy        # Tidy go.mod
```

## Test Commands

```bash
make test           # Run linting then all tests
go test ./...       # Run all tests without linting
go test -v ./...    # Run all tests with verbose output

# Run a single test by name
go test -v ./engine -run TestNewTask
go test -v ./... -run TestPerformTaskUnknownType

# Run tests matching a pattern
go test -v ./... -run "TestTask.*"

# Run tests in Docker (includes headless Chrome)
make test.docker
```

## Lint Commands

```bash
make check          # Run golangci-lint
golangci-lint run   # Direct lint invocation
gofmt -l .          # Check formatting (CI uses this)
```

The project uses `golangci-lint` v2 with nearly all linters enabled. See
`.golangci.yml` for the full configuration. Code must pass `gofmt` formatting.

## Project Structure

```
pisces/
├── cmd/cli/main.go           # CLI entry point
├── cmd/pisces-web/main.go    # Web server entry point
├── engine/                   # Core analysis engine (crawler, tasks, errors)
├── internal/browser/         # Browser configuration/profiles
├── internal/piscestest/      # Test utilities and fixtures
├── internal/rest/            # REST API server
├── rules/                    # Sigma detection rules (YAML)
└── logger.go                 # Logging setup (root package)
```

## Code Style Guidelines

### Import Ordering

Organize imports in three groups separated by blank lines:
1. Standard library packages
2. Third-party packages
3. Internal project packages

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Packages | lowercase, single word | `engine`, `browser`, `rest` |
| Variables/functions | camelCase | `userAgent`, `winWidth` |
| Exported names | PascalCase | `NewCrawler`, `AnalyzeResult` |
| Constants | SCREAMING_SNAKE_CASE | `SIZE_LARGE`, `PROFILE_DESKTOP` |
| Error variables | `Err` prefix + PascalCase | `ErrNoCrawlerVisit` |
| JSON struct tags | snake_case | `json:"requested_url"` |
| YAML struct tags | camelCase | `yaml:"browserProfile"` |

### Error Handling

1. **Define sentinel errors** at package level:
```go
var ErrNoCrawlerVisit = errors.New("no visit from crawler")
```

2. **Wrap errors with context**:
```go
return fmt.Errorf("create output file: %w", err)
```

3. **Log errors with zerolog**:
```go
logger.Warn().Err(err).Msg("file close error")
```

### Functional Options Pattern

Use functional options for configurable types:

```go
type Option func(*Engine)

func WithRemoteAllocator(host string, port int) Option {
    return func(e *Engine) {
        host := net.JoinHostPort(host, strconv.Itoa(port))
        e.config.remoteURL = fmt.Sprintf("http://%s/json/version", host)
    }
}
```

### Testing Patterns

1. **Use table-driven tests**:
```go
tests := []struct {
    name           string
    action         string
    expectedAction string
}{
    {name: "basic task", action: "navigate", expectedAction: "navigate"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        task := NewTask(tt.action, tt.input)
        assert.Equal(t, tt.expectedAction, task.action)
    })
}
```

2. **Use testify assertions**:
```go
assert.Equal(t, expected, actual)
require.NoError(t, err)  // Fails test immediately
require.Error(t, err)
```

3. **Mark parallel-safe tests** with `t.Parallel()`

4. **Use test utilities** from `internal/piscestest/`:
   - `NewTestWebServer()` - HTTP test server with embedded testdata
   - `NewTestContext()` - Browser context (respects env vars for remote/headfull)
   - `FindByID[T]()` - Generic helper for finding structs by ID

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/chromedp/chromedp` | Chrome DevTools Protocol automation |
| `github.com/bradleyjkemp/sigma-go` | Sigma rule detection |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/urfave/cli/v3` | CLI framework |
| `github.com/gin-gonic/gin` | HTTP web framework |
| `github.com/stretchr/testify` | Test assertions |

## Docker

```bash
make test.docker      # Run tests in Docker (includes headless Chrome)
make run.docker ARGS="--help"  # Run CLI in Docker
make chromedp.run     # Start headless Chrome for local development (port 9222)
make chromedp.pull    # Pull latest Chrome image
```
