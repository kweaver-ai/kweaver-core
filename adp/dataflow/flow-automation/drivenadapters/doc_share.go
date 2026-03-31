package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/doc_share.go -destination ../tests/mock_drivenadapters/doc_share_mock.go

// DocShare method interface
type DocShare interface {
	// GetDocConfig 获取文档读取策略配置
	GetDocConfig(ctx context.Context, userID, ip, docID, docLibType string) (map[string]interface{}, error)

	// SetPerm 设置共享权限, 不支持内部共享，外部共享权限设置
	SetPerm(ctx context.Context, docID string, perminfos []PermInfo, inherit bool, token string) (result float64, err error)

	// 设置共享权限, 不支持内部共享，外部共享权限设置
	SetPerm2(ctx context.Context, body *Perm2SetReqBody, token string) (result float64, err error)

	// BatchGetDocPerm 批量获取文档权限
	BatchGetDocPerm(ctx context.Context, method string, subject map[string]string, objects []map[string]string) (map[string]map[string][]string, error)

	// SetDocPerm 批量设置文档权限
	SetDocPerm(ctx context.Context, docID string, configs []map[string]interface{}) (int, error)

	// 检查文档所有者
	CheckOwner(ctx context.Context, docID string, ownerID string) (bool, error)

	// 获取文档全部所有者
	GetDocOwners(ctx context.Context, docID string) (owners []DocOwner, err error)

	// 获取文档权限
	GetPerm(ctx context.Context, docID string, token string) (result GetPermRes, err error)

	// SetDocPerm2 设置权限接口(PUT) 接口config权限信息不传递继承权限相关配置
	SetDocPerm2(ctx context.Context, body *DocPermConfig, token, id string) (respCode int, err error)
}

type docShare struct {
	address        string
	privateAddress string
	httpClient     otelHttp.HTTPClient
}

// DocPermInfo doc perm info struct
type DocPermInfo struct {
	ID    string   `json:"id"`
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
}

var (
	rOnce sync.Once
	r     DocShare
)

// NewDocShare 创建获取用户服务
func NewDocShare() DocShare {
	rOnce.Do(func() {
		config := common.NewConfig()
		r = &docShare{
			address:        fmt.Sprintf("http://%s:%v", config.DocShare.PublicHost, config.DocShare.PublicPort),
			privateAddress: fmt.Sprintf("http://%s:%v", config.DocShare.PrivateHost, config.DocShare.PrivatePort),
			httpClient:     NewOtelHTTPClient(),
		}
	})
	return r
}

// GetDocConfig 获取文档读取策略配置
func (r *docShare) GetDocConfig(ctx context.Context, userID, ip, docID, docLibType string) (map[string]interface{}, error) {
	target := fmt.Sprintf("%s/api/doc-share/v1/doc-config?user_id=%v&ip=%v&doc_id=%v&doc_lib_type=%v&accessed_by=accessed_by_users&read_restriction=download",
		r.privateAddress, userID, ip, docID, docLibType)
	_, respParam, err := r.httpClient.Get(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetDocConfig failed: %v, url: %v", err, target)
		return nil, err
	}

	result := respParam.(map[string]interface{})["result"].(map[string]interface{})

	return result, nil
}

// SetPerm 设置权限
func (r *docShare) SetPerm(ctx context.Context, docID string, perminfos []PermInfo, inherit bool, token string) (result float64, err error) {
	target := fmt.Sprintf("%v/api/eacp/v1/perm2/set", r.address)
	body := map[string]interface{}{"docid": docID, "perminfos": perminfos, "inherit": inherit}
	if !strings.HasPrefix(token, "Bearer") {
		token = fmt.Sprintf("Bearer %s", token)
	}
	_, respParam, err := r.httpClient.Post(ctx, target, map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetPerm failed: %v, url: %v", err, target)
		return
	}
	result = respParam.(map[string]interface{})["result"].(float64)
	return
}

type Perm2SetReqBody struct {
	DocID       string     `json:"docid"`
	PermInfos   []PermInfo `json:"perminfos"`
	Inherit     bool       `json:"inherit"`
	SendMessage bool       `json:"send_message"` // 值为false时将不会发送anyshare消息通知
}

func (r *docShare) SetPerm2(ctx context.Context, body *Perm2SetReqBody, token string) (result float64, err error) {
	target := fmt.Sprintf("%v/api/eacp/v1/perm2/set", r.address)
	if !strings.HasPrefix(token, "Bearer") {
		token = fmt.Sprintf("Bearer %s", token)
	}
	_, respParam, err := r.httpClient.Post(ctx, target, map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetPerm failed: %v, url: %v", err, target)
		return
	}
	result = respParam.(map[string]interface{})["result"].(float64)
	return
}

