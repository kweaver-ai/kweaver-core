package actions

import (
	"encoding/json"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
)

// DatasetWriteDocs 向指定 dataset 写入文档数据
type DatasetWriteDocs struct {
	DatasetID string `json:"dataset_id"`
	Documents any    `json:"documents"`
}

// Name 返回操作符名称
func (d *DatasetWriteDocs) Name() string {
	return common.OpDatasetWriteDocs
}

// ParameterNew 创建新的参数实例
func (d *DatasetWriteDocs) ParameterNew() any {
	return &DatasetWriteDocs{}
}

// normalizeDatasetDocuments 将输入文档标准化为 []map[string]any
func normalizeDatasetDocuments(documents any) []map[string]any {
	switch v := documents.(type) {
	case string:
		var parsed any
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil
		}
		return normalizeDatasetDocuments(parsed)
	case map[string]any:
		return []map[string]any{v}
	case []any:
		results := make([]map[string]any, 0, len(v))
		for _, item := range v {
			switch elem := item.(type) {
			case map[string]any:
				results = append(results, elem)
			case string:
				var parsed any
				if err := json.Unmarshal([]byte(elem), &parsed); err == nil {
					if nestedResults := normalizeDatasetDocuments(parsed); nestedResults != nil {
						results = append(results, nestedResults...)
					}
				}
			}
		}
		return results
	case []map[string]any:
		return v
	default:
		return nil
	}
}

// Run 执行写入操作
func (d *DatasetWriteDocs) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	log := traceLog.WithContext(ctx.Context())
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	input := params.(*DatasetWriteDocs)

	// 验证参数
	if input.DatasetID == "" {
		return nil, fmt.Errorf("dataset_id is required")
	}

	documents := normalizeDatasetDocuments(input.Documents)
	if len(documents) == 0 {
		return map[string]any{
			"total":   0,
			"success": 0,
			"failed":  0,
		}, nil
	}

	vegaBackend := drivenadapters.NewVegaBackend()

	result := map[string]any{
		"total": len(documents),
	}

	// 批量写入，每批 1000 条
	batchSize := 1000
	success, failed := 0, 0
	reasons := []string{}

	for i := 0; i < len(documents); i += batchSize {
		end := min(i+batchSize, len(documents))
		batch := documents[i:end]

		if err = vegaBackend.WriteDatasetDocuments(ctx.Context(), input.DatasetID, batch); err != nil {
			log.Warnf("[DatasetWriteDocs] batch %d-%d failed: %s", i, end, err.Error())
			reasons = append(reasons, fmt.Sprintf("[%d-%d] %s", i, end, err.Error()))
			failed += len(batch)
		} else {
			success += len(batch)
		}
	}

	result["success"] = success
	result["failed"] = failed

	if len(reasons) > 0 {
		result["reasons"] = reasons
	}

	ctx.ShareData().Set(ctx.GetTaskID(), result)
	return result, nil
}

// 确保 DatasetWriteDocs 实现了 Action 接口
var _ entity.Action = (*DatasetWriteDocs)(nil)
