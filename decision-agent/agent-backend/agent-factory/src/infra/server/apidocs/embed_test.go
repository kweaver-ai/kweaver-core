package apidocs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentFactoryJSON_AgentConfigSkillsIncludeSkillItems(t *testing.T) {
	t.Parallel()

	doc := mustLoadAgentFactoryDoc(t)

	createReqSchema := mustResolvePathSchemaRef(t, doc,
		"paths", "/api/agent-factory/v3/agent", "post", "requestBody", "content", "application/json", "schema", "$ref",
	)
	createReq := mustResolveSchemaRef(t, doc, createReqSchema)

	configSchema := mustResolveNestedSchema(t, doc, createReq, "properties", "config")

	skillsSchema := mustResolveNestedSchema(t, doc, configSchema, "properties", "skills")
	skillItems := mustNestedMap(t, skillsSchema, "properties", "skills")

	itemRef := mustNestedString(t, skillItems, "items", "$ref")
	skillSchema := mustResolveSchemaRef(t, doc, itemRef)

	skillID := mustNestedMap(t, skillSchema, "properties", "skill_id")
	assert.Equal(t, "string", skillID["type"])
}

func TestAgentFactoryJSON_CreateAgentResponseUsesCreateRes(t *testing.T) {
	t.Parallel()

	doc := mustLoadAgentFactoryDoc(t)

	createRespRef := mustResolvePathSchemaRef(t, doc,
		"paths", "/api/agent-factory/v3/agent", "post", "responses", "201", "content", "application/json", "schema", "$ref",
	)
	createResp := mustResolveSchemaRef(t, doc, createRespRef)

	assert.Contains(t, createResp, "properties")
	assert.NotContains(t, createResp, "config")

	properties := mustNestedMap(t, createResp, "properties")
	assert.Contains(t, properties, "id")
	assert.Contains(t, properties, "version")
	assert.Len(t, properties, 2)
}

func TestAgentFactoryJSON_CreateReactAgentUsesSameSchemasAsCreateAgent(t *testing.T) {
	t.Parallel()

	doc := mustLoadAgentFactoryDoc(t)

	createReqRef := mustResolvePathSchemaRef(t, doc,
		"paths", "/api/agent-factory/v3/agent", "post", "requestBody", "content", "application/json", "schema", "$ref",
	)
	createReactReqRef := mustResolvePathSchemaRef(t, doc,
		"paths", "/api/agent-factory/v3/agent/react", "post", "requestBody", "content", "application/json", "schema", "$ref",
	)
	assert.Equal(t, createReqRef, createReactReqRef)

	createRespRef := mustResolvePathSchemaRef(t, doc,
		"paths", "/api/agent-factory/v3/agent", "post", "responses", "201", "content", "application/json", "schema", "$ref",
	)
	createReactRespRef := mustResolvePathSchemaRef(t, doc,
		"paths", "/api/agent-factory/v3/agent/react", "post", "responses", "201", "content", "application/json", "schema", "$ref",
	)
	assert.Equal(t, createRespRef, createReactRespRef)
}

func TestAgentFactoryJSON_DetailAndUpdateSchemasIncludeSkillItems(t *testing.T) {
	t.Parallel()

	doc := mustLoadAgentFactoryDoc(t)

	updateReqRef := mustResolvePathSchemaRef(t, doc,
		"paths", "/api/agent-factory/v3/agent/{agent_id}", "put", "requestBody", "content", "application/json", "schema", "$ref",
	)
	updateReq := mustResolveSchemaRef(t, doc, updateReqRef)
	assertSkillIDSchemaExists(t, doc, updateReq)

	detailRespRef := mustResolvePathSchemaRef(t, doc,
		"paths", "/api/agent-factory/v3/agent/{agent_id}", "get", "responses", "200", "content", "application/json", "schema", "$ref",
	)
	detailResp := mustResolveSchemaRef(t, doc, detailRespRef)
	assertSkillIDSchemaExists(t, doc, detailResp)
}

func mustLoadAgentFactoryDoc(t *testing.T) map[string]any {
	t.Helper()

	var doc map[string]any

	require.NoError(t, json.Unmarshal(AgentFactoryJSON, &doc))

	return doc
}

func mustResolvePathSchemaRef(t *testing.T, doc map[string]any, keys ...string) string {
	t.Helper()

	value := mustNestedValue(t, doc, keys...)
	ref, ok := value.(string)
	require.True(t, ok, "expected string ref at %v", keys)

	return ref
}

func mustResolveSchemaRef(t *testing.T, doc map[string]any, ref string) map[string]any {
	t.Helper()

	const prefix = "#/components/schemas/"

	require.Contains(t, ref, prefix)

	return mustNestedMap(t, doc, "components", "schemas", ref[len(prefix):])
}

func mustResolveNestedSchema(t *testing.T, doc map[string]any, root map[string]any, keys ...string) map[string]any {
	t.Helper()

	schema := mustNestedMap(t, root, keys...)
	if ref, ok := schema["$ref"].(string); ok {
		return mustResolveSchemaRef(t, doc, ref)
	}

	allOf, ok := schema["allOf"].([]any)
	require.True(t, ok, "expected $ref or allOf at %v", keys)
	require.NotEmpty(t, allOf, "expected non-empty allOf at %v", keys)

	allOfSchema, ok := allOf[0].(map[string]any)
	require.True(t, ok, "expected schema object in allOf at %v", keys)

	ref, ok := allOfSchema["$ref"].(string)
	require.True(t, ok, "expected $ref in allOf at %v", keys)

	return mustResolveSchemaRef(t, doc, ref)
}

func assertSkillIDSchemaExists(t *testing.T, doc map[string]any, schema map[string]any) {
	t.Helper()

	configSchema := mustResolveNestedSchema(t, doc, schema, "properties", "config")
	skillsSchema := mustResolveNestedSchema(t, doc, configSchema, "properties", "skills")
	skillItems := mustNestedMap(t, skillsSchema, "properties", "skills")

	itemRef := mustNestedString(t, skillItems, "items", "$ref")
	skillSchema := mustResolveSchemaRef(t, doc, itemRef)

	skillID := mustNestedMap(t, skillSchema, "properties", "skill_id")
	assert.Equal(t, "string", skillID["type"])
}

func mustNestedMap(t *testing.T, root map[string]any, keys ...string) map[string]any {
	t.Helper()

	value := mustNestedValue(t, root, keys...)
	result, ok := value.(map[string]any)
	require.True(t, ok, "expected map at %v", keys)

	return result
}

func mustNestedString(t *testing.T, root map[string]any, keys ...string) string {
	t.Helper()

	value := mustNestedValue(t, root, keys...)
	result, ok := value.(string)
	require.True(t, ok, "expected string at %v", keys)

	return result
}

func mustNestedValue(t *testing.T, root map[string]any, keys ...string) any {
	t.Helper()

	var current any = root
	for _, key := range keys {
		asMap, ok := current.(map[string]any)
		require.True(t, ok, "expected map before key %q", key)

		value, exists := asMap[key]
		require.True(t, exists, "missing key %q", key)

		current = value
	}

	return current
}
