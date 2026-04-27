package releasereq

import (
	"strings"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/publishvo"
	"github.com/pkg/errors"
)

// UpdatePublishInfoReq 更新发布信息请求
type UpdatePublishInfoReq struct {
	publishvo.PublishInfo
}

// GetErrMsgMap 获取错误信息映射
func (req *UpdatePublishInfoReq) GetErrMsgMap() map[string]string {
	return map[string]string{}
}

// CustomCheck 自定义参数校验
func (req *UpdatePublishInfoReq) CustomCheck() (err error) {
	if req == nil {
		return errors.New("[UpdatePublishInfoReq]: request is required")
	}

	categoryIDs := make([]string, 0, len(req.CategoryIDs))

	for _, categoryID := range req.CategoryIDs {
		categoryID = strings.TrimSpace(categoryID)
		if categoryID == "" {
			continue
		}

		categoryIDs = append(categoryIDs, categoryID)
	}

	req.CategoryIDs = categoryIDs

	publishToWhere := make([]daenum.PublishToWhere, 0, len(req.PublishToWhere))

	// 校验发布目标
	for _, target := range req.PublishToWhere {
		if err = target.WriteEnumCheck(); err != nil {
			err = errors.Wrap(err, "[UpdatePublishInfoReq]: publish_to_where is invalid")
			return
		}

		publishToWhere = append(publishToWhere, target)
	}

	req.PublishToWhere = publishToWhere

	// 校验发布为标识
	for _, target := range req.PublishToBes {
		if err = target.EnumCheck(); err != nil {
			err = errors.Wrap(err, "[UpdatePublishInfoReq]: publish_to_bes is invalid")
			return
		}
	}

	return
}
