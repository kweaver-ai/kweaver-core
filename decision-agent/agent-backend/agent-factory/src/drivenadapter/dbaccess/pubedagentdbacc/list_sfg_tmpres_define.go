//nolint:unused // 预留给 sfg 优化使用
package pubedagentdbacc

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

var ErrSfgTmpResTypeNotMatch = errors.New("[sfg] tmp res type not match")

// 临时存储sfg产生的val
// 流程：
// 1. 将sfg包裹的内容转换为sfgTmpResPubList，然后返回
// 2. 通过LoadFromValByCopyWithJSON方法，将sfg返回的val转换回sfgTmpResPubList
// 作用：通过json的CopyStructUseJSON方法，实现sfgTmpResPubList的深拷贝，让每个goroutine都能拥有自己的sfgTmpResPubList实例，避免可能的并发问题
type sfgTmpResPubList struct {
	Pos []*dapo.PublishedJoinPo `json:"pos"`
}

func (r *sfgTmpResPubList) LoadFromValByCopyWithJSON(val interface{}) (err error) {
	v, ok := val.(*sfgTmpResPubList)
	if !ok {
		err = ErrSfgTmpResTypeNotMatch
		return
	}

	err = cutil.CopyStructUseJSON(r, v)

	return
}