type GetPermRes struct {
	Perminfos []PermInfo `json:"perminfos"`
	Inherit   bool       `json:"inherit"`
}

func (r *docShare) GetPerm(ctx context.Context, docID string, token string) (result GetPermRes, err error) {
	target := fmt.Sprintf("%v/api/eacp/v1/perm2/get", r.address)

	reqParams := map[string]interface{}{"docid": docID}
	body, _ := json.Marshal(reqParams)

	if !strings.HasPrefix(token, "Bearer") {
		token = fmt.Sprintf("Bearer %s", token)
	}

	_, respBody, err := r.httpClient.Request(ctx, target, "POST", map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, &body)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetPerm failed: %v, url: %v", err, target)
		return
	}

	err = json.Unmarshal(respBody, &result)

	return
}

// BatchGetDocPerm 批量获取文档权限
func (r *docShare) BatchGetDocPerm(ctx context.Context, method string, subject map[string]string, objects []map[string]string) (map[string]map[string][]string, error) {
	log := traceLog.WithContext(ctx)
	var docsPerm = make([]DocPermInfo, 0)
	var docsPermMap = make(map[string]map[string][]string)
	target := fmt.Sprintf("%v/api/doc-share/v1/access-control/operability", r.privateAddress)
	body := map[string]interface{}{"method": method, "subject": subject, "objects": objects}
	_, respParam, err := r.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		log.Warnf("BatchGetDocPerm failed: %v, url: %v", err, target)
		return docsPermMap, err
	}

	res := respParam.([]interface{})
	byteRes, _ := json.Marshal(res)
	err = json.Unmarshal(byteRes, &docsPerm)
	if err != nil {
		log.Warnf("BatchGetDocPerm unmarshal body faild, detail: %v", err.Error())
		return docsPermMap, err
	}
	for _, docperm := range docsPerm {
		docsPermMap[docperm.ID] = map[string][]string{
			"allow": docperm.Allow,
			"deny":  docperm.Deny,
		}
	}

	return docsPermMap, err
}

// SetDocPerm 批量设置文档权限
func (r *docShare) SetDocPerm(ctx context.Context, docID string, configs []map[string]interface{}) (int, error) {
	target := fmt.Sprintf("%s/api/doc-share/v1/doc-perm", r.privateAddress)
	body := map[string]interface{}{"doc_id": docID, "configs": configs}
	respCode, _, err := r.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetDocPerm failed: %v, url: %v", err, target)
		return respCode, err
	}

	return respCode, err
}

func (r *docShare) CheckOwner(ctx context.Context, docID string, ownerID string) (bool, error) {
	target := fmt.Sprintf("%s/api/doc-share/v1/check-owner?doc_id=%v&owner_id=%v&owner_type=user", r.privateAddress, docID, ownerID)
	_, respParam, err := r.httpClient.Get(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("CheckOwner failed: %v, url: %v", err, target)
		return false, err
	}

	result := respParam.(map[string]interface{})["value"].(map[string]interface{})["result"].(bool)
	return result, nil
}

type DocOwner struct {
	DocID string `json:"doc_id"`
	Owner Owner  `json:"owner"`
}

type Owner struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

func (r *docShare) GetDocOwners(ctx context.Context, docID string) (owners []DocOwner, err error) {
	target := fmt.Sprintf("%s/api/doc-share/v1/doc-owners/%v", r.privateAddress, url.QueryEscape(docID))
	_, body, err := r.httpClient.Request(ctx, target, "GET", map[string]string{"Content-Type": "application/json;charset=UTF-8"}, &[]byte{})
	if err != nil {
		return nil, err
	}

	err = jsoniter.Unmarshal(body, &owners)
	return owners, err
}

type PermConfig struct {
	Accessor  Accessor `json:"accessor"`
	Allow     []string `json:"allow"`
	Deny      []string `json:"deny"`
	ExpiresAt string   `json:"expires_at"`
	EndedTime int64    `json:"-"`
}

type DocPermConfig struct {
	Configs     []PermConfig `json:"configs"`
	Inherit     bool         `json:"inherit"`
	SendMessage bool         `json:"send_message"`
}

func (r *docShare) SetDocPerm2(ctx context.Context, body *DocPermConfig, token, id string) (respCode int, err error) {
	if utils.IsGNS(id) {
		id = id[len(id)-32:]
	}

	target := fmt.Sprintf("%v/api/doc-share/v1/doc-perm/%v", r.address, id)
	if !strings.HasPrefix(token, "Bearer") {
		token = fmt.Sprintf("Bearer %s", token)
	}
	respCode, _, err = r.httpClient.Put(ctx, target, map[string]string{"Authorization": token, "Content-Type": "application/json;charset=UTF-8"}, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetDocPerm2 failed: %v, url: %v", err, target)
		return
	}

	return
}
