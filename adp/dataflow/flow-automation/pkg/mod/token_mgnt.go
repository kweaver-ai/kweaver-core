package mod

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	lock "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/lock"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	rds "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/store"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

// tokenMgntMap key[ip:port:clientID:secret]
var tokenMgntMap sync.Map

// TokenMgnt   令牌管理器
type TokenMgnt interface {
	GetUserToken(tokenIn, userid string) (token *entity.Token, err error)
	GetAppToken() (tokenInfo *entity.Token, err error)
}

// NewTokenMgnt  创建 令牌管理器
func NewTokenMgnt(userID string) TokenMgnt {
	key := userID
	// 先判断  防止每次 NewHydraPublicWithAddress NewLogger
	if tokenMgntPtr, ok := tokenMgntMap.Load(key); ok {
		return tokenMgntPtr.(*tokenMgnt)
	}

	// 创建对象 少量并发 会走到这里
	logger := commonLog.NewLogger()
	config := common.NewConfig()
	address := utils.CreateAddress(config.OAuth.PublicHost, config.OAuth.PublicPort)
	hydra := drivenadapters.NewHydraPublicWithAddress(address, false)
	hydraAdmin := drivenadapters.NewHydraAdmin()
	appTokenMgnt := drivenadapters.NewAppTokenMgnt(config.OAuth.ClientID, config.OAuth.ClientSecret, false)
	tokenMgntPtr, _ := tokenMgntMap.LoadOrStore(key, &tokenMgnt{
		token:          "",
		refreshToken:   "",
		expireTime:     0,
		mutex:          sync.RWMutex{},
		hydra:          hydra,
		hydraAdmin:     hydraAdmin,
		ip:             config.OAuth.PublicHost,
		port:           config.OAuth.PublicPort,
		clientID:       config.OAuth.ClientID,
		clientSecret:   config.OAuth.ClientSecret,
		logger:         logger,
		store:          GetStore(),
		config:         common.NewConfig(),
		appTokenMgnt:   appTokenMgnt,
		authentication: drivenadapters.NewAuthentication(),
		rds:            rds.NewRedis(),
	})
	return tokenMgntPtr.(*tokenMgnt)
}

// SecretInfo 密钥信息 结构体
type tokenMgnt struct {
	token          string
	refreshToken   string
	expireTime     int64
	mutex          sync.RWMutex
	hydra          drivenadapters.HydraPublic
	hydraAdmin     drivenadapters.HydraAdmin
	appTokenMgnt   drivenadapters.AppTokenMgnt
	authentication drivenadapters.Authentication
	ip             string
	port           int
	clientID       string
	clientSecret   string
	logger         commonLog.Logger
	store          Store
	config         *common.Config
	rds            rds.RDB
}

func (t *tokenMgnt) GetUserToken(tokenIn, userid string) (tokenInfo *entity.Token, err error) {
	retryMaxTime := 10
	curRetryTime := 0
	rdsLockKey := fmt.Sprintf("flow-automation:refreshtoken_%s", userid)
	lockClient := lock.NewDistributeLock(t.rds, rdsLockKey, userid)
	for {
		if curRetryTime >= retryMaxTime {
			return
		}
		curRetryTime++
		// 传入字符串 不用每次判断是否过期
		curTime := time.Now().Unix()

		oldTokenInfo, err := t.store.GetTokenByUserID(userid)

		if err != nil {
			t.logger.Errorf(err.Error())
		}

		if curTime < oldTokenInfo.UpdatedAt+int64(oldTokenInfo.ExpiresIn)-60 {
			return oldTokenInfo, nil
		}
		err = lockClient.Lock(context.Background(), 60*time.Second)
		if err != nil {
			if !errors.Is(err, rds.ErrLockNotAcquired) {
				t.logger.Errorf("[GetUserToken] lock failed, detail: %s", err.Error())
			}
			time.Sleep(time.Second)
			continue
		}
		tokenInfo, err := t.RefreshToken(userid, oldTokenInfo)
		if err == nil {
			t.logger.Infof("****************%s:%s hydra RequestToken success expireIn:%v", t.ip, t.clientID, tokenInfo.ExpiresIn)
		} else {
			t.logger.Infof("****************%s:%s hydra RequestToken err:%v", t.ip, t.clientID, err)
		}

		if _, rErr := lockClient.Release(); rErr != nil {
			t.logger.Errorf("[GetUserToken] release lock failed, detail: %s", rErr.Error())
		}

		if tokenInfo != nil && tokenInfo.Token != "" {
			return tokenInfo, nil
		}
	}
}

func (t *tokenMgnt) RefreshToken(userid string, oldToken *entity.Token) (*entity.Token, error) {
	clientID := t.config.OAuth.ClientID
	clientSecret := t.config.OAuth.ClientSecret
	tokenID := ""
	expiresIn := 0

	var err error
	// 接口调用时使用传入的应用账户
	if oldToken.IsApp && oldToken.UserID != clientID {
		tokenInfo, tErr := t.hydraAdmin.Introspect(context.Background(), oldToken.Token)
		if tErr != nil {
			t.logger.Errorf("[RefreshToken] Introspect failed, detail: %s", tErr.Error())
			return nil, tErr
		}
		if !tokenInfo.Active {
			_ = t.store.DeleteToken(oldToken.ID)
			return oldToken, fmt.Errorf("token is not active")
		}
		tokenID = oldToken.Token
		expiresIn = int(tokenInfo.ExpiresIn)
	} else {
		// 获取应用账户的token
		appToken, appTokenExpiredTime, err := t.appTokenMgnt.GetAppToken("")
		if err != nil {
			t.logger.Errorf("get app token failed: %s", err.Error())
			return nil, err
		}
		if userid == clientID {
			tokenID = appToken
			expiresIn = int(appTokenExpiredTime)
		} else {
			// 请求断言
			assertion, aErr := t.authentication.GetAssertion(userid, appToken)
			if aErr != nil {
				t.logger.Errorf("get assertion failed: %s", aErr.Error())
				return nil, aErr
			}

			// 以断言获取用户token
			tokenInfo, code, tErr := t.hydra.RequestTokenWithAsserts(clientID, clientSecret, assertion)

			if tErr != nil {
				if code == http.StatusBadRequest {
					_ = t.store.DeleteToken(oldToken.ID)
				}
				return nil, tErr
			}

			tokenID = tokenInfo.Token
			expiresIn = tokenInfo.ExpiresIn
		}
	}

	tokenEntity := &entity.Token{
		Token:     tokenID,
		UserID:    userid,
		ExpiresIn: expiresIn,
	}

	if oldToken.UserID == "" {
		err = t.store.CreateToken(tokenEntity)
	} else {
		err = t.store.UpdateToken(tokenEntity)
	}

	return tokenEntity, err
}

func isAdmin(userid string) bool {
	if userid == common.SystemAuditAdmin ||
		userid == common.SystemOriginSysAdmin ||
		userid == common.SystemSecAdmin ||
		userid == common.SystemSysAdmin {
		return true
	}
	return false
}

func (t *tokenMgnt) GetAppToken() (tokenInfo *entity.Token, err error) {
	tokenStr, expireTime, err := t.appTokenMgnt.GetAppToken("")

	if err != nil {
		return nil, err
	}

	tokenInfo = &entity.Token{
		Token:     tokenStr,
		ExpiresIn: int(expireTime),
	}
	return
}
