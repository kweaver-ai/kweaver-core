# 03 · Action Lifecycle — Self-Evolving Knowledge Network

> A knowledge network that watches your production line and acts before things go wrong.

## The Problem

Every morning, a procurement engineer checks three systems — inventory records, material lists,
and production orders — to find which orders are at risk of material shortage. One missed order
means a production stoppage.

## What This Shows

A knowledge network is not a static query layer. Once you define **action types** and a
**schedule**, it operates autonomously:

- **Finds the right entities** — using inventory-material-order relationship context to identify
  production orders whose materials are critically low
- **Triggers follow-up actions** — calling your business systems
- **Records everything** — full audit trail in `action-log`

The engineer arrives at 08:00. The list is already there.

## Prerequisites

- `kweaver` CLI ≥ 0.6.3 and a logged-in session (`kweaver auth whoami`)
- MySQL accessible from the kweaver platform
- Python 3 and `curl` on your local machine

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
| 6 | Define action type: *"find production orders with critically low materials"* |
| 7 | Query confirms 5 at-risk orders found via inventory-material-order relationships |
| 8–9 | Schedule: runs every day at 08:00 automatically |
| 10 | Manual trigger: see results immediately |
| 11 | Audit log: the network's history of autonomous actions |

## Note on Execute Status

The demo tool backend does not perform real write-back (that is your business system's job).
Execute may show `failed` at the tool level — the execution record and audit log are still
written correctly. In production, replace the tool binding with your ERP or notification API.

## Cleanup

Resources are deleted automatically when the script exits (success or failure).
