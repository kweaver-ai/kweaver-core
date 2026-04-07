package dapo

import (
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

type ListForBenchmarkPo struct {
	ID        string         `json:"id" db:"f_id"`
	Key       string         `json:"key" db:"f_key"`
	Name      string         `json:"name" db:"f_name"`
	Version   string         `json:"version" db:"f_version"`
	Status    cdaenum.Status `json:"status" db:"f_status"`
	UpdatedAt int64          `json:"updated_at" db:"updated_at"`
}

type ListForBenchmarkAgentPo struct {
	ID        string         `json:"id" db:"f_id"`
	Key       string         `json:"key" db:"f_key"`
	Name      string         `json:"name" db:"f_name"`
	UpdatedAt int64          `json:"updated_at" db:"f_updated_at"`
	Version   string         `json:"version" db:"f_version"`
	Status    cdaenum.Status `json:"status" db:"f_status"`
}

type ListForBenchmarkReleasePo struct {
	ID        sql.NullString `json:"id" db:"f_agent_id"`
	Key       sql.NullString `json:"key" db:"f_agent_key"`
	Name      sql.NullString `json:"name" db:"f_agent_name"`
	UpdatedAt sql.NullInt64  `json:"updated_at" db:"f_published_at"`
	Version   sql.NullString `json:"version" db:"f_release_version"`
	Status    sql.NullString `json:"status" db:"f_release_status"`
}

func (rp *ListForBenchmarkReleasePo) ToListForBenchmarkPo() (po *ListForBenchmarkPo) {
	po = &ListForBenchmarkPo{
		ID:        rp.ID.String,
		Key:       rp.Key.String,
		Name:      rp.Name.String,
		Version:   rp.Version.String,
		Status:    cdaenum.Status(rp.Status.String),
		UpdatedAt: rp.UpdatedAt.Int64,
	}

	return
}

//nolint:govet // 嵌入结构体存在重复 json tag，暂不修改以避免影响现有逻辑
type ListForBenchmarkMerge struct {
	ListForBenchmarkAgentPo
	ListForBenchmarkReleasePo
}
