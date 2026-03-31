// Package auth logics 鉴权
package auth

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
)

//go:generate mockgen -package mock_logics -source ../../logics/auth/auth.go -destination ../../tests/mock_logics/auth_mock.go

// CheckAuthRes 鉴权结果信息
type CheckAuthRes struct {
	Status bool   `json:"status"`
	URL    string `json:"url"`
	IP     string `json:"ip"`
}

// AuthHandler method interfaces
type AuthHandler interface {
	RequestToken(ctx context.Context, code, ip string) error
	CheckAuth(userid string) (*CheckAuthRes, error)
	Auth(userid, token, ip string) error
}

var (
	aOnce sync.Once
	a     AuthHandler
)

type auth struct {
	log        commonLog.Logger
	config     *common.Config
	hydra      drivenadapters.HydraPublic
	hydraAdmin drivenadapters.HydraAdmin
	store      mod.Store
}

// NewAuth new auth instance
func NewAuth() AuthHandler {
	aOnce.Do(func() {
		a = &auth{
			log:        commonLog.NewLogger(),
			hydra:      drivenadapters.NewHydraPublic(),
			hydraAdmin: drivenadapters.NewHydraAdmin(),
			config:     common.NewConfig(),
			store:      mod.GetStore(),
		}
	})
	return a
}

func (a *auth) RequestToken(ctx context.Context, code, ip string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	clientID := a.config.OAuth.ClientID
	clientSecret := a.config.OAuth.ClientSecret
	redirecURI := fmt.Sprintf("https://%s:%v/%s", a.config.DeployService.Host, a.config.DeployService.Port, a.config.OAuth.RedirectURI)
	tokenInfo, _, err := a.hydra.RequestTokenWithCode(clientID, clientSecret, code, redirecURI)
	if err != nil {
		return err
	}

	userInfo, err := a.hydraAdmin.Introspect(ctx, tokenInfo.Token)
	if err != nil {
		return err
	}

	oldTokenInfo, err := a.store.GetTokenByUserID(userInfo.UserID)

	if err != nil {
		return err
	}

	if oldTokenInfo.UserID == "" {
		err = a.store.CreateToken(&entity.Token{
			RefreshToken: tokenInfo.RefreshToken,
			Token:        tokenInfo.Token,
			UserID:       userInfo.UserID,
			ExpiresIn:    tokenInfo.ExpiresIn,
			LoginIP:      userInfo.LoginIP,
		})
	} else {
		err = a.store.UpdateToken(&entity.Token{
			RefreshToken: tokenInfo.RefreshToken,
			Token:        tokenInfo.Token,
			UserID:       userInfo.UserID,
			ExpiresIn:    tokenInfo.ExpiresIn,
			LoginIP:      userInfo.LoginIP,
		})
	}

	return err
}

func (a *auth) RefreshToken(userid string) (*entity.Token, error) {
	clientID := a.config.OAuth.ClientID
	clientSecret := a.config.OAuth.ClientSecret
	oldTokenInfo, err := a.store.GetTokenByUserID(userid)

	if err != nil {
		return nil, err
	}

	if oldTokenInfo.UserID == "" {
		return nil, nil
	}

	tokenInfo, _, err := a.hydra.RequestTokenWithRefreshToken(clientID, clientSecret, oldTokenInfo.RefreshToken)

	if err != nil {
		return nil, err
	}

	tokenEntity := &entity.Token{
		RefreshToken: tokenInfo.RefreshToken,
		Token:        tokenInfo.Token,
		UserID:       userid,
		ExpiresIn:    tokenInfo.ExpiresIn,
	}

	err = a.store.UpdateToken(tokenEntity)

	return tokenEntity, err
}

func (a *auth) GetToken(userid string) (*entity.Token, error) {
	tokenInfo, err := a.store.GetTokenByUserID(userid)

	return tokenInfo, err
}

func (a *auth) CheckAuth(userid string) (*CheckAuthRes, error) {
	res := new(CheckAuthRes)
	tokenInfo, err := a.store.GetTokenByUserID(userid)

	if err != nil {
		a.log.Errorf("[CheckAuth] GetTokenByUserID failure, err=%v", err.Error())
		return nil, err
	}

	if tokenInfo.ID != "" && tokenInfo.LoginIP != "" {
		res.Status = true
		res.IP = tokenInfo.LoginIP
	}

	return res, nil
}

func (a *auth) Auth(userid, token, ip string) error {
	oldTokenInfo, err := a.store.GetTokenByUserID(userid)

	if err != nil {
		return err
	}

	tokenEntity := &entity.Token{
		Token:     token,
		UserID:    userid,
		ExpiresIn: 0,
		LoginIP:   ip,
	}

	if oldTokenInfo.UserID == "" {
		err = a.store.CreateToken(tokenEntity)
	} else {
		err = a.store.UpdateToken(tokenEntity)
	}

	return err
}
