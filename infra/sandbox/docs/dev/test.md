# 测试

## Control Plane

```bash
cd sandbox_control_plane
uv run pytest
uv run black src tests
uv run ruff check src tests
uv run mypy src
```

## Executor

```bash
cd runtime/executor
make test-unit
make test-cov
```

## Frontend

```bash
cd sandbox_web
npm run lint
npm run format
```
