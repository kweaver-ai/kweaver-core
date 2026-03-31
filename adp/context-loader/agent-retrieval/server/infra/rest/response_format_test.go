// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package rest

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/toon-format/toon-go"
)

func TestParseResponseFormat(t *testing.T) {
	convey.Convey("ParseResponseFormat", t, func() {
		convey.Convey("empty or json defaults to JSON", func() {
			f, err := ParseResponseFormat("")
			convey.So(err, convey.ShouldBeNil)
			convey.So(f, convey.ShouldEqual, FormatJSON)

			f, err = ParseResponseFormat("json")
			convey.So(err, convey.ShouldBeNil)
			convey.So(f, convey.ShouldEqual, FormatJSON)
		})
		convey.Convey("toon returns TOON", func() {
			f, err := ParseResponseFormat("toon")
			convey.So(err, convey.ShouldBeNil)
			convey.So(f, convey.ShouldEqual, FormatTOON)
		})
		convey.Convey("invalid value returns error", func() {
			_, err := ParseResponseFormat("xml")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid response_format")
		})
	})
}

func TestMarshalResponse_JSON(t *testing.T) {
	convey.Convey("MarshalResponse FormatJSON", t, func() {
		body := map[string]any{"a": 1, "b": "x"}
		ct, b, err := MarshalResponse(FormatJSON, body)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ct, convey.ShouldEqual, ContentTypeJSON)
		convey.So(string(b), convey.ShouldContainSubstring, `"a":1`)
		convey.So(string(b), convey.ShouldContainSubstring, `"b":"x"`)
	})
}

func TestMarshalResponse_TOON_RoundTrip(t *testing.T) {
	convey.Convey("MarshalResponse FormatTOON round-trip", t, func() {
		// 与 PRD 示例一致：同构对象数组
		body := map[string]any{
			"concepts": []map[string]any{
				{"concept_type": "object_type", "concept_id": "ot_1", "concept_name": "公司", "intent_score": 0.95},
				{"concept_type": "object_type", "concept_id": "ot_2", "concept_name": "产品", "intent_score": 0.88},
			},
			"hits_total": float64(2),
		}
		ct, toonBytes, err := MarshalResponse(FormatTOON, body)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ct, convey.ShouldEqual, ContentTypeTOON)
		convey.So(len(toonBytes), convey.ShouldBeGreaterThan, 0)

		// TOON -> Unmarshal -> map
		var decoded map[string]any
		err = toon.Unmarshal(toonBytes, &decoded)
		convey.So(err, convey.ShouldBeNil)

		// 再 Marshal 为 JSON 与原始结构语义一致（比较关键字段）
		jsonBytes, _ := json.Marshal(decoded)
		var decodedFromJSON map[string]any
		err = json.Unmarshal(jsonBytes, &decodedFromJSON)
		convey.So(err, convey.ShouldBeNil)
		convey.So(decodedFromJSON["hits_total"], convey.ShouldEqual, float64(2))
		concepts, ok := decodedFromJSON["concepts"].([]any)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(concepts), convey.ShouldEqual, 2)
		c0, _ := concepts[0].(map[string]any)
		convey.So(c0["concept_name"], convey.ShouldEqual, "公司")
		convey.So(c0["concept_id"], convey.ShouldEqual, "ot_1")
	})
}

func TestMarshalResponse_TOON_Struct(t *testing.T) {
	type Concept struct {
		ConceptType string  `json:"concept_type"`
		ConceptID   string  `json:"concept_id"`
		Score       float64 `json:"intent_score"`
	}
	type Resp struct {
		Concepts  []Concept `json:"concepts"`
		HitsTotal int       `json:"hits_total"`
	}
	convey.Convey("MarshalResponse FormatTOON with struct (json tag only, no toon tag)", t, func() {
		body := &Resp{
			Concepts: []Concept{
				{ConceptType: "object_type", ConceptID: "ot_1", Score: 0.95},
				{ConceptType: "object_type", ConceptID: "ot_2", Score: 0.88},
			},
			HitsTotal: 2,
		}
		ct, toonBytes, err := MarshalResponse(FormatTOON, body)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ct, convey.ShouldEqual, ContentTypeTOON)

		// TOON round-trip: field names should use json tag names, not Go field names
		var decoded map[string]any
		err = toon.Unmarshal(toonBytes, &decoded)
		convey.So(err, convey.ShouldBeNil)
		convey.So(decoded["hits_total"], convey.ShouldEqual, float64(2))
		concepts, ok := decoded["concepts"].([]any)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(concepts), convey.ShouldEqual, 2)
		c0, _ := concepts[0].(map[string]any)
		convey.So(c0["concept_type"], convey.ShouldEqual, "object_type")
		convey.So(c0["concept_id"], convey.ShouldEqual, "ot_1")
		// should NOT have Go field names
		convey.So(c0["ConceptType"], convey.ShouldBeNil)
	})
}

func TestMarshalResponse_TOON_MapSkipsJSONRoundTrip(t *testing.T) {
	convey.Convey("MarshalResponse FormatTOON with map skips JSON roundtrip", t, func() {
		body := map[string]any{"key": "value", "num": float64(42)}
		ct, toonBytes, err := MarshalResponse(FormatTOON, body)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ct, convey.ShouldEqual, ContentTypeTOON)

		var decoded map[string]any
		err = toon.Unmarshal(toonBytes, &decoded)
		convey.So(err, convey.ShouldBeNil)
		convey.So(decoded["key"], convey.ShouldEqual, "value")
		convey.So(decoded["num"], convey.ShouldEqual, float64(42))
	})
}

func TestMarshalResponse_NilBody(t *testing.T) {
	convey.Convey("MarshalResponse nil body", t, func() {
		ct, b, err := MarshalResponse(FormatJSON, nil)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ct, convey.ShouldEqual, ContentTypeJSON)
		convey.So(b, convey.ShouldBeNil)

		ct, b, err = MarshalResponse(FormatTOON, nil)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ct, convey.ShouldEqual, ContentTypeJSON) // nil 时 response_format 未改 contentType 逻辑
		convey.So(b, convey.ShouldBeNil)
	})
}
