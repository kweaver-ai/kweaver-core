package releasereq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/publishvo"
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
	// 校验发布目标
	for _, target := range req.PublishToWhere {
		if err = target.EnumCheck(); err != nil {
			err = errors.Wrap(err, "[UpdatePublishInfoReq]: publish_to_where is invalid")
			return
		}
	}

	// 校验发布为标识
	for _, target := range req.PublishToBes {
		if err = target.EnumCheck(); err != nil {
			err = errors.Wrap(err, "[UpdatePublishInfoReq]: publish_to_bes is invalid")
			return
		}
	}

	// 如果发布目标包含custom_space，则必须提供custom_space_ids
	// for _, target := range req.PublishToWhere {
	//	if target == daenum.PublishToWhereCustomSpace && len(req.CustomSpaceIDs) == 0 {
	//		err = errors.New("[UpdatePublishInfoReq]: custom_space_ids is required when publish_to_where contains custom_space")
	//		return
	//	}
	//}

	return
}
