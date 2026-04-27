// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/opensearch-project/opensearch-go/v2"
	. "github.com/smartystreets/goconvey/convey"

	"ontology-query/common"
)

var (
	testCtx = context.WithValue(context.Background(), rest.XLangKey, rest.DefaultLanguage)
)

// mockTransport 用于模拟 OpenSearch HTTP 响应
type mockTransport struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}, nil
}

func MockNewOpenSearchAccess(appSetting *common.AppSetting, transport *mockTransport) (*openSearchAccess, *mockTransport) {
	if transport == nil {
		transport = &mockTransport{}
	}

	// 创建 OpenSearch 客户端配置，使用自定义 Transport
	cfg := opensearch.Config{
		Addresses: []string{"http://localhost:9200"},
		Transport: transport,
	}

	client, _ := opensearch.NewClient(cfg)

	osa := &openSearchAccess{
		appSetting: appSetting,
		client:     client,
	}

	return osa, transport
}

func Test_openSearchAccess_CreateIndex(t *testing.T) {
	Convey("test CreateIndex\n", t, func() {
		appSetting := &common.AppSetting{}
		osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				So(req.Method, ShouldEqual, "PUT")
				So(req.URL.Path, ShouldContainSubstring, "test-index")

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"acknowledged": true}`)),
				}, nil
			},
		})

		indexBody := map[string]any{
			"settings": map[string]any{
				"number_of_shards": 1,
			},
		}

		Convey("CreateIndex Success \n", func() {
			err := osa.CreateIndex(testCtx, "test-index", indexBody)
			So(err, ShouldBeNil)
		})

		Convey("CreateIndex Failed - marshal error\n", func() {
			invalidBody := make(chan int)
			err := osa.CreateIndex(testCtx, "test-index", invalidBody)
			So(err, ShouldNotBeNil)
		})

		Convey("CreateIndex Failed - HTTP error\n", func() {
			osa2, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			})

			err := osa2.CreateIndex(testCtx, "test-index", indexBody)
			So(err, ShouldNotBeNil)
		})

		Convey("CreateIndex Failed - response error\n", func() {
			osa3, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
					}, nil
				},
			})

			err := osa3.CreateIndex(testCtx, "test-index", indexBody)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_openSearchAccess_IndexExists(t *testing.T) {
	Convey("test IndexExists\n", t, func() {
		appSetting := &common.AppSetting{}

		Convey("IndexExists Success - exists\n", func() {
			osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					So(req.Method, ShouldEqual, "HEAD")
					So(req.URL.Path, ShouldContainSubstring, "test-index")

					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("")),
					}, nil
				},
			})

			exists, err := osa.IndexExists(testCtx, "test-index")
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
		})

		Convey("IndexExists Success - not exists\n", func() {
			osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(strings.NewReader("")),
					}, nil
				},
			})

			exists, err := osa.IndexExists(testCtx, "test-index")
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
		})

		Convey("IndexExists Failed - HTTP error\n", func() {
			osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			})

			exists, err := osa.IndexExists(testCtx, "test-index")
			So(exists, ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("IndexExists Failed - other status code\n", func() {
			osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 500,
						Body:       io.NopCloser(strings.NewReader(`{"error": "internal error"}`)),
					}, nil
				},
			})

			exists, err := osa.IndexExists(testCtx, "test-index")
			So(exists, ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_openSearchAccess_DeleteIndex(t *testing.T) {
	Convey("test DeleteIndex\n", t, func() {
		appSetting := &common.AppSetting{}
		osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				So(req.Method, ShouldEqual, "DELETE")
				So(req.URL.Path, ShouldContainSubstring, "test-index")

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"acknowledged": true}`)),
				}, nil
			},
		})

		Convey("DeleteIndex Success \n", func() {
			err := osa.DeleteIndex(testCtx, "test-index")
			So(err, ShouldBeNil)
		})

		Convey("DeleteIndex Failed - HTTP error\n", func() {
			osa2, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			})

			err := osa2.DeleteIndex(testCtx, "test-index")
			So(err, ShouldNotBeNil)
		})

		Convey("DeleteIndex Failed - response error\n", func() {
			osa3, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
					}, nil
				},
			})

			err := osa3.DeleteIndex(testCtx, "test-index")
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_openSearchAccess_InsertData(t *testing.T) {
	Convey("test InsertData\n", t, func() {
		appSetting := &common.AppSetting{}
		osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				So(req.Method, ShouldEqual, "PUT")
				So(req.URL.Path, ShouldContainSubstring, "test-index")
				So(req.URL.Path, ShouldContainSubstring, "doc1")

				return &http.Response{
					StatusCode: 201,
					Body:       io.NopCloser(strings.NewReader(`{"_id": "doc1", "result": "created"}`)),
				}, nil
			},
		})

		data := map[string]any{
			"title":   "Test Document",
			"content": "This is a test",
		}

		Convey("InsertData Success \n", func() {
			err := osa.InsertData(testCtx, "test-index", "doc1", data)
			So(err, ShouldBeNil)
		})

		Convey("InsertData Failed - marshal error\n", func() {
			invalidData := make(chan int)
			err := osa.InsertData(testCtx, "test-index", "doc1", invalidData)
			So(err, ShouldNotBeNil)
		})

		Convey("InsertData Failed - HTTP error\n", func() {
			osa2, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			})

			err := osa2.InsertData(testCtx, "test-index", "doc1", data)
			So(err, ShouldNotBeNil)
		})

		Convey("InsertData Failed - response error\n", func() {
			osa3, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
					}, nil
				},
			})

			err := osa3.InsertData(testCtx, "test-index", "doc1", data)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_openSearchAccess_BulkInsertData(t *testing.T) {
	Convey("test BulkInsertData\n", t, func() {
		appSetting := &common.AppSetting{}
		osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				So(req.Method, ShouldEqual, "POST")
				So(req.URL.Path, ShouldEqual, "/_bulk")

				bulkResponse := map[string]any{
					"took":   10,
					"errors": false,
					"items":  []any{},
				}
				respBytes, _ := json.Marshal(bulkResponse)

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader(respBytes)),
				}, nil
			},
		})

		dataList := []any{
			map[string]any{
				"_id":   "doc1",
				"title": "Document 1",
			},
			map[string]any{
				"_id":   "doc2",
				"title": "Document 2",
			},
		}

		Convey("BulkInsertData Success \n", func() {
			err := osa.BulkInsertData(testCtx, "test-index", dataList)
			So(err, ShouldBeNil)
		})

		Convey("BulkInsertData Success - empty list\n", func() {
			err := osa.BulkInsertData(testCtx, "test-index", []any{})
			So(err, ShouldBeNil)
		})

		Convey("BulkInsertData Failed - HTTP error\n", func() {
			osa2, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			})

			err := osa2.BulkInsertData(testCtx, "test-index", dataList)
			So(err, ShouldNotBeNil)
		})

		Convey("BulkInsertData Failed - response error\n", func() {
			osa3, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
					}, nil
				},
			})

			err := osa3.BulkInsertData(testCtx, "test-index", dataList)
			So(err, ShouldNotBeNil)
		})

		Convey("BulkInsertData Failed - errors in response\n", func() {
			osa4, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					bulkResponse := map[string]any{
						"took":   10,
						"errors": true,
						"items":  []any{map[string]any{"error": "bulk error"}},
					}
					respBytes, _ := json.Marshal(bulkResponse)

					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader(respBytes)),
					}, nil
				},
			})

			err := osa4.BulkInsertData(testCtx, "test-index", dataList)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_openSearchAccess_SearchData(t *testing.T) {
	Convey("test SearchData\n", t, func() {
		appSetting := &common.AppSetting{}
		osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				So(req.Method, ShouldEqual, "POST")
				So(req.URL.Path, ShouldContainSubstring, "test-index")

				searchResponse := map[string]any{
					"hits": map[string]any{
						"hits": []any{
							map[string]any{
								"_source": map[string]any{
									"title":   "Document 1",
									"content": "Content 1",
								},
								"_score": 1.5,
								"sort":   []any{1},
							},
							map[string]any{
								"_source": map[string]any{
									"title":   "Document 2",
									"content": "Content 2",
								},
								"_score": 1.2,
								"sort":   []any{2},
							},
						},
					},
				}
				respBytes, _ := json.Marshal(searchResponse)

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader(respBytes)),
				}, nil
			},
		})

		query := map[string]any{
			"query": map[string]any{
				"match_all": map[string]any{},
			},
		}

		Convey("SearchData Success \n", func() {
			hits, err := osa.SearchData(testCtx, "test-index", query)
			So(err, ShouldBeNil)
			So(len(hits), ShouldEqual, 2)
			So(hits[0].Source["title"], ShouldEqual, "Document 1")
			So(hits[0].Score, ShouldEqual, 1.5)
		})

		Convey("SearchData Failed - marshal error\n", func() {
			invalidQuery := make(chan int)
			hits, err := osa.SearchData(testCtx, "test-index", invalidQuery)
			So(hits, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("SearchData Failed - HTTP error\n", func() {
			osa2, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			})

			hits, err := osa2.SearchData(testCtx, "test-index", query)
			So(hits, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("SearchData Failed - response error\n", func() {
			osa3, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
					}, nil
				},
			})

			hits, err := osa3.SearchData(testCtx, "test-index", query)
			So(hits, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("SearchData Failed - decode error\n", func() {
			osa4, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`invalid json`)),
					}, nil
				},
			})

			hits, err := osa4.SearchData(testCtx, "test-index", query)
			So(hits, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_openSearchAccess_DeleteData(t *testing.T) {
	Convey("test DeleteData\n", t, func() {
		appSetting := &common.AppSetting{}
		osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				So(req.Method, ShouldEqual, "DELETE")
				So(req.URL.Path, ShouldContainSubstring, "test-index")
				So(req.URL.Path, ShouldContainSubstring, "doc1")

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"_id": "doc1", "result": "deleted"}`)),
				}, nil
			},
		})

		Convey("DeleteData Success \n", func() {
			err := osa.DeleteData(testCtx, "test-index", "doc1")
			So(err, ShouldBeNil)
		})

		Convey("DeleteData Success - not found (404)\n", func() {
			osa2, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(strings.NewReader(`{"found": false}`)),
					}, nil
				},
			})

			err := osa2.DeleteData(testCtx, "test-index", "doc1")
			So(err, ShouldBeNil)
		})

		Convey("DeleteData Failed - HTTP error\n", func() {
			osa3, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			})

			err := osa3.DeleteData(testCtx, "test-index", "doc1")
			So(err, ShouldNotBeNil)
		})

		Convey("DeleteData Failed - other error status\n", func() {
			osa4, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 500,
						Body:       io.NopCloser(strings.NewReader(`{"error": "internal error"}`)),
					}, nil
				},
			})

			err := osa4.DeleteData(testCtx, "test-index", "doc1")
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_openSearchAccess_BulkDeleteData(t *testing.T) {
	Convey("test BulkDeleteData\n", t, func() {
		appSetting := &common.AppSetting{}
		osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				So(req.Method, ShouldEqual, "POST")
				So(req.URL.Path, ShouldEqual, "/_bulk")

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"took": 10, "errors": false}`)),
				}, nil
			},
		})

		docIDs := []string{"doc1", "doc2", "doc3"}

		Convey("BulkDeleteData Success \n", func() {
			err := osa.BulkDeleteData(testCtx, "test-index", docIDs)
			So(err, ShouldBeNil)
		})

		Convey("BulkDeleteData Success - empty list\n", func() {
			err := osa.BulkDeleteData(testCtx, "test-index", []string{})
			So(err, ShouldBeNil)
		})

		Convey("BulkDeleteData Failed - marshal error\n", func() {
			osa2, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					// 读取请求体后返回错误，模拟序列化问题
					return nil, errors.New("marshal error")
				},
			})

			err := osa2.BulkDeleteData(testCtx, "test-index", docIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("BulkDeleteData Failed - HTTP error\n", func() {
			osa3, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			})

			err := osa3.BulkDeleteData(testCtx, "test-index", docIDs)
			So(err, ShouldNotBeNil)
		})

		Convey("BulkDeleteData Failed - response error\n", func() {
			osa4, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
					}, nil
				},
			})

			err := osa4.BulkDeleteData(testCtx, "test-index", docIDs)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_openSearchAccess_Count(t *testing.T) {
	Convey("test Count\n", t, func() {
		appSetting := &common.AppSetting{}
		osa, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				So(req.Method, ShouldEqual, "POST")
				So(req.URL.Path, ShouldContainSubstring, "test-index")
				So(req.URL.Path, ShouldContainSubstring, "_count")

				countResponse := map[string]any{
					"count": 100,
				}
				respBytes, _ := json.Marshal(countResponse)

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewReader(respBytes)),
				}, nil
			},
		})

		query := map[string]any{
			"query": map[string]any{
				"match_all": map[string]any{},
			},
		}

		Convey("Count Success \n", func() {
			result, err := osa.Count(testCtx, "test-index", query)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)

			var countResp map[string]any
			err = json.Unmarshal(result, &countResp)
			So(err, ShouldBeNil)
			So(countResp["count"], ShouldEqual, float64(100))
		})

		Convey("Count Failed - marshal error\n", func() {
			invalidQuery := make(chan int)
			result, err := osa.Count(testCtx, "test-index", invalidQuery)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("Count Failed - HTTP error\n", func() {
			osa2, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				},
			})

			result, err := osa2.Count(testCtx, "test-index", query)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("Count Failed - response error\n", func() {
			osa3, _ := MockNewOpenSearchAccess(appSetting, &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
					}, nil
				},
			})

			result, err := osa3.Count(testCtx, "test-index", query)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}
