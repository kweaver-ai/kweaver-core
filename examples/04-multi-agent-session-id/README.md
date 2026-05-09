# 04 · Multi-Agent — Custom Input Propagation

> A platform feature audit shows a custom input field travels intact through `father → son1/son2 → SKILL`, every step verifiable.

## The Problem

When you compose multiple agents on the platform, you need to know that user-supplied
context — a `session_id`, an `org_id`, a tenant marker — actually reaches every nested
agent and skill. Reading the LLM's reply doesn't prove it: the LLM might just be
echoing what it saw in the prompt. You need evidence the **platform itself** routed
the field, not the model's good behavior.

## What This Shows

A minimal three-agent + one-skill setup:

- `exp_father` — Dolphin orchestration, calls `@exp_son1(...)` then `@exp_son2(...)`
- `exp_son1` / `exp_son2` — identical sons that read `input.session_id` and use the SKILL
- `exp_session_echo` — a SKILL that echoes the literal `session_id` value verbatim

The script invokes `exp_father` with `custom_querys.session_id=DEMO-XXX` and verifies:

1. **Literal echo** — each son's textual answer contains
   `[exp_session_echo] RECEIVED session_id=DEMO-XXX from <son_name>`.
2. **Platform-side routing** — each son's `input_message` (what the platform actually
   delivered to the son) contains the prefix `[input.session_id=DEMO-XXX]`.

The second assertion is the one that rules out LLM hallucination. Only the platform's
agent-executor injects that prefix via the son's Dolphin DSL.

## Prerequisites

- `kweaver` CLI ≥ 0.6.4 with an authenticated session (`kweaver auth whoami`)
- `jq`, `bash`, `curl` on your local machine
- Business domain `bd_public` (default)

No database, no CSV files, no `.env` setup. The example uses the platform's default
LLM and the toolboxes that ship with KWeaver Core.

## Quick Start

```bash
./run.sh                                  # ensure artifacts + invoke + verify
./run.sh --session-id MY-CUSTOM-XYZ       # custom session_id (default: DEMO-2026-XXXXXX)
./run.sh --cleanup                        # delete the SKILL + 3 agents from the platform
```

Successful run tails to:

```
[exp] son1 answer (first 200 chars): ... [exp_session_echo] RECEIVED session_id=DEMO-2026-A1B2C3 from exp_son1 ...
[exp] son2 answer (first 200 chars): ... [exp_session_echo] RECEIVED session_id=DEMO-2026-A1B2C3 from exp_son2 ...
[exp] assert_literal: exp_son1 — PASS
[exp] assert_literal: exp_son2 — PASS
[exp] assert_propagation: exp_son1 input contains [input.session_id=DEMO-2026-A1B2C3] — PASS
[exp] assert_propagation: exp_son2 input contains [input.session_id=DEMO-2026-A1B2C3] — PASS
[exp] ALL ASSERTIONS PASSED. session_id=DEMO-2026-A1B2C3, conversation=01KQ...
[exp] (artifacts kept on platform; run with --cleanup to remove)
```

End-to-end takes ~30 seconds.

## Architecture

```
user
  │  POST /api/agent-factory/v1/app/<father_key>/chat/completion
  │  body: { ..., query, custom_querys: { session_id: "DEMO-XXX" }, stream: false }
  ▼
exp_father  is_dolphin_mode=1
  dolphin DSL:
    @exp_son1(session_id=$session_id, query=$query) -> res_1
    @exp_son2(session_id=$session_id, query=$query) -> res_2
  ├─► exp_son1
  │     "[input.session_id=" + $session_id + "] " + $query -> q
  │     /explore/(history=true) <prompt>\n$q\n -> answer
  │     ├─ list_skills_v2() — discovers exp_session_echo
  │     ├─ builtin_skill_load(skill_id) — fetches SKILL.md
  │     └─ echoes "[exp_session_echo] RECEIVED session_id=DEMO-XXX from exp_son1"
  └─► exp_son2  (identical config except agent name)
```

The full final answer lives at `final_answer.answer_type_other.{res_1,res_2}.answer.answer`.
The `input_message` field on the same path proves the platform routed `session_id` from
father into the son's call.

## Flow

| Step | What happens |
|------|-------------|
| 1 | Register the `exp_session_echo` SKILL (or reuse if already registered) |
| 2 | Create + publish `exp_son1` and `exp_son2` from the jq template (or reuse) |
| 3 | Create + publish `exp_father` binding both sons via `skills.agents` |
| 4 | POST to `/chat/completion` with `custom_querys.session_id` set |
| 5 | Two assertions on the response (literal echo + routed-input prefix) |

Default behavior is **idempotent**: agents and skill are kept on the platform after a
successful run, so the next invocation just reuses them. Use `--cleanup` to remove.

## File Layout

```
04-multi-agent-session-id/
├── README.md / README.zh.md       # this file (and Chinese long-form notes)
├── run.sh                         # entry point
├── skills/
│   └── exp_session_echo/SKILL.md  # SKILL content
├── configs/
│   ├── base.config.json           # platform-template-derived base
│   ├── son.config.template.jq     # son renderer (Dolphin mode)
│   ├── father.config.template.jq  # father renderer (Dolphin orchestration)
│   └── .schema-notes.md           # swagger field-name notes
└── lib/
    ├── common.sh   # shared helpers + name vars
    ├── render.sh   # jq render functions
    ├── verify.sh   # the two assertions
    └── cleanup.sh  # opt-in teardown (only --cleanup invokes this)
```

## Where to Look for Evidence

After a run:

- `/tmp/exp_run_resp.json` — full chat completion response (the source of both assertions)
- `kweaver agent get <agent_id> --verbose` — current platform-side config of any agent
  (note the `--verbose` flag — without it, `agent get` returns a thin response)

Note: `kweaver agent trace` currently returns HTTP 500 on the demo platform
(Uniquery DataView issue). The chat completion response carries enough information for
the two assertions, so this example does not depend on it.

## Cleanup

Unlike the other examples in this folder, **artifacts are kept by default** — the SKILL
and three agents stay on the platform so you can browse them in the Web UI or invoke
them again without paying the create cost. Run `./run.sh --cleanup` when you're done.

## Platform Notes Discovered While Building This

- Custom input fields like `session_id` only land in `input.<name>` when sent inside
  `custom_querys: { ... }` — top-level body keys are ignored.
- `is_dolphin_mode=0` (plain prompt) ignores the `pre_dolphin` array entirely; the
  executor synthesizes its own program from `system_prompt`. To inject custom fields
  into what the LLM sees, switch to `is_dolphin_mode=1` and write the dolphin program.
- `list_skills_v2` (toolbox `tmp_skill_discovery_R1`) requires a `tool_input` entry
  mapping `X-Authorization` from `header.authorization` (`map_type=var`); without it
  the proxy returns 401.
- Schema field is `tool_box_id` (with the underscore between `box`), not `toolbox_id`.
- LLM id lives at `.llms[0].llm_config.id`, not `.llms[0].id`.

The Chinese long-form [README.zh.md](./README.zh.md) goes deeper on each of these.
