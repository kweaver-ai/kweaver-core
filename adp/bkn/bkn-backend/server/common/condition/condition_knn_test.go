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

func TestNewKnnCond(t *testing.T) {
	Convey("Test NewKnnCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_VECTOR,
			},
		}

		Convey("invalid value_from should return error", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Field,
					Value:     "value1",
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "does not support value_from type")
		})

		Convey("AllField should use _vector", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     AllField,
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			knnCond, ok := cond.(*KnnCond)
			So(ok, ShouldBeTrue)
			So(knnCond.mFilterFieldName, ShouldEqual, "_vector")
		})

		Convey("specific field should be used", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			knnCond, ok := cond.(*KnnCond)
			So(ok, ShouldBeTrue)
			So(knnCond.mFilterFieldName, ShouldContainSubstring, "field1")
		})

		Convey("sub conditions should be processed", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
				SubConds: []*CondCfg{
					{
						Operation: OperationEq,
						Field:     "field1",
						ValueOptCfg: ValueOptCfg{
							ValueFrom: ValueFrom_Const,
							Value:     "value1",
						},
					},
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			knnCond, ok := cond.(*KnnCond)
			So(ok, ShouldBeTrue)
			So(len(knnCond.mSubConds), ShouldEqual, 1)
		})

		Convey("nil sub condition should be skipped", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
				SubConds: []*CondCfg{
					nil,
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			knnCond, ok := cond.(*KnnCond)
			So(ok, ShouldBeTrue)
			So(len(knnCond.mSubConds), ShouldEqual, 0)
		})
	})
}

func TestKnnCond_Convert(t *testing.T) {
	Convey("Test KnnCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_VECTOR,
			},
		}

		Convey("should create knn query", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			knnCond := cond.(*KnnCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return []*VectorResp{
					{
						Object: "test",
						Vector: []float32{0.1, 0.2, 0.3},
						Index:  0,
					},
				}, nil
			}

			dsl, err := knnCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "knn")
			So(dsl, ShouldContainSubstring, "field1")
			So(dsl, ShouldContainSubstring, "k")
			So(dsl, ShouldContainSubstring, "10")
			So(dsl, ShouldContainSubstring, "vector")
		})

		Convey("vectorizer error should propagate", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			knnCond := cond.(*KnnCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, context.DeadlineExceeded
			}

			dsl, err := knnCond.Convert(ctx, vectorizer)
			So(err, ShouldNotBeNil)
			So(dsl, ShouldBeEmpty)
			So(err.Error(), ShouldContainSubstring, "vectorizer")
		})

		Convey("sub conditions should be included in filter", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
				SubConds: []*CondCfg{
					{
						Operation: OperationEq,
						Field:     "field1",
						ValueOptCfg: ValueOptCfg{
							ValueFrom: ValueFrom_Const,
							Value:     "value1",
						},
					},
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			knnCond := cond.(*KnnCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return []*VectorResp{
					{
						Object: "test",
						Vector: []float32{0.1, 0.2, 0.3},
						Index:  0,
					},
				}, nil
			}

			dsl, err := knnCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "filter")
			So(dsl, ShouldContainSubstring, "bool")
			So(dsl, ShouldContainSubstring, "must")
		})

		Convey("sub condition error should propagate", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
				SubConds: []*CondCfg{
					{
						Operation: OperationKNN,
						Field:     "field1",
						ValueOptCfg: ValueOptCfg{
							ValueFrom: ValueFrom_Const,
							Value:     "value1",
						},
						RemainCfg: map[string]any{
							"limit_key":   "k",
							"limit_value": 10,
						},
					},
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			knnCond := cond.(*KnnCond)

			// Use a vectorizer that returns error for sub condition
			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				// Return error for sub condition's vectorizer call
				if len(words) > 0 && words[0] == "value1" {
					return nil, context.DeadlineExceeded
				}
				return []*VectorResp{
					{
						Object: "test",
						Vector: []float32{0.1, 0.2, 0.3},
						Index:  0,
					},
				}, nil
			}

			dsl, err := knnCond.Convert(ctx, vectorizer)
			So(err, ShouldNotBeNil)
			So(dsl, ShouldBeEmpty)
		})
	})
}

func TestKnnCond_Convert2SQL(t *testing.T) {
	Convey("Test KnnCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_VECTOR,
			},
		}

		Convey("should return empty string", func() {
			cfg := &CondCfg{
				Operation: OperationKNN,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
				RemainCfg: map[string]any{
					"limit_key":   "k",
					"limit_value": 10,
				},
			}
			cond, err := NewKnnCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			knnCond := cond.(*KnnCond)

			sql, err := knnCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldBeEmpty)
		})
	})
}
