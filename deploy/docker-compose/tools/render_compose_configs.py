#!/usr/bin/env python3
"""Render configs/kweaver/**/*.tmpl -> configs/generated/...

Each ``*.tmpl`` is rendered with placeholder substitution from ``.env``.
Output path: same relative path under ``configs/generated/`` with ``.tmpl`` stripped
(e.g. ``bkn-backend/cm-x/foo.yaml.tmpl`` -> ``generated/bkn-backend/cm-x/foo.yaml``).

Files named ``*.env.tmpl`` render to ``*.env`` (e.g. ``mf-model-api/cm-kw-yaml.env.tmpl``
-> ``generated/mf-model-api/cm-kw-yaml.env``) for use as Compose ``env_file``.
"""
from __future__ import annotations

import sys
from pathlib import Path

DEPLOY_COMPOSE_ROOT = Path(__file__).resolve().parents[1]
KWEAVER_TPL = DEPLOY_COMPOSE_ROOT / "configs" / "kweaver"
GEN_ROOT = DEPLOY_COMPOSE_ROOT / "configs" / "generated"
ENV_PATH = DEPLOY_COMPOSE_ROOT / ".env"


def load_env(path: Path) -> dict:
    out: dict = {}
    if not path.is_file():
        return out
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        k, v = line.split("=", 1)
        out[k.strip()] = v.strip()
    return out


def subst(text: str, mapping: dict) -> str:
    for key in sorted(mapping.keys(), key=len, reverse=True):
        val = mapping[key]
        if val is None:
            val = ""
        text = text.replace(key, str(val))
    return text


def clear_generated(gen: Path) -> None:
    if not gen.is_dir():
        gen.mkdir(parents=True, exist_ok=True)
        return
    for p in sorted(gen.rglob("*"), reverse=True):
        if p.is_file() and p.name != ".gitignore":
            p.unlink()
    for p in sorted(gen.rglob("*"), reverse=True):
        if p.is_dir():
            try:
                p.rmdir()
            except OSError:
                pass


def render_templates(mapping: dict) -> None:
    for tmpl in sorted(KWEAVER_TPL.rglob("*.tmpl")):
        rel = tmpl.relative_to(KWEAVER_TPL)
        out = GEN_ROOT / rel.parent / tmpl.stem
        out.parent.mkdir(parents=True, exist_ok=True)
        text = subst(tmpl.read_text(encoding="utf-8"), mapping)
        out.write_text(text, encoding="utf-8")


def main() -> int:
    env = load_env(ENV_PATH)
    reg = (env.get("IMAGE_REGISTRY") or "").rstrip("/")
    dip = (env.get("DIP_NAMESPACE") or "dip").strip().strip("/")

    if not reg:
        print("ERROR: IMAGE_REGISTRY empty in .env", file=sys.stderr)
        return 1
    if reg.endswith("/" + dip):
        print(
            "ERROR: IMAGE_REGISTRY must not include DIP_NAMESPACE; "
            f"use IMAGE_REGISTRY=.../kweaver-ai and DIP_NAMESPACE={dip}",
            file=sys.stderr,
        )
        return 1

    maria_host = env.get("MARIADB_HOST", "mariadb")
    maria_port = env.get("MARIADB_PORT", "3306")
    kafka_host = env.get("KAFKA_HOST", "kafka")
    kafka_port = env.get("KAFKA_PORT", "9092")
    os_host = env.get("OPENSEARCH_HOST", "opensearch")
    os_port = env.get("OPENSEARCH_PORT", "9200")
    redis_host = env.get("REDIS_HOST", "redis")

    mapping = {
        "__MARIA_HOST__": maria_host,
        "__MARIA_PORT__": maria_port,
        "__MARIADB_USER__": env.get("MARIADB_USER", ""),
        "__MARIADB_PASSWORD__": env.get("MARIADB_PASSWORD", ""),
        "__MARIADB_DATABASE__": env.get("MARIADB_DATABASE", ""),
        "__KAFKA_HOST__": kafka_host,
        "__KAFKA_PORT__": kafka_port,
        "__OPENSEARCH_HOST__": os_host,
        "__OPENSEARCH_PORT__": os_port,
        "__REDIS_SENTINEL_HOST__": redis_host,
        "__ACCESS_HOST__": env.get("ACCESS_HOST", "localhost"),
        "__ACCESS_PORT__": env.get("ACCESS_PORT", "8080"),
        "__IMAGE_REGISTRY__": reg,
        "__MINIO_ROOT_USER__": env.get("MINIO_ROOT_USER", ""),
        "__MINIO_ROOT_PASSWORD__": env.get("MINIO_ROOT_PASSWORD", ""),
        "__COMPOSE_PROJECT_NAME__": env.get("COMPOSE_PROJECT_NAME", "kweaver-compose"),
    }

    clear_generated(GEN_ROOT)
    render_templates(mapping)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
