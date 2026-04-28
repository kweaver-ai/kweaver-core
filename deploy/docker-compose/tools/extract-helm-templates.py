#!/usr/bin/env python3
"""
One-shot / dev helper: extract ConfigMap and Secret stringData from local Helm charts
and write deploy/docker-compose/configs/kweaver/<service>/*.tmpl

Requires: charts unpacked under CHART_ROOT as <chartName>-<version>/

    /tmp/kc-helm/*.tgz + tar extraction to CHART_ROOT

Run from repo root:
  python3 deploy/docker-compose/tools/extract-helm-templates.py
"""
from __future__ import annotations

import base64
import re
import subprocess
import sys
from pathlib import Path

# deploy/docker-compose/ (this file lives in tools/)
DEPLOY_COMPOSE_ROOT = Path(__file__).resolve().parents[1]
OUT_BASE = DEPLOY_COMPOSE_ROOT / "configs" / "kweaver"
CHART_ROOT = Path("/tmp/kc-charts-unpacked")

# Single-key Secret files whose chart-default value is a placeholder ("xxxxxx",
# "root", "anyrobot", ...). Map the key name to a Compose placeholder so re-extracts
# don't leak chart fixtures into configs/kweaver/**/secret-*/<KEY>.tmpl.
SECRET_KEY_TO_PLACEHOLDER: dict[str, str] = {
    "DB_PASSWORD": "__MARIADB_PASSWORD__",
    "RDSPASS": "__MARIADB_PASSWORD__",
    "REDIS_PASSWORD": "__REDIS_PASSWORD__",
    "REDIS_SENTINEL_PASSWORD": "__REDIS_PASSWORD__",
    "REDIS_USERNAME": "__REDIS_USER__",
    "REDIS_SENTINEL_USERNAME": "__REDIS_USER__",
    "MQ_AUTH_USERNAME": "__KAFKA_USER__",
    "MQ_AUTH_PASSWORD": "__KAFKA_PASSWORD__",
    "MQ_AUTH_MECHANISM": "__MQ_AUTH_MECHANISM__",
    "MONGODB_PASSWORD": "__MONGODB_PASSWORD__",
}

# chart folder name (without path) -> logical output subdir
# Matches deploy/release-manifests/0.7.0/kweaver-core.yaml (kweaver-core --minimum installs
# the same release set; --minimum only sets auth.enabled=false businessDomain.enabled=false).
CHART_MAP = [
    ("kweaver-core-data-migrator-0.7.0", "kweaver-core-data-migrator"),
    ("mf-model-manager-0.7.0", "mf-model-manager"),
    ("mf-model-api-0.7.0", "mf-model-api"),
    ("bkn-backend-0.7.0", "bkn-backend"),
    ("ontology-query-0.7.0", "ontology-query"),
    ("vega-backend-0.7.0", "vega-backend"),
    ("data-connection-0.6.0", "data-connection"),
    ("vega-gateway-0.6.0", "vega-gateway"),
    ("vega-gateway-pro-0.6.0", "vega-gateway-pro"),
    ("mdl-data-model-0.6.0", "mdl-data-model"),
    ("mdl-uniquery-0.6.0", "mdl-uniquery"),
    ("mdl-data-model-job-0.6.0", "mdl-data-model-job"),
    ("agent-operator-integration-0.7.0", "agent-operator-integration"),
    ("agent-retrieval-0.7.0", "agent-retrieval"),
    ("agent-backend-0.7.0", "agent-backend"),
    ("dataflow-0.7.0", "dataflow"),
    ("coderunner-0.7.0", "coderunner"),
    ("doc-convert-0.7.0", "doc-convert"),
    ("sandbox-0.3.3", "sandbox"),
    ("oss-gateway-backend-0.7.0", "oss-gateway-backend"),
    ("otelcol-contrib-0.2.3", "otelcol-contrib"),
    ("agent-observability-0.2.3", "agent-observability"),
]

ENV_KEY_RE = re.compile(r"^[A-Z][A-Z0-9_]*$")


