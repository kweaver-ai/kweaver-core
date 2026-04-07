# Repository Guidelines

## Project Structure & Module Organization
This repository is organized as a small monorepo. `sandbox_control_plane/` contains the FastAPI control plane; core code lives in `src/` with hexagonal layers (`domain`, `application`, `infrastructure`, `interfaces`) and tests in `tests/`. `runtime/executor/` contains the executor service, split into `application/`, `domain/`, and `infrastructure/`, with unit tests under `tests/`. `sandbox_web/` is the React + TypeScript frontend with source in `src/` and static assets in `public/`. Deployment assets live in `deploy/`, container image definitions in `images/`, shared docs in `docs/`, automation scripts in `scripts/`, and design specs in `specs/`.

## Build, Test, and Development Commands
Start the full local stack with `docker-compose -f deploy/docker-compose/docker-compose.yml up -d`. For the control plane, run `cd sandbox_control_plane && uv sync` to install dependencies, then `uv run pytest` for tests. Use `uv run black src tests`, `uv run ruff check src tests`, and `uv run mypy src` before submitting Python changes. For the executor, `cd runtime/executor && make test-unit` runs unit tests and `make test-cov` enforces coverage. For the frontend, use `cd sandbox_web && npm run dev` for local development, `npm run build` for production bundles, `npm run lint` for ESLint, and `npm run format` for Prettier.

## Coding Style & Naming Conventions
Python uses 4-space indentation, explicit type hints, and a 100-character line limit enforced by Black and Ruff. Keep Python modules and functions in `snake_case`. Respect architectural boundaries in `sandbox_control_plane/src/`; domain logic should not depend on interface or infrastructure details. In `sandbox_web/src/`, name React components in `PascalCase`, hooks as `useX.ts`, and group API clients by resource under `apis/`.

## Testing Guidelines
Python services use `pytest`. Place fast isolated tests in `tests/unit/`, service or dependency-backed tests in `tests/integration/`, and API contract coverage in the control plane’s contract suites where applicable. The executor coverage target is 90%, enforced by `make test-cov`. The frontend currently defines linting and formatting scripts, but no first-class `npm test` command; do not add undocumented test steps to contributor docs or PRs.

## Commit & Pull Request Guidelines
Follow the existing Conventional Commit style seen in history: `feat:`, `fix(scope):`, `docs(scope):`, and `chore(release):`. Keep commits focused and descriptive. Pull requests should summarize behavior changes, note affected services, link related issues when available, and include screenshots for `sandbox_web` UI changes. Call out any required environment or deployment follow-up in the PR description.
