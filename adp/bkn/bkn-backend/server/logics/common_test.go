// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func Test_BuildDslQuery(t *testing.T) {
	Convey("Test BuildDslQuery\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid query string and query\n", func() {
			queryStr := `{"match_all": {}}`
			query := &interfaces.ConceptsQuery{
				Limit: 10,
				Sort: []*interfaces.SortParams{
					{
						Field:     "name",
						Direction: "asc",
					},
				},
			}
			dsl, err := BuildDslQuery(ctx, queryStr, query)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeNil)
			So(dsl["size"], ShouldEqual, 10)
			So(dsl["track_scores"], ShouldEqual, true)
			sort, ok := dsl["sort"].([]map[string]any)
			So(ok, ShouldBeTrue)
			So(len(sort), ShouldEqual, 1)
			So(sort[0]["name"], ShouldEqual, "asc")
		})

		Convey("Success with search_after parameter\n", func() {
			queryStr := `{"match_all": {}}`
			query := &interfaces.ConceptsQuery{
				Limit: 10,
				SearchAfterParams: interfaces.SearchAfterParams{
					SearchAfter: []any{"value1", 123},
				},
			}
			dsl, err := BuildDslQuery(ctx, queryStr, query)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeNil)
			So(dsl["search_after"], ShouldNotBeNil)
			So(query.NeedTotal, ShouldBeFalse)
		})

		Convey("Success with search_after and zero limit sets default limit\n", func() {
			queryStr := `{"match_all": {}}`
			query := &interfaces.ConceptsQuery{
				Limit: 0,
				SearchAfterParams: interfaces.SearchAfterParams{
					SearchAfter: []any{"value1", 123},
				},
			}
			dsl, err := BuildDslQuery(ctx, queryStr, query)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeNil)
			So(query.Limit, ShouldEqual, interfaces.SearchAfter_Limit)
		})

		Convey("Success with empty sort\n", func() {
			queryStr := `{"match_all": {}}`
			query := &interfaces.ConceptsQuery{
				Limit: 10,
				Sort:  []*interfaces.SortParams{},
			}
			dsl, err := BuildDslQuery(ctx, queryStr, query)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeNil)
			sort, ok := dsl["sort"].([]map[string]any)
			So(ok, ShouldBeTrue)
			So(len(sort), ShouldEqual, 0)
		})

		Convey("Success with multiple sort params\n", func() {
			queryStr := `{"match_all": {}}`
			query := &interfaces.ConceptsQuery{
				Limit: 20,
				Sort: []*interfaces.SortParams{
					{
						Field:     "name",
						Direction: "asc",
					},
					{
						Field:     "create_time",
						Direction: "desc",
					},
				},
			}
			dsl, err := BuildDslQuery(ctx, queryStr, query)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeNil)
			sort, ok := dsl["sort"].([]map[string]any)
			So(ok, ShouldBeTrue)
			So(len(sort), ShouldEqual, 2)
			So(sort[0]["name"], ShouldEqual, "asc")
			So(sort[1]["create_time"], ShouldEqual, "desc")
		})

		Convey("Failed with invalid JSON query string\n", func() {
			queryStr := `{"match_all": {invalid json}`
			query := &interfaces.ConceptsQuery{
				Limit: 10,
			}
			dsl, err := BuildDslQuery(ctx, queryStr, query)
			So(err, ShouldNotBeNil)
			So(dsl, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_InternalError_UnMarshalDataFailed)
		})

		Convey("Success with complex query string\n", func() {
			queryStr := `{"bool": {"must": [{"term": {"status": "active"}}]}}`
			query := &interfaces.ConceptsQuery{
				Limit: 15,
				Sort: []*interfaces.SortParams{
					{
						Field:     "id",
						Direction: "asc",
					},
				},
			}
			dsl, err := BuildDslQuery(ctx, queryStr, query)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeNil)
			queryMap, ok := dsl["query"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(queryMap, ShouldNotBeNil)
		})
	})
}
