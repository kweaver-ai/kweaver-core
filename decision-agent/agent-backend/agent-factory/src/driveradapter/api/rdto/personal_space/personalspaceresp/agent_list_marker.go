package personalspaceresp

import (
	"encoding/base64"
	"encoding/json"
)

type PAListPaginationMarker struct {
	UpdatedAt   int64  `json:"updated_at"`
	LastAgentID string `json:"last_agent_id"`
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
