package tplsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDataAgentTplSvc_Copy_SourceNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockBdAgentTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:           service.NewSvcBase(),
		agentTplRepo:      mockAgentTplRepo,
		bizDomainHttp:     mockBizDomainHttp,
		bdAgentTplRelRepo: mockBdAgentTplRelRepo,
	}

	ctx := context.Background()
	templateID := int64(123)

	notFoundErr := errors.New("record not found")
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(nil, notFoundErr)

	resp, auditLogInfo, err := svc.Copy(ctx, templateID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Empty(t, auditLogInfo.Name)
}

func TestDataAgentTplSvc_Copy_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockBdAgentTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:           service.NewSvcBase(),
		agentTplRepo:      mockAgentTplRepo,
		bizDomainHttp:     mockBizDomainHttp,
		bdAgentTplRelRepo: mockBdAgentTplRelRepo,
	}

	ctx := context.Background()
	templateID := int64(123)

	dbErr := errors.New("database connection failed")
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(nil, dbErr)

	resp, auditLogInfo, err := svc.Copy(ctx, templateID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Empty(t, auditLogInfo.Name)
}

func TestDataAgentTplSvc_Copy_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockBdAgentTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:           service.NewSvcBase(),
		agentTplRepo:      mockAgentTplRepo,
		bizDomainHttp:     mockBizDomainHttp,
		bdAgentTplRelRepo: mockBdAgentTplRelRepo,
	}

	ctx := context.Background()
	templateID := int64(123)

	sourcePo := &dapo.DataAgentTplPo{
		ID:   123,
		Name: "Test Template",
		Key:  "test-tpl-key",
	}

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(sourcePo, nil)

	txErr := errors.New("transaction begin failed")
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, txErr)

	resp, auditLogInfo, err := svc.Copy(ctx, templateID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
	assert.Nil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.NotEmpty(t, auditLogInfo.Name)
}

func TestDataAgentTplSvc_Copy_PanicsWithoutAgentTplRepo(t *testing.T) {
	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	templateID := int64(123)

	assert.Panics(t, func() {
		_, _, _ = svc.Copy(ctx, templateID)
	})
}

func TestDataAgentTplSvc_Copy_Success_ClearsPublishedFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockBdAgentTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:           service.NewSvcBase(),
		agentTplRepo:      mockAgentTplRepo,
		bizDomainHttp:     mockBizDomainHttp,
		bdAgentTplRelRepo: mockBdAgentTplRelRepo,
	}

	ctx := createTplCtxWithUserID("test-user-id")
	templateID := int64(123)
	templateName := "Source Template_副本" // 代码会自动添加"_副本"后缀

	// 设置源模板，包含发布信息
	publishedAt := int64(1640995200000) // 2022-01-01 00:00:00
	publishedBy := "original-publisher"
	sourcePo := &dapo.DataAgentTplPo{
		ID:           123,
		Name:         "Source Template",
		Key:          "source-tpl-key",
		ProductKey:   "test-product",
		Status:       cdaenum.StatusPublished,
		PublishedAt:  &publishedAt,
		PublishedBy:  &publishedBy,
		Config:       "{}",
	}

	// Mock 期望
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(sourcePo, nil)
	
	// Mock 事务
	db, mockSql, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	
	mockSql.ExpectBegin()
	mockTx, err := db.Begin()
	require.NoError(t, err)
	
	mockSql.ExpectCommit()
	
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(mockTx, nil)
	
	// 用于捕获传入 Create 方法的参数
	var capturedPo *dapo.DataAgentTplPo
	
	// Mock 创建新模板
	mockAgentTplRepo.EXPECT().Create(gomock.Any(), mockTx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentTplPo) error {
			// 捕获传入的参数用于验证
			capturedPo = po
			return nil
		},
	)
	
	// Mock 创建新模板后的返回
	createdPo := &dapo.DataAgentTplPo{
		ID:      456,
		Name:    templateName,
		Key:     "new-tpl-key", // ULID 生成的 key
		Status:  cdaenum.StatusUnpublished,
		Config:  "{}",
	}
	mockAgentTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), mockTx, gomock.Any()).Return(createdPo, nil)
	
	// Mock 业务域关联
	mockBdAgentTplRelRepo.EXPECT().BatchCreate(gomock.Any(), mockTx, gomock.Any()).Return(nil)
	mockBizDomainHttp.EXPECT().AssociateResource(gomock.Any(), gomock.Any()).Return(nil)

	// 执行测试
	resp, auditLogInfo, err := svc.Copy(ctx, templateID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(456), resp.ID)
	assert.Equal(t, templateName, resp.Name)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Equal(t, "Source Template", auditLogInfo.Name)

	// 验证 Create 被调用时，PublishedAt 和 PublishedBy 被清空
	assert.NotNil(t, capturedPo, "Create 方法应该被调用并捕获参数")
	
	// 验证 PublishedAt 被清空为 0
	assert.Equal(t, int64(0), capturedPo.GetPublishedAtInt64(), "PublishedAt 应该被清空为 0")
	
	// 验证 PublishedBy 被清空为空字符串
	assert.Equal(t, "", capturedPo.GetPublishedByString(), "PublishedBy 应该被清空为空字符串")
	
	// 验证状态为未发布
	assert.Equal(t, cdaenum.StatusUnpublished, capturedPo.Status, "状态应该是未发布")
	
	// 验证其他字段正确设置
	assert.Equal(t, templateName, capturedPo.Name, "模板名称应该正确")
	assert.Equal(t, "test-user-id", capturedPo.CreatedBy, "创建者应该是当前用户")
	assert.Equal(t, "test-user-id", capturedPo.UpdatedBy, "更新者应该是当前用户")
}
