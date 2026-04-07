package utils

import (
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
)

var (
	errorConfig = NewConfiger()
)

// NewECronError 生成定时服务错误
func NewECronError(cause string, code int, detail map[string]interface{}) (err *common.ECronError) {
	message := ""
	switch code {
	case common.BadRequest:
		message = errorConfig.Lang().GetString("IDS_BAD_REQUEST")
	case common.Unauthorized:
		message = errorConfig.Lang().GetString("IDS_UNAUTHORIZED")
	case common.NotFound:
		message = errorConfig.Lang().GetString("IDS_NOT_FOUND")
	case common.TooManyRequests:
		message = errorConfig.Lang().GetString("IDS_TOO_MANY_REQUESTS")
	case common.InternalError:
		message = errorConfig.Lang().GetString("IDS_INTERNAL_ERROR")
	}

	err = &common.ECronError{
		Cause:   cause,
		Code:    code,
		Message: message,
		Detail:  detail,
	}
	return
}
