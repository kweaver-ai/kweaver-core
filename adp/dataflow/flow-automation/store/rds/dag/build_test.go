package dagmodel

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_Convert(t *testing.T) {
	conv := NewConverter("users")

	tests := []struct {
		name  string
		query map[string]interface{}
	}{
		{
			name: "1. 简单等值查询",
			query: map[string]interface{}{
				"name": "john",
			},
			// → SELECT * FROM users WHERE name = ?
			// params: ["john"]
		},
		{
			name: "2. 多字段 AND",
			query: map[string]interface{}{
				"_id":  "abc123",
				"type": "admin",
			},
			// → SELECT * FROM users WHERE _id = ? AND type = ?
			// params: ["abc123", "admin"]
		},
		{
			name: "3. 比较操作符",
			query: map[string]interface{}{
				"age": map[string]interface{}{
					"$gt":  18,
					"$lte": 65,
				},
				"status": "active",
			},
			// → SELECT * FROM users WHERE age > ? AND age <= ? AND status = ?
			// params: [18, 65, "active"]
		},
		{
			name: "4. $or 查询",
			query: map[string]interface{}{
				"_id": "abc123",
				"$or": []interface{}{
					map[string]interface{}{"userid": "u001"},
					map[string]interface{}{"type": "admin"},
				},
			},
			// → SELECT * FROM users WHERE _id = ? AND (userid = ? OR type = ?)
			// params: ["abc123", "u001", "admin"]
		},
		{
			name: "5. $in 查询",
			query: map[string]interface{}{
				"status": map[string]interface{}{
					"$in": []interface{}{"active", "pending", "review"},
				},
			},
			// → SELECT * FROM users WHERE status IN (?, ?, ?)
			// params: ["active", "pending", "review"]
		},
		{
			name: "6. $ne 和 $exists",
			query: map[string]interface{}{
				"removed": map[string]interface{}{
					"$ne": true,
				},
				"email": map[string]interface{}{
					"$exists": true,
				},
			},
			// → SELECT * FROM users WHERE email IS NOT NULL AND removed != ?
			// params: [true]
		},
		{
			name: "7. 复杂嵌套查询",
			query: map[string]interface{}{
				"name":    "john",
				"userid":  "u001",
				"removed": false,
				"$or": []interface{}{
					map[string]interface{}{
						"age": map[string]interface{}{"$gte": 18},
					},
					map[string]interface{}{
						"type": map[string]interface{}{
							"$in": []interface{}{"admin", "superadmin"},
						},
					},
				},
			},
		},
		{
			name: "8. NULL 查询",
			query: map[string]interface{}{
				"deleted_at": nil,
				"name":       "john",
			},
			// → SELECT * FROM users WHERE deleted_at IS NULL AND name = ?
		},
		{
			name: "9. $regex 模糊查询",
			query: map[string]interface{}{
				"name": map[string]interface{}{
					"$regex": "^john",
				},
			},
			// → SELECT * FROM users WHERE name LIKE ?
			// params: ["john%"]
		},
		{
			name: "10. $nin 查询",
			query: map[string]interface{}{
				"type": map[string]interface{}{
					"$nin": []interface{}{"banned", "deleted"},
				},
			},
			// → SELECT * FROM users WHERE type NOT IN (?, ?)
		},
		{
			name: "11. $nin 查询",
			query: map[string]interface{}{
				"_id":  "dagID",
				"type": "type",
				"removed": bson.M{
					"$ne": true,
				},
			},
			// → SELECT * FROM users WHERE _id = ? AND removed != ? AND type = ?
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.query)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}

			queryJSON, _ := json.MarshalIndent(tt.query, "  ", "  ")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("📋 %s\n", tt.name)
			fmt.Printf("  MongoDB: %s\n", string(queryJSON))
			fmt.Printf("  SQL:     %s\n", result.SQL)
			fmt.Printf("  Params:  %v\n", result.Params...)
			fmt.Println()
		})
	}
}

func TestBuildDagStepIndex(t *testing.T) {
	dag := &entity.Dag{BaseInfo: entity.BaseInfo{ID: "1"}, Steps: []entity.Step{
		{Operator: "op1", Parameters: map[string]interface{}{"docid": "d1"}},
		{Operator: "op2", Parameters: map[string]interface{}{"docids": []interface{}{"d2", "d3"}}, DataSource: &entity.DataSource{}},
	}}
	rows := BuildDagStepIndex(dag)
	if len(rows) < 3 {
		t.Fatalf("expected >= 3 rows, got %d", len(rows))
	}
}

func TestBuildDagAccessorIndex(t *testing.T) {
	dag := &entity.Dag{BaseInfo: entity.BaseInfo{ID: "1"}, Accessors: []entity.Accessor{{ID: "a1"}, {ID: "a2"}}}
	rows := BuildDagAccessorIndex(dag)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestBuildDagIndexSubquery_TriggerSources(t *testing.T) {
	input := &mod.ListDagInput{Trigger: []string{"op"}, Sources: []string{"s1"}}
	sql, args := BuildDagIndexSubquery(input)
	if !strings.Contains(sql, "t_flow_dag_step") || !strings.Contains(sql, "UNION") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) == 0 {
		t.Fatalf("expected args")
	}
}

func TestBuildDagIndexSubquery_ScopeAll(t *testing.T) {
	input := &mod.ListDagInput{Scope: "all", UserID: "u1", Accessors: []string{"a1"}, TriggerExclude: []string{"op"}}
	sql, _ := BuildDagIndexSubquery(input)
	if !strings.Contains(sql, "t_flow_dag_accessor") || !strings.Contains(sql, "t_flow_dag_step") {
		t.Fatalf("unexpected sql: %s", sql)
	}
}
