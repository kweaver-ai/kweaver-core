# <Recipe title — verb-first, states the outcome in one line>

> - **Difficulty**: ⭐ Beginner / ⭐⭐ Intermediate / ⭐⭐⭐ Advanced
> - **Time**: ~ N minutes
> - **Modules touched**: `<bkn|datasource|dataflow|...>`
> - **CLI version**: `kweaver >= x.y`

## 1. Goal

> Lead with "**After this recipe you will have:** ..." — outcome-oriented and observable; do not just restate the title.

## 2. Prerequisites

- Logged in via `kweaver auth login <platform-url>`.
- Business domain: `kweaver config show` to confirm; switch with `kweaver config set-bd <uuid>` if needed.
- <List recipe-specific dependencies: datasource / files / an existing KN, etc.>

## 3. Steps

> Split with `### 3.x` once you have more than one step; each `### 3.x` keeps at most one paragraph + one code block — wrap further detail under `### 3.x.y`.
> Put alternative or advanced paths inside `<details>` so they do not interrupt the main flow.

### 3.1 <Step name>

```bash
# the kweaver CLI commands a reader must run
```

### 3.2 <Step name>

```bash
# ...
```

Add a quick-reference table when the command has many flags:

| Parameter | Required | Description |
| --- | --- | --- |
| `<param>` | yes/no | <one line> |

## 4. Expected output

> **Success criterion**: <one explicit observable, e.g. `total > 0` and `datas[0]` contains field X>

```jsonc
{
  // a trimmed, real snippet — strip secrets
}
```

## 5. Troubleshooting

> The "Symptom" column should be the **literal output or error a reader will see**, so they can search-match it.

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| `<concrete error or output>` | <cause> | <one-line command or action> |

## 6. See also

- References: [<manual page>](../manual/<x>.md) · [Quick start](../quick-start.md)
- End-to-end sample project: [`examples/<NN-slug>/`](../../../examples/<NN-slug>/)
- Related recipe: [<other cookbook>](./<other-recipe>.md)
