#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
提交监听：根据变更文件列表 + path_scope_mapping.yaml 产出 scope、suggested_suites 等，
供智能体与执行器使用。可与 GitHub Webhook 或 GitHub Actions 配合使用。
用法:
  python scripts/commit_scope_mapping.py --changed-list changed_list.txt \
    --mapping testcase/vega/_config/path_scope_mapping.yaml \
    --repo owner/repo --branch main --sha abc123 --message "feat: xxx" --output listener_output.json
changed_list.txt 每行: 文件路径 [added|modified|removed]，空格分隔；若仅一路径则视为 modified。
"""
import argparse
import json
import os
import sys

try:
    import yaml
except ImportError:
    yaml = None


def _read_yaml(path, default=None):
    if default is None:
        default = {}
    if not os.path.isfile(path):
        return default
    with open(path, "r", encoding="utf-8") as f:
        return yaml.safe_load(f) or default


def path_matches(file_path, pattern):
    """匹配仓库相对路径与 path_pattern（支持 **）。统一用 /。"""
    file_path = file_path.replace("\\", "/")
    # pathlib.Path.match 在 3.12+ 对 ** 更严格，这里用 fnmatch 兼容：** 视为 *.*
    from fnmatch import fnmatch
    # path_patterns 如 vega/data-connection/**，filename 如 vega/data-connection/src/x.java
    if pattern.endswith("/**"):
        prefix = pattern[:-3]
        return file_path == prefix or file_path.startswith(prefix + "/")
    return fnmatch(file_path, pattern)


def main():
    ap = argparse.ArgumentParser(description="Map commit changed files to scope/suites for test pipeline")
    ap.add_argument("--changed-list", required=True, help="File with lines: path [added|modified|removed]")
    ap.add_argument("--mapping", required=True, help="path_scope_mapping.yaml path")
    ap.add_argument("--repo", default="", help="Repo full name, e.g. owner/repo")
    ap.add_argument("--branch", default="", help="Branch name, e.g. main")
    ap.add_argument("--sha", default="", help="Commit SHA")
    ap.add_argument("--message", default="", help="Commit message")
    ap.add_argument("--output", default="listener_output.json", help="Output JSON path")
    ap.add_argument("--path-to-api", default="", help="Optional path_to_api_mapping.yaml for affected_api_*")
    args = ap.parse_args()

    if not yaml:
        print("Need PyYAML: pip install pyyaml", file=sys.stderr)
        sys.exit(1)

    # 解析变更列表
    changed_files = []
    change_summary = {"added": [], "modified": [], "removed": []}
    with open(args.changed_list, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            parts = line.split(None, 1)
            path = parts[0].replace("\\", "/")
            status = (parts[1].lower() if len(parts) > 1 else "modified")
            if status not in change_summary:
                status = "modified"
            change_summary[status].append(path)
            changed_files.append({"path": path, "status": status})

    mapping = _read_yaml(args.mapping, {})
    subsystems = mapping.get("subsystems", [])

    scopes = []
    scope_tags = []
    suggested_suites = []
    for fp in changed_files:
        path = fp["path"]
        for sub in subsystems:
            for pat in sub.get("path_patterns", []):
                if path_matches(path, pat):
                    sid = sub.get("id")
                    if sid and sid not in scopes:
                        scopes.append(sid)
                    for t in sub.get("scope_tags", []):
                        if t not in scope_tags:
                            scope_tags.append(t)
                    for s in sub.get("suggested_suites", []):
                        if s not in suggested_suites:
                            suggested_suites.append(s)
                    break

    need_add_cases = bool(change_summary["added"]) or False

    # 可选：path_to_api 映射得到 affected_api_names / affected_api_paths
    affected_api_names = []
    affected_api_paths = []
    if args.path_to_api and os.path.isfile(args.path_to_api):
        path_to_api = _read_yaml(args.path_to_api, {}).get("path_to_api", [])
        for fp in changed_files:
            path = fp["path"]
            for item in path_to_api:
                pat = item.get("path_pattern", "")
                if path_matches(path, pat):
                    if item.get("api_name") and item["api_name"] not in affected_api_names:
                        affected_api_names.append(item["api_name"])
                    if item.get("api_path") and item["api_path"] not in affected_api_paths:
                        affected_api_paths.append(item["api_path"])

    out = {
        "repo": args.repo,
        "branch": (args.branch or "").strip(),
        "commit_sha": args.sha,
        "commit_message": (args.message or "").strip(),
        "commit_url": "https://github.com/%s/commit/%s" % (args.repo, args.sha) if args.repo and args.sha else "",
        "changed_files": changed_files,
        "change_summary": change_summary,
        "scopes": scopes,
        "scope_tags": scope_tags,
        "suggested_suites": suggested_suites,
        "affected_api_names": affected_api_names,
        "affected_api_paths": affected_api_paths,
        "need_add_cases": need_add_cases,
        "diffs": [],
    }
    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(out, f, ensure_ascii=False, indent=2)
    print("Wrote %s" % args.output)


if __name__ == "__main__":
    main()
