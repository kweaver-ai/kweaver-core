# Quick start

This walkthrough assumes KWeaver Core is already [deployed](installation/deploy.md) and [verified](installation/verify.md).

## 1. Install the `kweaver` CLI

```bash
npm install -g @kweaver-ai/kweaver-sdk
kweaver --help
```

## 2. Authenticate

```bash
kweaver auth login https://<access-address> -k
```

Use the same host you configured as `--access_address` during install. With `--minimum` install, auth may be relaxed; follow prompts from the CLI.

## 3. Explore the Business Knowledge Network (BKN)

```bash
kweaver bkn list
```

See [BKN Engine](bkn.md) for modeling, instances, and APIs.

## 4. Inspect data plane (VEGA)

After you have connections and views configured, query through VEGA-backed APIs or SDKs. See [VEGA Engine](vega.md).

## 5. Run a Decision Agent workflow

Create or select an agent template, then chat or invoke runs via CLI/SDK. See [Decision Agent](decision-agent.md).

## 6. Observe with Trace AI

When tracing is enabled, use Trace AI APIs or the observability query service to inspect spans linked to agent runs. See [Trace AI](trace-ai.md).

## Where to go next

| Goal | Doc |
| --- | --- |
| Model business knowledge | [bkn.md](bkn.md) |
| Unified data access | [vega.md](vega.md) |
| Pipelines | [dataflow.md](dataflow.md) |
| Tools for agents | [execution-factory.md](execution-factory.md) |
| Context for LLMs | [context-loader.md](context-loader.md) |
| Security layer | [isf.md](isf.md) |

Programmatic examples for each module list **CLI**, **Python SDK**, **TypeScript SDK**, and **curl** in the same file.