def patch_compose_placeholders(blob: str) -> str:
    """Normalize k8s DNS names and inject __KEY__ placeholders for .env substitution."""
    s = blob

    repl = [
        (r"kafka-headless\.resource\.svc\.cluster\.local\.", "__KAFKA_HOST__"),
        (r"kafka-headless\.resource\.svc\.cluster\.local", "__KAFKA_HOST__"),
        (r"opensearch-master\.resource\.svc\.cluster\.local\.", "__OPENSEARCH_HOST__"),
        (r"opensearch-cluster-master\.resource\.svc\.cluster\.local\.", "__OPENSEARCH_HOST__"),
        (r"mariadb-mariadb-master\.resource\.svc\.cluster\.local\.", "__MARIA_HOST__"),
        (r"mariadb-mariadb-cluster\.resource\.svc\.cluster\.local\.", "__MARIA_HOST__"),
        (r"proton-redis-proton-redis-sentinel\.resource\.svc\.cluster\.local", "__REDIS_SENTINEL_HOST__"),
        (r"mdl-data-model-svc", "mdl-data-model"),
        (r"mdl-uniquery-svc", "mdl-uniquery"),
        (r"mdl-data-model-job-svc", "mdl-data-model-job"),
        (r"ontology-query-svc", "ontology-query"),
        (r"vega-backend-svc", "vega-backend"),
        (r"bkn-backend-svc", "bkn-backend"),
        (r"vega-gateway-pro-svc", "vega-gateway-pro"),
        # Optional pipeline not in compose — route via nginx (may 502)
        (r"flow-stream-data-pipeline-svc", "nginx"),
    ]
    for pat, to in repl:
        s = re.sub(pat, to, s)

    for h in (
        "authorization-private",
        "user-management-private",
        "hydra-admin",
        "business-system-service",
        "kafka-connect-svc",
    ):
        s = re.sub(
            rf"^(\s*host:\s*){re.escape(h)}\s*$",
            r"\1nginx",
            s,
            flags=re.MULTILINE,
        )

    s = s.replace("feed-ingester-service", "nginx")

    s = re.sub(r"(\n\s+port:\s*)3330", r"\g<1>__MARIA_PORT__", s)
    s = re.sub(r"(\n\s+portRead:\s*)3330", r"\g<1>__MARIA_PORT__", s)
    s = re.sub(r"(\n\s+mqPort:\s*)9097", r"\g<1>__KAFKA_PORT__", s)

    s = s.replace("protocol: sasl_plaintext", "protocol: plaintext")
    s = s.replace("mechanism: PLAIN", 'mechanism: ""')
    s = s.replace("username: admin", 'username: ""')

    s = re.sub(
        r"(\n\s+rds:\n(?:.*\n)*?\s+user:\s*)root",
        r"\g<1>__MARIADB_USER__",
        s,
    )
    s = re.sub(
        r"(\n\s+rds:\n(?:.*\n)*?\s+password:\s*)\"\"",
        r'\g<1>"__MARIADB_PASSWORD__"',
        s,
        count=1,
    )

    s = re.sub(
        r"(\n\s+opensearch:\n(?:.*\n)*?\s+user:\s*)admin",
        r"\g<1>''",
        s,
    )
    s = re.sub(
        r"(\n\s+password:\s*)eisoo\.com123",
        r'\g<1>""',
        s,
    )

    return s


def _secret_value(key: str, raw: str) -> str:
    """Single-key Secret values: prefer placeholder when key matches a known one and
    chart shipped a fixture (xxxxxx / root / anyrobot / minioadmin), else keep raw."""
    text = patch_compose_placeholders(raw)
    placeholder = SECRET_KEY_TO_PLACEHOLDER.get(key)
    if placeholder is None:
        return text
    stripped = text.strip()
    if stripped in {"xxxxxx", "root", "anyrobot", "minioadmin", "PLAIN", ""}:
        return placeholder
    return text


def helm_template(chart_dir: Path) -> str:
    p = subprocess.run(
        ["helm", "template", "kw", str(chart_dir)],
        capture_output=True,
        text=True,
    )
    if p.returncode != 0:
        print(p.stderr, file=sys.stderr)
        raise RuntimeError(f"helm template failed: {chart_dir}")
    return p.stdout


