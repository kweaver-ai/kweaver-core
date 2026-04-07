// Package drivenadapters 当前微服务依赖的其他服务
package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/doc.go -destination ../tests/mock_drivenadapters/doc_mock.go

// EntryDocLib 入口文档库信息
type EntryDocLib struct {
	ID   string
	Type string
	Name string
}

type DocAttrSubtype struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DocAttr 文档属性
type DocAttr struct {
	CreateTime float64         `json:"create_time"`
	Creator    string          `json:"creator"`
	CreatorID  string          `json:"creator_id"`
	Modified   float64         `json:"modified"`
	Editor     string          `json:"editor"`
	EditorID   string          `json:"editor_id"`
	Name       string          `json:"name"`
	DocID      string          `json:"docid"`
	ID         string          `json:"id"`
	Rev        string          `json:"rev"`
	Path       string          `json:"path"`
	Size       float64         `json:"size"`
	CsfLevel   float64         `json:"csflevel"`
	DocLibType string          `json:"doc_lib_type"`
	Subtype    *DocAttrSubtype `json:"subtype"`
	CustomType *DocAttrSubtype `json:"custom_type"`
}

// UserDocLibInfo 个人文档库信息
type UserDocLibInfo struct {
	Entries []Entry `json:"entries"`
}

// Quota 配额空间
type Quota struct {
	Allocated int64 `json:"allocated"`
	Used      int64 `json:"used"`
}

// Storage 存储信息
type Storage struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Entry entry
type Entry struct {
	Name    string  `json:"name"`
	ID      string  `json:"id"`
	Quota   Quota   `json:"quota"`
	Storage Storage `json:"storage"`
	Type    string  `json:"type"`
}

// DownloadInfo 下载信息
type DownloadInfo struct {
	Name   string
	Rev    string
	Method string
	URL    string
	Size   int64
}

// TemplateStructInfo 编目结构信息
type TemplateStructInfo struct {
	Fields []TemplateField `json:"fields"`
}

// TemplateField 编目模板字段信息
type TemplateField struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

type DocSetSubdocAbstractRes struct {
	DocID   string `json:"doc_id"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Url     string `json:"url"`
	Data    string `json:"data"`
}

type DocSetSubdocFulltextRes struct {
	DocID   string `json:"doc_id"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Url     string `json:"url"`
	ErrMsg  string `json:"err_msg"`
}

// Efast efast服务处理接口
type Efast interface {
	// GetEntryDocLib 获取入口文档库信息
	GetEntryDocLib(ctx context.Context, doclibType, token, ip string) (ids []EntryDocLib, err error)

	// CreateDir 创建目录
	CreateDir(ctx context.Context, docID, name string, ondup int, token, ip string) (doc DocAttr, err error)

	// DeleteDir 删除目录
	DeleteDir(ctx context.Context, docID, token, ip string) (err error)

	// CopyDir 复制目录
	CopyDir(ctx context.Context, docID, destparent string, ondup int, token, ip string) (doc map[string]string, err error)

	// MoveDir 移动目录
	MoveDir(ctx context.Context, docID, destparent string, ondup int, token, ip string) (doc map[string]string, err error)

	// GetDirAttr 获取目录属性
	GetDirAttr(ctx context.Context, docID, token, ip string) (attr DocAttr, err error)

	// RenameDir 重命名目录
	RenameDir(ctx context.Context, docID, name string, ondup int, token, ip string) (doc DocAttr, err error)

	// CopyFile 复制文件
	CopyFile(ctx context.Context, docID, destparent string, ondup int, token, ip string) (doc map[string]string, err error)

	// MoveFile 移动文件
	MoveFile(ctx context.Context, docID, destparent string, ondup int, token, ip string) (doc map[string]string, err error)

	// GetFileAttr 获取文件属性
	GetFileAttr(ctx context.Context, docID, token, ip string) (attr DocAttr, err error)

	// GetFileMetadata 获取文件元数据
	GetFileMetadata(ctx context.Context, docID, token, ip string) (attr DocAttr, err error)

	// GetDocMsg 获取指定文件/目录的元数据
	GetDocMsg(ctx context.Context, docID string) (attr *DocAttr, err error)

	// 根据 object_id 获取指定文件/目录元数据
	GetObjectMsg(ctx context.Context, objectID string) (attr *DocAttr, err error)

	// RenameDir 重命名文件
	RenameFile(ctx context.Context, docID, name string, ondup int, token, ip string) (doc DocAttr, err error)

	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, docID, token, ip string) (err error)

	// SetTag 设置标签
	SetTag(ctx context.Context, docID string, tags []string, token, ip string) (respParam interface{}, err error)

	// ConvertPath 转换路径
	ConvertPath(ctx context.Context, docID, token, ip string) (doc DocAttr, err error)

	// ListDir 列举文件
	ListDir(ctx context.Context, docID string, token, ip string) ([]interface{}, []interface{}, error)

	// GetDocMetaData 获取指定docid的元数据信息
	GetDocMetaData(ctx context.Context, docid string, params []string) (*DocAttr, error)

	// GetDocLibsName 批量转换docGns为文档库名称
	GetDocLibsName(ctx context.Context, docGns []string) (int, interface{}, error)

	// CheckPerm 检查权限
	CheckPerm(ctx context.Context, docID, action, token, ip string) (float64, error)

	// OSDownload OSS 下载
	OSDownload(ctx context.Context, docid, rev, token, ip string) (downloadInfo DownloadInfo, err error)

	// InnerOSDownload OSS 下载
	InnerOSDownload(ctx context.Context, docid, rev string) (downloadInfo DownloadInfo, err error)

	// SetCsfLevel 设置密级
	SetCsfLevel(ctx context.Context, docid string, csflevel int, token, ip string) (float64, error)

	// SetCsfInfo 设置定密信息
	SetCsfInfo(ctx context.Context, docID string, scope, screason, secrecyperiod, token, ip string) error

	// 设置上传审核暂存权限
	SetUploadProcessPerm(ctx context.Context, pid string, docid string, perms []UploadProcessPerm) error

	// SetTemplate 设置编目
	SetTemplate(ctx context.Context, docID, key string, tpl map[string]interface{}, token, ip string) (respParam interface{}, err error)

	// GetTemplates 获取编目
	GetTemplates(ctx context.Context, docID, token, ip string) (tpls map[string]map[string]interface{}, err error)

	// GetTemplateStruct 获取编目模板信息
	GetTemplateStruct(ctx context.Context, templateID, token, ip string) (tpl TemplateStructInfo, err error)

	// DownloadFile 下载文件
	DownloadFile(ctx context.Context, docid, rev, token, ip string) (*[]byte, error)

	DocSetSubdoc(ctx context.Context, params DocSetSubdocParams, sizeLimit float64) (res DocSetSubdocRes, err error)

	// DocSetSubdocWithRetry(ctx context.Context, params DocSetSubdocParams, retryCount int64, waitTime time.Duration, sizeLimit float64) (res DocSetSubdocRes, err error)

	GetDocSetSubdocContent(ctx context.Context, params DocSetSubdocParams, retryCount int64, waitTime time.Duration, sizeLimit float64) (result string, err error)

	// 批量获取文件信息
	BatchGetMetadata(ctx context.Context, ids []string, fields []string) (metadatas []map[string]interface{}, err error)

	// 添加关联文档
	AddRelevance(ctx context.Context, params RelevanceParams, token, ip string) (err error)

	// 设置文件夹属性
	SetDirAttributes(ctx context.Context, id string, fields []string, attributes map[string]interface{}, token, ip string) (err error)

	// 获取目录/文件最高密级
	GetMaxCsfLevel(ctx context.Context, docID string) (int, error)

	// GetUserDocLib 根据用户id获取用户个人文档库信息
	GetUserDocLib(ctx context.Context, userID, token string) (Entry, error)

	// SetUserDocLibQuota 用户个人文档库扩容
	SetUserDocLibQuota(ctx context.Context, docID, token string, scale int64) error

	// GetDocLibs 获取文档库列表
	GetDocLibs(ctx context.Context, docLibType string, token, ip string) ([]Entry, error)

	// 获取摘要
	DocSetSubdocAbstract(ctx context.Context, docID string, version string) (result *DocSetSubdocAbstractRes, err error)

	// 获取纯文本
	DocSetSubdocFulltext(ctx context.Context, docID string, version string) (result *DocSetSubdocFulltextRes, err error)

	// 获取OSS信息
	GetOssInfo(ctx context.Context, docID string, version string) (ossInfo *OssInfo, err error)
}

