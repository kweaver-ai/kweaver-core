package datasourcevalobj

import (
	"github.com/pkg/errors"
)

type RetrieverAdvancedConfig struct {
	KG  *KGAdvancedConfig  `json:"kg"`
	Doc *DocAdvancedConfig `json:"doc"`
}

func NewRetrieverAdvancedConfig() (c *RetrieverAdvancedConfig) {
	return &RetrieverAdvancedConfig{}
}

func (c *RetrieverAdvancedConfig) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{}
}

func (c *RetrieverAdvancedConfig) ValObjCheck() (err error) {
	if c.KG != nil {
		err = c.KG.ValObjCheck()
		if err != nil {
			err = errors.Wrap(err, "[RetrieverAdvancedConfig]: kg is invalid")
			return
		}
	}

	if c.Doc != nil {
		err = c.Doc.ValObjCheck()
		if err != nil {
			err = errors.Wrap(err, "[RetrieverAdvancedConfig]: doc is invalid")
			return
		}
	}

	return
}
