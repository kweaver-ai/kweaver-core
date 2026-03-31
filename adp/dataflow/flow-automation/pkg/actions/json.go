package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

type JsonGet struct {
	Json   any          `json:"json"`
	Fields []*JsonField `json:"fields"`
}

func (a *JsonGet) Name() string {
	return common.OpJsonGet
}

func (a *JsonGet) ParameterNew() interface{} {
	return &JsonGet{}
}

func (a *JsonGet) Run(ctx entity.ExecuteContext, input interface{}, _ *entity.Token) (interface{}, error) {
	fields := make(map[string]interface{})

	params := input.(*JsonGet)
	var obj interface{}
	if raw, ok := params.Json.(string); ok {
		_ = json.Unmarshal([]byte(raw), &obj)
	} else {
		obj = params.Json
	}

	for index, field := range params.Fields {
		var fieldValueRaw string

		if fieldValue, ok := field.Value.(string); ok {
			fieldValueRaw = fieldValue
		} else {
			bytes, _ := json.Marshal(field.Value)
			fieldValueRaw = string(bytes)
		}

		var value interface{}
		var fieldValue interface{}
		_ = json.Unmarshal([]byte(fieldValueRaw), &fieldValue)

		paths := keyToPath(field.Key)
		switch field.Type {
		case "string":
			if _, ok := fieldValue.(string); !ok {
				fieldValue = fieldValueRaw
			}
			value = lookupJson(obj, paths, fieldValue)
			if _, ok := value.(string); !ok {
				value = fieldValue
			}

		case "number":
			if _, ok := fieldValue.(float64); !ok {
				fieldValue = 0
			}
			value = lookupJson(obj, paths, fieldValue)
			if _, ok := value.(float64); !ok {
				value = fieldValue
			}
		case "boolean":
			if _, ok := fieldValue.(bool); !ok {
				fieldValue = false
			}
			value = lookupJson(obj, paths, fieldValue)
			if _, ok := value.(bool); !ok {
				value = fieldValue
			}
		case "array":
			if _, ok := fieldValue.([]interface{}); !ok {
				fieldValue = make([]interface{}, 0)
			}
			value = lookupJson(obj, paths, fieldValue)
			if _, ok := value.([]interface{}); !ok {
				value = fieldValue
			}
		case "object":
			if _, ok := fieldValue.(map[string]interface{}); !ok {
				fieldValue = make(map[string]interface{})
			}
			value = lookupJson(obj, paths, fieldValue)
			if _, ok := value.(map[string]interface{}); !ok {
				value = fieldValue
			}
		}

		if _, ok := value.(string); !ok {
			bytes, _ := json.Marshal(value)
			value = string(bytes)
		}

		fields[fmt.Sprintf("_%d", index)] = value
	}

	result := map[string]interface{}{
		"fields": fields,
	}

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)
	ctx.Trace(ctx.Context(), "run end")

	return result, nil
}

type JsonField struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type JsonSet struct {
	Json   any          `json:"json"`
	Fields []*JsonField `json:"fields"`
}

func (a *JsonSet) Name() string {
	return common.OpJsonSet
}

func (a *JsonSet) ParameterNew() interface{} {
	return &JsonSet{}
}

func (a *JsonSet) Run(ctx entity.ExecuteContext, input interface{}, _ *entity.Token) (interface{}, error) {
	params := input.(*JsonSet)
	result := map[string]interface{}{
		"json": params.Json,
	}

	var obj interface{}
	if raw, ok := params.Json.(string); ok {
		_ = json.Unmarshal([]byte(raw), &obj)
	} else {
		obj = params.Json
	}

	if obj == nil {
		obj = map[string]interface{}{}
	}

	for _, field := range params.Fields {

		var fieldValueRaw string

		if fieldValue, ok := field.Value.(string); ok {
			fieldValueRaw = fieldValue
		} else {
			bytes, _ := json.Marshal(field.Value)
			fieldValueRaw = string(bytes)
		}

		var fieldValue interface{}
		_ = json.Unmarshal([]byte(fieldValueRaw), &fieldValue)

		paths := keyToPath(field.Key)
		switch field.Type {
		case "string":
			if _, ok := fieldValue.(string); !ok {
				fieldValue = fieldValueRaw
			}

		case "number":
			if _, ok := fieldValue.(float64); !ok {
				fieldValue = 0
			}

		case "boolean":
			if _, ok := fieldValue.(bool); !ok {
				fieldValue = false
			}

		case "array":
			if _, ok := fieldValue.([]interface{}); !ok {
				fieldValue = make([]interface{}, 0)
			}

		case "object":
			if _, ok := fieldValue.(map[string]interface{}); !ok {
				fieldValue = make(map[string]interface{})
			}
		}

		obj = setJson(obj, paths, fieldValue)
	}

	bytes, _ := json.Marshal(obj)
	result["json"] = string(bytes)

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)
	ctx.Trace(ctx.Context(), "run end")

	return result, nil
}

