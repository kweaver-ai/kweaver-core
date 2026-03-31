package skill

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestSkillHandler(t *testing.T) {
	Convey("SkillHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("RegisterSkill binds multipart form and calls registry", func() {
			mockRegistry := mocks.NewMockSkillRegistry(ctrl)
			mockMarket := mocks.NewMockSkillMarket(ctrl)
			mockReader := mocks.NewMockSkillReader(ctrl)
			handler := &skillHandler{
				Registry: mockRegistry,
				Market:   mockMarket,
				Reader:   mockReader,
			}
			mockRegistry.EXPECT().RegisterSkill(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ any, req *interfaces.RegisterSkillReq) (*interfaces.RegisterSkillResp, error) {
					// So(req.BusinessDomainID, ShouldEqual, "bd-test")
					// So(req.UserID, ShouldEqual, "user-1")
					So(req.FileType, ShouldEqual, "content")
					return &interfaces.RegisterSkillResp{SkillID: "skill-1", Status: "active"}, nil
				},
			)

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			So(writer.WriteField("file_type", "content"), ShouldBeNil)
			filePart, err := writer.CreateFormFile("file", "SKILL.md")
			So(err, ShouldBeNil)
			_, err = filePart.Write([]byte("---\nname: demo\ndescription: desc\n---\nbody"))
			So(err, ShouldBeNil)
			So(writer.Close(), ShouldBeNil)

			recorder := performSkillRequest(http.MethodPost, "/skills", writer.FormDataContentType(), body.String(), map[string]string{
				"x-business-domain": "bd-test",
				"user_id":           "user-1",
			}, handler.RegisterSkill)

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Body.String(), ShouldContainSubstring, `"skill_id":"skill-1"`)
		})

		Convey("RegisterSkill rejects unsupported content type", func() {
			handler := &skillHandler{}
			recorder := performSkillRequest(http.MethodPost, "/skills", "text/plain", "raw", map[string]string{
				"x-business-domain": "bd-test",
				"user_id":           "user-1",
			}, handler.RegisterSkill)

			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetSkillContent binds uri and calls reader", func() {
			mockRegistry := mocks.NewMockSkillRegistry(ctrl)
			mockMarket := mocks.NewMockSkillMarket(ctrl)
			mockReader := mocks.NewMockSkillReader(ctrl)
			handler := &skillHandler{
				Registry: mockRegistry,
				Market:   mockMarket,
				Reader:   mockReader,
			}
			mockReader.EXPECT().GetSkillContent(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ any, req *interfaces.GetSkillContentReq) (*interfaces.GetSkillContentResp, error) {
					// So(req.BusinessDomainID, ShouldEqual, "bd-test")
					So(req.SkillID, ShouldEqual, "skill-2")
					return &interfaces.GetSkillContentResp{SkillID: "skill-2", URL: "https://download/skill-2/SKILL.md"}, nil
				},
			)

			recorder := performSkillRequest(http.MethodGet, "/skills/:skill_id/content", "", "", map[string]string{
				"x-business-domain": "bd-test",
			}, handler.GetSkillContent, "skill-2")

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Body.String(), ShouldContainSubstring, `"skill_content":"guide"`)
		})

		Convey("ReadSkillFile binds body and calls reader", func() {
			mockRegistry := mocks.NewMockSkillRegistry(ctrl)
			mockMarket := mocks.NewMockSkillMarket(ctrl)
			mockReader := mocks.NewMockSkillReader(ctrl)
			handler := &skillHandler{
				Registry: mockRegistry,
				Market:   mockMarket,
				Reader:   mockReader,
			}
			mockReader.EXPECT().ReadSkillFile(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ any, req *interfaces.ReadSkillFileReq) (*interfaces.ReadSkillFileResp, error) {
					// So(req.BusinessDomainID, ShouldEqual, "bd-test")
					So(req.SkillID, ShouldEqual, "skill-3")
					So(req.RelPath, ShouldEqual, "refs/guide.md")
					return &interfaces.ReadSkillFileResp{SkillID: "skill-3", RelPath: "refs/guide.md", URL: "https://download/skill-3/refs/guide.md"}, nil
				},
			)

			recorder := performSkillRequest(http.MethodPost, "/skills/:skill_id/files/read", "application/json", `{"rel_path":"refs/guide.md"}`, map[string]string{
				"x-business-domain": "bd-test",
			}, handler.ReadSkillFile, "skill-3")

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Body.String(), ShouldContainSubstring, `"content":"body"`)
		})

		Convey("DownloadSkill binds uri and returns zip response", func() {
			mockRegistry := mocks.NewMockSkillRegistry(ctrl)
			mockMarket := mocks.NewMockSkillMarket(ctrl)
			mockReader := mocks.NewMockSkillReader(ctrl)
			handler := &skillHandler{
				Registry: mockRegistry,
				Market:   mockMarket,
				Reader:   mockReader,
			}
			mockRegistry.EXPECT().DownloadSkill(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ any, req *interfaces.DownloadSkillReq) (*interfaces.DownloadSkillResp, error) {
					// So(req.BusinessDomainID, ShouldEqual, "bd-test")
					So(req.SkillID, ShouldEqual, "skill-4")
					return &interfaces.DownloadSkillResp{
						SkillID:  "skill-4",
						FileName: "demo-skill.zip",
						Content:  []byte("zip-bytes"),
					}, nil
				},
			)

			recorder := performSkillRequest(http.MethodGet, "/skills/:skill_id/download", "", "", map[string]string{
				"x-business-domain": "bd-test",
			}, handler.DownloadSkill, "skill-4")

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Header().Get("Content-Type"), ShouldEqual, "application/zip")
			So(recorder.Header().Get("Content-Disposition"), ShouldContainSubstring, `filename="demo-skill.zip"`)
			So(recorder.Body.String(), ShouldEqual, "zip-bytes")
		})

		Convey("QuerySkillMarketList binds query and calls market", func() {
			mockRegistry := mocks.NewMockSkillRegistry(ctrl)
			mockMarket := mocks.NewMockSkillMarket(ctrl)
			mockReader := mocks.NewMockSkillReader(ctrl)
			handler := &skillHandler{
				Registry: mockRegistry,
				Market:   mockMarket,
				Reader:   mockReader,
			}
			mockMarket.EXPECT().QuerySkillMarketList(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ any, req *interfaces.QuerySkillMarketListReq) (*interfaces.QuerySkillMarketListResp, error) {
					// So(req.BusinessDomainID, ShouldEqual, "bd-test")
					So(req.Page, ShouldEqual, 2)
					So(req.PageSize, ShouldEqual, 5)
					return &interfaces.QuerySkillMarketListResp{
						CommonPageResult: interfaces.CommonPageResult{
							Page:       2,
							PageSize:   5,
							TotalCount: 1,
						},
						Data: []*interfaces.SkillSummary{
							{SkillID: "skill-market-1", Name: "market-demo"},
						},
					}, nil
				},
			)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Handle(http.MethodGet, "/skills/market", handler.QuerySkillMarketList)
			req := httptest.NewRequest(http.MethodGet, "/skills/market?page=2&page_size=5", nil)
			req.Header.Set("x-business-domain", "bd-test")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Body.String(), ShouldContainSubstring, `"skill_id":"skill-market-1"`)
			So(recorder.Body.String(), ShouldContainSubstring, `"page":2`)
		})

		Convey("GetSkillMarketDetail binds uri and calls market", func() {
			mockRegistry := mocks.NewMockSkillRegistry(ctrl)
			mockMarket := mocks.NewMockSkillMarket(ctrl)
			mockReader := mocks.NewMockSkillReader(ctrl)
			handler := &skillHandler{
				Registry: mockRegistry,
				Market:   mockMarket,
				Reader:   mockReader,
			}
			mockMarket.EXPECT().GetSkillMarketDetail(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ any, req *interfaces.GetSkillMarketDetailReq) (*interfaces.SkillInfo, error) {
					// So(req.BusinessDomainID, ShouldEqual, "bd-test")
					So(req.SkillID, ShouldEqual, "skill-market-2")
					return &interfaces.SkillInfo{
						SkillID:     "skill-market-2",
						Name:        "market-detail",
						Description: "detail-desc",
						Status:      "published",
					}, nil
				},
			)

			recorder := performSkillRequest(http.MethodGet, "/skills/market/:skill_id", "", "", map[string]string{
				"x-business-domain": "bd-test",
			}, handler.GetSkillMarketDetail, "skill-market-2")

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Body.String(), ShouldContainSubstring, `"skill_id":"skill-market-2"`)
			So(recorder.Body.String(), ShouldContainSubstring, `"status":"published"`)
		})
	})
}

func performSkillRequest(method, path, contentType, body string, headers map[string]string, handler func(c *gin.Context), pathParams ...string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Handle(method, path, func(c *gin.Context) {
		for i, param := range pathParams {
			paramName := strings.Split(path, "/")[i+1][1:]
			c.Params = append(c.Params, gin.Param{Key: paramName, Value: param})
		}
		handler(c)
	})

	formattedPath := path
	for _, param := range pathParams {
		start := strings.Index(formattedPath, ":")
		if start == -1 {
			break
		}
		end := strings.Index(formattedPath[start:], "/")
		if end == -1 {
			end = len(formattedPath)
		} else {
			end += start
		}
		formattedPath = formattedPath[:start] + param + formattedPath[end:]
	}

	req := httptest.NewRequest(method, formattedPath, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
