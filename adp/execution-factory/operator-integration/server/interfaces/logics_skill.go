package interfaces

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
)

//go:generate mockgen -source=logics_skill.go -destination=../mocks/logics_skill.go -package=mocks

// RegisterSkillReq 注册 Skill 请求
type RegisterSkillReq struct {
	BusinessDomainID string          `header:"x-business-domain" validate:"required"`
	UserID           string          `header:"user_id" validate:"required"`
	FileType         string          `form:"file_type" validate:"required,oneof=zip content"`
	File             json.RawMessage `form:"file" validate:"required"`
	Category         BizCategory     `form:"category" default:"other_category" validate:"required"`
	Source           string          `form:"source" default:"custom" validate:"oneof=custom internal"`
	ExtendInfo       json.RawMessage `form:"extend_info"`
}

// RegisterSkillResp 注册 Skill 响应
type RegisterSkillResp struct {
	SkillID     string    `json:"skill_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Status      BizStatus `json:"status"`
	Files       []string  `json:"files"`
}

// DeleteSkillReq 删除 Skill 请求
type DeleteSkillReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id" validate:"required"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// DownloadSkillReq 下载 Skill 请求
type DownloadSkillReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// DownloadSkillResp 下载 Skill 响应
type DownloadSkillResp struct {
	SkillID  string `json:"skill_id"`
	FileName string `json:"file_name"`
	Content  []byte `json:"content"`
}

// QuerySkillListReq Skill 列表查询
type QuerySkillListReq struct {
	BusinessDomainID string      `header:"x-business-domain" validate:"required"`
	UserID           string      `header:"user_id"`
	Name             string      `form:"name"`
	Status           BizStatus   `form:"status" validate:"omitempty,oneof=unpublish published offline"`
	Category         BizCategory `form:"category"`
	CreateUser       string      `form:"create_user"`
	CommonPageParams `json:",inline"`
}

// SkillInfo Skill 详情
type SkillInfo struct {
	SkillID          string         `json:"skill_id"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Version          string         `json:"version"`
	Status           BizStatus      `json:"status"`
	Source           string         `json:"source"`
	Dependencies     map[string]any `json:"dependencies,omitempty"`
	ExtendInfo       map[string]any `json:"extend_info,omitempty"`
	CreateUser       string         `json:"create_user"`
	CreateTime       int64          `json:"create_time"`
	UpdateUser       string         `json:"update_user"`
	UpdateTime       int64          `json:"update_time"`
	Category         BizCategory    `json:"category,omitempty"`
	CategoryName     string         `json:"category_name,omitempty"`
	BusinessDomainID string         `json:"business_domain_id"`
}

