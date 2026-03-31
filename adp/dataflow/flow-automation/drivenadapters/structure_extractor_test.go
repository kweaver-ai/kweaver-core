package drivenadapters

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/go-playground/assert/v2"
// 	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
// 	. "github.com/smartystreets/goconvey/convey"
// )

// func TestFileParse(t *testing.T) {
// 	Convey("TestFileParse", t, func() {
// 		parseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			assert.Equal(t, r.Method, "POST")
// 			result := `{
// 				"results": {
// 					"f1": {
// 						"md_content": "hello",
// 						"content_list": "[]"
// 					}
// 				}
// 			}`
// 			fmt.Fprint(w, result)
// 		}))
// 		defer parseServer.Close()

// 		fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			fmt.Fprint(w, "file data")
// 		}))
// 		defer fileServer.Close()

// 		se := &structureExtractor{
// 			baseURL: parseServer.URL,
// 			config:  &common.StructureExtractor{},
// 			client:  parseServer.Client(),
// 		}

// 		item, items, err := se.FileParse(context.Background(), fileServer.URL, "f.pdf")
// 		assert.Equal(t, err, nil)
// 		assert.Equal(t, item.MdContent, "hello")
// 		assert.Equal(t, len(items), 0)
// 	})
// }
