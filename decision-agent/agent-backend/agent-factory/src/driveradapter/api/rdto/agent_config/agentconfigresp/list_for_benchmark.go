package agentconfigresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

type ListForBenchmarkItem struct {
	dapo.ListForBenchmarkPo
}

type ListForBenchmarkResp struct {
	Entries []*ListForBenchmarkItem `json:"entries"`
	Total   int64                   `json:"total"`
}

func NewListForBenchmarkResp() *ListForBenchmarkResp {
	return &ListForBenchmarkResp{
		Entries: make([]*ListForBenchmarkItem, 0),
	}
}

func (l *ListForBenchmarkResp) LoadFromListForBenchmark(pos []*dapo.ListForBenchmarkPo) (err error) {
	for _, po := range pos {
		item := &ListForBenchmarkItem{}

		err = cutil.CopyStructUseJSON(item, po)
		if err != nil {
			return
		}

		l.Entries = append(l.Entries, item)
	}

	return
}
