# 📒 Cookbook (English)

Task-oriented recipes for KWeaver: each entry is a self-contained "**one goal / a few commands / one output**" guide that you can copy and run.

> The sibling [module docs](../README.md) are the reference manual organized **by subsystem**. The cookbook is organized **by "what do I want to do"**. They cross-link instead of duplicating each other.

## Index

| Recipe | One-line goal |
| --- | --- |
| [Build a knowledge network from CSV in one shot](./cookbook_example.md) | Use `kweaver bkn create-from-csv` to turn local CSV files into a queryable KN |

## Template for a new recipe

Copy [`_TEMPLATE.md`](./_TEMPLATE.md) and adapt it to your scenario; see the worked-out [`cookbook_example.md`](./cookbook_example.md) for reference.

Name new files `NN-short-slug.md` and keep the structure consistent:

0. **Metadata card** (top blockquote) — difficulty, time, modules touched, CLI version
1. **Goal** — open with "**After this recipe you will have:** ..."; outcome-oriented and observable
2. **Prerequisites** — versions, login, business domain, recipe-specific dependencies
3. **Steps** — numbered steps with runnable commands; split into `### 3.x` once you have more than one step; put alternative or advanced paths inside `<details>`
4. **Expected output** — start with one "**Success criterion**" line, then a trimmed real snippet
5. **Troubleshooting** — the "Symptom" column should be the **literal output or error a reader will see**
6. **See also** — links to the [module docs](../README.md), [`examples/`](../examples/README.md), and related recipes

> Prefer the **`kweaver`** CLI; show an equivalent `curl` only when needed. Never paste private tokens or real customer data into examples.
