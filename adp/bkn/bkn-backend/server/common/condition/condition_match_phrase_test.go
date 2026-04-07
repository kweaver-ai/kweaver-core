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

func TestNewMatchPhraseCond(t *testing.T) {
	Convey("Test NewMatchPhraseCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("invalid value_from should return error", func() {
			cfg := &CondCfg{
				Operation: OperationMatchPhrase,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Field,
					Value:     "value1",
				},
			}
			cond, err := NewMatchPhraseCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "does not support value_from type")
		})

		Convey("AllField should use default fields", func() {
			cfg := &CondCfg{
				Operation: OperationMatchPhrase,
				Name:      AllField,
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMatchPhraseCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			matchPhraseCond, ok := cond.(*MatchPhraseCond)
			So(ok, ShouldBeTrue)
			So(len(matchPhraseCond.mFilterFieldNames), ShouldBeGreaterThan, 0)
			So(matchPhraseCond.mFilterFieldNames[0], ShouldEqual, "id")
		})

		Convey("specific field should use that field", func() {
			cfg := &CondCfg{
				Operation: OperationMatchPhrase,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewMatchPhraseCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			matchPhraseCond, ok := cond.(*MatchPhraseCond)
			So(ok, ShouldBeTrue)
			So(len(matchPhraseCond.mFilterFieldNames), ShouldEqual, 1)
			So(matchPhraseCond.mFilterFieldNames[0], ShouldContainSubstring, "field1")
		})
	})
}

func TestMatchPhraseCond_Convert(t *testing.T) {
	Convey("Test MatchPhraseCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("string value should be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationMatchPhrase,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewMatchPhraseCond(ctx, cfg, CUSTOM, fieldsMap)
			matchPhraseCond := cond.(*MatchPhraseCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := matchPhraseCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "multi_match")
			So(dsl, ShouldContainSubstring, "phrase")
			So(dsl, ShouldContainSubstring, `"test"`)
		})

		Convey("AllField should include multiple fields", func() {
			cfg := &CondCfg{
				Operation: OperationMatchPhrase,
				Name:      AllField,
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewMatchPhraseCond(ctx, cfg, CUSTOM, fieldsMap)
			matchPhraseCond := cond.(*MatchPhraseCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := matchPhraseCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "multi_match")
			So(dsl, ShouldContainSubstring, "phrase")
			So(dsl, ShouldContainSubstring, "fields")
		})
	})
}

func TestMatchPhraseCond_Convert2SQL(t *testing.T) {
	Convey("Test MatchPhraseCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("should return empty string", func() {
			cfg := &CondCfg{
				Operation: OperationMatchPhrase,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewMatchPhraseCond(ctx, cfg, CUSTOM, fieldsMap)
			matchPhraseCond := cond.(*MatchPhraseCond)

			sql, err := matchPhraseCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldBeEmpty)
		})
	})
}
