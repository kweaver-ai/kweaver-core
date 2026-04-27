package agentinoutsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
)

func buildJSONFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", "application/json")
	part, err := writer.CreatePart(h)
	assert.NoError(t, err)
	_, err = part.Write(content)
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	req := httptest.NewRequest("POST", "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	err = req.ParseMultipartForm(20 << 20)
	assert.NoError(t, err)

	return req.MultipartForm.File["file"][0]
}

func registerCheckAgentAndTplNameValidator(t *testing.T) {
	t.Helper()

	v, ok := binding.Validator.Engine().(*validator.Validate)
	assert.True(t, ok)

	_ = v.RegisterValidation("checkAgentAndTplName", func(fl validator.FieldLevel) bool {
		return true
	})
}

func validImportJSON(t *testing.T) []byte {
	t.Helper()

	exportData := &agentinoutresp.ExportResp{
		Agents: []*agentinoutresp.ExportAgentItem{
			{
				DataAgentPo: &dapo.DataAgentPo{
					Key:        "k-valid",
					Name:       "n-valid",
					Profile:    func() *string { s := "p-valid"; return &s }(),
					AvatarType: cdaenum.AvatarTypeBuiltIn,
					Avatar:     "a-valid",
					ProductKey: "DIP",
					Config: `{
						"input":{"fields":[{"name":"q","type":"string"}]},
						"llms":[{"is_default":true,"llm_config":{"name":"m1","model_type":"llm","max_tokens":100}}],
						"output":{"default_format":"markdown"}
					}`,
				},
			},
		},
	}
	bys, err := json.Marshal(exportData)
	assert.NoError(t, err)

	return bys
}

