#!/usr/bin/env python3
"""Generate BKN v2.0-ish tree from 06-world-cup CSV headers."""

from __future__ import annotations

import csv
import hashlib
import shutil
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parent.parent / "data"
OUT_DIR = Path(__file__).resolve().parent.parent / "worldcup-bkn"


# Column suffix -> referenced object_type id (table stem)
FK_MAP: dict[str, str] = {
    "tournament_id": "tournaments",
    "team_id": "teams",
    "home_team_id": "teams",
    "away_team_id": "teams",
    "match_id": "matches",
    "player_id": "players",
    "manager_id": "managers",
    "referee_id": "referees",
    "stadium_id": "stadiums",
    "confederation_id": "confederations",
    "award_id": "awards",
    "opponent_id": "teams",  # joins teams.team_id
    "player_team_id": "teams",
}

# Prefer natural PK columns (not surrogate key_id) when present
PK_PREFERENCE = ("tournament_id", "match_id", "team_id", "player_id", "manager_id", "referee_id",
                 "stadium_id", "confederation_id", "award_id", "goal_id", "booking_id",
                 "substitution_id", "penalty_kick_id")


def sniff_type(header: str) -> str:
    h = header.lower()
    if h == "key_id" or h.endswith("_id"):
        return "integer"
    if "_date" in h:
        return "date"
    if "_time" in h:
        return "time"
    if h in {"female", "mens_team", "womens_team", "home_team", "away_team", "group_stage",
             "knockout_stage", "replayed", "replay", "extra_time", "penalty_shootout",
             "own_goal", "penalty", "starter", "substitute", "yellow_card", "red_card",
             "second_yellow_card", "sending_off", "converted", "going_off", "coming_on",
             "home_team_win", "away_team_win", "draw", "win", "lose", "shared", "advanced"}:
        return "boolean"
    if any(x in h for x in ("_score", "margin", "_goals_", "played", "wins", "draws",
                            "losses", "points", "position", "shirt_number", "stage_number",
                            "count_", "minute_regulation", "minute_stoppage", "capacity",
                            "goal_difference")):
        return "integer"
    if "difference" in h or "margin" in h:
        return "integer"
    if "link" in h or "_name" == h[-5:] == False:
        pass
    if "_link" in h or "wikipedia" in h:
        return "string"
    if any(x in h for x in ("description", "_text", "list_tournaments")):
        return "text"
    return "string"


def display_name_cn(header: str) -> str:
    return header.replace("_", " ").strip().title()


def short_object_name(stem: str) -> str:
    """BKN display name <= 40 codepoints (SDK local check)."""
    base = display_name_cn(stem.replace("_", " "))
    if len(base) > 40:
        return base[:37] + "..."
    return base


def relation_display_name(rid: str, src_prop: str) -> str:
    """Must satisfy 1..40 codepoints."""
    base = f"{rid} {src_prop}"
    if len(base) <= 40:
        return base
    return rid  # fk_xxxxxxxxxx (≤14)



def pick_pk(headers: list[str], stem: str) -> tuple[str, str]:
    for pref in PK_PREFERENCE:
        if pref in headers:
            return pref, pref
    # Fallback: singular_id from stem e.g. teams -> team_id not always
    guessed = stem[:-1] + "_id" if stem.endswith("s") else f"{stem}_id"
    if guessed in headers:
        return guessed, guessed
    if "key_id" in headers:
        return "key_id", "key_id"
    return headers[0], headers[0]


def pick_display_key(headers: list[str], pk: str) -> str:
    priority = ("tournament_name", "team_name", "match_name", "award_name",
                "confederation_name", "stadium_name", "stage_name")
    for p in priority:
        if p in headers:
            return p
    for h in headers:
        if h.endswith("_name"):
            return h
    return pk


def relation_id(stem: str, col: str, target: str) -> str:
    """Platform IDs: <=40 chars, lowercase [a-z0-9_], no leading underscore."""
    h = hashlib.sha1(f"{stem}:{col}:{target}".encode("utf-8")).hexdigest()[:10]
    return f"fk_{h}"


