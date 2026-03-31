package drivenadapters

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	osapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	osutil "github.com/opensearch-project/opensearch-go/v2/opensearchutil"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestBulkIndexer_Add(t *testing.T) {
	Convey("Test BulkIndexer Add", t, func() {
		client, err := opensearch.NewClient(opensearch.Config{})
		if err != nil {
			t.Fatalf("Failed to create opensearch client: %v", err)
		}

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		indexer := &BulkIndexer{
			Client: client,
			Buf:    bytes.NewBuffer(make([]byte, 0, 5*1024*1024)),
			Aux:    make([]byte, 0, 512),
		}

		root, _ := sonic.Get([]byte("{\"test\":true}"))
		raw, _ := root.Raw()
		RequireAlias := true
		item := &osutil.BulkIndexerItem{
			Index:        "ar-test",
			Action:       "index",
			DocumentID:   "ar-test",
			Body:         strings.NewReader(raw),
			RequireAlias: &RequireAlias,
		}

		Convey("WriteMeta failed", func() {
			patches := ApplyMethodReturn(indexer, "WriteMeta", errors.New("error"))
			defer patches.Reset()

			err := indexer.Add(item)
			So(err, ShouldNotBeNil)
		})

		Convey("WriteBody failed", func() {
			patches := ApplyMethodReturn(indexer, "WriteBody", errors.New("error"))
			defer patches.Reset()

			err := indexer.Add(item)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			err := indexer.Add(item)
			So(err, ShouldBeNil)
		})
	})
}

func TestBulkIndexer_WriteMeta(t *testing.T) {
	Convey("Test BulkIndexer WriteMeta", t, func() {
		client, err := opensearch.NewClient(opensearch.Config{})
		if err != nil {
			t.Fatalf("Failed to create opensearch client: %v", err)
		}

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		indexer := &BulkIndexer{
			Client: client,
			Buf:    bytes.NewBuffer(make([]byte, 0, 5*1024*1024)),
			Aux:    make([]byte, 0, 512),
		}

		root, _ := sonic.Get([]byte("{\"test\":true}"))
		raw, _ := root.Raw()
		RequireAlias := true
		item := &osutil.BulkIndexerItem{
			Index:        "ar-test",
			Action:       "index",
			DocumentID:   "ar-test",
			Body:         strings.NewReader(raw),
			RequireAlias: &RequireAlias,
		}

		Convey("sonic.Marshal failed", func() {
			patches := ApplyFuncReturn(sonic.Marshal, nil, errors.New("error"))
			defer patches.Reset()

			err := indexer.WriteMeta(item)
			So(err, ShouldNotBeNil)
		})

		Convey("Buf.Write failed", func() {
			patches := ApplyMethodReturn(indexer.Buf, "Write", 0, errors.New("error"))
			defer patches.Reset()

			err := indexer.WriteMeta(item)
			So(err, ShouldNotBeNil)
		})

		Convey("Buf.WriteRune failed", func() {
			patches := ApplyMethodReturn(indexer.Buf, "WriteRune", 0, errors.New("error"))
			defer patches.Reset()

			err := indexer.WriteMeta(item)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			err := indexer.WriteMeta(item)
			So(err, ShouldBeNil)
		})
	})
}

func TestBulkIndexer_WriteBody(t *testing.T) {
	Convey("Test BulkIndexer WriteBody", t, func() {
		client, err := opensearch.NewClient(opensearch.Config{})
		if err != nil {
			t.Fatalf("Failed to create opensearch client: %v", err)
		}

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		indexer := &BulkIndexer{
			Client: client,
			Buf:    bytes.NewBuffer(make([]byte, 0, 5*1024*1024)),
			Aux:    make([]byte, 0, 512),
		}

		root, _ := sonic.Get([]byte("{\"test\":true}"))
		raw, _ := root.Raw()
		RequireAlias := true
		item := &osutil.BulkIndexerItem{
			Index:        "ar-test",
			Action:       "index",
			DocumentID:   "ar-test",
			Body:         strings.NewReader(raw),
			RequireAlias: &RequireAlias,
		}

		Convey("Buf.ReadFrom failed", func() {
			patches := ApplyMethodReturn(indexer.Buf, "ReadFrom", int64(0), errors.New("error"))
			defer patches.Reset()

			err := indexer.WriteBody(item)
			So(err, ShouldNotBeNil)
		})

		Convey("Buf.WriteRune failed", func() {
			patches := ApplyMethodReturn(item.Body, "Seek", int64(0), errors.New("error"))
			defer patches.Reset()

			err := indexer.WriteBody(item)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			err := indexer.WriteMeta(item)
			So(err, ShouldBeNil)
		})
	})
}

func TestBulkIndexer_Flush(t *testing.T) {
	Convey("Test BulkIndexer Flush", t, func() {
		ctx := context.Background()

		client, err := opensearch.NewClient(opensearch.Config{})
		if err != nil {
			t.Fatalf("Failed to create opensearch client:%v", err)
		}

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		indexer := &BulkIndexer{
			Client: client,
			Buf:    bytes.NewBuffer(make([]byte, 0, 5*1024*1024)),
			Aux:    make([]byte, 0, 512),
		}
		indexer.Buf.WriteByte('x')

		Convey("Buf.Len() is 0", func() {
			indexer.Buf.Reset()
			_, err := indexer.Flush(ctx)
			So(err, ShouldBeNil)
		})

		Convey("req.Do failed", func() {
			req := osapi.BulkRequest{}

			patches := ApplyMethodReturn(req, "Do", nil, errors.New("error"))
			defer patches.Reset()

			_, err := indexer.Flush(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("res.IsError() true", func() {
			req := osapi.BulkRequest{}
			res := osapi.Response{
				StatusCode: http.StatusInternalServerError,
			}

			patches := ApplyMethodReturn(req, "Do", &res, nil)
			defer patches.Reset()

			_, err := indexer.Flush(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("sonic Decode failed", func() {
			req := osapi.BulkRequest{}
			res := osapi.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("not a json")),
			}

			patches := ApplyMethodReturn(req, "Do", &res, nil)
			defer patches.Reset()

			_, err := indexer.Flush(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("info.Error.Type == xxx", func() {

			item := osutil.BulkIndexerResponseItem{
				Status: http.StatusInternalServerError,
			}
			item.Error.Type = "document_missing_exception"
			item.Error.Reason = "[_doc][5]: document missing"

			birsp := osutil.BulkIndexerResponse{
				Took:      10,
				HasErrors: true,
				Items: []map[string]osutil.BulkIndexerResponseItem{
					{
						"index": item,
					},
				},
			}
			birspBytes, _ := sonic.Marshal(birsp)

			req := osapi.BulkRequest{}
			res := osapi.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(birspBytes)),
			}

			patches := ApplyMethodReturn(req, "Do", &res, nil)
			defer patches.Reset()

			_, err := indexer.Flush(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {

			item := osutil.BulkIndexerResponseItem{
				Status: http.StatusCreated,
			}

			birsp := osutil.BulkIndexerResponse{
				Took:      10,
				HasErrors: false,
				Items: []map[string]osutil.BulkIndexerResponseItem{
					{
						"index": item,
					},
				},
			}
			birspBytes, _ := sonic.Marshal(birsp)

			req := osapi.BulkRequest{}
			res := osapi.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(birspBytes)),
			}

			patches := ApplyMethodReturn(req, "Do", &res, nil)
			defer patches.Reset()

			_, err := indexer.Flush(ctx)
			So(err, ShouldBeNil)
		})
	})
}
