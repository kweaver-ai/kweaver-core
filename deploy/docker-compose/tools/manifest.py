#!/usr/bin/env python3
"""Single source of truth for Compose ↔ chart ↔ image bindings.

Reads ``deploy/docker-compose/compose-manifest.yaml`` and offers helpers used by
the other scripts (``compose.sh``, ``setup.sh``, ``extract-helm-templates.py``).

CLI:
    python3 tools/manifest.py services [--phase=app|infra|all]
    python3 tools/manifest.py charts          # one (folder, out_sub) per line
    python3 tools/manifest.py env-defaults    # KEY=VAL lines for tagEnv defaults
    python3 tools/manifest.py check-compose   # diff manifest vs docker-compose.yml
    python3 tools/manifest.py check-env [path] # diff manifest tagEnv vs an .env file
"""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

DEPLOY_COMPOSE_ROOT = Path(__file__).resolve().parents[1]
MANIFEST_PATH = DEPLOY_COMPOSE_ROOT / "compose-manifest.yaml"
COMPOSE_FILE = DEPLOY_COMPOSE_ROOT / "docker-compose.yml"


def _require_yaml():
    try:
        import yaml  # type: ignore

        return yaml
    except ImportError:
        print("ERROR: pip install pyyaml", file=sys.stderr)
        sys.exit(1)


def load_manifest(path: Path = MANIFEST_PATH) -> dict:
    yaml = _require_yaml()
    if not path.is_file():
        print(f"ERROR: manifest missing: {path}", file=sys.stderr)
        sys.exit(1)
    return yaml.safe_load(path.read_text(encoding="utf-8")) or {}


def services(manifest: dict, phase: str = "all") -> list[str]:
    """Return ordered service names. ``phase`` ∈ {app, infra, all}.

    Entries marked as non-Compose (``note`` mentions "rather than as a Compose
    service image") are skipped automatically.
    """
    out: list[str] = []
    for name, meta in (manifest.get("images") or {}).items():
        if not isinstance(meta, dict):
            continue
        note = (meta.get("note") or "").lower()
        if "rather than as a compose service" in note:
            continue
        p = meta.get("phase")
        if phase != "all" and p != phase:
            continue
        out.append(name)
    return out


def chart_map(manifest: dict) -> list[tuple[str, str]]:
    """Return ``[(folder, out_sub), ...]`` for ``extract-helm-templates.py``.

    Deduplicated by chart so multi-image charts (dataflow / coderunner / sandbox
    / doc-convert) only show up once. ``folder`` is ``<chart>-<chartVersion>``;
    ``out_sub`` is the chart name (matches existing ``configs/kweaver/<chart>/``).
    """
    seen: set[str] = set()
    out: list[tuple[str, str]] = []
    for meta in (manifest.get("images") or {}).values():
        if not isinstance(meta, dict):
            continue
        chart = meta.get("chart")
        ver = meta.get("chartVersion")
        if not chart or not ver:
            continue
        if chart in seen:
            continue
        seen.add(chart)
        out.append((f"{chart}-{ver}", chart))
    return out


def env_defaults(manifest: dict) -> dict[str, str]:
    """Return ``{tagEnv: tag}`` collected from manifest entries."""
    out: dict[str, str] = {}
    for meta in (manifest.get("images") or {}).values():
        if not isinstance(meta, dict):
            continue
        key = meta.get("tagEnv")
        tag = meta.get("tag")
        if not key or tag is None:
            continue
        out.setdefault(str(key), str(tag))
    return out


_COMPOSE_SERVICE_RE = re.compile(r"^  ([a-z0-9][a-z0-9._-]*):\s*$")


def compose_services(path: Path = COMPOSE_FILE) -> list[str]:
    """Parse top-level ``services:`` keys without requiring full YAML traversal."""
    out: list[str] = []
    if not path.is_file():
        return out
    in_services = False
    for raw in path.read_text(encoding="utf-8").splitlines():
        if raw.startswith("services:"):
            in_services = True
            continue
        if in_services and raw and not raw.startswith((" ", "\t", "#")):
            break
        if not in_services:
            continue
        m = _COMPOSE_SERVICE_RE.match(raw)
        if m:
            out.append(m.group(1))
    return out


def check_compose(manifest: dict) -> int:
    expected = set(services(manifest, "all"))
    actual = set(compose_services())
    missing = sorted(expected - actual)
    extra = sorted(actual - expected)
    rc = 0
    if missing:
        print("manifest services missing in docker-compose.yml:", file=sys.stderr)
        for s in missing:
            print(f"  - {s}", file=sys.stderr)
        rc = 1
    if extra:
        print("docker-compose.yml services not in compose-manifest.yaml:", file=sys.stderr)
        for s in extra:
            print(f"  - {s}", file=sys.stderr)
        rc = 1
    if rc == 0:
        print(f"OK: {len(actual)} services in sync with manifest")
    return rc


def _read_env_file(path: Path) -> dict[str, str]:
    out: dict[str, str] = {}
    if not path.is_file():
        return out
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        k, v = line.split("=", 1)
        out[k.strip()] = v.strip()
    return out


def check_env(manifest: dict, env_path: Path) -> int:
    expected = env_defaults(manifest)
    if not env_path.is_file():
        print(f"ERROR: env file missing: {env_path}", file=sys.stderr)
        return 1
    actual = _read_env_file(env_path)
    missing: list[str] = []
    drifted: list[tuple[str, str, str]] = []
    for key, exp in expected.items():
        cur = actual.get(key, "")
        if not cur:
            missing.append(key)
        elif cur != exp:
            drifted.append((key, cur, exp))
    rc = 0
    if missing:
        print(f"{env_path}: missing tagEnv values from manifest:", file=sys.stderr)
        for k in missing:
            print(f"  - {k}={expected[k]}", file=sys.stderr)
        rc = 1
    if drifted:
        print(f"{env_path}: tagEnv values drifted from manifest:", file=sys.stderr)
        for k, cur, exp in drifted:
            print(f"  - {k}: env={cur}  manifest={exp}", file=sys.stderr)
        rc = 1
    if rc == 0:
        print(f"OK: {len(expected)} tagEnv values in {env_path.name} match manifest")
    return rc


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser(prog="manifest.py")
    sub = ap.add_subparsers(dest="cmd", required=True)
    p_svc = sub.add_parser("services")
    p_svc.add_argument("--phase", default="all", choices=("all", "app", "infra"))
    sub.add_parser("charts")
    sub.add_parser("env-defaults")
    sub.add_parser("check-compose")
    p_env = sub.add_parser("check-env")
    p_env.add_argument("path", nargs="?", default=str(DEPLOY_COMPOSE_ROOT / ".env"))
    args = ap.parse_args(argv)

    m = load_manifest()
    if args.cmd == "services":
        for s in services(m, args.phase):
            print(s)
        return 0
    if args.cmd == "charts":
        for folder, out_sub in chart_map(m):
            print(f"{folder}\t{out_sub}")
        return 0
    if args.cmd == "env-defaults":
        for k, v in sorted(env_defaults(m).items()):
            print(f"{k}={v}")
        return 0
    if args.cmd == "check-compose":
        return check_compose(m)
    if args.cmd == "check-env":
        return check_env(m, Path(args.path))
    return 2


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
