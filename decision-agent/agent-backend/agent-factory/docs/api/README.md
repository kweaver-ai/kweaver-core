# Agent Factory API Documentation

This directory contains the final OpenAPI documentation artifacts for external viewing.

**Language versions:**
- [中文](README.zh.md)

## Files in This Directory

- `agent-factory.json`
  - OpenAPI 3 JSON version, suitable for tool integration
- `agent-factory.yaml`
  - OpenAPI 3 YAML version, suitable for manual reading and distribution
- `agent-factory.html`
  - Scalar-style static documentation page, suitable for Try it Out
- `agent-factory-redoc.html`
  - Redoc-style static documentation page, suitable for read-only display
- `favicon.png`
  - Icon for static documentation pages

## How to Generate

Execute in the `agent-backend/agent-factory` directory:

```bash
make gen-api-docs
```

## How to Validate

```bash
make validate-api-docs
```

This command will simultaneously check:

- Whether the OpenAPI structure is valid
- Whether the number of paths and endpoints meets expectations
- Whether runtime HTML still only depends on local embed resources
- Whether JSON / YAML / favicon are consistent with runtime copies

## APIChat Path Description

- Recommended to use `/api/agent-factory/v1/api/chat/completion`
  - Specify the target agent via `agent_key` in the body
- `/api/agent-factory/v1/app/{app_key}/api/chat/completion`
  - Deprecated, retained only for backward compatibility
  - New integrations should migrate to the main entry without `app_key`

The static templates used by `GetAPIDoc` are also maintained in `src/static/agent-api.json` and `src/static/agent-api.yaml` following the recommended paths above.

## v3 Agent Config Mode Description

The `config` of `/api/agent-factory/v3/agent` now uses `mode` as the primary mode field:

- `default`
  - Uses `system_prompt`, `plan_mode`
- `dolphin`
  - Uses `dolphin`
  - `is_dolphin_mode` is retained for backward compatibility with old requests
  - Added `is_use_tool_id_in_dolphin`
- `react`
  - Uses `system_prompt`, `react_config`, `plan_mode`

External documentation and detail responses uniformly use `react_config`.

## React Agent Dedicated Creation Endpoint

Added a React Agent dedicated creation endpoint:

- `/api/agent-factory/v3/agent/react`
  - Request body is consistent with `/api/agent-factory/v3/agent`
  - Only used for creating ReAct mode agents
  - Handler layer will additionally validate `config.mode`, returning `400` if it's not `react`

## How to View

### View Static Pages Directly

Open any page directly in the current directory:

- `agent-factory.html`
- `agent-factory-redoc.html`

These pages load frontend documentation JS resources from `https://cdn.jsdmirror.com/`.

Applicable scenarios:

- Need to distribute a single HTML file directly to others for viewing
- Want to avoid including the `ui/` local resource directory

Note:

- Network access to the CDN is required on first open
- The OpenAPI documentation content within the page is still embedded in the HTML file itself

### View After Starting Service

After starting Agent Factory, visit:

- `http://127.0.0.1:30777/scalar`
- `http://127.0.0.1:30777/redoc`
- `http://127.0.0.1:30777/scalar/doc.json`
- `http://127.0.0.1:30777/scalar/doc.yaml`

This set of runtime pages continues to use service-embedded local UI resources and does not depend on external CDNs.

Recommended approach:

- Need to send requests, Try it Out: Use Scalar page
- Need better read-only documentation display: Use Redoc page

## In-Depth Maintenance Guide

If you need to understand the generation pipeline, overlay, baseline, or Swagger intermediate artifacts, please refer to:

- `../../cmd/openapi-docs/README.simple.md`
- `../../cmd/openapi-docs/docs/OPENAPI_AUTOMATION_GUIDE.md`
