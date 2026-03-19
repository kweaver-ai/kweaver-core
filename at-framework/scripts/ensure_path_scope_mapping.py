#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
根据模板与仓库目录结构，为**所有模块**生成或更新 path_scope_mapping.yaml（无人工介入）。
- 解析不写死 vega：按仓库顶层目录及二级目录通用扫描，生成 subsystems。
- 若目标不存在：从模板 + 仓库扫描生成；若已存在：在已有基础上增补缺失的 scope，不删除已有条目。
- 可选智能体补充：通过 --supplement 传入智能体产出的 YAML，合并 suggested_suites、path_to_api 等。
"""
from __future__ import annotations

import argparse
from pathlib import Path


def load_yaml(path: Path) -> dict | None:
    if not path.exists():
        return None
    try:
        import yaml
        with open(path, "r", encoding="utf-8") as f:
            return yaml.safe_load(f)
    except Exception:
        return None


def save_yaml(path: Path, data: dict) -> None:
    try:
        import yaml
        path.parent.mkdir(parents=True, exist_ok=True)
        with open(path, "w", encoding="utf-8") as f:
            yaml.safe_dump(data, f, allow_unicode=True, default_flow_style=False, sort_keys=False)
    except Exception as e:
        raise SystemExit(f"Failed to write {path}: {e}") from e


def scan_repo_subsystems(repo_root: Path) -> list[dict]:
    """根据仓库顶层目录通用扫描：每个顶层目录若有子目录则展开为 顶层-子目录 scope，否则为单 scope。不写死模块名。"""
    subsystems = []
    repo_root = Path(repo_root)
    if not repo_root.is_dir():
        return subsystems

    seen = set()
    for top in sorted(repo_root.iterdir()):
        if not top.is_dir() or top.name.startswith("."):
            continue
        name = top.name
        subdirs = [d for d in sorted(top.iterdir()) if d.is_dir() and not d.name.startswith(".")]
        if subdirs:
            for second in subdirs:
                scope_id = f"{name}-{second.name}".replace("_", "-")
                if scope_id in seen:
                    continue
                seen.add(scope_id)
                subsystems.append({
                    "id": scope_id,
                    "name": f"{name.upper()}-{second.name.replace('-', ' ').title()}",
                    "path_patterns": [f"{name}/{second.name}/**"],
                    "scope_tags": ["regression"],
                    "suggested_suites": [],
                })
        else:
            scope_id = name.replace("_", "-")
            if scope_id in seen:
                continue
            seen.add(scope_id)
            subsystems.append({
                "id": scope_id,
                "name": name.replace("-", " ").title(),
                "path_patterns": [f"{name}/**"],
                "scope_tags": ["regression"],
                "suggested_suites": [],
            })
    return subsystems


def merge_subsystems(template_subs: list[dict], scanned_subs: list[dict], existing_subs: list[dict] | None) -> list[dict]:
    """合并：模板中的虚拟 scope + 已有条目（保留）+ 扫描出的新 scope（仅追加尚未存在的 id）。"""
    by_id = {}
    for s in template_subs:
        by_id[s.get("id")] = dict(s)
    if existing_subs:
        for s in existing_subs:
            by_id[s.get("id")] = dict(s)
    for s in scanned_subs:
        sid = s.get("id")
        if sid and sid not in by_id:
            by_id[sid] = dict(s)
    order = []
    seen = set()
    for s in template_subs:
        sid = s.get("id")
        if sid and sid not in seen:
            order.append(sid)
            seen.add(sid)
    for s in (existing_subs or []):
        sid = s.get("id")
        if sid and sid not in seen:
            order.append(sid)
            seen.add(sid)
    for s in scanned_subs:
        sid = s.get("id")
        if sid and sid not in seen:
            order.append(sid)
            seen.add(sid)
    return [by_id[k] for k in order if k in by_id]


def apply_supplement(merged_data: dict, supplement: dict) -> dict:
    """智能体补充：用 supplement 中的 suggested_suites、path_to_api 等覆盖或追加。"""
    out = dict(merged_data)
    subs = out.get("subsystems") or []
    supp_subs = supplement.get("subsystems") or []
    by_id = {s.get("id"): dict(s) for s in subs}
    for s in supp_subs:
        sid = s.get("id")
        if not sid:
            continue
        if sid in by_id:
            for k, v in s.items():
                if v is not None and k != "id":
                    by_id[sid][k] = v
        else:
            by_id[sid] = dict(s)
    out["subsystems"] = list(by_id.values())
    if supplement.get("path_to_api") is not None:
        existing = out.get("path_to_api") or []
        out["path_to_api"] = existing + list(supplement["path_to_api"])
    return out


def discover_module_dirs(testcase_root: Path) -> list[Path]:
    """返回 testcase 下各模块目录（含 _config 或可写入的子目录，排除 _config、spec 等）。"""
    testcase_root = Path(testcase_root)
    if not testcase_root.is_dir():
        return []
    modules = []
    for d in sorted(testcase_root.iterdir()):
        if not d.is_dir() or d.name.startswith("."):
            continue
        if d.name in ("_config", "spec"):
            continue
        modules.append(d)
    return modules


def main() -> None:
    parser = argparse.ArgumentParser(
        description="从模板与仓库扫描为所有模块生成/更新 path_scope_mapping.yaml，可选智能体补充"
    )
    parser.add_argument("--repo-root", type=Path, required=True, help="被测仓库根目录（用于扫描目录结构）")
    parser.add_argument("--template", type=Path, required=True, help="path_scope_mapping.template.yaml 路径（testcase/_config/spec/ 下）")
    parser.add_argument("--output", type=Path, default=None, help="单文件输出路径（与 --testcase-root 二选一）")
    parser.add_argument("--testcase-root", type=Path, default=None,
                        help="testcase 根目录；指定则为所有模块分别写入 testcase/<模块>/_config/path_scope_mapping.yaml")
    parser.add_argument("--supplement", type=Path, default=None,
                        help="智能体补充 YAML：含 subsystems 的 suggested_suites 等、path_to_api，将合并入最终结果")
    args = parser.parse_args()

    if args.output is None and args.testcase_root is None:
        raise SystemExit("请指定 --output 或 --testcase-root 之一")
    if args.output is not None and args.testcase_root is not None:
        raise SystemExit("--output 与 --testcase-root 不可同时指定")

    template_data = load_yaml(args.template)
    if not template_data:
        raise SystemExit(f"Template not found or invalid: {args.template}")

    supplement_data = load_yaml(args.supplement) if args.supplement else None

    if args.output is not None:
        output_paths = [args.output]
        existing_data = load_yaml(args.output) if args.output.exists() else None
    else:
        modules = discover_module_dirs(args.testcase_root)
        output_paths = [m / "_config" / "path_scope_mapping.yaml" for m in modules]
        existing_data = None
        for p in output_paths:
            if p.exists():
                existing_data = load_yaml(p)
                break

    template_subs = template_data.get("subsystems") or []
    existing_subs = (existing_data.get("subsystems") or []) if existing_data else None
    scanned_subs = scan_repo_subsystems(args.repo_root)

    merged_subs = merge_subsystems(template_subs, scanned_subs, existing_subs)
    out = {
        "subsystems": merged_subs,
        "path_to_api": (existing_data or template_data).get("path_to_api") or [],
        "change_actions": (existing_data or template_data).get("change_actions") or {},
        "test_type_tags": (existing_data or template_data).get("test_type_tags") or {},
    }
    if supplement_data:
        out = apply_supplement(out, supplement_data)

    for out_path in output_paths:
        save_yaml(out_path, out)
        print(f"Wrote {out_path} ({len(out['subsystems'])} subsystems)")
    if supplement_data:
        print("Applied agent supplement.")


if __name__ == "__main__":
    main()