func TestAgentInOutSvc_Import_MoreBranches(t *testing.T) {
	t.Run("open file failed", func(t *testing.T) {
		svc := &agentInOutSvc{SvcBase: service.NewSvcBase()}
		req := agentinoutreq.NewImportReq()
		req.ImportType = agentinoutreq.ImportTypeCreate
		req.File = &multipart.FileHeader{
			Size: 1,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
		}

		ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: "u1"}) //nolint:staticcheck // SA1029
		resp, err := svc.Import(ctx, req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, err.Error(), "无法打开上传文件")
	})

	t.Run("invalid json", func(t *testing.T) {
		svc := &agentInOutSvc{SvcBase: service.NewSvcBase()}
		req := agentinoutreq.NewImportReq()
		req.ImportType = agentinoutreq.ImportTypeCreate
		req.File = buildJSONFileHeader(t, "bad.json", []byte("{bad-json"))

		ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: "u1"}) //nolint:staticcheck // SA1029
		resp, err := svc.Import(ctx, req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, err.Error(), "文件格式错误，无法解析JSON")
	})

	t.Run("no agents in file", func(t *testing.T) {
		svc := &agentInOutSvc{SvcBase: service.NewSvcBase()}
		req := agentinoutreq.NewImportReq()
		req.ImportType = agentinoutreq.ImportTypeCreate

		bys, _ := json.Marshal(&agentinoutresp.ExportResp{Agents: []*agentinoutresp.ExportAgentItem{}})
		req.File = buildJSONFileHeader(t, "empty.json", bys)

		ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: "u1"}) //nolint:staticcheck // SA1029
		resp, err := svc.Import(ctx, req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, err.Error(), "导入文件中没有agent数据")
	})

	t.Run("max size exceeded", func(t *testing.T) {
		svc := &agentInOutSvc{SvcBase: service.NewSvcBase()}
		req := agentinoutreq.NewImportReq()
		req.ImportType = agentinoutreq.ImportTypeCreate

		agents := make([]*agentinoutresp.ExportAgentItem, 0, daconstant.AgentInoutMaxSize+1)
		for i := 0; i < daconstant.AgentInoutMaxSize+1; i++ {
			agents = append(agents, &agentinoutresp.ExportAgentItem{
				DataAgentPo: &dapo.DataAgentPo{Key: "k", Name: "n", Config: "{}"},
			})
		}

		bys, _ := json.Marshal(&agentinoutresp.ExportResp{Agents: agents})
		req.File = buildJSONFileHeader(t, "many.json", bys)

		ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: "u1"}) //nolint:staticcheck // SA1029
		resp, err := svc.Import(ctx, req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, err.Error(), "单次导入最多导入")
	})

	t.Run("import check fail returns resp", func(t *testing.T) {
		svc := &agentInOutSvc{
			SvcBase: service.NewSvcBase(),
		}
		req := agentinoutreq.NewImportReq()
		req.ImportType = agentinoutreq.ImportTypeCreate

		bys, _ := json.Marshal(&agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{
					DataAgentPo: &dapo.DataAgentPo{
						Key:    "k1",
						Name:   "n1",
						Config: "{bad-json",
					},
				},
			},
		})
		req.File = buildJSONFileHeader(t, "invalid_config.json", bys)

		ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: "u1"}) //nolint:staticcheck // SA1029
		resp, err := svc.Import(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.IsSuccess)
		assert.NotEmpty(t, resp.ConfigInvalid)
	})

	t.Run("import type create branch returns repo error", func(t *testing.T) {
		registerCheckAgentAndTplNameValidator(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		svc := &agentInOutSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
			logger:        noopAgentLogger{},
		}
		req := agentinoutreq.NewImportReq()
		req.ImportType = agentinoutreq.ImportTypeCreate
		req.File = buildJSONFileHeader(t, "ok.json", validImportJSON(t))

		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k-valid"}).Return([]*dapo.DataAgentPo{}, nil)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, assert.AnError)

		ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: "u1"}) //nolint:staticcheck // SA1029
		resp, err := svc.Import(ctx, req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.IsSuccess)
	})

	t.Run("import type upsert branch returns repo error", func(t *testing.T) {
		registerCheckAgentAndTplNameValidator(t)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &agentInOutSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
			bizDomainHttp: mockBiz,
			logger:        noopAgentLogger{},
		}
		req := agentinoutreq.NewImportReq()
		req.ImportType = agentinoutreq.ImportTypeUpsert
		req.File = buildJSONFileHeader(t, "ok.json", validImportJSON(t))

		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{""}).Return([]string{}, map[string]string{}, nil)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k-valid"}).Return([]*dapo.DataAgentPo{}, nil).Times(2)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, assert.AnError)

		ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: "u1"}) //nolint:staticcheck // SA1029
		resp, err := svc.Import(ctx, req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.IsSuccess)
	})

	t.Run("unknown import type returns success after checks", func(t *testing.T) {
		registerCheckAgentAndTplNameValidator(t)

		svc := &agentInOutSvc{
			SvcBase: service.NewSvcBase(),
		}
		req := agentinoutreq.NewImportReq()
		req.ImportType = "unknown"
		req.File = buildJSONFileHeader(t, "ok.json", validImportJSON(t))

		ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: "u1"}) //nolint:staticcheck // SA1029
		resp, err := svc.Import(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.IsSuccess)
	})
}

func TestAgentInOutSvc_checkAgentConfigValid_And_importCheck(t *testing.T) {
	t.Run("invalid config is recorded without panic", func(t *testing.T) {
		svc := &agentInOutSvc{SvcBase: service.NewSvcBase()}
		resp := agentinoutresp.NewImportResp()
		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{
					DataAgentPo: &dapo.DataAgentPo{
						Key:        "k1",
						Name:       "n1",
						Profile:    func() *string { s := "p1"; return &s }(),
						AvatarType: cdaenum.AvatarTypeBuiltIn,
						Avatar:     "a1",
						ProductKey: "DIP",
						Config:     "{bad-json",
					},
				},
			},
		}

		assert.NotPanics(t, func() {
			svc.checkAgentConfigValid(context.Background(), exportData, resp)
		})
		assert.Len(t, resp.ConfigInvalid, 1)
	})

	t.Run("importCheck permission error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		svc := &agentInOutSvc{
			SvcBase: service.NewSvcBase(),
			pmsSvc:  mockPms,
		}
		isSystem := cenum.YesNoInt8Yes
		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{
					DataAgentPo: &dapo.DataAgentPo{
						Key:           "k1",
						Name:          "n1",
						Config:        "{bad-json",
						IsSystemAgent: &isSystem,
					},
				},
			},
		}
		resp := agentinoutresp.NewImportResp()

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentCreateSystemAgent).
			Return(false, assert.AnError)

		err := svc.importCheck(context.Background(), exportData, resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "check system agent create permission failed")
	})
}