var (
	efastOnce sync.Once
	efast     Efast
)

type efastSvc struct {
	baseURL            string
	metadataURL        string
	privateBaseURL     string
	docshareURL        string
	documentPublicURL  string
	documentPrivateURL string
	docsetPublicURL    string
	docsetPrivateURL   string
	metadataPublicURL  string
	metadataPrivateURL string
	httpClient         otelHttp.HTTPClient
}

// NewEfast 创建efast服务处理对象
func NewEfast() Efast {
	efastOnce.Do(func() {
		config := common.NewConfig()
		efast = &efastSvc{
			baseURL:            fmt.Sprintf("http://%s:%v/api/efast", config.Efast.PublicHost, config.Efast.PublicPort),
			metadataURL:        fmt.Sprintf("http://%s:%v/api/metadata", config.Emetadata.Host, config.Emetadata.Port),
			privateBaseURL:     fmt.Sprintf("http://%s:%v/api/efast", config.Efast.PrivateHost, config.Efast.PrivatePort),
			docshareURL:        fmt.Sprintf("http://%s:%v", config.DocShare.PublicHost, config.DocShare.PublicPort),
			documentPublicURL:  fmt.Sprintf("http://%s:%v", config.Document.PublicHost, config.Document.PublicPort),
			documentPrivateURL: fmt.Sprintf("http://%s:%v", config.Document.PrivateHost, config.Document.PrivatePort),
			httpClient:         NewOtelHTTPClient(),
			docsetPublicURL:    fmt.Sprintf("http://%s:%v", config.DocSet.PublicHost, config.DocSet.PublicPort),
			docsetPrivateURL:   fmt.Sprintf("http://%s:%v", config.DocSet.PrivateHost, config.DocSet.PrivatePort),
			metadataPublicURL:  fmt.Sprintf("http://%s:%v", config.Metadata.PublicHost, config.Metadata.PublicPort),
			metadataPrivateURL: fmt.Sprintf("http://%s:%v", config.Metadata.PrivateHost, config.Metadata.PrivatePort),
		}
	})

	return efast
}

// CheckPerm 检查指定docGns对象是否有指定操作权限
func (e *efastSvc) CheckPerm(ctx context.Context, docID, action, token, ip string) (float64, error) {
	var hasPerm float64
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%v/api/eacp/v1/perm1/check", e.docshareURL)
	body := map[string]interface{}{"perm": action, "docid": docID}
	if !strings.HasPrefix(token, "Bearer ") {
		token = fmt.Sprintf("Bearer %s", token)
	}
	headers := map[string]string{
		"Authorization":   token,
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		log.Warnf("CheckPerm failed: %v, url: %v", err, target)
		return hasPerm, err
	}

	return respParam.(map[string]interface{})["result"].(float64), nil
}

