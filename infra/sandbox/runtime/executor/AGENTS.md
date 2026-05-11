# Repository Guidelines

## Project Structure & Module Organization
`runtime/executor/` contains the code execution daemon. The service is split into `application/` for use-case orchestration, `domain/` for core models and ports, `infrastructure/` for isolation, HTTP, logging, and persistence adapters, and `interfaces/` for external entry points. Tests live under `tests/`, while longer-form developer docs are in `docs/`. Container image build assets live under the repository-level `images/` directory.

## Build, Test, and Development Commands
Use `make build` as a convenience wrapper for the repository-level `images/build.sh` template build. Run `make test-unit` for the unit suite and `make test-cov` for coverage-checked tests. Use `uv run pytest tests/unit/ -v --tb=short` directly when you need a narrower loop. `make test` performs a basic import check in the local Python environment.

## Coding Style & Naming Conventions
Target Python 3.11 with 4-space indentation and explicit typing for new code. Black enforces a 100-character line length; Mypy is enabled, and existing config still includes Flake8 in the dev dependencies. Name modules in `snake_case` and keep service or port names aligned with their responsibilities, for example `heartbeat_service.py` or `callback_port.py`. Keep platform-specific isolation details contained in `infrastructure/isolation/`.

## Testing Guidelines
Use `pytest` for all tests and mark long-running cases with `@pytest.mark.slow`. Keep fast unit tests under `tests/unit/`; only add broader scenarios when the behavior cannot be validated in isolation. The configured coverage threshold is 90%, enforced by `make test-cov`, so new logic should include coverage for success paths and important failures.

## Commit & Pull Request Guidelines
Use Conventional Commits consistent with repo history, including `feat:`, `fix(scope):`, and `docs(scope):`. Keep commits narrowly scoped to one executor behavior or subsystem. Pull requests should explain runtime impact, mention isolation or callback changes explicitly, and include any verification steps needed to reproduce the change locally.