def _envfile_line(key: str, val: str) -> str:
    v = val.replace("\r", "").strip()
    if "\n" in v:
        raise ValueError(f"multiline env value for {key}")
    if re.search(r"[\n\0]", v):
        raise ValueError(f"bad env value for {key}")
    return f"{key}={v}"


def write_configmap_data(out_sub: str, cm_name: str, data: dict) -> None:
    """Write ConfigMap keys: env-style keys -> ``<service>/cm-<safe>.env.tmpl`` (Compose env_file)."""
    safe_cm = re.sub(r"[^a-zA-Z0-9_.-]+", "_", cm_name)
    cm_prefix = f"cm-{safe_cm}"
    out_dir = OUT_BASE / out_sub / cm_prefix
    out_dir.mkdir(parents=True, exist_ok=True)

    env_pairs: list[tuple[str, str]] = []
    file_entries: list[tuple[str, str]] = []

    for fname, content in (data or {}).items():
        if content is None:
            continue
        text = patch_compose_placeholders(str(content))
        stem = str(fname)
        if "." not in stem and ENV_KEY_RE.match(stem):
            if "\n" in text.replace("\r", "").strip() or "\n" in text:
                file_entries.append((fname, text))
            else:
                env_pairs.append((stem, text))
        else:
            file_entries.append((fname, text))

    if env_pairs:
        body = "\n".join(_envfile_line(k, v) for k, v in sorted(env_pairs)) + "\n"
        (OUT_BASE / out_sub / f"{cm_prefix}.env.tmpl").write_text(
            body, encoding="utf-8"
        )

    for fname, text in file_entries:
        out_name = f"{fname}.tmpl" if not str(fname).endswith(".tmpl") else str(fname)
        (out_dir / out_name).write_text(text, encoding="utf-8")


def main() -> int:
    try:
        import yaml  # type: ignore
    except ImportError:
        print("ERROR: pip install pyyaml", file=sys.stderr)
        return 1

    if not CHART_ROOT.is_dir():
        print(f"ERROR: CHART_ROOT missing: {CHART_ROOT}", file=sys.stderr)
        return 1

    OUT_BASE.mkdir(parents=True, exist_ok=True)

    for folder, out_sub in CHART_MAP:
        cdir = CHART_ROOT / folder
        if not cdir.is_dir():
            print(f"skip (missing chart): {cdir}", file=sys.stderr)
            continue
        raw = helm_template(cdir)
        out_dir = OUT_BASE / out_sub
        out_dir.mkdir(parents=True, exist_ok=True)

        for doc in yaml.safe_load_all(raw):
            if not doc or not isinstance(doc, dict):
                continue
            kind = doc.get("kind")
            meta = doc.get("metadata", {}) or {}
            if kind == "ConfigMap":
                cm_name = meta.get("name", "configmap")
                write_configmap_data(out_sub, cm_name, doc.get("data") or {})
            elif kind == "Secret":
                sec_name = meta.get("name", "secret")
                safe_sec = re.sub(r"[^a-zA-Z0-9_.-]+", "_", sec_name)
                sec_dir = out_dir / f"secret-{safe_sec}"
                sec_dir.mkdir(parents=True, exist_ok=True)
                sdata = doc.get("stringData") or {}
                seen = set()
                for fname, content in sdata.items():
                    seen.add(fname)
                    text = _secret_value(fname, str(content))
                    out_name = f"{fname}.tmpl"
                    (sec_dir / out_name).write_text(text, encoding="utf-8")
                for fname, content in (doc.get("data") or {}).items():
                    if fname in seen:
                        continue
                    if content is None:
                        continue
                    raw_b = str(content)
                    try:
                        decoded = base64.b64decode(raw_b, validate=True).decode("utf-8")
                        raw_use = decoded
                    except Exception:
                        raw_use = raw_b
                    text = _secret_value(fname, raw_use)
                    out_name = f"{fname}.tmpl"
                    (sec_dir / out_name).write_text(text, encoding="utf-8")

        print("OK", out_sub)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
