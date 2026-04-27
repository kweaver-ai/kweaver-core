package interfaces

import "context"

//
//go:generate mockgen -source=dbaccess_metadata.go -destination=../mocks/dbaccess_metadata.go -package=mocks

// IMetadataDB 元数据通用接口
type IMetadataDB interface {
	GetType() string
	GetSummary() string
	SetSummary(summary string)
	GetDescription() string
	SetDescription(description string)
	GetVersion() string
	SetVersion(version string)
	GetServerURL() string
	SetServerURL(serverURL string)
	GetAPISpec() string
	SetAPISpec(apiSpec string)
	GetMethod() string
	SetMethod(method string)
	GetPath() string
	SetPath(path string)
	Validate(ctx context.Context) error
	GetUpdateUser() (user string)
	SetUpdateInfo(user string)
	GetCreateUser() (user string)
	SetCreateInfo(user string)
	// UpdataMetadata(metadata interface{}) error
	// 获取ErrMessage信息
	GetErrMessage() string
	GetCode() string
	SetCode(code string)
	GetScriptType() string
	SetScriptType(scriptType string)
	GetDependenciesURL() string
	SetDependenciesURL(dependenciesURL string)
	SetDependencies(dependencies string)
	GetDependencies() string
}