type JsonTemplate struct {
	Json     any    `json:"json"`
	Template string `json:"template"`
}

func (a *JsonTemplate) Name() string {
	return common.OpJsonTemplate
}

func (a *JsonTemplate) ParameterNew() interface{} {
	return &JsonTemplate{}
}

func (a *JsonTemplate) Run(ctx entity.ExecuteContext, input interface{}, _ *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)

	params := input.(*JsonTemplate)
	result := map[string]interface{}{
		"result": "",
	}

	var data interface{}
	if raw, ok := params.Json.(string); ok {
		_ = json.Unmarshal([]byte(raw), &data)
	} else {
		data = params.Json
	}

	if data == nil {
		data = map[string]any{}
	}

	// Render template
	t, err := template.New("example").Funcs(template.FuncMap{
		"dict":   dict,
		"merge":  merge,
		"get":    get,
		"set":    set,
		"unset":  unset,
		"hasKey": hasKey,
		"list":   list,
		"append": my_append,
		"keys":   keys,
		"values": values,
		"pick":   pick,
		"omit":   omit,
		"pluck":  pluck,
		"add":    func(a, b int) int { return a + b },
		"sub":    func(a, b int) int { return a - b },
		"mul":    func(a, b int) int { return a * b },
		"div":    func(a, b int) int { return a / b },
	}).Parse(params.Template)
	if err != nil {
		ctx.Trace(newCtx, fmt.Sprintf("run error: %v", err))
		return nil, err
	}

	var output bytes.Buffer
	if err := t.Execute(&output, data); err != nil {
		ctx.Trace(newCtx, fmt.Sprintf("run error: %v", err))
		return nil, err
	}

	result["result"] = output.String()

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)

	ctx.Trace(newCtx, "run end")
	return result, nil
}

// Helper functions for template
func dict(pairs ...interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for i := 0; i < len(pairs); i += 2 {
		key := pairs[i]
		value := pairs[i+1]
		m[key.(string)] = value
	}
	return m
}

func pluck(key string, d ...map[string]interface{}) []interface{} {
	res := []interface{}{}
	for _, dict := range d {
		if val, ok := dict[key]; ok {
			res = append(res, val)
		}
	}
	return res
}

func pick(dict map[string]interface{}, keys ...string) map[string]interface{} {
	res := map[string]interface{}{}
	for _, k := range keys {
		if v, ok := dict[k]; ok {
			res[k] = v
		}
	}
	return res
}

func omit(dict map[string]interface{}, keys ...string) map[string]interface{} {
	res := map[string]interface{}{}

	omit := make(map[string]bool, len(keys))
	for _, k := range keys {
		omit[k] = true
	}

	for k, v := range dict {
		if _, ok := omit[k]; !ok {
			res[k] = v
		}
	}
	return res
}

func merge(dist map[string]interface{}, source ...map[string]interface{}) map[string]interface{} {
	for _, d := range source {
		for k, v := range d {
			dist[k] = v
		}
	}
	return dist
}

func keys(dicts ...map[string]interface{}) []string {
	k := []string{}
	for _, dict := range dicts {
		for key := range dict {
			k = append(k, key)
		}
	}
	return k
}

func values(dict map[string]interface{}) []interface{} {
	values := []interface{}{}
	for _, value := range dict {
		values = append(values, value)
	}

	return values
}

func get(d map[string]interface{}, key string) interface{} {
	if val, ok := d[key]; ok {
		return val
	}
	return ""
}

func set(d map[string]interface{}, key string, value interface{}) map[string]interface{} {
	d[key] = value
	return d
}

func unset(d map[string]interface{}, key string) map[string]interface{} {
	delete(d, key)
	return d
}

func hasKey(d map[string]interface{}, key string) bool {
	_, ok := d[key]
	return ok
}

func list(values ...interface{}) []interface{} {
	return values
}

func my_append(s []interface{}, values ...interface{}) []interface{} {
	return append(s, values...)
}

type JsonParse struct {
	Json any `json:"json"`
}

func (a *JsonParse) Name() string {
	return common.OpJsonParse
}

func (a *JsonParse) ParameterNew() interface{} {
	return &JsonParse{}
}

func (a *JsonParse) Run(ctx entity.ExecuteContext, input interface{}, _ *entity.Token) (interface{}, error) {
	params := input.(*JsonParse)

	var data any

	if raw, ok := params.Json.(string); ok {
		raw = strings.TrimSpace(raw)
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimSuffix(raw, "```")
		_ = json.Unmarshal([]byte(raw), &data)
	} else {
		data = params.Json
	}

	result := map[string]any{
		"data": data,
	}

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)
	ctx.Trace(ctx.Context(), "run end")

	return result, nil
}

var (
	_ entity.Action = (*JsonGet)(nil)
	_ entity.Action = (*JsonSet)(nil)
	_ entity.Action = (*JsonTemplate)(nil)
)
