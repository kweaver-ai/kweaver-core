<p align="center">
  <img alt="KWeaver" src="./assets/logo/banner.png" width="800" />
</p>

[中文](README.zh.md) | English

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt) [![skills.sh kweaver-core](https://img.shields.io/badge/skills.sh-kweaver--core-blue)](https://skills.sh/kweaver-ai/kweaver-sdk/kweaver-core) [![skills.sh create-bkn](https://img.shields.io/badge/skills.sh-create--bkn-green)](https://skills.sh/kweaver-ai/kweaver-sdk/create-bkn)

KWeaver Core is a harness-first foundation for enterprise decision agents. It turns fragmented data, knowledge, tools, and policies into governed context, safe execution, and verifiable feedback loops. With semantic modeling, real-time access, runtime control, and TraceAI, it helps AI systems reason, adapt, and act reliably in complex enterprises.

**On this page:** [📚 Quick links](#toc-quick-links) · [🎬 Demo video](#toc-demo) · [🚀 Quick start](#toc-quick-start) · [🛠️ KWeaver SDK](#toc-kweaver-sdk) · [🛡️ KWeaver Admin CLI](#toc-kweaver-admin) · [🏗️ KWeaver Core](#toc-kweaver-core) · [📐 BKN Lang](#toc-bkn-lang) · [📊 Benchmarks](#toc-benchmarks) · [💬 Community](#toc-community)

> **Note:** KWeaver Core is a **backend-only framework** — it does not include a web UI. All interactions are through the CLI, SDK, or API. If you need a graphical interface, please install [**KWeaver DIP**](https://github.com/kweaver-ai/kweaver).

Want to have an intuitive experience of the core functions of KWeaver DIP? Click the link below to sign up and start your trial experience immediately to quickly unlock the product's value! 👉 Apply for a trial: https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf

<a id="toc-quick-links"></a>

## 📚 Quick Links

- 🌐 [KWeaver DIP](https://dip-poc.aishu.cn/studio/agent/development/my-agent-list) - KWeaver DIP demo environment, [Apply for a trial](https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf)
- 🛠️ [KWeaver SDK](https://github.com/kweaver-ai/kweaver-sdk) - End-user / agent `kweaver` CLI, TypeScript / Python SDK, and AI agent skills
- 🛡️ [kweaver-admin](https://github.com/kweaver-ai/kweaver-admin) - Platform administrator CLI (users / orgs / roles / models / audit) for full installs
- 🤝 [Contributing](rules/CONTRIBUTING.md) - Guidelines for contributing to the project
- 🚢 [Deployment](deploy/README.md) - One-click deploy to Kubernetes
- 📘 [Documentation](help/README.md) - Product documentation and usage guides ([EN](help/en/README.md) / [中文](help/zh/README.md))
- 📦 [Examples](examples/README.md) - End-to-end CLI walkthroughs (DB / CSV / actions)
- 📝 [Blog](https://kweaver-ai.github.io/kweaver-core/) - KWeaver technical articles and updates
- 🧾 [Release Notes](release-notes/) - All notable changes

<a id="toc-demo"></a>

## 🎬 Demo Video

[Watch the KWeaver demo on Bilibili](https://www.bilibili.com/video/BV1nGXVBTEmo/?vd_source=4cdad687b2ac18a0b25e434f1fafe2f7).

<a id="toc-quick-start"></a>

## 🚀 Quick Start

1. **Prerequisites & planning** — read the [Deployment Guide](deploy/README.md) and satisfy its prerequisites.
2. **Run installation scripts**:

```bash
git clone https://github.com/kweaver-ai/kweaver-core.git
cd kweaver-core/deploy
chmod +x deploy.sh

# Minimum installation — recommended for first-time experience
./deploy.sh kweaver-core install --minimum
# Equivalent to:
# ./deploy.sh kweaver-core install --set auth.enabled=false --set businessDomain.enabled=false

# Full installation (includes auth & business-domain modules)
./deploy.sh kweaver-core install

# Or specify addresses explicitly (skips interactive prompts):
#   --access_address       Address for clients to reach KWeaver services (can be IP or domain)
#   --api_server_address   IP bound to a local network interface for K8s API server (must be a real NIC IP)
./deploy.sh kweaver-core install \
  --access_address=<your-ip> \
  --api_server_address=<your-ip>

./deploy.sh --help
```

3. **Verify the deployment**:

```bash
# Check cluster status
kubectl get nodes
kubectl get pods -A

# Check service status
./deploy.sh kweaver-core status
```

4. **Verify API access**

   KWeaver Core is backend-only and does not provide a web console. On the machine you use to reach the cluster (laptop, bastion, etc.), use the `kweaver` CLI from [**kweaver-sdk**](https://github.com/kweaver-ai/kweaver-sdk): either `npm install -g @kweaver-ai/kweaver-sdk` or `npx kweaver` (no global install; see [KWeaver SDK](#toc-kweaver-sdk) below). Then run:

```bash
kweaver auth login https://<node-ip> -k
kweaver bkn list
# or: npx kweaver auth login https://<node-ip> -k
#     npx kweaver bkn list
```

5. **View help**:

```bash
kweaver --help                   # list all commands
kweaver <command> --help         # help for a specific command, e.g. kweaver bkn --help
```

For full product documentation, see the [Documentation](help/README.md) ([EN](help/en/README.md) / [中文](help/zh/README.md)).

> **Did a full install (without `--minimum`)?** Also install [`kweaver-admin`](#toc-kweaver-admin) to manage users, organizations, roles, models, and audit logs — see [help/en/install.md — Administrator tool after a full install](help/en/install.md#-administrator-tool-after-a-full-install-kweaver-admin).

> **No deployment yet?** KWeaver DIP / public demo access may require prior signup — [Apply for a trial](https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf). After you have access, open the [KWeaver DIP](https://dip-poc.aishu.cn/studio/agent/development/my-agent-list) web UI or connect your CLI/SDK to the demo environment (see below).

<a id="toc-kweaver-core"></a>

## 🏗️ KWeaver Core

**KWeaver Core** is the AI-native platform foundation for autonomous decision-making. It sits between AI Agents (above) and AI/Data infrastructure (below), with the **Business Knowledge Network (BKN)** at its center, providing unified data access, execution, and security governance for Agents.

```text
            ┌─────────────────────────────────┐
            │     AI Agents (Decision Agent,   │
            │     Data Agent, HiAgent, ...)    │
            └───────────────┬─────────────────┘
                            │
            ┌───────────────▼─────────────────┐
            │     Business Knowledge Network   │
            │         KWeaver Core             │
            └───────────────┬─────────────────┘
                            │
            ┌───────────────▼─────────────────┐
            │   AI Infrastructure & Data       │
            │   Infrastructure                 │
            └─────────────────────────────────┘
```

KWeaver Core solves two critical pain points when connecting proprietary data with autonomous AI Agents:

### Context Engineering — High-Quality Context for Agents

In long-running agent scenarios, context inevitably faces explosion, decay, pollution, and high token costs. KWeaver Core addresses these through the Business Knowledge Network:

- **Context explosion containment** — Multi-source candidates are first organized and aggregated via the BKN semantic network, then unified by Context Loader (recall → coarse ranking → fine ranking) to retain only key evidence and constraints, avoiding massive prompt fragments that cause decision drift. Overall accuracy reaches **93%+**.
- **Context decay mitigation** — Replaces long-text stacking with "real-time facts + evidence citations", keeping reasoning grounded around specific objects and reducing forgetting and hallucination risks in long inputs. Accuracy improves **15%+** over baselines across scenario types.
- **Context pollution isolation** — Builds precise enterprise digital twins through the BKN network, blocking unreliable content and potential injection risks outside the knowledge and execution boundary, ensuring a clean and controllable reasoning chain.
- **Token cost compression** — Converts multi-source materials into structured object information fetched on demand (not full-text concatenation), improving information density within the same budget. Token consumption reduced **30%+** while improving accuracy.

### Harness Engineering — Safe & Controllable Execution

Beyond "seeing more", Agents must "do it right". KWeaver Core provides constraint engineering capabilities for enterprise-grade safe execution:

- **Explainable decisions** — Uses "Object → Action → Rule → Constraint" knowledge structures to express business intent graphs, grounding tool invocation and parameter selection to explicit semantic boundaries and rule dependencies, making it clear "why this action was taken".
- **Traceable evidence chain** — From action intent → knowledge node → data source → mapping/operator → final invocation, full-chain tracing is supported. Entities and relationships can be traced back to source data and active rules, enabling audit and review.
- **Controllable execution loop** — Unifies identity and access control down to knowledge network object/action permissions, with pre-execution validation, mid-execution policy interception, and post-execution audit logging, achieving "authorizable, approvable, revocable" security loops.
- **Risk prevention mechanism** — Models risks as "Risk Types" linked to Action Types; performs risk assessment and simulation before execution, with automatic downgrade/blocking/secondary confirmation when thresholds are hit, blocking high-risk actions before they execute.

### Core Architecture

```text
┌──────────────────────── KWeaver Core ────────────────────────┐
│         │                                          │         │
│         │  Decision Agent                          │         │
│  Info   │  (Dolphin Runtime / Agent Executor)       │  Trace │
│  Secu-  │──────────────────────────────────────────│         │
│  rity   │                                          │   AI    │
│         │  AI Data Platform                        │         │
│ Fabric  │  ┌────────────────────────────────────┐  │  Obse-  │
│         │  │ Context Loader                     │  │  rvab-  │
│ (ISF)   │  │  ┌───────────┐   ┌─────────────┐  │  │  ility  │
│         │  │  │ Retrieval │ → │   Ranker    │  │  │    /    │
│ Access  │  │  └───────────┘   └─────────────┘  │  │  Evid-  │
│ Control │  ├────────────────────────────────────┤  │  ence   │
│    /    │  │ Business Knowledge Network         │  │         │
│ Secu-   │  │  ┌────────────────────────────┐    │  │         │
│  rity   │  │  │ BKN Engine                 │    │  │         │
│         │  │  │ (Data/Logic/Risk/Action)    │    │  │         │
│         │  │  └──────┬─────────────┬───────┘    │  │         │
│         │  │         ↓ mapping     ↓ mapping    │  │         │
│         │  │  ┌────────┐ ┌───────────┐ ┌──────┐ │  │         │
│         │  │  │  VEGA  │ │ Execution │ │ Data │ │  │         │
│         │  │  │ Engine │ │  Factory  │ │ flow │ │  │         │
│         │  │  └────────┘ └───────────┘ └──────┘ │  │         │
│         │  └────────────────────────────────────┘  │         │
│         │                                          │         │
└─────────┴──────────────────────────────────────────┴─────────┘
               ↕                ↕                ↕
       Multi-source & Multi-modal Data (30+ data sources)
```

| Component | Description |
| --- | --- |
| **AI Data Platform** | Non-intrusive access architecture — unified data access, unified execution, and unified security governance through the Business Knowledge Network |
| **Decision Agent** | Goal-oriented autonomous task planning — acquires high-quality context from AI Data Platform, manages runtime effectively, suppresses hallucination and context decay, invokes tools and skills under permission control, forming a safe "reason → risk-assess → execute → feedback" business loop |
| **Info Security Fabric** | Unified identity, permissions, and policies as a single entry point — end-to-end control and audit over data access, model output, and tool invocation, reducing privilege escalation, leakage, and prompt injection risks |
| **Trace AI** | Full-chain observability and evidence chain tracing — supports issue localization and automatic optimization recommendations, enabling explainable and auditable AI applications |

<a id="toc-bkn-lang"></a>

### 📐 BKN Lang

BKN Lang is a Markdown-based business knowledge modeling language, designed for human-machine bidirectional friendliness:

- **Easy to develop** — Based on standard Markdown syntax, eliminating code barriers entirely. Business experts write, read, and modify definitions via WYSIWYG editors, with rapid version diff and collaborative review — modifying system rules is like editing a document.
- **Easy to understand** — "Object-Relationship-Risk-Action" four-in-one model perfectly maps enterprise business models. Humans read business semantics; Agents parse precise context constraints in real-time. Logic is made explicit, rejecting black boxes, fundamentally reducing LLM reasoning hallucination and logical deviation.
- **Easy to integrate** — Definitions stored as full-text in specific database fields with no complex underlying table coupling. Context Loader dynamically loads on demand, discarding static hardcoding. Plug-and-play across systems and Agents, flowing as lightweight assets through AI Data Platform.

### Key Value Summary

| Metric | Value |
| --- | --- |
| **Scenario Coverage** | Q&A, workflow execution, intelligence analysis, decision judgment, exploration |
| **TCO Reduction** | 70% lower with integrated platform |
| **BKN Build Efficiency** | 300% improvement in knowledge network construction |
| **Token Cost Savings** | 50% reduction through context optimization and compression |

<a id="toc-kweaver-sdk"></a>

## 🛠️ KWeaver SDK

<a id="toc-kweaver-core-and-sdk"></a>

### Install the SDK on the client

After deploying KWeaver Core, we recommend installing [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk) as your next step. The SDK provides the `kweaver` CLI and AI Agent Skills — the primary way to interact with the platform.

[**kweaver-sdk**](https://github.com/kweaver-ai/kweaver-sdk) gives AI agents (Claude Code, GPT, custom agents, etc.) access to KWeaver knowledge networks and Decision Agents via the `kweaver` CLI. It also provides Python and TypeScript SDKs for programmatic integration.

Install the CLI with:

```bash
npm install -g @kweaver-ai/kweaver-sdk
```

Or run it without a global install:

```bash
npx kweaver --help
```

### AI Agent Skills

Install skills from [**kweaver-sdk**](https://github.com/kweaver-ai/kweaver-sdk) with [`npx skills`](https://www.npmjs.com/package/skills).

**Install both skills** (recommended):

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk \
  --skill kweaver-core --skill create-bkn
```

- **`kweaver-core`** — full KWeaver APIs and CLI conventions so assistants can operate KWeaver on your behalf. See [skills/kweaver-core/SKILL.md](https://github.com/kweaver-ai/kweaver-sdk/blob/main/skills/kweaver-core/SKILL.md).
- **`create-bkn`** — guided workflow and tooling to create and manage **Business Knowledge Networks (BKN)** from your AI coding assistant. See [skills/create-bkn/SKILL.md](https://github.com/kweaver-ai/kweaver-sdk/blob/main/skills/create-bkn/SKILL.md).

**Install one skill** (optional):

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk --skill kweaver-core
# or
npx skills add https://github.com/kweaver-ai/kweaver-sdk --skill create-bkn
```

**Before using any skill**, authenticate with your KWeaver instance:

```bash
kweaver auth login https://your-kweaver-instance.com
```

> **Self-signed certificate?** If your instance uses a self-signed or untrusted TLS certificate (common for fresh deployments without a CA-issued cert), add `-k` to skip certificate verification:
>
> ```bash
> kweaver auth login https://your-kweaver-instance.com -k
> ```

### Try with Demo Environment

No deployment needed — [Apply for a trial](https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf) first, then connect your AI agent to the demo environment (for the web UI, visit [KWeaver DIP](https://dip-poc.aishu.cn/studio/agent/development/my-agent-list)).

```bash
npx skills add https://github.com/kweaver-ai/kweaver-sdk \
  --skill kweaver-core --skill create-bkn

npm install -g @kweaver-ai/kweaver-sdk
kweaver auth login https://dip-poc.aishu.cn -k
```

Then ask your AI agent (Cursor, Claude Code, etc.) using natural language:

```
List all knowledge networks
What object types are in the supply chain knowledge network?
Search the supply chain knowledge network for "supply chain risks"
Show 2 sample customer records
List all Decision Agents
Chat with Agent xxx, ask "What is the current inventory status?"
```

Or use `/kweaver-core` slash commands (the skill takes over automatically):

```
/kweaver-core List all knowledge networks
/kweaver-core What's in the supply chain knowledge network?
/kweaver-core Search knowledge network for "supply chain risks"
/kweaver-core Show 2 sample customer records from the knowledge network
/kweaver-core List all Decision Agents
/kweaver-core Chat with Agent <agent-id>, ask "What is the current inventory status?"
```

> **Demo access:** [Apply for a trial](https://kweaver-ai.feishu.cn/share/base/form/shrcni732cNDY4x3A5SYTncrguf) if needed.

### Headless / no-browser authentication (SSH, CI, containers)

The `kweaver` CLI supports authenticating without a local browser or without pasting callback URLs.

**Which option to use**

| Your situation | Use | Notes |
| --- | --- | --- |
| **Username and password** sign-in is available | **Method 1** (HTTP, `-u` / `-p`) | One command on the host; no need to copy an OAuth callback from elsewhere. |
| **[kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk) is installed** (the `kweaver` command works) | **Method 2** (`auth export` / replay) | After browser login, run `kweaver auth export` and **replay** the one-line command on the headless target. |
| kweaver-sdk is **not** installed; you usually run **`npx kweaver`** | **Method 3** (`--no-browser`) | After signing in on another device, an **extra step**: **copy the full callback URL** or **only the authorization code** (*copy code*), then paste at `Paste URL or code` in the headless terminal. |

**Method 1 — Username/password HTTP sign-in (fully non-interactive on the host, no browser required)**

No Node/Chromium needed — the CLI calls the platform's `/oauth2/signin` endpoint over HTTPS and stores the returned tokens. Suitable for CI runners, minimal Linux containers, and any host without a browser:

```bash
kweaver auth login https://your-instance -u <username> -p <password> -k
```

`-u` / `-p` together select this path automatically (you can also add `--http-signin` explicitly). If you omit `-u` / `-p`, the CLI prompts for them on stdin (password input is hidden on a TTY). The CLI saves tokens under `~/.kweaver/` including a `refresh_token` when the IdP returns one — same auto-refresh behavior as a normal browser login.

**Method 2 — Export & replay (kweaver-sdk installed; export and replay)**

On a machine that **has** the `kweaver` CLI, sign in with the browser once, then `export` a one-line command for the headless host — you do not need to transcribe the OAuth callback URL or code in the terminal by hand.

1. On a machine **with** a browser, run `kweaver auth login https://your-instance`. After success, export credentials:

```bash
kweaver auth export              # prints a one-line command you can paste on the headless host
```

2. On the **headless** host, paste the exported command. It uses `--client-id`, `--client-secret`, and `--refresh-token` to exchange tokens and save credentials under `~/.kweaver/`:

```bash
kweaver auth login https://your-instance \
  --client-id <ID> --client-secret <SECRET> --refresh-token <TOKEN>
```

**Method 3 — `--no-browser` (when kweaver-sdk is not installed; extra copy/paste of URL or code)**

Use this when the CLI is not installed globally and you run via `npx kweaver` (or similar). Compared with **Method 2**, you must **manually copy** the URL or code from the browser after login.

```bash
kweaver auth login https://your-instance --no-browser
# or: npx kweaver auth login https://your-instance --no-browser
```

The CLI prints an OAuth URL instead of opening a local browser window. Open that URL on **any device** with a browser (phone, laptop, etc.). After login, the browser redirects to a `localhost` callback — an error page is normal. Copy the **full URL** from the address bar, **or only the authorization code**, and paste it at the prompt below (the extra **Paste URL or code** step):

```
Open this URL on any device (use a private/incognito window if you need the full sign-in form):

  https://your-instance/oauth2/auth?redirect_uri=...&client_id=...

After login, the browser may show an error page (this is expected if nothing listens on localhost).
Copy the FULL URL from the address bar and paste it here, or paste only the authorization code.

Paste URL or code>
```

> With saved `~/.kweaver/` sessions, the CLI automatically exchanges `refresh_token` for a new access token when it expires — no extra flags needed. You can also set environment variables (`KWEAVER_BASE_URL`, `KWEAVER_TOKEN`) instead of persisting credentials to disk.

Full details: [kweaver-sdk — Authentication](https://github.com/kweaver-ai/kweaver-sdk#authentication) and [Headless / Server Authentication](https://github.com/kweaver-ai/kweaver-sdk/blob/main/packages/typescript/README.md#headless--server-authentication). The Python `kweaver` CLI still uses interactive browser login; reuse the `~/.kweaver/` directory from a machine where the Node CLI finished login, or set the environment variables above.

### CLI

```bash
kweaver auth login https://your-kweaver.com -k    # authenticate (-k for self-signed TLS)
kweaver bkn list                                 # list knowledge networks
kweaver bkn search <kn-id> "query"               # semantic search
kweaver agent chat <agent-id> -m "Hello"         # chat with a Decision Agent
kweaver --help                                   # all subcommands
```

### TypeScript & Python SDK

Minimal example (after CLI login or equivalent credentials):

```typescript
import kweaver from "@kweaver-ai/kweaver-sdk/kweaver";
kweaver.configure({ config: true, bknId: "your-bkn-id", agentId: "your-agent-id" });

const results = await kweaver.search("What risks exist in the supply chain?");
const reply   = await kweaver.chat("Summarise the top 3 risks");
await kweaver.weaver({ wait: true });   // rebuild BKN index
```

```python
import kweaver
kweaver.configure(config=True, bkn_id="your-bkn-id", agent_id="your-agent-id")

results = kweaver.search("What risks exist in the supply chain?")
reply   = kweaver.chat("Summarise the top 3 risks")
```

For streaming, `KWeaverClient`, and the full API surface, see the [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk) repository docs and examples.

<a id="toc-kweaver-admin"></a>

## 🛡️ KWeaver Admin CLI

[`kweaver-admin`](https://github.com/kweaver-ai/kweaver-admin) is a separate npm CLI for **platform administrators**, complementary to the `kweaver` CLI from `kweaver-sdk`:

| CLI | Audience | Scope |
| --- | --- | --- |
| `kweaver` (`@kweaver-ai/kweaver-sdk`) | End users / Agents | BKN, Decision Agent, Action, Skill, query |
| `kweaver-admin` (`@kweaver-ai/kweaver-admin`) | Platform administrators | Users, organizations, roles, models, audit, raw HTTP |

> Most `kweaver-admin` commands target services that come with a **full install** (`auth.enabled=true`, `businessDomain.enabled=true`): `user-management`, `deploy-manager`, `deploy-auth`, `eacp`, `mf-model-manager`, OAuth2 (Hydra). On a `--minimum` install most commands return 401 / 404 — expected.

### Install

Requires Node.js 22+ (kweaver CLIs; matches `@kweaver-ai/kweaver-sdk` `engines` on npm). Credentials are stored under `~/.kweaver-admin/platforms/`, isolated from `~/.kweaver/`.

```bash
npm install -g @kweaver-ai/kweaver-admin
kweaver-admin --version
```

### Login

```bash
# Browser OAuth2 (skip TLS for self-signed certs)
kweaver-admin auth login https://<access-address> -k

# Username/password (CI / headless)
kweaver-admin auth login https://<access-address> -u <user> -p <password> -k

# Or via environment variables
export KWEAVER_BASE_URL=https://<access-address>
export KWEAVER_ADMIN_TOKEN=<bearer-token>   # falls back to KWEAVER_TOKEN
```

### Common admin tasks

```bash
kweaver-admin org tree                           # list departments
kweaver-admin user create --login alice          # default password: 123456 (forced change at first login)
kweaver-admin user reset-password -u alice       # admin reset
kweaver-admin role list
kweaver-admin role add-member <roleId> -u alice
kweaver-admin llm add                            # register an LLM
kweaver-admin small-model add                    # register an embedding model
kweaver-admin audit list --user alice --start 2026-04-01 --end 2026-04-30
kweaver-admin call /api/user-management/v1/management/users -X GET   # raw HTTP with auth header
```

> New users created by `user create` always start with the platform default password **`123456`** and are forced to change it on first sign-in (this is documented upstream behavior of the ISF user store, not a CLI choice). For lost-password flows, prefer `kweaver-admin user reset-password`.

> Respect the **separation-of-duties** built-in accounts (`system`, `admin`, `security`, `audit`) — operators should use individual accounts, not the shared `admin`.

Full command tree, security notes, and `auth change-password` (EACP `modifypassword`, same `401001017` first-login flow as the `kweaver` CLI): see [kweaver-admin README](https://github.com/kweaver-ai/kweaver-admin) and [help/en/install.md — Administrator tool after a full install](help/en/install.md#-administrator-tool-after-a-full-install-kweaver-admin).

### AI Agent Skill — `kweaver-admin`

`kweaver-admin` ships its own progressive-disclosure skill so AI coding assistants can drive admin tasks (auth, org, user, role, model, audit, raw HTTP) on your behalf. Install with `npx skills`:

```bash
npx skills add https://github.com/kweaver-ai/kweaver-admin --skill kweaver-admin
```

After installing, authenticate once with `kweaver-admin auth login https://your-instance -k`, then ask your assistant in natural language (`/kweaver-admin` slash also works):

```text
List all roles
Create user alice and assign every role from role list
Reset alice's password
Show login audit for alice in the last 7 days
Register an embedding model named bge-m3 against https://api.siliconflow.cn
```

Skill source: [skills/kweaver-admin/SKILL.md](https://github.com/kweaver-ai/kweaver-admin/blob/main/skills/kweaver-admin/SKILL.md). It complements the `kweaver-core` / `create-bkn` skills above (which target the `kweaver` CLI).

<a id="toc-benchmarks"></a>

## 📊 Benchmarks & Experiments

See more details: [KWeaver Blog](https://kweaver-ai.github.io/kweaver-core/)

### Unstructured Data Q&A — Cross-Platform Comparison

Based on 145 HR scenario samples (resume corpus with 118 multi-format PDFs), covering simple information lookup, cross-section experience analysis, and multi-hop comprehensive reasoning. All platforms used DeepSeek V3.2 + BGE M3-Embedding with identical data sources, tested in Agentic mode.

| Metric | KWeaver Core (v0.3.0) | BiSheng | Dify (v0.15.3) | RAGFlow (v0.17.0) |
| --- | --- | --- | --- | --- |
| **Accuracy** | **99.31%** (144/145) | 86.90% (126/145) | 96.55% (140/145) | 86.90% (126/145) |
| **Avg Latency** | 43.69s | **19.52s** | 63.82s | 71.56s |
| **P90 Latency** | 56.92s | 32.53s | 79.15s | 95.37s |
| **Avg Token** | 21.36K | **4.98K** | 36.25K | 16.28K |

KWeaver Core is the only platform that breaks the traditional RAG "performance impossible triangle" — achieving >99% accuracy while keeping inference cost and latency at production-ready levels. Dify trades high token consumption (1.7x) for decent accuracy; BiSheng sacrifices reasoning depth for speed; RAGFlow falls behind on both accuracy and latency.

### Ablation Studies — Key Technical Levers

The following ablation experiments identify the contribution of each KWeaver Core component:

**Retrieval Depth** — Increasing retrieval limit from 10 to 20 raised accuracy from 96.67% to 100%, with only +6.48s latency. Context Loader's semantic reranking and compression enable KWeaver Core to effectively handle the increased context without "Lost in the Middle" effects.

**Schema Preloading** — With Context Loader preloading the BKN schema:

| Configuration | Accuracy | Avg Steps | Avg Token |
| --- | --- | --- | --- |
| Schema preloaded | **100.0%** | **3.2** | **20.07K** |
| No schema | 93.33% | 5.8 | 38.54K |

Schema provides Agents with a clear "map" of entity relationships, reducing reasoning steps by 44.8% and token consumption by 47.9%.

**Tool Curation** — Precise tool selection outperforms providing all available tools:

| Configuration | Accuracy | Avg Steps | Avg Token |
| --- | --- | --- | --- |
| Full toolset (6 tools) | 75.0% | 7.1 | 42.3K |
| Curated toolset (3 tools) | **100.0%** | **2.4** | **12.8K** |

With excessive tools, Agents favor "seemingly powerful" broad-search tools whose noise triggers reflection loops and path divergence. Curated tools constrain Agents onto the correct path, achieving one-shot resolution.

**Path Guidance** — Encoding domain expert experience into executable reasoning templates:

| Configuration | Path Guidance | Tools | Accuracy | Avg Latency | Avg Token |
| --- | --- | --- | --- | --- | --- |
| Explore-kn_search | Yes | 3 | **100.0%** | **37.82s** | **15,420** |
| No guidance, 2 tools | No | 2 | 100.0% | 53.06s | 23,287 |
| No guidance, 3 tools | No | 3 | 80.0% | 53.28s | 19,870 |

Path guidance tells Agents "how to walk" for efficiency; tool curation "reduces wrong turns" for stability. Combined, they deliver optimal production performance.

### Heterogeneous Data Reasoning — F1 Bench

F1 Bench is based on the BIRD test set with the Formula-1 database mixed with 30 unstructured documents, testing Agent capabilities in structured + unstructured heterogeneous data reasoning.

| Metric | KWeaver Core | Dify Retrieval Baseline |
| --- | --- | --- |
| **Overall Accuracy** | **92.96%** | 78.87% |
| **SQL Waste Rate** | **8.2%** | 24.5% |
| **SQL Hit Efficiency** | **0.226** | 0.137 |
| **Total SQL Calls** | **292** | 408 |

<a id="toc-community"></a>

## 🤝 Community

**KWeaver WeChat Group QR Code**

<img src="./help/qrcode.png" width="200" alt="KWeaver WeChat group QR code" />

Scan to join the KWeaver community group.