def main() -> None:
    tables: dict[str, list[str]] = {}
    paths = sorted(DATA_DIR.glob("*.csv"))
    if len(paths) < 27:
        raise SystemExit(f"Expected CSVs under {DATA_DIR}")

    for p in paths:
        with p.open(newline="", encoding="utf-8") as fp:
            r = csv.reader(fp)
            hdr = next(r)
        stem = p.stem
        tables[stem] = hdr

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    for sub in ("object_types", "relation_types", "concept_groups"):
        p = OUT_DIR / sub
        if p.is_dir():
            shutil.rmtree(p)
        p.mkdir(parents=True)

    ot_ids = sorted(tables.keys())

    # Object types
    for stem, headers in sorted(tables.items()):
        pk, _ = pick_pk(headers, stem)
        dk = pick_display_key(headers, pk)
        lines_prop = []
        for h in headers:
            lines_prop.append(
                f"| {h} | {display_name_cn(h)} | {sniff_type(h)} | Fjelstul World Cup CSV 字段 {h} | {h} |"
            )
        body = f'''---
type: object_type
id: {stem}
name: {short_object_name(stem)}
tags: [worldcup, fjelstul]
---

## ObjectType: {short_object_name(stem)}

来自 `data/{stem}.csv` 的世界杯数据结构（Fjelstul World Cup Database）。

### Data Properties

| Name | Display Name | Type | Description | Mapped Field |
|------|--------------|------|-------------|--------------|
{chr(10).join(lines_prop)}

### Keys

Primary Keys: {pk}
Display Key: {dk}
Incremental Key:

### Logic Properties

'''
        path = OUT_DIR / "object_types" / f"{stem}.bkn"
        path.write_text(body.strip() + "\n", encoding="utf-8")

    # Relations from FK_MAP on each table's columns (skip self-id as PK?)
    seen_rel: dict[str, tuple[str, str, str, str]] = {}
    for stem, headers in sorted(tables.items()):
        for col in headers:
            target = FK_MAP.get(col)
            if not target or target == stem:
                continue
            if target not in tables:
                continue
            tgt_pk, _ = pick_pk(tables[target], target)
            rid = relation_id(stem, col, target)
            if rid in seen_rel:
                continue
            # Target property for join (opponent joins teams.team_id)
            tgt_join = tgt_pk
            seen_rel[rid] = (stem, target, col, tgt_join)

    rel_ids = sorted(seen_rel.keys())
    for rid, (src_ot, tgt_ot, src_prop, tgt_prop) in seen_rel.items():
        rn = relation_display_name(rid, src_prop)
        body = f'''---
type: relation_type
id: {rid}
name: {rn}
tags: [worldcup]
---

## RelationType: {rn}

CSV `{src_ot}.{src_prop}` ↔ `{tgt_ot}.{tgt_prop}`。

### Endpoint

| Source | Target | Type |
|--------|--------|------|
| {src_ot} | {tgt_ot} | direct |

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| {src_prop} | {tgt_prop} |
'''
        path = OUT_DIR / "relation_types" / f"{rid}.bkn"
        path.write_text(body.strip() + "\n", encoding="utf-8")

    cg_id = "worldcup_entities"
    cg_body = f'''---
type: concept_group
id: {cg_id}
name: 世界杯核心实体
tags: [worldcup]
---

## ConceptGroup: 世界杯核心实体

常用分析入口：赛事、场次、国家队、球员与进球事件。

### Object Types

| ID | Name | Description |
|----|------|-------------|
| tournaments | tournaments | 历届世界杯赛事 |
| teams | teams | 国家队 |
| players | players | 球员档案 |
| matches | matches | 比赛场次 |
| goals | goals | 进球明细 |
'''

    (OUT_DIR / "concept_groups" / f"{cg_id}.bkn").write_text(cg_body.strip() + "\n", encoding="utf-8")

    net_body = f"""---
type: knowledge_network
id: worldcup_fjelstul_bkn
name: Fjelstul 世界杯数据集（CSV）
tags: [worldcup, fjelstul, CSV]
---

# Fjelstul 世界杯数据集（CSV）

由 `examples/06-world-cup/data/` 共 {len(ot_ids)} 张 CSV 表头自动推导对象类型与外键关联；数据源版权与许可见 README。

## Network Overview

- **ObjectTypes** (object_types/): {", ".join(ot_ids)}
- **RelationTypes** (relation_types/): {", ".join(rel_ids)}
- **ConceptGroups** (concept_groups/): {cg_id}
"""

    (OUT_DIR / "network.bkn").write_text(net_body.strip() + "\n", encoding="utf-8")

    skill_md = """# 世界杯 CSV 知识网络（BKN）

本目录对应 Fjelstul World Cup Database 导出 CSV，列即 `Mapped Field`。灌库时请使用与其它示例一致的表前缀（如 `wc_*`）。

## Object types（路径）

"""

    ot_rows = [f"| `{bid}` | `object_types/{bid}.bkn` |" for bid in ot_ids]
    skill_md += "\n".join(["| Object | Path |"] + ["|--------|------|"] + ot_rows)

    rr = sorted(rel_ids)
    skill_md += "\n## Relation types\n\n"
    lines = ["| Relation ID | Path |"] + ["|---------------|------|"]
    lines += [f"| `{r}` | `relation_types/{r}.bkn` |" for r in rr]
    skill_md += "\n".join(lines)
    skill_md += f"\n\n共 **{len(rr)}** 条关系。\n"
    skill_md += "\n## Concept groups\n\n| Group | Path |\n|-------|------|\n"
    skill_md += f"| {cg_id} | `concept_groups/{cg_id}.bkn` |\n"
    skill_md += "\n## Validation\n\n`kweaver bkn validate ./worldcup-bkn`\n"

    (OUT_DIR / "SKILL.md").write_text(skill_md, encoding="utf-8")


if __name__ == "__main__":
    main()
