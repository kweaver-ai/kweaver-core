package admin

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

type AdminHandler interface {
	CreateAdmin(ctx context.Context, params *AdminReq) error
	ListAdmins(ctx context.Context) ([]AdminInfo, error)
	DeleteAdmin(ctx context.Context, ID string) error
	UpdateAdmin(ctx context.Context, userID, userName string) error
	IsAdmin(ctx context.Context, userID string) (bool, error)
}

var (
	aOnce sync.Once
	ah    AdminHandler
)

type admin struct {
	log      commonLog.Logger
	adminDao rds.ContentAmdinDao
}

// AdminReq 添加管理员请求参数
type AdminReq struct {
	Users []UserInfo `json:"users"`
}

// UserInfo 用户信息
type UserInfo struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// 已添加管理员信息
type AdminInfo struct {
	ID string `json:"id"`
	UserInfo
}

// NewAuth new auth instance
func NewAdmin() AdminHandler {
	aOnce.Do(func() {
		ah = &admin{
			log:      commonLog.NewLogger(),
			adminDao: rds.GetContentAdminDao(),
		}
	})
	return ah
}

// CreateAdmin 创建管理员
func (a *admin) CreateAdmin(ctx context.Context, params *AdminReq) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	var datas []*rds.ContentAdmin
	var userIDs []string
	for _, user := range params.Users {
		userIDs = append(userIDs, user.UserID)
		id, _ := utils.GetUniqueID()
		data := &rds.ContentAdmin{
			ID:       id,
			UserID:   user.UserID,
			UserName: user.UserName,
		}
		datas = append(datas, data)
	}
	admins, err := a.adminDao.ListAdminsByUserID(ctx, userIDs)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[logic.CreateAdmin] ListAdminsByUserID err, detail: %s", err.Error())
		return errors.NewIError(errors.InternalError, "", nil)
	}

	if len(admins) > 0 {
		var existIDs []string
		for _, admin := range admins {
			existIDs = append(existIDs, admin.UserID)
		}
		return errors.NewIError(errors.DuplicatedAdmin, "", map[string]interface{}{"ids": existIDs})
	}

	err = a.adminDao.CreateAdmin(ctx, datas)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[logic.CreateAdmin] CreateAdmin err, detail: %s", err.Error())
		return errors.NewIError(errors.InternalError, "", nil)
	}
	return nil
}

// ListAdmins 列举管理员
func (a *admin) ListAdmins(ctx context.Context) ([]AdminInfo, error) {
	var (
		err       error
		adminList []AdminInfo
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	admins, err := a.adminDao.ListAdmins(ctx)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[logic.ListAdmins] ListAdmins err, detail: %s", err.Error())
		return adminList, errors.NewIError(errors.InternalError, "", nil)
	}

	if len(admins) == 0 {
		adminList = make([]AdminInfo, 0)
	}

	for _, admin := range admins {
		adminList = append(adminList, AdminInfo{
			ID: fmt.Sprintf("%v", admin.ID),
			UserInfo: UserInfo{
				UserID:   admin.UserID,
				UserName: admin.UserName,
			},
		})
	}

	return adminList, nil
}

// DeleteAdmin 根据id删除管理员
func (a *admin) DeleteAdmin(ctx context.Context, ID string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = a.adminDao.DeleteAdminByID(ctx, ID)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[logic.DeleteAdmin] DeleteAdmin err, detail: %s", err.Error())
		return errors.NewIError(errors.InternalError, "", nil)
	}

	return nil
}

func (a *admin) UpdateAdmin(ctx context.Context, userID, userName string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	err = a.adminDao.UpdateAdminByUserID(ctx, userID, userName)
	return err
}

func (a *admin) IsAdmin(ctx context.Context, userID string) (bool, error) {
	admins, err := a.adminDao.ListAdminsByUserID(ctx, []string{userID})
	if err != nil {
		traceLog.WithContext(ctx).Errorf("[logic.IsAdmin] ListAdminsByUserID err, detail: %s", err.Error())
		return false, errors.NewIError(errors.InternalError, "", nil)
	}
	return len(admins) > 0, nil
}
