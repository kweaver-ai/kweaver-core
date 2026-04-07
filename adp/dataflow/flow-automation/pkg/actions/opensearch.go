package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

type OpenSearchBulkUpsert struct {
	BaseType  string                 `json:"base_type"`
	DataType  string                 `json:"data_type"`
	Category  string                 `json:"category"`
	Documents any                    `json:"documents"`
	Template  string                 `json:"template"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
	Mappings  map[string]interface{} `json:"mappings,omitempty"`
}

func (b *OpenSearchBulkUpsert) Name() string {
	return common.OpOpenSearchBulkUpsert
}

func (b *OpenSearchBulkUpsert) ParameterNew() any {
	return &OpenSearchBulkUpsert{}
}

// getDefaultIndexTemplate 获取默认的索引配置模板
func getDefaultIndexTemplate() (map[string]interface{}, map[string]interface{}) {
	settings := map[string]interface{}{
		"number_of_shards":         1,
		"number_of_replicas":       0,
		"knn":                      true,
		"knn.algo_param.ef_search": 100,
		"refresh_interval":         "30s",
	}

	mappings := map[string]interface{}{
		"dynamic": false,
		"properties": map[string]interface{}{
			"doc_name": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 256,
					},
				},
			},
			"doc_md5": map[string]interface{}{
				"type": "keyword",
			},
			"slice_md5": map[string]interface{}{
				"type": "keyword",
			},
			"id": map[string]interface{}{
				"type": "keyword",
			},
			"deduplication_id": map[string]interface{}{
				"type": "keyword",
			},
			"document_id": map[string]interface{}{
				"type": "keyword",
			},
			"slice_type": map[string]interface{}{
				"type": "integer",
			},
			"pages": map[string]interface{}{
				"type": "integer",
			},
			"segment_id": map[string]interface{}{
				"type": "integer",
			},
			"slice_content": map[string]interface{}{
				"type":     "text",
				"analyzer": "standard",
			},
			"text_vector": map[string]interface{}{
				"type":      "knn_vector",
				"dimension": 768,
			},
			"img_path": map[string]interface{}{
				"type":  "keyword",
				"index": false,
			},
			"image_vector": map[string]interface{}{
				"type":      "knn_vector",
				"dimension": 512,
			},
			"created_at": map[string]interface{}{
				"type": "date",
			},
			"updated_at": map[string]interface{}{
				"type": "date",
			},
		},
	}

	return settings, mappings
}

// getAdvancedIndexTemplate 获取高级索引配置模板，支持动态字段类型预设
func getAdvancedIndexTemplate() (map[string]interface{}, map[string]interface{}) {
	settings := map[string]interface{}{
		"number_of_shards":         1,
		"number_of_replicas":       0,
		"knn":                      true,
		"knn.algo_param.ef_search": 100,
		"refresh_interval":         "30s",
	}

	mappings := map[string]interface{}{
		"properties": map[string]interface{}{
			"lat_lon": map[string]interface{}{
				"type":  "geo_point",
				"store": "true",
			},
		},
		"date_detection": "true",
		"dynamic_templates": []map[string]interface{}{
			{
				"int": map[string]interface{}{
					"match": "*_int",
					"mapping": map[string]interface{}{
						"type":  "integer",
						"store": "true",
					},
				},
			},
			{
				"ulong": map[string]interface{}{
					"match": "*_ulong",
					"mapping": map[string]interface{}{
						"type":  "unsigned_long",
						"store": "true",
					},
				},
			},
			{
				"long": map[string]interface{}{
					"match": "*_long",
					"mapping": map[string]interface{}{
						"type":  "long",
						"store": "true",
					},
				},
			},
			{
				"short": map[string]interface{}{
					"match": "*_short",
					"mapping": map[string]interface{}{
						"type":  "short",
						"store": "true",
					},
				},
			},
			{
				"numeric": map[string]interface{}{
					"match": "*_flt",
					"mapping": map[string]interface{}{
						"type":  "float",
						"store": true,
					},
				},
			},
			{
				"tks": map[string]interface{}{
					"match": "*_tks",
					"mapping": map[string]interface{}{
						"type":       "text",
						"similarity": "scripted_sim",
						"analyzer":   "whitespace",
						"store":      true,
					},
				},
			},
			{
				"ltks": map[string]interface{}{
					"match": "*_ltks",
					"mapping": map[string]interface{}{
						"type":     "text",
						"analyzer": "whitespace",
						"store":    true,
					},
				},
			},
			{
				"kwd": map[string]interface{}{
					"match_pattern": "regex",
					"match":         "^(.*_(kwd|id|ids|uid|uids)|uid)$",
					"mapping": map[string]interface{}{
						"type":       "keyword",
						"similarity": "boolean",
						"store":      true,
					},
				},
			},
			{
				"nested": map[string]interface{}{
					"match": "*_nst",
					"mapping": map[string]interface{}{
						"type": "nested",
					},
				},
			},
			{
				"object": map[string]interface{}{
					"match": "*_obj",
					"mapping": map[string]interface{}{
						"type":    "object",
						"dynamic": "true",
					},
				},
			},
			{
				"string": map[string]interface{}{
					"match_pattern": "regex",
					"match":         "^.*_(with_weight|list)$",
					"mapping": map[string]interface{}{
						"type":  "text",
						"index": "false",
						"store": true,
					},
				},
			},
			{
				"rank_feature": map[string]interface{}{
					"match": "*_fea",
					"mapping": map[string]interface{}{
						"type": "rank_feature",
					},
				},
			},
			{
				"rank_features": map[string]interface{}{
					"match": "*_feas",
					"mapping": map[string]interface{}{
						"type": "rank_features",
					},
				},
			},
			{
				"knn_vector_512": map[string]interface{}{
					"match": "*_512_vec",
					"mapping": map[string]interface{}{
						"type":      "knn_vector",
						"dimension": 512,
					},
				},
			},
			{
				"knn_vector_768": map[string]interface{}{
					"match": "*_768_vec",
					"mapping": map[string]interface{}{
						"type":      "knn_vector",
						"dimension": 768,
					},
				},
			},
			{
				"knn_vector_1024": map[string]interface{}{
					"match": "*_1024_vec",
					"mapping": map[string]interface{}{
						"type":      "knn_vector",
						"dimension": 1024,
					},
				},
			},
			{
				"knn_vector_1536": map[string]interface{}{
					"match": "*_1536_vec",
					"mapping": map[string]interface{}{
						"type":      "knn_vector",
						"dimension": 1536,
					},
				},
			},
			{
				"binary": map[string]interface{}{
					"match": "*_bin",
					"mapping": map[string]interface{}{
						"type": "binary",
					},
				},
			},
		},
	}

	return settings, mappings
}

func normalizeDocuments(documents any, baseType, dataType, category string) (results []map[string]any) {
	switch v := documents.(type) {
	case string:
		var parsed any
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil
		}
		return normalizeDocuments(parsed, baseType, dataType, category)
	case map[string]any:
		v["__base_type"] = baseType
		v["__data_type"] = dataType
		v["__catetory"] = category

		writeTime := time.Now().UnixMilli()
		v["__write_time"] = writeTime
		if _, ok := v["@timestamp"]; !ok {
			v["@timestamp"] = writeTime
		}
		return []map[string]any{v}
	case []any:
		for _, item := range v {
			switch elem := item.(type) {
			case map[string]any:
				elem["__base_type"] = baseType
				elem["__data_type"] = dataType
				elem["__catetory"] = category

				writeTime := time.Now().UnixMilli()
				elem["__write_time"] = writeTime
				if _, ok := elem["@timestamp"]; !ok {
					elem["@timestamp"] = writeTime
				}
				results = append(results, elem)
			case string:
				var parsed any
				if err := json.Unmarshal([]byte(elem), &parsed); err == nil {
					if nestedResults := normalizeDocuments(parsed, baseType, dataType, category); nestedResults != nil {
						results = append(results, nestedResults...)
					}
				}
			}
		}
		return results
	default:
		return nil
	}
}

func (b *OpenSearchBulkUpsert) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	log := traceLog.WithContext(ctx.Context())
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	taskIns := ctx.GetTaskInstance()
	input := params.(*OpenSearchBulkUpsert)
	openSearch := drivenadapters.NewOpenSearch()
	documents := normalizeDocuments(input.Documents, input.BaseType, input.DataType, input.Category)

	index := "mdl-" + input.BaseType

	// 确定使用的 settings 和 mappings
	settings := input.Settings
	mappings := input.Mappings

	// 如果用户没有提供 settings 和 mappings，使用内置模板
	if settings == nil && mappings == nil {
		if input.Template != "" {
			// 如果指定了模板，使用指定的模板
			switch input.Template {
			case "default":
				settings, mappings = getDefaultIndexTemplate()
				log.Infof("[OpenSearchBulkUpsert] taskInsID %s, using default index template", taskIns.TaskID)
			case "advanced":
				settings, mappings = getAdvancedIndexTemplate()
				log.Infof("[OpenSearchBulkUpsert] taskInsID %s, using advanced index template", taskIns.TaskID)
			}
		} else {
			// 如果没有指定模板，默认使用高级模板
			settings, mappings = getAdvancedIndexTemplate()
			log.Infof("[OpenSearchBulkUpsert] taskInsID %s, using default advanced index template", taskIns.TaskID)
		}
	}

	result := map[string]any{}
	success, failed := 0, 0
	reasons := []string{}
	batchSize := 1000
	for i := 0; i < len(documents); i += batchSize {
		end := min(i+batchSize, len(documents))
		batch := documents[i:end]
		err = openSearch.BulkUpsert(ctx.Context(), index, batch, settings, mappings)

		if err != nil {
			log.Warnf("[OpenSearchBulkUpsert] taskInsID %s, total %d, range %d-%d, err: %s", taskIns.TaskID, len(documents), i, end, err.Error())
			reasons = append(reasons, fmt.Sprintf("[%d-%d] %s", i, end, err.Error()))
			failed += len(batch)
		} else {
			success += len(batch)
		}

		for j := range batch {
			batch[j] = nil
		}
	}

	result["total"] = len(documents)
	result["success"] = success
	result["failed"] = failed

	if len(reasons) > 0 {
		result["reasons"] = reasons
	}

	ctx.ShareData().Set(ctx.GetTaskID(), result)
	return result, nil
}

var _ entity.Action = (*OpenSearchBulkUpsert)(nil)
