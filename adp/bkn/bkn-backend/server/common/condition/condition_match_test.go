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

func TestNewMatchCond(t *testing.T) {
	Convey("Test NewMatchCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("invalid value_from should return error", func() {
			cfg := &CondCfg{
				Operation: OperationMatch,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Field,
					Value:     "value1",
				},
			}
			cond, err := NewMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "does not support value_from type")
		})

		Convey("AllField should use default fields", func() {
			cfg := &CondCfg{
				Operation: OperationMatch,
				Field:     AllField,
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			matchCond, ok := cond.(*MatchCond)
			So(ok, ShouldBeTrue)
			So(len(matchCond.mFilterFieldNames), ShouldBeGreaterThan, 0)
			So(matchCond.mFilterFieldNames[0], ShouldEqual, "id")
		})

		Convey("specific field should use that field", func() {
			cfg := &CondCfg{
				Operation: OperationMatch,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			matchCond, ok := cond.(*MatchCond)
			So(ok, ShouldBeTrue)
			So(len(matchCond.mFilterFieldNames), ShouldEqual, 1)
			So(matchCond.mFilterFieldNames[0], ShouldContainSubstring, "field1")
		})
	})
}

func TestMatchCond_Convert(t *testing.T) {
	Convey("Test MatchCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("string value should be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationMatch,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			matchCond := cond.(*MatchCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := matchCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "multi_match")
			So(dsl, ShouldContainSubstring, "best_fields")
			So(dsl, ShouldContainSubstring, `"test"`)
		})

		Convey("numeric value should not be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationMatch,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, _ := NewMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			matchCond := cond.(*MatchCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := matchCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "multi_match")
			So(dsl, ShouldContainSubstring, "123")
		})

		Convey("AllField should include multiple fields", func() {
			cfg := &CondCfg{
				Operation: OperationMatch,
				Field:     AllField,
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			matchCond := cond.(*MatchCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := matchCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "multi_match")
			So(dsl, ShouldContainSubstring, "fields")
		})
	})
}

func TestMatchCond_Convert2SQL(t *testing.T) {
	Convey("Test MatchCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("should return empty string", func() {
			cfg := &CondCfg{
				Operation: OperationMatch,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewMatchCond(ctx, cfg, CUSTOM, fieldsMap)
			matchCond := cond.(*MatchCond)

			sql, err := matchCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldBeEmpty)
		})
	})
}