// GetEntryDocLib 获取入口文档库信息
func (e *efastSvc) GetEntryDocLib(ctx context.Context, doclibType, token, ip string) (ids []EntryDocLib, err error) {
	target := fmt.Sprintf("%v/v1/entry-doc-lib?type=%s", e.baseURL, doclibType)
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetEntryDocLib failed: %v, url: %v", err, target)
		return
	}

	results := respParam.([]interface{})

	for _, item := range results {
		id := item.(map[string]interface{})["id"].(string)
		docType := item.(map[string]interface{})["type"].(string)
		name := item.(map[string]interface{})["name"].(string)
		ids = append(ids, EntryDocLib{
			ID:   id,
			Type: docType,
			Name: name,
		})
	}

	return
}

// CreateDir 创建目录
func (e *efastSvc) CreateDir(ctx context.Context, docID, name string, ondup int, token, ip string) (doc DocAttr, err error) {
	target := fmt.Sprintf("%v/v1/dir/create", e.baseURL)
	body := map[string]interface{}{"docid": docID, "name": name, "ondup": ondup}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("CreateDir failed: %v, url: %v", err, target)
		return
	}

	doc.DocID = respParam.(map[string]interface{})["docid"].(string)
	doc.Name = name
	doc.Editor = respParam.(map[string]interface{})["editor"].(string)
	doc.Creator = respParam.(map[string]interface{})["creator"].(string)
	doc.Modified = respParam.(map[string]interface{})["modified"].(float64)
	doc.CreateTime = respParam.(map[string]interface{})["create_time"].(float64)

	return
}

// DeleteDir 删除目录
func (e *efastSvc) DeleteDir(ctx context.Context, docID, token, ip string) (err error) {
	target := fmt.Sprintf("%v/v1/dir/delete", e.baseURL)
	body := map[string]interface{}{"docid": docID}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, _, err = e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("DeleteDir failed: %v, url: %v", err, target)
		return
	}

	return
}

// CopyDir 复制目录
func (e *efastSvc) CopyDir(ctx context.Context, docID, destparent string, ondup int, token, ip string) (doc map[string]string, err error) {
	doc = make(map[string]string)
	target := fmt.Sprintf("%v/v1/dir/copy", e.baseURL)
	body := map[string]interface{}{"docid": docID, "destparent": destparent, "ondup": ondup}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("CopyDir failed: %v, url: %v", err, target)
		return
	}
	doc["docid"] = docID
	docid, ok := respParam.(map[string]interface{})["docid"].(string)
	if ok {
		doc["docid"] = docid
	}
	name, ok := respParam.(map[string]interface{})["name"].(string)
	if ok {
		doc["name"] = name
	}

	return
}

// MoveDir 移动目录
func (e *efastSvc) MoveDir(ctx context.Context, docID, destparent string, ondup int, token, ip string) (doc map[string]string, err error) {
	doc = make(map[string]string)
	target := fmt.Sprintf("%v/v1/dir/move", e.baseURL)
	body := map[string]interface{}{"docid": docID, "destparent": destparent, "ondup": ondup}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("MoveDir failed: %v, url: %v", err, target)
		return
	}

	doc["docid"] = respParam.(map[string]interface{})["docid"].(string)

	return
}

// GetDirAttr 获取目录属性
func (e *efastSvc) GetDirAttr(ctx context.Context, docID, token, ip string) (attr DocAttr, err error) {
	target := fmt.Sprintf("%v/v1/dir/attribute", e.baseURL)
	body := map[string]string{"docid": docID}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetDirAttr failed: %v, url: %v", err, target)
		return
	}

	attr = DocAttr{
		Creator:    respParam.(map[string]interface{})["creator"].(string),
		Name:       respParam.(map[string]interface{})["name"].(string),
		CreateTime: respParam.(map[string]interface{})["create_time"].(float64),
		DocID:      docID,
	}

	return
}

// RenameDir 重命名目录
func (e *efastSvc) RenameDir(ctx context.Context, docID, name string, ondup int, token, ip string) (doc DocAttr, err error) {
	target := fmt.Sprintf("%v/v1/dir/rename", e.baseURL)
	body := map[string]interface{}{"docid": docID, "name": name, "ondup": ondup}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RenameDir failed: %v, url: %v", err, target)
		return
	}

	doc.DocID = docID
	if param, ok := respParam.(map[string]interface{}); ok {
		if name, ok := param["name"].(string); ok {
			doc.Name = name
		}
	}
	return
}

// DeleteFile 删除文件
func (e *efastSvc) DeleteFile(ctx context.Context, docID, token, ip string) (err error) {
	target := fmt.Sprintf("%v/v1/file/delete", e.baseURL)
	body := map[string]interface{}{"docid": docID}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, _, err = e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("DeleteFile failed: %v, url: %v", err, target)
		return
	}

	return
}

// CopyFile 复制文件
func (e *efastSvc) CopyFile(ctx context.Context, docID, destparent string, ondup int, token, ip string) (doc map[string]string, err error) {
	doc = make(map[string]string)
	target := fmt.Sprintf("%v/v1/file/copy", e.baseURL)
	body := map[string]interface{}{"docid": docID, "destparent": destparent, "ondup": ondup}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("CopyFile failed: %v, url: %v", err, target)
		return
	}

	docAttr := respParam.(map[string]interface{})

	doc["docid"] = docAttr["docid"].(string)
	if name, ok := docAttr["name"].(string); ok {
		doc["name"] = name
	}

	return
}

