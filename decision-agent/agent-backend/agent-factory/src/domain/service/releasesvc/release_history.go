package releasesvc

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/release/releaseresp"
	"github.com/pkg/errors"
)

// GetPublishHistoryInfo implements iv3portdriver.IReleaseSvc.
func (r *releaseSvc) GetPublishHistoryInfo(ctx context.Context, req interface{}) (res string, err error) {
	return "", nil
}

// GetPublishHistoryList implements iv3portdriver.IReleaseSvc.
func (r *releaseSvc) GetPublishHistoryList(ctx context.Context, agentID string) (res releaseresp.HistoryListResp, total int64, err error) {
	poList, total, err := r.releaseHistoryRepo.ListByAgentID(ctx, agentID)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "get release history by agent id %s", agentID)
	}

	res = make(releaseresp.HistoryListResp, 0, len(poList))

	for _, po := range poList {
		res = append(res, releaseresp.HistoryListItemResp{
			AgentId:      po.AgentID,
			AgentVersion: po.AgentVersion,
			AgentDesc:    po.AgentDesc,
			CreateTime:   po.CreateTime,
			HistoryId:    po.ID,
		})
	}

	// 按版本号倒序排序
	sort.SliceStable(res, func(i, j int) bool {
		return compareVersion(res[i].AgentVersion, res[j].AgentVersion) > 0
	})

	return
}

// compareVersion 比较两个版本号，返回值大于0表示v1>v2，等于0表示相等，小于0表示v1<v2
func compareVersion(v1, v2 string) int {
	// 去掉版本号前的'v'字符
	version1 := strings.TrimPrefix(v1, "v")
	version2 := strings.TrimPrefix(v2, "v")

	// 转换为整数进行比较
	num1, err1 := strconv.Atoi(version1)
	num2, err2 := strconv.Atoi(version2)

	// 如果都能成功转换为整数，则按整数比较
	if err1 == nil && err2 == nil {
		return num1 - num2
	}

	// 如果不能转换为整数，则按字符串比较
	if version1 > version2 {
		return 1
	} else if version1 < version2 {
		return -1
	}

	return 0
}
