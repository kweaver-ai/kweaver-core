package pubedresp

import (
	"encoding/base64"
	"encoding/json"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

type PAListPaginationMarker struct {
	PublishedAt   int64  `json:"published_at"`
	LastReleaseID string `json:"last_release_id"`
}

func NewPAListPaginationMarker() *PAListPaginationMarker {
	return &PAListPaginationMarker{}
}

func (m *PAListPaginationMarker) ToString() (str string, err error) {
	// json and to base64
	jsonStr, err := json.Marshal(m)
	if err != nil {
		return
	}

	str = base64.StdEncoding.EncodeToString(jsonStr)

	return
}

func (m *PAListPaginationMarker) LoadFromStr(str string) (err error) {
	if str == "" {
		return
	}

	jsonStr, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return
	}

	err = json.Unmarshal(jsonStr, m)
	if err != nil {
		return
	}

	return
}

func (m *PAListPaginationMarker) LoadFromPos(pos []*dapo.PublishedJoinPo) {
	if len(pos) == 0 {
		return
	}

	// 1. 取最后一个
	lastItem := pos[len(pos)-1]

	// 2. 设置 marker
	m.PublishedAt = lastItem.PublishedAt
	m.LastReleaseID = lastItem.ReleaseID
}
