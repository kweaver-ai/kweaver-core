package opensearch

import (
	"testing"

	"vega-backend/interfaces"
)

func TestFlattenNestedGroupByRows_TwoDimensions(t *testing.T) {
	conn := &OpenSearchConnector{}
	params := &interfaces.ResourceDataQueryParams{
		Limit: 10,
		Aggregation: &interfaces.Aggregation{
			Property: "id",
			Aggr:     "count",
			Alias:    "__value",
		},
		GroupBy: []*interfaces.GroupByItem{
			{Property: "kn_id"},
			{Property: "module_type"},
		},
	}

	rootAgg := map[string]any{
		"buckets": []any{
			map[string]any{
				"key": "yzm_mock_system",
				"group_by_module_type": map[string]any{
					"buckets": []any{
						map[string]any{
							"key": "a",
							"__value": map[string]any{
								"value": float64(3),
							},
						},
						map[string]any{
							"key": "b",
							"__value": map[string]any{
								"value": float64(2),
							},
						},
					},
				},
			},
		},
	}

	rows := conn.flattenNestedGroupByRows(rootAgg, params, "__value")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0]["kn_id"] != "yzm_mock_system" || rows[0]["module_type"] != "a" || rows[0]["__value"] != float64(3) {
		t.Fatalf("unexpected first row: %#v", rows[0])
	}

	if rows[1]["kn_id"] != "yzm_mock_system" || rows[1]["module_type"] != "b" || rows[1]["__value"] != float64(2) {
		t.Fatalf("unexpected second row: %#v", rows[1])
	}
}
