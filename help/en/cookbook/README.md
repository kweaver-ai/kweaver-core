# 📒 Cookbook (English)

Task-oriented recipes for KWeaver: each entry is a self-contained "**one goal / a few commands / one output**" guide that you can copy and run.

> The sibling [module docs](../README.md) are the reference manual organized **by subsystem**. The cookbook is organized **by "what do I want to do"**. They cross-link instead of duplicating each other.

## Index

| Recipe | One-line goal |
| --- | --- |
| [Build a knowledge network from CSV in one shot](./cookbook_example.md) | Use `kweaver bkn create-from-csv` to turn local CSV files into a queryable KN |

## Template for a new recipe

Name new files `NN-short-slug.md` and use the same six sections:

1. **Goal** — one sentence on what you'll have after the recipe
2. **Prerequisites** — versions, login, business domain, etc.
3. **Steps** — numbered steps with runnable commands
4. **Expected output** — a trimmed real-world snippet
5. **Troubleshooting** — common 401 / 403 / business-domain mismatches
6. **See also** — links to the [module docs](../README.md) and matching items in [`examples/`](../examples/README.md)

> Prefer the **`kweaver`** CLI; show an equivalent `curl` only when needed. Never paste private tokens or real customer data into examples.
