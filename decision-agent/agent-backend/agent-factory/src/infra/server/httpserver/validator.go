package httpserver

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/customvalidator"
)

func init() {
	// 注册自定义校验器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("checkAgentAndTplName", customvalidator.CheckAgentAndTplName)
	}
}