func TestAgentInOutSvc_importByCreateCheck_More(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	svc := &agentInOutSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentRepo,
	}

	t.Run("repo error", func(t *testing.T) {
		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{DataAgentPo: &dapo.DataAgentPo{Key: "k1", Name: "n1"}},
			},
		}
		resp := agentinoutresp.NewImportResp()

		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).Return(nil, assert.AnError)

		err := svc.importByCreateCheck(context.Background(), exportData, resp)
		assert.Error(t, err)
	})

	t.Run("conflict adds fail item", func(t *testing.T) {
		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{DataAgentPo: &dapo.DataAgentPo{Key: "k1", Name: "n1"}},
			},
		}
		resp := agentinoutresp.NewImportResp()

		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).
			Return([]*dapo.DataAgentPo{{Key: "k1", Name: "old"}}, nil)

		err := svc.importByCreateCheck(context.Background(), exportData, resp)
		assert.NoError(t, err)
		assert.True(t, resp.HasFail())
		assert.Len(t, resp.AgentKeyConflict, 1)
	})
}

func TestAgentInOutSvc_importByUpsertCheck_More(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	svc := &agentInOutSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentRepo,
		bizDomainHttp: mockBiz,
	}
	ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029

	t.Run("biz domain conflict causes early return", func(t *testing.T) {
		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{DataAgentPo: &dapo.DataAgentPo{Key: "k1", Name: "n1"}},
			},
		}
		resp := agentinoutresp.NewImportResp()

		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-1"}).
			Return([]string{}, map[string]string{}, nil)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).
			Return([]*dapo.DataAgentPo{{ID: "a1", Key: "k1", Name: "n1"}}, nil)

		existingMap, err := svc.importByUpsertCheck(ctx, exportData, "u1", resp)
		assert.NoError(t, err)
		assert.Nil(t, existingMap)
		assert.True(t, resp.HasFail())
		assert.NotEmpty(t, resp.BizDomainConflict)
	})

	t.Run("upsert check repo error", func(t *testing.T) {
		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{DataAgentPo: &dapo.DataAgentPo{Key: "k2", Name: "n2"}},
			},
		}
		resp := agentinoutresp.NewImportResp()

		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-1"}).
			Return([]string{}, map[string]string{}, nil)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k2"}).
			Return([]*dapo.DataAgentPo{}, nil)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k2"}).
			Return(nil, assert.AnError)

		existingMap, err := svc.importByUpsertCheck(ctx, exportData, "u1", resp)
		assert.Error(t, err)
		assert.Nil(t, existingMap)
	})
}

func TestAgentInOutSvc_Export_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	svc := &agentInOutSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}
	req := &agentinoutreq.ExportReq{AgentIDs: []string{"a1"}}

	mockAgentConfRepo.EXPECT().GetByIDsAndCreatedBy(gomock.Any(), []string{"a1"}, "u1").
		Return([]*dapo.DataAgentPo{
			{
				ID:     "a1",
				Key:    "k1",
				Name:   "n1",
				Config: "{}",
			},
		}, nil)

	ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: "u1"}) //nolint:staticcheck // SA1029
	resp, filename, err := svc.Export(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Agents, 1)
	assert.True(t, strings.HasPrefix(filename, "agent_export_"))
}
