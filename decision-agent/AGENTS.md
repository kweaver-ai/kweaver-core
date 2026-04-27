# AGENTS.md - Decision Agent Development Guide

## Project Overview

Decision Agent is a microservices architecture with:
- **Go Services**: agent-factory (DDD with hexagonal architecture)
- **Python Services**: agent-executor, agent-memory, data-retrieval (FastAPI)
- **Frontend**: agent-web (React 18.3.1 + TypeScript + Ant Design)

## Development Commands

### Python Services (agent-executor, agent-memory, data-retrieval)

```bash
# Setup
make dev-setup          # Set up development environment
make uv-sync            # Sync UV dependencies
make run               # Run the application

# Testing - agent-executor & agent-memory
make test              # Run all tests; make test-unit / test-integration
uv run pytest test/unit/test_file.py::test_function -v  # Single test

# Testing - data-retrieval
uv run pytest tests/unit_tests/ -v  # Run all tests; --cov for coverage

# Code Quality - agent-executor & agent-memory
make lint              # Ruff linting; make format to format code
uv run ruff check src  # Check specific directory; ruff format to format

# Code Quality - data-retrieval
uv run flake8 src/     # Flake8 linting (max line length 120); uv run ruff format src/
```

### Go Services (agent-factory)

```bash
# Code Quality
make all               # Format (gofumpt + goimports) and lint
make fmt               # Format with gofumpt and goimports
make ciLint            # golangci-lint checks; make ciLintFix to auto-fix

# Testing
make goTest            # Run all tests; make ut for coverage report
go test -v ./...       # Run tests with verbose output
go test ./path/to/package -run TestSpecificFunction  # Single test
go test ./... -run TestSpecificFunction/SubTest  # Subtest
go test -v ./src/infra/common/util/string_test.go  # Specific file

# Build & Run
make localRun          # Build and run locally
make goGenerate        # Generate mocks and code
make killLocal         # Kill local instance on port 13020
make reRunLocal        # Restart local instance
```

### Frontend (agent-web)

```bash
# Development
npm run dev            # Development server
npm run build          # Production build
npm run preview        # Preview production build

# Testing
npm run test           # Jest tests
npm run test -- --testNamePattern="SpecificTest"  # Single test
npm run test -- --testPathPattern="specific/file"  # Single file

# Code Quality
npm run lint           # ESLint checks
```

## Code Style Guidelines

### Python (FastAPI Services)

**Linting & Formatting**:
- agent-executor & agent-memory: Ruff (configured in pyproject.toml)
- data-retrieval: flake8 + ruff format, type hints required

**Import Order**: standard library → third-party → local

**Naming**:
- snake_case (variables/functions)
- PascalCase (classes/Pydantic models)
- UPPER_SNAKE_CASE (constants)

**Patterns**:
- async/await for endpoints
- Pydantic for validation
- custom exception classes
- structlog for logging

**Max line length**: 120 characters

**Testing**: pytest with asyncio, markers (unit, integration, slow), coverage ≥90%

### Go (DDD Architecture)

**Linting & Formatting**: golangci-lint, gofumpt, goimports (see Makefile)

**Import Order**: standard library → third-party → local

**Naming**:
- PascalCase (exported)
- camelCase (private)
- snake_case (packages)

**DDD Structure**:
- domain (entity/service/valueobject/aggregate)
- drivenadapter (infrastructure)
- driveradapter (API)
- port (interfaces)

**Patterns**:
- Repository pattern
- singleton with sync.Once
- service layer with DTO injection
- interface-based design

**Error handling**: pkg/errors.Wrap for wrapping errors

**Testing**: standard go test with testify, table-driven tests, go generate for mocks

### TypeScript/React (Frontend)

**Linting & Formatting**: ESLint + TypeScript rules + Prettier (see eslint.config.mjs)

**Import Order**: React → third-party → local (use @/ alias for absolute imports)

**Naming**:
- PascalCase (components/interfaces)
- camelCase (variables/functions)
- kebab-case (files/CSS)

**Patterns**:
- Functional components with hooks
- MobX observer pattern
- Ant Design library
- custom hooks for business logic

**Testing**: Jest with React Testing Library, test files: src/**/*.test.tsx

## Important Notes

- Always run tests and linting before committing
- Python: UV package manager, max line length 120
- Go: sync.Once for singletons, context propagation for tracing, port 13020
- Frontend: absolute imports with @/, MobX for state management
- No Cursor or Copilot rules in this repository
