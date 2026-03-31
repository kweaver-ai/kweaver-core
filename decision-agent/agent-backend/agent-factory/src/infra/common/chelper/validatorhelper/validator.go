package validatorhelper

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/validatorhelper/persrecvalid"
)

// CommonCustomValidator 注册自定义common校验器
func CommonCustomValidator() (err error) {
	// 注册自定义校验器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err = v.RegisterValidation("persCheckName", persrecvalid.CheckName)
		if err != nil {
			return
		}

		err = v.RegisterValidation("persCheckCode", persrecvalid.CheckCode)
		if err != nil {
			return
		}
	}

	return
}