// MoveFile 移动文件
func (e *efastSvc) MoveFile(ctx context.Context, docID, destparent string, ondup int, token, ip string) (doc map[string]string, err error) {
	doc = make(map[string]string)
	target := fmt.Sprintf("%v/v1/file/move", e.baseURL)
	body := map[string]interface{}{"docid": docID, "destparent": destparent, "ondup": ondup}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("MoveFile failed: %v, url: %v", err, target)
		return
	}

	doc["docid"] = respParam.(map[string]interface{})["docid"].(string)

	return
}

// GetFileAttr 获取文件属性
func (e *efastSvc) GetFileAttr(ctx context.Context, docID, token, ip string) (attr DocAttr, err error) {
	target := fmt.Sprintf("%v/v1/file/attribute", e.baseURL)
	body := map[string]string{"docid": docID}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetFileAttr failed: %v, url: %v", err, target)
		return
	}

	parsedTime, err := time.Parse(time.RFC3339, respParam.(map[string]interface{})["modified_at"].(string))
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetFileAttr failed: %v, url: %v", err, target)
		return
	}

	// 转换为时间戳
	modified := parsedTime.UnixNano() / 1000

	attr = DocAttr{
		Creator:    respParam.(map[string]interface{})["creator"].(string),
		Name:       respParam.(map[string]interface{})["name"].(string),
		CreateTime: respParam.(map[string]interface{})["create_time"].(float64),
		Modified:   float64(modified),
		DocID:      docID,
		Rev:        respParam.(map[string]interface{})["rev"].(string),
	}

	return
}

// GetFileMetadata 获取文件元数据
func (e *efastSvc) GetFileMetadata(ctx context.Context, docID, token, ip string) (attr DocAttr, err error) {
	target := fmt.Sprintf("%v/v1/file/metadata", e.baseURL)
	body := map[string]string{"docid": docID}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetFileMetadata failed: %v, url: %v", err, target)
		return
	}

	attr = DocAttr{
		Editor: respParam.(map[string]interface{})["editor"].(string),
		Name:   respParam.(map[string]interface{})["name"].(string),
		Size:   respParam.(map[string]interface{})["size"].(float64),
		DocID:  docID,
	}

	return
}

// GetDocMsg 获取文件完整信息
func (e *efastSvc) GetDocMsg(ctx context.Context, docID string) (attr *DocAttr, err error) {
	curID := utils.GetDocCurID(docID)
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	targetMetadata := fmt.Sprintf("%v/v1/items/%s/name,path,created_at,created_by,modified_at,modified_by,size,csflevel,rev,doc_lib_type", e.privateBaseURL, curID)
	_, respMetadata, err := e.httpClient.Get(ctx, targetMetadata, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetDocMsg failed: %v, url: %v", err, targetMetadata)
		return
	}

	metadata := respMetadata.(map[string]interface{})
	attr = mapMetadataToDocAttr(metadata, docID)
	return
}

func (e *efastSvc) GetObjectMsg(ctx context.Context, objectID string) (attr *DocAttr, err error) {
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	url := fmt.Sprintf("%v/v1/objects/%s/id", e.privateBaseURL, objectID)
	_, obj, err := e.httpClient.Get(ctx, url, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetObjectMsg failed: %v, url: %v", err, url)
		return
	}

	doc := obj.(map[string]interface{})

	return e.GetDocMsg(ctx, doc["id"].(string))
}

// GetDocMsg 获取文件夹完整信息
// func (e *efastSvc) GetDirMsg(docID, token string) (attr DocAttr, err error) {

// }

// RenameFile 重命名文件
func (e *efastSvc) RenameFile(ctx context.Context, docID, name string, ondup int, token, ip string) (doc DocAttr, err error) {
	target := fmt.Sprintf("%v/v1/dir/rename", e.baseURL)
	body := map[string]interface{}{"docid": docID, "name": name, "ondup": ondup}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RenameDir failed: %v, url: %v", err, target)
		return
	}

	doc.DocID = docID
	if param, ok := respParam.(map[string]interface{}); ok {
		if name, ok := param["name"].(string); ok {
			doc.Name = name
		}
	}

	return
}

// SetTag 设置标签
func (e *efastSvc) SetTag(ctx context.Context, docID string, tags []string, token, ip string) (respParam interface{}, err error) {
	target := fmt.Sprintf("%v/v1/item/%s/tag", e.metadataURL, docID)
	body := tags
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err = e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetTag failed: %v, url: %v", err, target)
		return
	}

	return
}

// SetTemplate 设置编目
func (e *efastSvc) SetTemplate(ctx context.Context, docID, key string, tpl map[string]interface{}, token, ip string) (respParam interface{}, err error) {
	target := fmt.Sprintf("%v/v1/item/%s/mdata/aishu/%s", e.metadataURL, docID, key)
	body := tpl
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err = e.httpClient.Put(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetTemplate failed: %v, url: %v", err, target)
		return
	}

	return
}

// GetTemplates 获取编目
func (e *efastSvc) GetTemplates(ctx context.Context, docID, token, ip string) (tpls map[string]map[string]interface{}, err error) {
	target := fmt.Sprintf("%v/v1/item/%s/mdata/aishu", e.metadataURL, docID)
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetTemplates failed: %v, url: %v", err, target)
		return
	}

	resMap, ok := respParam.(map[string]interface{})
	tpls = make(map[string]map[string]interface{})
	if ok {
		entries := resMap["entries"].([]interface{})
		for _, e := range entries {
			tpl := e.(map[string]interface{})
			templatKey, ok := tpl["template_key"].(string)
			if !ok {
				continue
			}
			tpls[templatKey] = tpl
		}
	}

	return
}

