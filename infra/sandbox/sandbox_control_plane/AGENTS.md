# Repository Guidelines

## Project Structure & Module Organization
`sandbox_control_plane/` is a FastAPI service organized with hexagonal architecture. Core code is under `src/`: `domain/` holds business rules and entities, `application/` coordinates use cases, `infrastructure/` implements adapters, and `interfaces/` exposes REST and other entry points. Tests live in `tests/`, split into `unit/`, `integration/`, and contract-oriented suites. Utility scripts live in `scripts/`.

## Build, Test, and Development Commands
Run `uv sync` to install and lock dependencies into the local environment. Start the API locally with `uvicorn src.interfaces.rest.main:app --reload --port 8000`. Run `uv run pytest` for the full test suite, or target `uv run pytest tests/unit` and `uv run pytest tests/integration` during development. Before opening a PR, run `uv run black src tests`, `uv run ruff check src tests`, and `uv run mypy src`.

## Coding Style & Naming Conventions
Use Python 3.11+, 4-space indentation, and explicit type hints. Black, Ruff, and Mypy enforce the baseline style, with a 100-character line limit. Keep modules, functions, and variables in `snake_case`; classes use `PascalCase`. Preserve dependency direction: `interfaces -> application -> domain`, while `infrastructure` implements domain and application ports without leaking framework details into core logic.

## Testing Guidelines
Use `pytest` markers consistently: `unit`, `integration`, `contract`, and `slow`. Put isolated logic tests in `tests/unit/`; place database, storage, or scheduler-backed cases in `tests/integration/`. Use coverage only when needed, for example `uv run pytest --cov=src --cov-report=html`. Add tests alongside behavior changes, especially for services and repository adapters.

## Commit & Pull Request Guidelines
Match the repository’s commit style: `feat:`, `fix(scope):`, `docs(scope):`, `chore(release):`. Keep commits narrow and explain the behavioral change, not just the file touched. Pull requests should summarize affected flows, mention schema or deployment implications, link related issues, and note any required environment variables or migration steps.
