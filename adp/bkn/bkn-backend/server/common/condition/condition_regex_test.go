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

func TestNewRegexCond(t *testing.T) {
	Convey("Test NewRegexCond", t, func() {
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
			"field3": {
				Name: "field3",
				Type: dtype.DATATYPE_INTEGER,
			},
		}

		Convey("non-string field should return error", func() {
			cfg := &CondCfg{
				Operation: OperationRegex,
				Field:     "field3",
				NameField: fieldsMap["field3"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "not a string field")
		})

		Convey("invalid value_from should return error", func() {
			cfg := &CondCfg{
				Operation: OperationRegex,
				Field:     "field1",
				NameField: fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Field,
					Value:     "value1",
				},
			}
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "does not support value_from type")
		})

		Convey("non-string value should return error", func() {
			cfg := &CondCfg{
				Operation: OperationRegex,
				Field:     "field1",
				NameField: fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "not a string value")
		})

		Convey("invalid regex pattern should return error", func() {
			cfg := &CondCfg{
				Operation: OperationRegex,
				Field:     "field1",
				NameField: fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "[invalid",
				},
			}
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "regular expression error")
		})

		Convey("valid regex pattern should create RegexCond", func() {
			cfg := &CondCfg{
				Operation: OperationRegex,
				Field:     "field1",
				NameField: fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test.*",
				},
			}
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			regexCond, ok := cond.(*RegexCond)
			So(ok, ShouldBeTrue)
			So(regexCond.mValue, ShouldEqual, "test.*")
			So(regexCond.mRegexp, ShouldNotBeNil)
		})

		Convey("text field should work", func() {
			cfg := &CondCfg{
				Operation: OperationRegex,
				Field:     "field2",
				NameField: fieldsMap["field2"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test.*",
				},
			}
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})
	})
}

func TestRegexCond_Convert(t *testing.T) {
	Convey("Test RegexCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("should create regexp query", func() {
			cfg := &CondCfg{
				Operation: OperationRegex,
				Field:     "field1",
				NameField: fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test.*",
				},
			}
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			regexCond := cond.(*RegexCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := regexCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "regexp")
			So(dsl, ShouldContainSubstring, "field1")
			So(dsl, ShouldContainSubstring, "test.*")
		})
	})
}

func TestRegexCond_Convert2SQL(t *testing.T) {
	Convey("Test RegexCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("should return empty string", func() {
			cfg := &CondCfg{
				Operation: OperationRegex,
				Field:     "field1",
				NameField: fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test.*",
				},
			}
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			regexCond := cond.(*RegexCond)

			sql, err := regexCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldBeEmpty)
		})
	})
}