// GetTemplateStruct 获取编目模板
func (e *efastSvc) GetTemplateStruct(ctx context.Context, templateID, token, ip string) (tpl TemplateStructInfo, err error) {
	target := fmt.Sprintf("%v/v1/templates/aishu/%s", e.metadataURL, templateID)
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetTemplates failed: %v, url: %v", err, target)
		return
	}
	tplBytes, _ := json.Marshal(respParam)
	err = json.Unmarshal(tplBytes, &tpl)

	return
}

// ConvertPath 转换路径
func (e *efastSvc) ConvertPath(ctx context.Context, docID, token, ip string) (doc DocAttr, err error) {
	target := fmt.Sprintf("%v/v1/file/convertpath", e.baseURL)
	body := map[string]interface{}{"docid": docID}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("ConvertPath failed: %v, url: %v", err, target)
		return
	}

	doc.Path = respParam.(map[string]interface{})["namepath"].(string)
	doc.DocID = docID

	return
}

// ListDir 列举目录下文件
func (e *efastSvc) ListDir(ctx context.Context, docID, token, ip string) (files, dirs []interface{}, err error) {
	target := fmt.Sprintf("%v/v1/dir/list", e.baseURL)
	body := map[string]string{"docid": docID}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("ListDir failed: %v, url: %v, docid: %s", err, target, docID)
		return
	}

	files = respParam.(map[string]interface{})["files"].([]interface{})

	dirs = respParam.(map[string]interface{})["dirs"].([]interface{})

	return files, dirs, err
}

// SetCsfLevel 设置文件密级
func (e *efastSvc) SetCsfLevel(ctx context.Context, docID string, csflevel int, token, ip string) (result float64, err error) {
	target := fmt.Sprintf("%v/api/efast/v1/file/setcsflevel", e.documentPublicURL)
	body := map[string]interface{}{"docid": docID, "csflevel": csflevel}
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetCsfLevel failed: %v, url: %v", err, target)
		return
	}

	resp, ok := respParam.(map[string]interface{})
	if !ok {
		return
	}
	respResult, ok := resp["result"].(float64)
	if !ok {
		return
	}

	return respResult, nil
}

func (e *efastSvc) SetCsfInfo(ctx context.Context, docID string, scope, screason, secrecyperiod, token, ip string) error {
	url := fmt.Sprintf("%v/v1/file/setcsfinfo", e.baseURL)
	body := map[string]interface{}{
		"docid": docID,
		"csfinfo": map[string]interface{}{
			"scope":         scope,
			"screason":      screason,
			"secrecyperiod": secrecyperiod,
		},
	}

	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, _, err := e.httpClient.Post(ctx, url, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetCsfInfo failed: %v, url: %v", err, url)
		return err
	}

	return nil
}

type Accessor struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type UploadProcessPerm struct {
	Allow    []string `json:"allow"`
	Deny     []string `json:"deny"`
	Accessor Accessor `json:"accessor"`
	Endtime  int64    `json:"endtime"`
}

func (e *efastSvc) SetUploadProcessPerm(ctx context.Context, pid string, docid string, perms []UploadProcessPerm) error {
	reqParam := make(map[string]interface{})
	reqParam["doc_id"] = docid
	reqParam["configs"] = perms

	target := fmt.Sprintf("%v/api/document/v1/upload-process-perm/%v", e.documentPrivateURL, pid)

	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	_, _, err := e.httpClient.Put(ctx, target, headers, reqParam)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetUploadProcessPerm failed: %v, url: %v", err, target)
	}

	return err
}

// GetDocMetaData 获取指定docid的元数据信息
func (e *efastSvc) GetDocMetaData(ctx context.Context, docid string, params []string) (*DocAttr, error) {
	curID := utils.GetDocCurID(docid)
	params2str := strings.Join(params, ",")
	target := fmt.Sprintf("%v/v1/items/%s/%s", e.privateBaseURL, curID, params2str)
	_, respParam, err := e.httpClient.Get(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetDocMetaData failed: %v, url: %v", err, target)
		return nil, err
	}
	metadata := respParam.(map[string]interface{})
	attr := mapMetadataToDocAttr(metadata, docid)
	return attr, nil
}

// GetDocLibsName 批量转换docGns为文档库名称
func (e *efastSvc) GetDocLibsName(ctx context.Context, docGns []string) (respCode int, respParam interface{}, err error) {
	target := fmt.Sprintf("%v/v1/doc-libs/name", e.privateBaseURL)
	respCode, respParam, err = e.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, docGns)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetDocLibsName failed: %v, url: %v", err, target)
	}
	return respCode, respParam, err
}

// OSDownload OSS 下载
func (e *efastSvc) OSDownload(ctx context.Context, docid, rev, token, ip string) (downloadInfo DownloadInfo, err error) {
	reqParam := make(map[string]interface{})
	reqParam["docid"] = docid
	reqParam["rev"] = rev
	reqParam["authtype"] = "QUERY_STRING"
	reqParam["externalrequest"] = false
	target := fmt.Sprintf("%v/v1/file/osdownload", e.baseURL)
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, reqParam)
	if err == nil {
		downloadInfo.Name = respParam.(map[string]interface{})["name"].(string)
		downloadInfo.Rev = respParam.(map[string]interface{})["rev"].(string)
		downloadInfo.Size = int64(respParam.(map[string]interface{})["size"].(float64))
		downloadInfo.Method = respParam.(map[string]interface{})["authrequest"].([]interface{})[0].(string)
		downloadInfo.URL = respParam.(map[string]interface{})["authrequest"].([]interface{})[1].(string)
	} else {
		traceLog.WithContext(ctx).Warnf("OSDownload failed: %v, url: %v", err, target)
	}
	return
}

