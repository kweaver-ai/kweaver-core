package capimiddleware

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil/crest"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/pkg/errors"
)

// ErrorHandler gin错误处理中间件
// 【注意】：
// 1. 只有添加到gin的错误处理链中，才能生效
// 2. 只会处理第一个错误
// 3. 会清空错误
// 4. 会记录日志
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 只处理第一个错误
		if len(c.Errors) > 0 {
			err := c.Errors[0]

			_err := err.Unwrap()

			// 1. 响应错误
			crest.ReplyError2(c, _err)

			c.Errors = c.Errors[0:0] // 清空错误

			// 2. 记录日志
			_logger := logger.GetLogger()

			// 2.1 debug模式下，打印错误信息
			if cenvhelper.IsDebugMode() {
				errCause := fmt.Sprintf("%v", errors.Cause(err))
				errTrace := fmt.Sprintf("%+v", err)

				errMSg := fmt.Sprintf("[ErrorHandler][ErrMsg]: \nerror cause: %v \n err trace: %+v\n", errCause, errTrace)

				log.Print(errMSg)

				_ = cutil.PrintFormatJSON(_err, "[ErrorHandler][PrintFormatJSON]: request error")
			}

			// 2.2 根据响应状态码记录不同级别的日志
			if c.Writer.Status() >= 500 {
				_logger.Errorf("[ErrorHandler][error]: %v", err)
			} else {
				_logger.Warnf("[ErrorHandler][warning]: %v", err)
			}
		}
	}
}
