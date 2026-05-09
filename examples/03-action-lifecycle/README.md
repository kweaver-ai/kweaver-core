# 03 · Action Lifecycle — Self-Evolving Knowledge Network

> A knowledge network that monitors material inventory and acts before a stockout hits the line.

## The Problem

Every morning, a procurement engineer scans the inventory ledger to find which materials have
dropped below safety stock and need urgent replenishment. One missed material means a
production stoppage.

## What This Shows

A knowledge network is not a static query layer. Once you define **action types** and a
**schedule**, it operates autonomously:

- **Finds the right entities** — identifies material objects where `material_risk == critical`
- **Triggers follow-up actions** — calling your business system to raise a replenishment alert
- **Records everything** — full audit trail in `action-log`

The engineer arrives at 08:00. The replenishment list is already there.

## Prerequisites

- `kweaver` CLI ≥ 0.6.4 and a logged-in session (`kweaver auth whoami`)
- MySQL accessible from the kweaver platform
- Python 3 on your local machine

## Quick Start

```bash
cp env.sample .env
# Edit .env with your DB credentials
./run.sh
```

## Flow

| Step | What happens |
|------|-------------|
| 1 | Connect MySQL datasource |
| 2 | Import CSVs → build knowledge network (inventory + production orders) |
| 3–5 | Register action tool backend |
| 6 | Define action type: *"find materials where `material_risk == critical`, trigger replenishment alert"* |
| 7 | Query confirms 3 critically low materials (MAT-001/003/005) |
| 8–9 | Schedule: runs every day at 08:00 automatically |
| 10 | Manual trigger: see results immediately |
| 11 | Audit log: the network's history of autonomous actions |

## Note on Execute Status

The demo tool backend does not perform real write-back (that is your business system's job).
Execute may show `failed` at the tool level — the execution record and audit log are still
written correctly. In production, replace the tool binding with your ERP or notification API.

## Cleanup

Resources are deleted automatically when the script exits (success or failure).
