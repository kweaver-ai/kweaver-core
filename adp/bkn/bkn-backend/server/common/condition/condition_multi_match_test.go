// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	dtype "bkn-backend/interfaces/data_type"
)

func TestNewMultiMatchCond(t *testing.T) {
	Convey("Test NewMultiMatchCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
			"field2": {
				Name: "field2",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("non-array fields should return error", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields": "field1",
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "should be an array")
		})

		Convey("non-string element in fields array should return error", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields": []any{"field1", 123},
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "should be a string array")
		})

		Convey("AllField in fields should use default fields", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields": []any{AllField},
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			multiMatchCond, ok := cond.(*MultiMatchCond)
			So(ok, ShouldBeTrue)
			So(len(multiMatchCond.mFilterFieldNames), ShouldBeGreaterThan, 0)
			So(multiMatchCond.mFilterFieldNames[0], ShouldEqual, "id")
		})

		Convey("specific fields should be used", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields": []any{"field1", "field2"},
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			multiMatchCond, ok := cond.(*MultiMatchCond)
			So(ok, ShouldBeTrue)
			So(len(multiMatchCond.mFilterFieldNames), ShouldEqual, 2)
		})

		Convey("empty fields should work", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			multiMatchCond, ok := cond.(*MultiMatchCond)
			So(ok, ShouldBeTrue)
			So(len(multiMatchCond.mFilterFieldNames), ShouldEqual, 0)
		})

		Convey("invalid match_type should return error", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields":     []any{"field1"},
					"match_type": "invalid_type",
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "match_type")
		})

		Convey("non-string match_type should return error", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields":     []any{"field1"},
					"match_type": 123,
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "should be a string")
		})

		Convey("valid match_type should work", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields":     []any{"field1"},
					"match_type": "best_fields",
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})
	})
}

func TestMultiMatchCond_Convert(t *testing.T) {
	Convey("Test MultiMatchCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("string value should be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields": []any{"field1"},
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			multiMatchCond := cond.(*MultiMatchCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := multiMatchCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "multi_match")
			So(dsl, ShouldContainSubstring, "best_fields")
			So(dsl, ShouldContainSubstring, `"test"`)
			So(dsl, ShouldContainSubstring, "fields")
		})

		Convey("numeric value should not be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields": []any{"field1"},
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			multiMatchCond := cond.(*MultiMatchCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := multiMatchCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "multi_match")
			So(dsl, ShouldContainSubstring, "123")
		})

		Convey("empty fields should not include fields in DSL", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			multiMatchCond := cond.(*MultiMatchCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := multiMatchCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "multi_match")
			So(dsl, ShouldNotContainSubstring, `"fields"`)
		})

		Convey("custom match_type should be used", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields":     []any{"field1"},
					"match_type": "phrase",
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			multiMatchCond := cond.(*MultiMatchCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := multiMatchCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "phrase")
		})

		Convey("non-string match_type in Convert should return error", func() {
			// First create a valid cond, then manually set invalid match_type
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields":     []any{"field1"},
					"match_type": "best_fields", // Valid at creation
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			multiMatchCond := cond.(*MultiMatchCond)

			// Now set invalid match_type
			multiMatchCond.mCfg.RemainCfg["match_type"] = 123

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := multiMatchCond.Convert(ctx, vectorizer)
			So(err, ShouldNotBeNil)
			So(dsl, ShouldBeEmpty)
			So(err.Error(), ShouldContainSubstring, "should be a string")
		})
	})
}

func TestMultiMatchCond_Convert2SQL(t *testing.T) {
	Convey("Test MultiMatchCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("should return empty string", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields": []any{"field1"},
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewMultiMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			multiMatchCond := cond.(*MultiMatchCond)

			sql, err := multiMatchCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldBeEmpty)
		})
	})
}
