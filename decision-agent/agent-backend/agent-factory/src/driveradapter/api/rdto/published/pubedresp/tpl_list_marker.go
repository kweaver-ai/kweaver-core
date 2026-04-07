package pubedresp

import (
	"encoding/base64"
	"encoding/json"
)

type PTplListPaginationMarker struct {
	LastPubedTplID int64 `json:"last_pubed_tpl_id"`
}

func NewPTplListPaginationMarker() *PTplListPaginationMarker {
	return &PTplListPaginationMarker{}
}

func (m *PTplListPaginationMarker) ToString() (str string, err error) {
	// json and to base64
	jsonStr, err := json.Marshal(m)
	if err != nil {
		return
	}

	str = base64.StdEncoding.EncodeToString(jsonStr)

	return
}

func (m *PTplListPaginationMarker) LoadFromStr(str string) (err error) {
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
