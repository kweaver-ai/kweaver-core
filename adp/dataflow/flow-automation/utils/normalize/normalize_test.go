package normalize

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNormalizeContainer(t *testing.T) {
	ts := primitive.NewDateTimeFromTime(primitive.DateTime(0).Time())
	input := primitive.D{
		{Key: "group_id", Value: "g1"},
		{Key: "nested", Value: primitive.M{
			"items": primitive.A{
				primitive.D{{Key: "name", Value: "alice"}},
				bson.M{"created_at": ts},
			},
		}},
	}

	got := NormalizeContainer(input)
	result, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", got)
	}

	if result["group_id"] != "g1" {
		t.Fatalf("expected group_id g1, got %v", result["group_id"])
	}

	nested, ok := result["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected nested map, got %T", result["nested"])
	}

	items, ok := nested["items"].([]interface{})
	if !ok {
		t.Fatalf("expected items slice, got %T", nested["items"])
	}

	first, ok := items[0].(map[string]interface{})
	if !ok || first["name"] != "alice" {
		t.Fatalf("expected normalized first item, got %#v", items[0])
	}

	second, ok := items[1].(map[string]interface{})
	if !ok {
		t.Fatalf("expected normalized second item, got %T", items[1])
	}

	if second["created_at"] != ts {
		t.Fatalf("expected bson scalar leaf to be preserved, got %#v", second["created_at"])
	}
}

func TestAsMap(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantOK  bool
		wantKey string
		wantVal interface{}
	}{
		{
			name:    "primitive d",
			input:   primitive.D{{Key: "group_id", Value: "g1"}},
			wantOK:  true,
			wantKey: "group_id",
			wantVal: "g1",
		},
		{
			name:    "primitive m",
			input:   primitive.M{"group_id": "g2"},
			wantOK:  true,
			wantKey: "group_id",
			wantVal: "g2",
		},
		{
			name:    "plain map",
			input:   map[string]interface{}{"group_id": "g3"},
			wantOK:  true,
			wantKey: "group_id",
			wantVal: "g3",
		},
		{
			name:   "non object",
			input:  primitive.A{"not", "an", "object"},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := AsMap(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if !tt.wantOK {
				return
			}
			if got[tt.wantKey] != tt.wantVal {
				t.Fatalf("expected %s=%v, got %v", tt.wantKey, tt.wantVal, got[tt.wantKey])
			}
		})
	}
}

func TestAsSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantOK  bool
		wantLen int
	}{
		{
			name:    "primitive a",
			input:   primitive.A{"a", "b"},
			wantOK:  true,
			wantLen: 2,
		},
		{
			name:    "plain slice",
			input:   []interface{}{"a", "b", "c"},
			wantOK:  true,
			wantLen: 3,
		},
		{
			name:   "non array",
			input:  primitive.D{{Key: "a", Value: "b"}},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := AsSlice(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if !tt.wantOK {
				return
			}
			if len(got) != tt.wantLen {
				t.Fatalf("expected len=%d, got %d", tt.wantLen, len(got))
			}
		})
	}
}

func TestPrimitiveToMap(t *testing.T) {
	got := PrimitiveToMap(primitive.D{
		{Key: "nested", Value: primitive.D{{Key: "name", Value: "alice"}}},
	})

	nested, ok := got["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected nested map, got %T", got["nested"])
	}
	if nested["name"] != "alice" {
		t.Fatalf("expected nested name alice, got %v", nested["name"])
	}
}
