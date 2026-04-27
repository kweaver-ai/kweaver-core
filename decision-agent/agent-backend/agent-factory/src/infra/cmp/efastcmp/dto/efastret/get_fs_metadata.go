package efastret

import "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"

type FsMetadata struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	DocLibType cenum.DocLibType `json:"doc_lib_type"`
	Path       string           `json:"path"`
	Size       int64            `json:"size"`
}

// GetFsMetadataRetDto 响应dto
type GetFsMetadataRetDto []*FsMetadata