// InnerOSDownload OSS 下载
func (e *efastSvc) InnerOSDownload(ctx context.Context, docid, rev string) (downloadInfo DownloadInfo, err error) {
	reqParam := make(map[string]interface{})
	reqParam["docid"] = docid
	reqParam["rev"] = rev
	reqParam["authtype"] = "QUERY_STRING"
	reqParam["externalrequest"] = false
	target := fmt.Sprintf("%v/v1/file/osdownload", e.privateBaseURL)
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}
	_, respParam, err := e.httpClient.Post(ctx, target, headers, reqParam)
	if err == nil {
		downloadInfo.Name = respParam.(map[string]interface{})["name"].(string)
		downloadInfo.Rev = respParam.(map[string]interface{})["rev"].(string)
		downloadInfo.Size = int64(respParam.(map[string]interface{})["size"].(float64))
		downloadInfo.Method = respParam.(map[string]interface{})["authrequest"].([]interface{})[0].(string)
		downloadInfo.URL = respParam.(map[string]interface{})["authrequest"].([]interface{})[1].(string)
	} else {
		traceLog.WithContext(ctx).Warnf("InnerOSDownload failed: %v, url: %v", err, target)
	}
	return
}

// DownloadFile 下载文件
func (e *efastSvc) DownloadFile(ctx context.Context, docid, rev, token, ip string) (*[]byte, error) {
	downLoadInfo, err := e.OSDownload(ctx, docid, rev, token, ip)
	if err != nil {
		return nil, err
	}
	_, resp, err := e.httpClient.Get(ctx, downLoadInfo.URL, nil)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("DownloadFile failed: %v, url: %v", err, downLoadInfo.URL)
		return nil, err
	}
	var content []byte
	respByte, _ := json.Marshal(resp)
	_ = json.Unmarshal(respByte, &content)
	return &content, nil
}

type DocSetSubdocFormat string

const (
	DocSetSubdocFormatRaw = "raw"
	DocSetSubdocFormatUrl = "url"
)

type DocSetSubdocParams struct {
	DocID    string             `json:"doc_id"`
	Version  string             `json:"version,omitempty"`
	Type     string             `json:"type"`
	Custom   interface{}        `json:"custom,omitempty"`
	Format   DocSetSubdocFormat `json:"format,omitempty"`
	Priority int                `json:"priority,omitempty"`
}

type DocSetSubdocRes struct {
	DocID  string      `json:"doc_id"`
	Rev    string      `json:"rev"`
	Status string      `json:"status"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data"`
	Url    string      `json:"url"`
}

func (e *efastSvc) DocSetSubdoc(ctx context.Context, params DocSetSubdocParams, sizeLimit float64) (res DocSetSubdocRes, err error) {
	target := fmt.Sprintf("%v/api/docset/v1/subdoc", e.docsetPrivateURL)
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	var docAttr *DocAttr

	if strings.HasPrefix(params.DocID, "gns://") {
		docAttr, err = e.GetDocMsg(ctx, params.DocID)

		segments := strings.Split(params.DocID, "/")
		params.DocID = segments[len(segments)-1]
	} else {
		docAttr, err = e.GetObjectMsg(ctx, params.DocID)
	}

	if err != nil {
		return
	}

	if docAttr.Size == -1 {
		err = ierrors.NewIError(errors.FileTypeNotSupported, "", map[string]interface{}{
			"docid": docAttr.DocID,
		})
		return
	}

	if docAttr.Size == 0 {
		err = ierrors.NewIError(errors.FileIsEmpty, "", map[string]interface{}{
			"docid": docAttr.DocID,
		})
		return
	}

	// 限制最大处理文件大小为10M
	if docAttr.Size >= sizeLimit && sizeLimit > 0 {
		err = ierrors.NewIError(errors.FileSizeExceed, "", map[string]interface{}{
			"docid": docAttr.DocID,
			"limit": sizeLimit,
		})
		return
	}

	_, respParam, err := e.httpClient.Post(ctx, target, headers, params)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("DocSetSubdoc failed: %v, url: %v", err, target)
		return
	}
	if parsedParam, ok := respParam.(map[string]interface{}); ok {
		if docid, ok := parsedParam["doc_id"].(string); ok {
			res.DocID = docid
		}
		if rev, ok := parsedParam["version"].(string); ok {
			res.Rev = rev
		}
		if status, ok := parsedParam["status"].(string); ok {
			res.Status = status
		}
		if errMsg, ok := parsedParam["err_msg"].(string); ok {
			res.ErrMsg = errMsg
		}
		if url, ok := parsedParam["url"].(string); ok {
			res.Url = url
		}
		if data, ok := parsedParam["data"]; ok {
			res.Data = data
		}
	}

	return
}

