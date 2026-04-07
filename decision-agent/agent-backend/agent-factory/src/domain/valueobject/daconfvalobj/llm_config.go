package daconfvalobj

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/pkg/errors"
)

// LlmConfig 表示大模型配置
type LlmConfig struct {
	ID   string `json:"id"`                      // 模型ID，预留参数，目前无用
	Name string `json:"name" binding:"required"` // 模型名

	ModelType        cdaenum.ModelType `json:"model_type"`                               // 模型类型
	Temperature      float64           `json:"temperature" binding:"gte=0,lte=2"`        // 温度参数
	TopP             float64           `json:"top_p" binding:"gte=0,lte=1"`              // Top-p参数
	TopK             int               `json:"top_k" binding:"gte=0"`                    // Top-k参数
	FrequencyPenalty float64           `json:"frequency_penalty" binding:"gte=-2,lte=2"` // 频率惩罚（-2~2）
	PresencePenalty  float64           `json:"presence_penalty" binding:"gte=-2,lte=2"`  // 存在惩罚（-2~2）
	MaxTokens        int               `json:"max_tokens" binding:"required,gte=0"`      // 最大token数

	// RetrievalMaxTokens int `json:"retrieval_max_tokens" binding:"gt=0"` // 召回的最大token数量（单位K）
}

// Validate 对 LlmConfig 进行参数校验
func (c *LlmConfig) Validate() (err error) {
	// 获取验证器引擎
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		// 如果验证器引擎类型不正确，直接抛出panic
		panic("binding.Validator.Engine() is not *validator.Validate")
	}

	// 使用验证器对结构体进行验证
	err = v.Struct(c)
	if err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[LlmConfig] invalid")
		return
	}

	return
}

func (c *LlmConfig) ValObjCheck() (err error) {
	// 1. 检查MaxTokens是否为0
	if c.MaxTokens == 0 {
		c.MaxTokens = 500
	}

	// 2. 验证LlmConfig的有效性
	if err = c.Validate(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[LlmConfig]: Validate failed")
		return
	}

	// 3. 检查ModelType
	if c.ModelType == "" {
		c.ModelType = cdaenum.ModelTypeLlm
	}

	if err = c.ModelType.EnumCheck(); err != nil {
		err = errors.Wrap(err, "[LlmConfig]: ModelType is invalid")
		return
	}

	return
}
