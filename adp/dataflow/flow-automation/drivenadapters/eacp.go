// Package drivenadapters 当前微服务依赖的其他服务
package drivenadapters

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/eacp.go -destination ../tests/mock_drivenadapters/eacp_mock.go

const (
	// ErrorCode403001011 forbidden
	ErrorCode403001011 = float64(403001011)
)

// DocumentVisitor 请求信息
type DocumentVisitor struct {
	TokenID string
	IP      string
}

// PermInfo 权限信息
type PermInfo struct {
	Allow        []string `json:"allow"`
	Deny         []string `json:"deny"`
	AccessorID   string   `json:"accessorid"`
	AccessorType string   `json:"accessortype"`
	Endtime      int64    `json:"endtime"`
	InheritDocID string   `json:"inheritdocid,omitempty"`
}

func (pi *PermInfo) ToPermConfig() (*PermConfig, error) {
	var expiresAt string
	var err error
	if pi.Endtime == -1 {
		expiresAt = "1970-01-01T08:00:00+08:00"
	} else {
		expiresAt, err = utils.TimestampToRFC3339(pi.Endtime, "us", false, "Asia/Shanghai")
		if err != nil {
			return nil, err
		}
	}
	return &PermConfig{
		Accessor: Accessor{
			ID:   pi.AccessorID,
			Type: pi.AccessorType,
		},
		Allow:     pi.Allow,
		Deny:      pi.Deny,
		ExpiresAt: expiresAt,
		EndedTime: pi.Endtime,
	}, nil
}

// ShareConfig 共享配置
type ShareConfig struct {
	EnableUserDocInnerLinkShare bool `json:"enableUserDocInnerLinkShare"`
	EnableUserDocOutLinkShare   bool `json:"enableUserDocOutLinkShare"`
}

// Eacp eacp服务处理接口
type Eacp interface {

	// GetUserInfo 检查用户是否登录
	GetUserInfo(token string) (string, string, error)

	// CheckUserByID 检查userid对应的用户是否存在
	CheckUserByID(userID, token string) (bool, error)

	// SetPerm 设置权限
	SetPerm(docid string, perminfos []PermInfo, token string) (float64, error)

	// GetContactorUsers 获取联系人组用户
	GetContactorUsers(contactorID, token string) ([]string, error)

	// GetShareConfig 获取共享配置
	GetShareConfig(token string) (ShareConfig, error)

	// CheckOwner 检查所有者权限
	CheckOwner(docID, token string) (bool, error)

	// CheckPerm 检查指定docGns对象是否有指定操作权限
	CheckPerm(docID, action, token string) (float64, error)
}

var (
	eacpOnce sync.Once
	eacp     Eacp
)

type eacpSvc struct {
	baseURL    string
	log        commonLog.Logger
	httpClient HTTPClient
}

// NewEacp 创建eacp服务处理对象
func NewEacp() Eacp {
	eacpOnce.Do(func() {
		eacp = &eacpSvc{
			baseURL:    fmt.Sprintf("http://%s:%s/api/eacp", os.Getenv("EacpHost"), os.Getenv("EacpPort")),
			log:        commonLog.NewLogger(),
			httpClient: NewHTTPClient(),
		}
	})

	return eacp
}

// GetUserInfo 获取用户信息
func (e *eacpSvc) GetUserInfo(token string) (string, string, error) { //nolint
	target := fmt.Sprintf("%v/v1/user/get", e.baseURL)
	respParam, err := e.httpClient.Get(target, map[string]string{"Authorization": token})
	if err != nil {
		e.log.Errorf("GetUserInfo failed: %v, url: %v", err, target)
		return "", "", err
	}
	userID := respParam.(map[string]interface{})["userid"].(string)
	name := respParam.(map[string]interface{})["name"].(string)
	return userID, name, nil
}

// CheckUserByID 检查用户是否存在
func (e *eacpSvc) CheckUserByID(userID, token string) (bool, error) {
	target := fmt.Sprintf("%v/v1/user/getbasicinfo", e.baseURL)
	body := map[string]string{"userid": userID}
	_, _, err := e.httpClient.Post(target, map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, body)

	if err == nil {
		return true, nil
	}

	httpError, ok := err.(errors.ExHTTPError)
	var httpErrorBody map[string]interface{}

	if !ok {
		return false, err
	}

	parseErr := json.Unmarshal([]byte(httpError.Body), &httpErrorBody)

	if parseErr != nil {
		return false, err
	}

	if httpErrorBody["code"] != ErrorCode403001011 {
		e.log.Errorf("CheckUserByID failed: %v, url: %v", err, target)
		return false, err
	}

	return false, nil
}

// SetPerm 设置权限
func (e *eacpSvc) SetPerm(docID string, perminfos []PermInfo, token string) (result float64, err error) {
	target := fmt.Sprintf("%v/v1/perm2/set", e.baseURL)
	body := map[string]interface{}{"docid": docID, "perminfos": perminfos, "inherit": true}
	_, respParam, err := e.httpClient.Post(target, map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		e.log.Errorf("SetPerm failed: %v, url: %v", err, target)
		return
	}
	result = respParam.(map[string]interface{})["result"].(float64)
	return
}

// GetContactorUsers 获取联系人组用户
func (e *eacpSvc) GetContactorUsers(contactorID, token string) (users []string, err error) {
	target := fmt.Sprintf("%v/v1/contactor/getpersons", e.baseURL)
	body := map[string]interface{}{"groupid": contactorID, "start": 0, "limit": -1}
	_, respParam, err := e.httpClient.Post(target, map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		e.log.Errorf("GetContactorUsers failed: %v, url: %v", err, target)
		return
	}
	userInfos := respParam.(map[string]interface{})["userinfos"].([]interface{})
	for _, userInfo := range userInfos {
		userID := userInfo.(map[string]interface{})["userid"].(string)
		users = append(users, userID)
	}
	return
}

// GetShareConfig 获取共享配置
func (e *eacpSvc) GetShareConfig(token string) (config ShareConfig, err error) {
	target := fmt.Sprintf("%v/v1/perm1/getsharedocconfig", e.baseURL)
	body := map[string]interface{}{}
	_, respParam, err := e.httpClient.Post(target, map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		e.log.Errorf("GetShareConfig failed: %v, url: %v", err, target)
		return
	}
	inner := respParam.(map[string]interface{})["enable_user_doc_inner_link_share"].(bool)
	out := respParam.(map[string]interface{})["enable_user_doc_inner_link_share"].(bool)

	config.EnableUserDocInnerLinkShare = inner
	config.EnableUserDocOutLinkShare = out
	return
}

// CheckOwner 检查所有者权限
func (e *eacpSvc) CheckOwner(docID, token string) (isowner bool, err error) {
	target := fmt.Sprintf("%v/v1/owner/check", e.baseURL)
	body := map[string]interface{}{"docid": docID}
	_, respParam, err := e.httpClient.Post(target, map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		e.log.Errorf("CheckOwner failed: %v, url: %v", err, target)
		return
	}
	isowner = respParam.(map[string]interface{})["isowner"].(bool)
	return
}

// CheckPerm 检查指定docGns对象是否有指定操作权限
func (e *eacpSvc) CheckPerm(docID, action, token string) (float64, error) {
	var hasPerm float64
	target := fmt.Sprintf("%v/v1/perm1/check", e.baseURL)
	body := map[string]interface{}{"perm": action, "docid": docID}
	_, respParam, err := e.httpClient.Post(target, map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		e.log.Errorf("CheckOwner failed: %v, url: %v", err, target)
		return hasPerm, err
	}
	return respParam.(map[string]interface{})["result"].(float64), nil
}
