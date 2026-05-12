# Changelog

Chinese version: [CHANGELOG.zh-CN.md](CHANGELOG.zh-CN.md)

## 0.8.0

### Features & Improvements

- Add concept group scoped Schema discovery in `search_schema`
  - Support `search_scope.concept_groups` to limit object, relation, and action schema discovery to selected BKN concept groups
  - Keep existing `search_schema` behavior unchanged when concept groups are not provided
  - Return referenced object types together with scoped relation and action schemas, so callers receive a complete Schema context
  - Note: metric schema requests carry the same concept group scope, but actual metric filtering depends on BKN metrics support
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/304
- Embed the ContextLoader standard toolset in the service startup flow
  - Automatically sync the built-in toolset to execution-factory during ContextLoader startup
  - Use `ContextLoader 标准内置工具集；契约版本: 0.8.0` as the toolset contract description
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/306

### Compatibility

- Keep compatible and legacy HTTP paths unchanged; new integrations that need concept group scope should use `search_schema`

### Documentation

- Update API, MCP schema, toolset, and release documentation for concept group scoped Schema discovery
- Document the built-in ContextLoader toolset delivery model and contract version rule

## 0.7.0

### Features & Improvements

- Add `search_schema` as the standard Schema discovery entry for MCP and HTTP callers
  - Support object, relation, action, and metric schema discovery from one interface
  - Use request body `kn_id` for the HTTP `search_schema` API
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/189
  - issue: https://github.com/kweaver-ai/kweaver-core/issues/234
- Consolidate MCP Schema discovery around `search_schema` to reduce tool selection ambiguity for agents

### Compatibility

- Keep `kn_search` as a compatible HTTP path and `semantic-search` as a legacy HTTP path

### Documentation

- Update release overview and tool usage documentation for the `search_schema` entry unification and metric schema recall contract
