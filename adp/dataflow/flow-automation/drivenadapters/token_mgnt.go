package drivenadapters

import (
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	commonLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/token_mgnt.go -destination ../tests/mock_drivenadapters/token_mgnt_mock.go

// tokenMgntMap key[ip:port:clientID:secret]
var tokenMgntMap sync.Map

// AppTokenMgnt   令牌管理器
type AppTokenMgnt interface {
	GetAppToken(tokenIn string) (tokenStr string, expireTime int64, err error)
}

// NewAppTokenMgnt  创建 令牌管理器
func NewAppTokenMgnt(clientID, clientSecret string, useHTTPS bool) AppTokenMgnt {
	if !common.NewConfig().IsAuthEnabled() {
		return &mockAppTokenMgnt{}
	}

	key := fmt.Sprintf("%s:%s", clientID, clientSecret)
	// 先判断  防止每次 NewHydraPublicWithAddress NewLogger
	if tokenMgntPtr, ok := tokenMgntMap.Load(key); ok {
		return tokenMgntPtr.(*tokenMgnt)
	}

	// 创建对象 少量并发 会走到这里
	logger := commonLog.NewLogger()
	hydra := NewHydraPublic()
	tokenMgntPtr, _ := tokenMgntMap.LoadOrStore(key, &tokenMgnt{
		token:        "",
		expireTime:   0,
		updatedAt:    0,
		mutex:        sync.RWMutex{},
		hydra:        hydra,
		clientID:     clientID,
		clientSecret: clientSecret,
		logger:       logger,
	})
	return tokenMgntPtr.(*tokenMgnt)
}

// SecretInfo 密钥信息 结构体
type tokenMgnt struct {
	token        string
	expireTime   int64
	updatedAt    int64
	mutex        sync.RWMutex
	hydra        HydraPublic
	clientID     string
	clientSecret string
	logger       commonLog.Logger
}

func (t *tokenMgnt) GetAppToken(tokenIn string) (tokenStr string, tokenExpireTime int64, err error) {
	// 读锁
	t.mutex.RLock()
	curToken := t.token
	tokenExpireTime = t.expireTime
	t.mutex.RUnlock()

	// 传入字符串 不用每次判断是否过期
	curTime := time.Now().Unix()
	if curToken != tokenIn && curTime < t.updatedAt+t.expireTime-60 {
		return curToken, tokenExpireTime, nil
	}

	tokenStr = ""
	tokenInfo, _, err := t.hydra.RequestTokenWithCredential(t.clientID, t.clientSecret, []string{"all"})
	if err == nil {
		t.mutex.Lock()
		t.token = tokenInfo.Token
		tokenStr = t.token
		t.expireTime = int64(tokenInfo.ExpiresIn)
		tokenExpireTime = t.expireTime
		t.mutex.Unlock()
		t.logger.Infof("****************%s hydra RequestToken success expireIn:%d", t.clientID, tokenInfo.ExpiresIn)
	} else {
		t.logger.Infof("****************%s hydra RequestToken err:%v", t.clientID, err)
	}
	return
}

func is401Error(err error) (err401 bool) { //nolint
	err401 = false
	switch e := err.(type) {
	case errors.ExHTTPError:
		err401 = e.Status == 401
	default:
		err401 = false
	}
	return
}
