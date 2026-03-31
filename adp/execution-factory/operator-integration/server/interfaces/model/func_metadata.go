package model

import (
	"context"
	"database/sql"
	"time"
)

// FunctionMetadataDB 函数元数据数据库
type FunctionMetadataDB struct {
	ID              int64  `json:"f_id" db:"f_id"`
	Summary         string `json:"f_summary" db:"f_summary"`
	Version         string `json:"f_version" db:"f_version"`
	Description     string `json:"f_description" db:"f_description"`
	Path            string `json:"f_path" db:"f_path"`
	ServerURL       string `json:"f_svc_url" db:"f_svc_url"`
	Method          string `json:"f_method" db:"f_method"`
	APISpec         string `json:"f_api_spec" db:"f_api_spec"`
	CreateUser      string `json:"f_create_user" db:"f_create_user"`
	CreateTime      int64  `json:"f_create_time" db:"f_create_time"`
	UpdateUser      string `json:"f_update_user" db:"f_update_user"`
	UpdateTime      int64  `json:"f_update_time" db:"f_update_time"`
	ScriptType      string `json:"f_script_type" db:"f_script_type"`   // 脚本类型，如 Python、Node.js
	Code            string `json:"f_code" db:"f_code"`                 // 函数代码
	Dependencies    string `json:"f_dependencies" db:"f_dependencies"` // 第三方库依赖，如 Python 中的 requests 库
	DependenciesURL string `json:"f_dependencies_url" db:"f_dependencies_url"`
	ErrMessage      string `json:"-"` // 错误信息
}

// IFunctionMetadataDB 函数元数据数据库接口
type IFunctionMetadataDB interface {
	InsertFuncMetadata(ctx context.Context, tx *sql.Tx, metadata *FunctionMetadataDB) (version string, err error)
	SelectByVersion(ctx context.Context, version string) (exist bool, metadata *FunctionMetadataDB, err error)
	UpdateByVersion(ctx context.Context, tx *sql.Tx, metadata *FunctionMetadataDB) error
	DeleteByVersion(ctx context.Context, tx *sql.Tx, version string) error
	DeleteByVersions(ctx context.Context, tx *sql.Tx, versions []string) error
	InsertFuncMetadatas(ctx context.Context, tx *sql.Tx, metadatas []*FunctionMetadataDB) (versions []string, err error)
	SelectListByVersion(ctx context.Context, versions []string) ([]*FunctionMetadataDB, error)
}

// GetType 获取资源类型
func (f *FunctionMetadataDB) GetType() string {
	return string(SourceTypeFunction)
}

// GetSummary 获取摘要
func (f *FunctionMetadataDB) GetSummary() string {
	return f.Summary
}

// GetDescription 获取函数描述
func (f *FunctionMetadataDB) GetDescription() string {
	if f.Description == "" {
		return f.Summary
	}
	return f.Description
}

// GetVersion 获取版本
func (f *FunctionMetadataDB) GetVersion() string {
	return f.Version
}
func (f *FunctionMetadataDB) GetMethod() string {
	return f.Method
}
func (f *FunctionMetadataDB) GetPath() string {
	return f.Path
}

func (f *FunctionMetadataDB) GetScriptType() string {
	return f.ScriptType
}

func (f *FunctionMetadataDB) Validate(ctx context.Context) error {
	return nil
}

func (f *FunctionMetadataDB) SetSummary(summary string) {
	f.Summary = summary
}
func (f *FunctionMetadataDB) SetDescription(description string) {
	f.Description = description
}

func (f *FunctionMetadataDB) SetMethod(method string) {
	f.Method = method
}
func (f *FunctionMetadataDB) SetPath(path string) {
	f.Path = path
}
func (f *FunctionMetadataDB) SetVersion(version string) {
	f.Version = version
}
func (f *FunctionMetadataDB) SetScriptType(scriptType string) {
	f.ScriptType = scriptType
}

func (f *FunctionMetadataDB) GetServerURL() string {
	return f.ServerURL
}
func (f *FunctionMetadataDB) SetServerURL(serverURL string) {
	f.ServerURL = serverURL
}

// GetAPISpec 获取API规范
func (f *FunctionMetadataDB) GetAPISpec() string {
	return f.APISpec
}

func (f *FunctionMetadataDB) SetAPISpec(apiSpec string) {
	f.APISpec = apiSpec
}

// GetUpdateUser 获取更新用户
func (f *FunctionMetadataDB) GetUpdateUser() (user string) {
	return f.UpdateUser
}

func (f *FunctionMetadataDB) SetUpdateInfo(user string) {
	f.UpdateUser = user
	f.UpdateTime = time.Now().UnixNano()
}

// GetCreateUser 获取创建用户
func (f *FunctionMetadataDB) GetCreateUser() (user string) {
	return f.CreateUser
}

// SetCreateInfo 设置创建信息
func (f *FunctionMetadataDB) SetCreateInfo(user string) {
	f.CreateUser = user
	f.CreateTime = time.Now().UnixNano()
}

func (f *FunctionMetadataDB) GetErrMessage() string {
	return f.ErrMessage
}

func (f *FunctionMetadataDB) GetCode() string {
	return f.Code
}
func (f *FunctionMetadataDB) SetCode(code string) {
	f.Code = code
}

func (f *FunctionMetadataDB) GetDependenciesURL() string {
	return f.DependenciesURL
}
func (f *FunctionMetadataDB) SetDependenciesURL(dependenciesURL string) {
	f.DependenciesURL = dependenciesURL
}
func (f *FunctionMetadataDB) SetDependencies(dependencies string) {
	f.Dependencies = dependencies
}
func (f *FunctionMetadataDB) GetDependencies() string {
	if f.Dependencies == "null" {
		return ""
	}
	return f.Dependencies
}