func (e *efastSvc) DocSetSubdocWithRetry(ctx context.Context, params DocSetSubdocParams, retryCount int64, waitTime time.Duration, sizeLimit float64) (res DocSetSubdocRes, err error) {
	remainCount := retryCount
	count := 0
	log := traceLog.WithContext(ctx)

	for {
		fmt.Printf("DocSetSubdocWithRetry: %v, retry: %v, remain: %v\n", params, count, remainCount)
		count += 1
		res, err = e.DocSetSubdoc(ctx, params, sizeLimit)

		if err != nil {
			log.Warnf("DocSetSubdocWithRetry failed: %v", err)
			return
		}

		if res.Status == "processing" || res.Status == "ready" {
			if remainCount == -1 {
				time.Sleep(waitTime)
				continue
			} else if remainCount > 0 {
				time.Sleep(waitTime)
				remainCount -= 1
				continue
			} else {
				log.Warnf("DocSetSubdocWithRetry retry time exceed, status: %v, err_msg: %v", res.Status, res.ErrMsg)
				return res, ierrors.NewIError(errors.InternalError, errors.ErrorDepencyService, map[string]interface{}{
					"status":  res.Status,
					"err_msg": res.ErrMsg,
				})
			}
		} else if res.Status == "failed" {
			log.Warnf("DocSetSubdocWithRetry failed, status: %v, err_msg: %v", res.Status, res.ErrMsg)
			return res, ierrors.NewIError(errors.FileContentUnknow, "", map[string]interface{}{
				"status":  res.Status,
				"err_msg": res.ErrMsg,
			})
		} else {
			return res, nil
		}
	}
}

func (e *efastSvc) GetDocSetSubdocContent(ctx context.Context, params DocSetSubdocParams, retryCount int64, waitTime time.Duration, sizeLimit float64) (result string, err error) {
	res, err := e.DocSetSubdocWithRetry(ctx, params, retryCount, waitTime, sizeLimit)
	if err != nil {
		return
	}
	client := NewRawHTTPClient()
	client.Timeout = 60 * time.Second
	resp, err := client.Get(res.Url)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			traceLog.WithContext(ctx).Warnln(closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	respCode := resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		err = ierrors.ExHTTPError{
			Body:   string(body),
			Status: respCode,
		}
		return
	}

	return string(body), nil
}

type Object struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type RelevanceItem struct {
	Details string `json:"details"`
	Item    Object `json:"item"`
}

type RelevanceParams struct {
	Item      Object          `json:"item"`
	Relevance []RelevanceItem `json:"relevance"`
}

func (e *efastSvc) AddRelevance(ctx context.Context, params RelevanceParams, token, ip string) (err error) {

	target := fmt.Sprintf("%v/api/metadata/v1/relevance", e.metadataPublicURL)
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, _, err = e.httpClient.Post(ctx, target, headers, params)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("AddRelevance failed: %v, url: %v", err, target)
		return
	}

	return
}

func (e *efastSvc) BatchGetMetadata(ctx context.Context, ids []string, fields []string) (metadatas []map[string]interface{}, err error) {

	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	target := fmt.Sprintf("%v/v1/items-batch/%v", e.privateBaseURL, strings.Join(fields, ","))

	_, respParam, err := e.httpClient.Post(ctx, target, headers, map[string]interface{}{
		"method": "GET",
		"ids":    ids,
	})

	if err != nil {
		return nil, err
	}

	items, ok := respParam.([]interface{})

	if ok {
		for _, item := range items {
			metadatas = append(metadatas, item.(map[string]interface{}))
		}
		return metadatas, nil
	}

	return nil, fmt.Errorf("invalid type")
}

func (e *efastSvc) SetDirAttributes(ctx context.Context, id string, fields []string, attributes map[string]interface{}, token, ip string) (err error) {
	headers := map[string]string{
		"Content-Type":    "application/json;charset=UTF-8",
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"X-Forwarded-For": ip,
	}
	target := fmt.Sprintf("%v/api/document/v1/dirs/%v/attributes/%v", e.documentPublicURL, id, strings.Join(fields, ","))

	_, _, err = e.httpClient.Put(ctx, target, headers, attributes)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetDirAttributes failed: %v, url: %v", err, target)
		return
	}

	return
}

// 获取目录/文件最高密级
func (e *efastSvc) GetMaxCsfLevel(ctx context.Context, docID string) (int, error) {

	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	target := fmt.Sprintf("%v/v1/max-csflevel?id=%v", e.privateBaseURL, url.QueryEscape(docID))
	_, resp, err := e.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetDocMsg failed: %v, url: %v", err, target)
		return 0, err
	}

	return int(resp.(map[string]interface{})["csf_level"].(float64)), nil
}

// GetUserDocLib 根据用户id获取用户个人文档库信息
func (e *efastSvc) GetUserDocLib(ctx context.Context, userID, token string) (Entry, error) {
	var userDocLibInfo UserDocLibInfo
	target := fmt.Sprintf("%v/v1/doc-lib/user?offset=0&limit=200&search_in_users=%s", e.baseURL, userID)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json;charset=UTF-8",
	}
	_, resp, err := e.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetDocMsg failed: %v, url: %v", err, target)
		return Entry{}, err
	}

	respBytes, _ := json.Marshal(resp)
	_ = json.Unmarshal(respBytes, &userDocLibInfo)

	if len(userDocLibInfo.Entries) == 0 {
		return Entry{}, nil
	}

	return userDocLibInfo.Entries[0], nil
}

// GetDocLibs 获取文档库列表信息
func (e *efastSvc) GetDocLibs(ctx context.Context, docLibType string, token, ip string) ([]Entry, error) {
	var userDocLibInfo UserDocLibInfo
	target := fmt.Sprintf("%v/v1/doc-lib/%s?offset=0&limit=200", e.baseURL, docLibType)
	headers := map[string]string{
		"Authorization":   fmt.Sprintf("Bearer %s", token),
		"Content-Type":    "application/json;charset=UTF-8",
		"X-Forwarded-For": ip,
	}
	_, resp, err := e.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetDocLibs failed: %v, url: %v", err, target)
		return []Entry{}, err
	}

	respBytes, _ := json.Marshal(resp)
	_ = json.Unmarshal(respBytes, &userDocLibInfo)

	if len(userDocLibInfo.Entries) == 0 {
		return []Entry{}, nil
	}

	return userDocLibInfo.Entries, nil
}

