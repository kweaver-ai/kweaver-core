// Package model 定义数据库操作接口
// @file api_metadata.go
// @description: 定义t_metadata_api表操作接口
package model

//go:generate mockgen -source=api_metadata.go -destination=../../mocks/model_api_metadata.go -package=mocks
import (
	"context"
	"database/sql"
	"time"
)

// APIMetadataDB API元数据数据库
type APIMetadataDB struct {
	ID          int64  `json:"f_id" db:"f_id"`
	Summary     string `json:"f_summary" db:"f_summary"`
	Version     string `json:"f_version" db:"f_version"`
	Description string `json:"f_description" db:"f_description"`
	Path        string `json:"f_path" db:"f_path"`
	ServerURL   string `json:"f_svc_url" db:"f_svc_url"`
	Method      string `json:"f_method" db:"f_method"`
	APISpec     string `json:"f_api_spec" db:"f_api_spec"`
	CreateUser  string `json:"f_create_user" db:"f_create_user"`
	CreateTime  int64  `json:"f_create_time" db:"f_create_time"`
	UpdateUser  string `json:"f_update_user" db:"f_update_user"`
	UpdateTime  int64  `json:"f_update_time" db:"f_update_time"`
	ErrMessage  string `json:"-"` // 错误信息
}

// IAPIMetadataDB API元数据数据库
type IAPIMetadataDB interface {
	InsertAPIMetadata(ctx context.Context, tx *sql.Tx, metadata *APIMetadataDB) (version string, err error)
	SelectByVersion(ctx context.Context, version string) (has bool, metadata *APIMetadataDB, err error)
	UpdateByVersion(ctx context.Context, tx *sql.Tx, version string, metadata *APIMetadataDB) error
	UpdateByID(ctx context.Context, tx *sql.Tx, id int64, metadata *APIMetadataDB) error
	DeleteByVersion(ctx context.Context, tx *sql.Tx, version string) error
	DeleteByVersions(ctx context.Context, tx *sql.Tx, versions []string) error
	InsertAPIMetadatas(ctx context.Context, tx *sql.Tx, metadatas []*APIMetadataDB) (versions []string, err error)
	SelectListByVersion(ctx context.Context, versions []string) ([]*APIMetadataDB, error)
}

// GetType 获取资源类型
func (a *APIMetadataDB) GetType() string {
	return string(SourceTypeOpenAPI)
}

// GetSummary 获取摘要
func (a *APIMetadataDB) GetSummary() string {
	return a.Summary
}

// GetDescription 获取函数描述
func (a *APIMetadataDB) GetDescription() string {
	if a.Description == "" {
		return a.Summary
	}
	return a.Description
}

// GetVersion 获取版本
func (a *APIMetadataDB) GetVersion() string {
	return a.Version
}
func (a *APIMetadataDB) GetMethod() string {
	return a.Method
}
func (a *APIMetadataDB) GetPath() string {
	return a.Path
}
func (a *APIMetadataDB) GetScriptType() string {
	return ""
}

func (a *APIMetadataDB) Validate(ctx context.Context) error {
	return nil
}

func (a *APIMetadataDB) SetSummary(summary string) {
	a.Summary = summary
}
func (a *APIMetadataDB) SetDescription(description string) {
	a.Description = description
}

func (a *APIMetadataDB) SetMethod(method string) {
	a.Method = method
}
func (a *APIMetadataDB) SetPath(path string) {
	a.Path = path
}
func (a *APIMetadataDB) SetVersion(version string) {
	a.Version = version
}
func (a *APIMetadataDB) SetScriptType(scriptType string) {
	// 不支持设置运行时
}
func (a *APIMetadataDB) GetServerURL() string {
	return a.ServerURL
}
func (a *APIMetadataDB) SetServerURL(serverURL string) {
	a.ServerURL = serverURL
}

// GetAPISpec 获取API规范
func (a *APIMetadataDB) GetAPISpec() string {
	return a.APISpec
}

func (a *APIMetadataDB) SetAPISpec(apiSpec string) {
	a.APISpec = apiSpec
}

// GetUpdateUser 获取更新用户
func (a *APIMetadataDB) GetUpdateUser() (user string) {
	return a.UpdateUser
}

func (a *APIMetadataDB) SetUpdateInfo(user string) {
	a.UpdateUser = user
	a.UpdateTime = time.Now().UnixNano()
}

// GetCreateUser 获取创建用户
func (a *APIMetadataDB) GetCreateUser() (user string) {
	return a.CreateUser
}

// SetCreateInfo 设置创建信息
func (a *APIMetadataDB) SetCreateInfo(user string) {
	a.CreateUser = user
	a.CreateTime = time.Now().UnixNano()
}

func (a *APIMetadataDB) GetErrMessage() string {
	return a.ErrMessage
}

// func (a *APIMetadataDB) GetFunctionContent() (code, scriptType, dependencies string) {
// 	// 不支持获取函数内容
// 	return code, scriptType, dependencies
// }
// func (a *APIMetadataDB) SetFunctionContent(code, scriptType, dependencies string) {
// 	// 不支持设置函数内容
// }

func (a *APIMetadataDB) GetCode() string {
	// 不支持获取函数内容
	return ""
}
func (a *APIMetadataDB) SetCode(code string) {
	// 不支持设置函数内容
}
func (a *APIMetadataDB) GetDependenciesURL() string {
	// 不支持获取函数内容
	return ""
}
func (a *APIMetadataDB) SetDependenciesURL(dependenciesURL string) {
	// 不支持设置函数依赖URL
}
func (a *APIMetadataDB) SetDependencies(dependencies string) {
	// 不支持设置函数依赖
}
func (a *APIMetadataDB) GetDependencies() string {
	// 不支持获取函数依赖
	return ""
}
