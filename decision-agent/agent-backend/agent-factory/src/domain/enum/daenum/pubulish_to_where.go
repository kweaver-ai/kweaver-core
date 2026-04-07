package daenum

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// PublishToWhere 发布到
type PublishToWhere string

const (
	// PublishToWhereCustomSpace 自定义空间
	PublishToWhereCustomSpace PublishToWhere = "custom_space"

	// PublishToWhereSquare 广场
	PublishToWhereSquare PublishToWhere = "square"
)

func (b PublishToWhere) EnumCheck() (err error) {
	if !cutil.ExistsGeneric([]PublishToWhere{PublishToWhereCustomSpace, PublishToWhereSquare}, b) {
		err = errors.New("[PublishToWhere]: invalid publish to where")
		return
	}

	return
}