// SetUserDocLibQuota 用户个人文档库扩容
func (e *efastSvc) SetUserDocLibQuota(ctx context.Context, docID, token string, scale int64) error {
	target := fmt.Sprintf("%v/v1/doc-lib/user/%s/storage_location,quota_allocated", e.baseURL, url.QueryEscape(docID))
	body := map[string]interface{}{
		"storage_location": map[string]interface{}{"type": "unspecified"},
		"quota_allocated":  scale,
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json;charset=UTF-8",
	}
	_, _, err := e.httpClient.Put(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("SetUserDocLibQuota failed: %v, url: %v", err, target)
		return err
	}

	return nil
}

func (e *efastSvc) DocSetSubdocAbstract(ctx context.Context, docID string, version string) (result *DocSetSubdocAbstractRes, err error) {
	target := fmt.Sprintf("%v/api/docset/v1/subdoc/abstract", e.docsetPrivateURL)
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}
	result = &DocSetSubdocAbstractRes{}

	if len(docID) >= 32 {
		docID = docID[len(docID)-32:]
	}

	body := map[string]interface{}{
		"doc_id":      docID,
		"doc_md_type": "abstract",
		"format":      "raw",
	}

	if version != "" {
		body["version"] = version
	}
	_, resp, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("DocSetSubdocAbstract failed: %v, url: %v", err, target)
		return nil, err
	}
	respBytes, _ := json.Marshal(resp)
	_ = json.Unmarshal(respBytes, result)

	return
}

func (e *efastSvc) DocSetSubdocFulltext(ctx context.Context, docID string, version string) (result *DocSetSubdocFulltextRes, err error) {
	target := fmt.Sprintf("%v/api/docset/v1/subdoc/full_text", e.docsetPrivateURL)
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}
	result = &DocSetSubdocFulltextRes{}

	if len(docID) >= 32 {
		docID = docID[len(docID)-32:]
	}

	body := map[string]interface{}{
		"doc_id":      docID,
		"doc_md_type": "abstract",
		"format":      "raw",
	}

	if version != "" {
		body["version"] = version
	}
	_, resp, err := e.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("DocSetSubdocAbstract failed: %v, url: %v", err, target)
		return nil, err
	}
	respBytes, _ := json.Marshal(resp)
	_ = json.Unmarshal(respBytes, result)

	return
}

// mapMetadataToDocAttr 将metadata映射到DocAttr结构体
func mapMetadataToDocAttr(metadata map[string]interface{}, docID string) *DocAttr { //nolint:funlen
	attr := &DocAttr{
		DocID: docID,
		ID:    docID,
	}

	// 安全地获取嵌套值
	if createdBy, ok := metadata["created_by"].(map[string]interface{}); ok {
		if name, ok := createdBy["name"].(string); ok {
			attr.Creator = name
		}
		if id, ok := createdBy["id"].(string); ok {
			attr.CreatorID = id
		}
	}

	if modifiedBy, ok := metadata["modified_by"].(map[string]interface{}); ok {
		if name, ok := modifiedBy["name"].(string); ok {
			attr.Editor = name
		}
		if id, ok := modifiedBy["id"].(string); ok {
			attr.EditorID = id
		}
	}

	// 安全地获取基本值
	if createTime, ok := metadata["created_at"].(float64); ok {
		attr.CreateTime = createTime
	}
	if modified, ok := metadata["modified_at"].(float64); ok {
		attr.Modified = modified
	}
	if name, ok := metadata["name"].(string); ok {
		attr.Name = name
	}
	if size, ok := metadata["size"].(float64); ok {
		attr.Size = size
	}
	if csfLevel, ok := metadata["csflevel"].(float64); ok {
		attr.CsfLevel = csfLevel
	}
	if path, ok := metadata["path"].(string); ok {
		attr.Path = path
	}
	if rev, ok := metadata["rev"].(string); ok {
		attr.Rev = rev
	}
	if docLibType, ok := metadata["doc_lib_type"].(string); ok {
		attr.DocLibType = docLibType
	}

	// 安全地获取subtype
	if subtype, ok := metadata["subtype"]; ok {
		if m, ok := subtype.(map[string]interface{}); ok {
			if id, ok := m["id"].(string); ok && m["name"] != nil {
				attr.Subtype = &DocAttrSubtype{
					ID:   id,
					Name: m["name"].(string),
				}
			}
		}
	}

	// 安全地获取custom_type
	if customType, ok := metadata["custom_type"]; ok {
		if m, ok := customType.(map[string]interface{}); ok {
			if id, ok := m["id"].(string); ok && m["name"] != nil {
				attr.CustomType = &DocAttrSubtype{
					ID:   id,
					Name: m["name"].(string),
				}
			}
		}
	}

	return attr
}

type OssInfo struct {
	OssID      string `json:"oss_id"`
	ObjectName string `json:"object_name"`
}

func (e *efastSvc) GetOssInfo(ctx context.Context, docID string, version string) (ossInfo *OssInfo, err error) {
	target := fmt.Sprintf("%v/v1/files/%s/oss-infos", e.privateBaseURL, url.QueryEscape(docID))
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	if version != "" {
		target = target + "?version_id=" + version
	}

	_, resp, err := e.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetOssInfo failed: %v, url: %v", err, target)
		return nil, err
	}

	ossInfo = &OssInfo{}
	respBytes, _ := json.Marshal(resp)
	err = json.Unmarshal(respBytes, ossInfo)

	if err != nil {
		traceLog.WithContext(ctx).Warnf("GetOssInfo failed: %v, url: %v", err, target)
		return nil, err
	}

	return
}
