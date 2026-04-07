package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	"github.com/toon-format/toon-go"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
)

func TestReplyOK_WithTOONResponseFormatFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	convey.Convey("ReplyOK respects TOON response_format from context", t, func() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

		// 设置 context 中的 response_format 为 TOON
		ctx := common.SetResponseFormatToCtx(req.Context(), FormatTOON)
		req = req.WithContext(ctx)
		c.Request = req

		body := map[string]any{
			"hits_total": float64(1),
			"concepts": []map[string]any{
				{"concept_id": "ot_1"},
			},
		}

		ReplyOK(c, http.StatusOK, body)

		convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
		convey.So(w.Header().Get(ContentTypeKey), convey.ShouldEqual, ContentTypeTOON)
		convey.So(w.Body.Len(), convey.ShouldBeGreaterThan, 0)

		// 验证 TOON 内容可被正常反序列化
		var decoded map[string]any
		err := toon.Unmarshal(w.Body.Bytes(), &decoded)
		convey.So(err, convey.ShouldBeNil)
		convey.So(decoded["hits_total"], convey.ShouldEqual, float64(1))
	})
}