// SkillSummary Skill 列表摘要
type SkillSummary struct {
	SkillID          string      `json:"skill_id"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Version          string      `json:"version"`
	Status           BizStatus   `json:"status"`
	Source           string      `json:"source"`
	CreateUser       string      `json:"create_user"`
	CreateTime       int64       `json:"create_time"`
	UpdateUser       string      `json:"update_user"`
	UpdateTime       int64       `json:"update_time"`
	BusinessDomainID string      `json:"business_domain_id"`
	Category         BizCategory `json:"category,omitempty"`
	CategoryName     string      `json:"category_name,omitempty"`
}

// SkillFileSummary Skill 文件摘要
type SkillFileSummary struct {
	RelPath  string `json:"rel_path"`
	FileType string `json:"file_type"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// QuerySkillListResp Skill 列表响应
type QuerySkillListResp struct {
	CommonPageResult `json:",inline"`
	Data             []*SkillSummary `json:"data"`
}

// QuerySkillMarketListReq Skill 市场列表查询
type QuerySkillMarketListReq struct {
	BusinessDomainID string      `header:"x-business-domain" validate:"required"`
	UserID           string      `header:"user_id"`
	Name             string      `form:"name"`
	Category         BizCategory `form:"category"`
	CreateUser       string      `form:"create_user"`
	CommonPageParams `json:",inline"`
}

// QuerySkillMarketListResp Skill 市场列表响应
type QuerySkillMarketListResp struct {
	CommonPageResult `json:",inline"`
	Data             []*SkillSummary `json:"data"`
}

// GetSkillDetailReq Skill 详情查询
type GetSkillDetailReq struct {
	BusinessDomainID string `header:"x-business-domain"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// GetSkillMarketDetailReq Skill 市场详情查询
type GetSkillMarketDetailReq struct {
	BusinessDomainID string `header:"x-business-domain"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// GetSkillContentReq Skill 内容查询
type GetSkillContentReq struct {
	BusinessDomainID string `header:"x-business-domain"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
}

// GetSkillContentResp Skill 内容响应
type GetSkillContentResp struct {
	SkillID string              `json:"skill_id"`
	URL     string              `json:"url"`
	Files   []*SkillFileSummary `json:"files"`
	Status  BizStatus           `json:"status"`
}

// ReadSkillFileReq 读取 Skill 文件请求
type ReadSkillFileReq struct {
	BusinessDomainID string `header:"x-business-domain"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
	RelPath          string `json:"rel_path" validate:"required"`
}

// ReadSkillFileResp 读取 Skill 文件响应
type ReadSkillFileResp struct {
	SkillID  string `json:"skill_id"`
	RelPath  string `json:"rel_path"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	FileType string `json:"file_type"`
}

// UpdateSkillStatusReq 更新 Skill 状态请求
type UpdateSkillStatusReq struct {
	BusinessDomainID string    `header:"x-business-domain" validate:"required"`
	UserID           string    `header:"user_id"`
	SkillID          string    `uri:"skill_id" validate:"required"`
	Status           BizStatus `json:"status" validate:"required,oneof=unpublish published offline"`
}

// UpdateSkillStatusResp 更新 Skill 状态响应
type UpdateSkillStatusResp struct {
	SkillID string    `json:"skill_id"`
	Status  BizStatus `json:"status"`
}

// ExecuteSkillReq 执行 Skill 请求
type ExecuteSkillReq struct {
	BusinessDomainID string `header:"x-business-domain"`
	UserID           string `header:"user_id"`
	SkillID          string `uri:"skill_id" validate:"required"`
	EntryShell       string `json:"entry_shell" validate:"required"`
	Timeout          int    `json:"timeout,omitempty"`
}

// ExecuteSkillResp 执行 Skill 响应
type ExecuteSkillResp struct {
	SkillID       string `json:"skill_id"`
	SessionID     string `json:"session_id"`
	WorkDir       string `json:"work_dir"`
	FileName      string `json:"file_name"`
	UploadedPath  string `json:"uploaded_path"`
	Command       string `json:"command"`
	ExitCode      int    `json:"exit_code"`
	Stdout        string `json:"stdout"`
	Stderr        string `json:"stderr"`
	ExecutionTime int64  `json:"execution_time"`
	Mocked        bool   `json:"mocked"`
}

// SkillRegistry Skill 管理接口
type SkillRegistry interface {
	RegisterSkill(ctx context.Context, req *RegisterSkillReq) (*RegisterSkillResp, error)
	DeleteSkill(ctx context.Context, req *DeleteSkillReq) error
	DownloadSkill(ctx context.Context, req *DownloadSkillReq) (*DownloadSkillResp, error)
	ExecuteSkill(ctx context.Context, req *ExecuteSkillReq) (*ExecuteSkillResp, error)
	QuerySkillList(ctx context.Context, req *QuerySkillListReq) (*QuerySkillListResp, error)
	GetSkillDetail(ctx context.Context, req *GetSkillDetailReq) (*SkillInfo, error)
	// 更新 Skill 状态
	UpdateSkillStatus(ctx context.Context, req *UpdateSkillStatusReq) (*UpdateSkillStatusResp, error)
}

// SkillMarket Skill 市场接口
type SkillMarket interface {
	QuerySkillMarketList(ctx context.Context, req *QuerySkillMarketListReq) (*QuerySkillMarketListResp, error)
	GetSkillMarketDetail(ctx context.Context, req *GetSkillMarketDetailReq) (*SkillInfo, error)
}

// SkillReader Skill 只读接口
type SkillReader interface {
	GetSkillContent(ctx context.Context, req *GetSkillContentReq) (*GetSkillContentResp, error)
	ReadSkillFile(ctx context.Context, req *ReadSkillFileReq) (*ReadSkillFileResp, error)
}

type SkillIndexSyncService interface {
	Init(ctx context.Context) error
	EnsureInitialized(ctx context.Context) error
	UpsertSkill(ctx context.Context, skill *model.SkillRepositoryDB) error
	DeleteSkill(ctx context.Context, skillID string) error
}
