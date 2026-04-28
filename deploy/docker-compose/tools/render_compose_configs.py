#!/usr/bin/env python3
"""Render configs/kweaver/**/*.tmpl -> configs/generated/...

Merges some chart envFrom patterns into single ``*.env`` files used by Compose ``env_file``.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

DEPLOY_COMPOSE_ROOT = Path(__file__).resolve().parents[1]
KWEAVER_TPL = DEPLOY_COMPOSE_ROOT / "configs" / "kweaver"
GEN_ROOT = DEPLOY_COMPOSE_ROOT / "configs" / "generated"
ENV_PATH = DEPLOY_COMPOSE_ROOT / ".env"

ENV_KEY_RE = re.compile(r"^[A-Z][A-Z0-9_]*$")


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


def parse_env_file_body(path: Path) -> dict:
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


def merge_key_files_from_dir(d: Path, merged: dict) -> None:
    if not d.is_dir():
        return
    for f in sorted(d.iterdir()):
        if not f.is_file():
            continue
        if f.suffix in (".yaml", ".yml", ".pem", ".conf") or f.name == "s3fs-passwd":
            continue
        stem = f.stem
        if not ENV_KEY_RE.match(stem):
            continue
        val = f.read_text(encoding="utf-8").strip()
        if "\n" in val:
            continue
        merged[stem] = val.replace("\n", "\\n")


def write_merged_env(rel_path: str, merged: dict) -> None:
    out = GEN_ROOT / rel_path
    out.parent.mkdir(parents=True, exist_ok=True)
    body = "\n".join(f"{k}={merged[k]}" for k in sorted(merged))
    out.write_text(body + ("\n" if body else ""), encoding="utf-8")


def merge_dataflow_env() -> None:
    merged: dict = {}
    base = GEN_ROOT / "dataflow" / "cm-flow-automation.env"
    merged.update(parse_env_file_body(base))
    merge_key_files_from_dir(GEN_ROOT / "dataflow" / "secret-flow-automation", merged)
    write_merged_env("dataflow/flow-automation.env", merged)


def merge_sandbox_env() -> None:
    merged: dict = {}
    merged.update(parse_env_file_body(GEN_ROOT / "sandbox" / "cm-sandbox-config.env"))
    merge_key_files_from_dir(GEN_ROOT / "sandbox" / "secret-sandbox-secrets", merged)
    if (GEN_ROOT / "sandbox" / "secret-s3fs-passwd" / "s3fs-passwd").is_file():
        v = (GEN_ROOT / "sandbox" / "secret-s3fs-passwd" / "s3fs-passwd").read_text(
            encoding="utf-8"
        ).strip()
        if v and "\n" not in v:
            merged.setdefault("S3FS_PASSWD", v)
    write_merged_env("sandbox/sandbox.env", merged)


def merge_coderunner_env() -> None:
    cr: dict = {}
    cr.update(parse_env_file_body(GEN_ROOT / "coderunner" / "cm-kw-coderunner.env"))
    merge_key_files_from_dir(GEN_ROOT / "coderunner" / "cm-kw-coderunner", cr)
    write_merged_env("coderunner/coderunner.env", cr)

    dt: dict = {}
    dt.update(parse_env_file_body(GEN_ROOT / "coderunner" / "cm-dataflowtools.env"))
    merge_key_files_from_dir(GEN_ROOT / "coderunner" / "cm-dataflowtools", dt)
    merge_key_files_from_dir(GEN_ROOT / "coderunner" / "secret-dataflowtools", dt)
    write_merged_env("coderunner/dataflowtools.env", dt)


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
    sandbox_ver = env.get("SANDBOX_VERSION", "0.3.3")
    sandbox_tpl_img = f"{reg}/{dip}/sandbox-template-python-basic:{sandbox_ver}"

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
        "__SANDBOX_DATABASE_URL__": env.get("SANDBOX_DATABASE_URL", ""),
        "__SANDBOX_TEMPLATE_IMAGE__": sandbox_tpl_img,
        "__SANDBOX_S3_BUCKET__": env.get("SANDBOX_S3_BUCKET", "sandbox-workspace"),
        "__REDIS_PASSWORD__": env.get("REDIS_PASSWORD", ""),
        "__REDIS_USER__": env.get("REDIS_USER", "default"),
        "__KAFKA_USER__": env.get("KAFKA_USER", ""),
        "__KAFKA_PASSWORD__": env.get("KAFKA_PASSWORD", ""),
        "__MQ_AUTH_MECHANISM__": env.get("MQ_AUTH_MECHANISM", ""),
        "__MONGODB_PASSWORD__": env.get("MONGODB_PASSWORD", ""),
    }

    clear_generated(GEN_ROOT)
    render_templates(mapping)

    merge_dataflow_env()
    merge_sandbox_env()
    merge_coderunner_env()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
