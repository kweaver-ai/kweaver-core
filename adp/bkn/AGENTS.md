# Agent Collaboration Rules

This file applies globally to the `bkn` subsystem (unless overridden by a more specific `AGENTS.md` in subdirectories).

## Agent Persona

You are an experienced programming master, focused on efficient, scalable, compatible, maintainable, well-commented, and low-entropy code.

## Subsystem Overview

BKN Engine is a subsystem of ADP (AISHU Data Platform), consisting of two Go microservices:

| Service | Port | Purpose |
|---------|------|---------|
| `bkn-backend` | 13014 | BKN modeling, knowledge network management, CRUD operations |
| `ontology-query` | 13018 | Graph queries, semantic search, data retrieval |

## Code Standards

- All code comments (including docstrings) MUST be in English
- All new log messages MUST be in English
- Follow Go official code style, use `go fmt` for formatting
- Use `go vet` and `golangci-lint` for static analysis
- Follow the existing clean architecture pattern (interfaces → logics → adapters)

### Architecture Layers

```
server/
├── common/              # Shared utilities, constants, settings
├── config/              # Configuration files and loaders
├── drivenadapters/      # Data access layer (driven adapters - DB, OpenSearch, external APIs)
├── driveradapters/      # Interface adapters (driver adapters - HTTP handlers)
├── errors/              # Error definitions and codes
├── interfaces/          # Interface definitions and DTOs
├── locale/              # i18n support (TOML files)
├── logics/              # Business logic layer
├── main.go              # Application entry point
├── version/             # Version information
└── worker/              # Background tasks (bkn-backend only)
```

## Entropy Reduction Principle

Goal: Every change should make the system more "ordered" — easier to understand, more consistent, more maintainable; avoid introducing unnecessary complexity and noise.

### Scope

- Code, tests, configuration, scripts, documentation, directory structure, and naming

### Specific Rules

- Prioritize "root cause fixes" over temporary patch stacking
- Unless necessary, avoid large-scale unrelated formatting/renaming/file moves (reduce diff noise)
- Maintain consistency: follow existing architecture, naming, directory hierarchy, and style; new patterns require justification and migration strategy
- Reduce complexity: delete what can be deleted (dead code/unused deps/duplicate logic); merge what can be merged (duplicate configs/docs)
- Improve readability: clear abstraction levels, explicit interface boundaries, predictable default behavior; avoid "clever but hard to understand" patterns
- Verifiable changes: new/modified behavior should have minimal necessary tests or runnable examples; update docs accordingly

### PR Self-Check List

- [ ] Does the diff only include changes related to the objective?
- [ ] Does it reduce duplication/coupling/temporary logic rather than increase it?
- [ ] Are naming, directory structure, and style consistent with existing patterns?
- [ ] Are necessary tests/docs/examples included?

## Testing

### Environment Variables

| Variable | Value | Description |
|----------|-------|-------------|
| `I18N_MODE_UT` | `true` | Required for unit tests. Enables locale loading from source directory instead of working directory. |

### Running Tests

```bash
# Run all unit tests for bkn-backend
cd bkn-backend
go test ./... -v

# Run all unit tests for ontology-query (requires I18N_MODE_UT)
cd ontology-query/server

# PowerShell
$env:I18N_MODE_UT = "true"; go test ./... -v

# Bash / Linux / macOS
I18N_MODE_UT=true go test ./... -v

# Run tests with coverage
I18N_MODE_UT=true go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package tests
I18N_MODE_UT=true go test ./logics/... -v

# Run tests with race detection
I18N_MODE_UT=true go test ./... -race
```

### Test Conventions

- Test files: `*_test.go` in the same package as the code being tested
- Use `goconvey` (already in dependencies) for BDD-style tests
- Use `gomock` for mocking interfaces
- Integration tests: use `-tags=integration` build tag

## Documentation

### Directory Structure

Documentation for Ontology is located one level up from this subsystem:

```
adp/
├── docs/
│   └── bkn/
│       ├── bkn-backend/     # API specs (OpenAPI YAML)
│       └── ontology-query/       # API specs (OpenAPI YAML)
└── bkn/
    ├── AGENTS.md                 # This file
    ├── README.md                 # English documentation
    ├── README.zh.md              # Chinese documentation
    ├── bkn-backend/
    │   └── README.md             # Service-specific docs
    └── ontology-query/
        └── README.md             # Service-specific docs
```

### Language Policy

- **README files**: Bilingual (English primary, Chinese version with `.zh.md` suffix)
- **Code comments**: English only
- **API documentation**: OpenAPI YAML specs in `../docs/bkn/`
- **Design documents**: Chinese acceptable for internal team discussions

### Documentation Rules

- File names: lowercase with underscores, e.g., `getting_started.md`
- Use `$PROJECT_ROOT` or relative paths, avoid hardcoded personal paths
- Include runnable code examples
- Keep API docs in sync with code changes

## Configuration

Configuration files are located at:

- `bkn-backend/server/config/bkn-backend-config.yaml`
- `ontology-query/server/config/ontology-query-config.yaml`

Key configuration items:
- Database connections (MariaDB/DM8)
- OpenSearch connection
- Dependent service addresses
- Service port and logging level

## Dependencies

- **Go version**: 1.24.0+
- **Database**: MariaDB 11.4+ or DM8
- **Search engine**: OpenSearch 2.x
- **Key libraries**: Gin (HTTP), Viper (config), Zap (logging), OpenTelemetry (tracing)

## Build & Deploy

```bash
# Build binary
cd bkn-backend
go build -o bkn-backend .

# Docker build
docker build -t bkn-backend:latest -f docker/Dockerfile .

# Helm deploy
helm3 install bkn-backend helm/bkn-backend/
```

## Common Commands

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Generate mocks (if using mockgen)
go generate ./...
```
