package tplsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Helper function to create context with user ID
func createTplCtxWithUserID(userID string) context.Context {
	visitor := &rest.Visitor{
		ID: userID,
	}

	return context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029
}

func TestDetail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	templateID := int64(123)
	profile := "Test Profile"
	expectedPo := &dapo.DataAgentTplPo{
		ID:         templateID,
		Name:       "Test Template",
		Key:        "test-template",
		Profile:    &profile,
		ProductKey: "test-product",
		Status:     cdaenum.StatusUnpublished,
		Config:     "{}",
	}

	productPo := &dapo.ProductPo{
		ID:   1,
		Name: "Test Product",
		Key:  "test-product",
	}

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(expectedPo, nil)
	mockProductRepo.EXPECT().GetByKey(gomock.Any(), "test-product").Return(productPo, nil)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		productRepo:  mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.Detail(ctx, templateID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestDetail_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	templateID := int64(999)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(nil, sql.ErrNoRows)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		productRepo:  mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.Detail(ctx, templateID)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "模板不存在")
}

func TestDetail_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	templateID := int64(123)
	expectedErr := errors.New("database error")
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(nil, expectedErr)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		productRepo:  mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.Detail(ctx, templateID)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, expectedErr, err)
}

func TestDetailByKey_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	templateKey := "test-template"
	profile := "Test Profile"
	expectedPo := &dapo.DataAgentTplPo{
		ID:         123,
		Name:       "Test Template",
		Key:        templateKey,
		Profile:    &profile,
		ProductKey: "test-product",
		Status:     cdaenum.StatusUnpublished,
		Config:     "{}",
	}

	productPo := &dapo.ProductPo{
		ID:   1,
		Name: "Test Product",
		Key:  "test-product",
	}

	mockAgentTplRepo.EXPECT().GetByKey(gomock.Any(), templateKey).Return(expectedPo, nil)
	mockProductRepo.EXPECT().GetByKey(gomock.Any(), "test-product").Return(productPo, nil)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		productRepo:  mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.DetailByKey(ctx, templateKey)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestDetailByKey_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	templateKey := "nonexistent-template"
	expectedErr := errors.New("not found")
	mockAgentTplRepo.EXPECT().GetByKey(gomock.Any(), templateKey).Return(nil, expectedErr)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		productRepo:  mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.DetailByKey(ctx, templateKey)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "get template by key")
}

func TestUpdate_Success(t *testing.T) {
	t.Skip("Skipping transaction-heavy test - requires real sql.Tx handling")
}

func TestUpdate_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	templateID := int64(999)
	userID := "user-123"

	profile := "Updated Profile"
	req := &agenttplreq.UpdateReq{
		Name:    "Updated Template",
		Profile: &profile,
		Config:  &daconfvalobj.Config{},
	}

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(nil, sql.ErrNoRows)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
	}

	ctx := createTplCtxWithUserID(userID)
	auditLog, err := svc.Update(ctx, req, templateID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "模板不存在")
	assert.Empty(t, auditLog.ID)
}

func TestUpdate_NameAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	templateID := int64(123)
	userID := "user-123"

	profile := "Updated Profile"
	req := &agenttplreq.UpdateReq{
		Name:    "Existing Template",
		Profile: &profile,
		Config:  &daconfvalobj.Config{},
	}

	oldPo := &dapo.DataAgentTplPo{
		ID:        templateID,
		Name:      "Old Template",
		Key:       "test-template",
		CreatedBy: userID,
		Config:    "{}",
	}

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(oldPo, nil)
	mockAgentTplRepo.EXPECT().ExistsByNameExcludeID(gomock.Any(), req.Name, templateID).Return(true, nil)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
	}

	ctx := createTplCtxWithUserID(userID)
	auditLog, err := svc.Update(ctx, req, templateID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "模板名称已存在")
	assert.Equal(t, "123", auditLog.ID)
}

func TestUpdate_NotOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	templateID := int64(123)
	userID := "user-123"
	ownerID := "user-456"

	profile := "Updated Profile"
	req := &agenttplreq.UpdateReq{
		Name:    "Updated Template",
		Profile: &profile,
		Config:  &daconfvalobj.Config{},
	}

	oldPo := &dapo.DataAgentTplPo{
		ID:        templateID,
		Name:      "Old Template",
		Key:       "test-template",
		CreatedBy: ownerID, // Different user
		Config:    "{}",
	}

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(oldPo, nil)
	mockAgentTplRepo.EXPECT().ExistsByNameExcludeID(gomock.Any(), req.Name, templateID).Return(false, nil)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
	}

	ctx := createTplCtxWithUserID(userID)
	auditLog, err := svc.Update(ctx, req, templateID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无权限更新，非创建人")
	assert.Equal(t, "123", auditLog.ID)
}

func TestUpdate_RepositoryErrorOnExistsCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	templateID := int64(123)
	userID := "user-123"

	profile := "Updated Profile"
	req := &agenttplreq.UpdateReq{
		Name:    "Updated Template",
		Profile: &profile,
		Config:  &daconfvalobj.Config{},
	}

	oldPo := &dapo.DataAgentTplPo{
		ID:        templateID,
		Name:      "Old Template",
		Key:       "test-template",
		CreatedBy: userID,
		Config:    "{}",
	}

	expectedErr := errors.New("database error")

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), templateID).Return(oldPo, nil)
	mockAgentTplRepo.EXPECT().ExistsByNameExcludeID(gomock.Any(), req.Name, templateID).Return(false, expectedErr)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
	}

	ctx := createTplCtxWithUserID(userID)
	auditLog, err := svc.Update(ctx, req, templateID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check template name exists")
	assert.Equal(t, "123", auditLog.ID)
}

func TestUpdate_BeginTxError(t *testing.T) {
	t.Skip("Skipping transaction-heavy test - requires real sql.Tx handling")
}
