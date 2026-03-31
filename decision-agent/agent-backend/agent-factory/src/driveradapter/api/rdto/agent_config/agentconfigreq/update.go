package agentconfigreq

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/customvalidator"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// UpdateReq 表示更新agent的请求
type UpdateReq struct {
	Name    string `json:"name" binding:"required,checkAgentAndTplName,max=50"` // 名字
	Profile string `json:"profile" binding:"required,max=500"`                  // 简介

	AvatarType cdaenum.AvatarType `json:"avatar_type" binding:"required"` // 头像类型: 1-内置头像, 2-用户上传头像, 3-AI生成头像
	Avatar     string             `json:"avatar" binding:"required"`      // 头像信息

	ProductKey string `json:"product_key" binding:"required"` // 所属产品标识

	Config *daconfvalobj.Config `json:"config" binding:"required"` // agent配置

	CreatedBy string `json:"created_by"` // 创建人
	UpdatedBy string `json:"updated_by"` // 更新人

	IsBuiltIn *cdaenum.BuiltIn `json:"is_built_in"` // 是否内置

	IsInternalAPI bool `json:"-"` // 是否是内部接口api。此字段不是前端传入，后端自动设置
}

func (p *UpdateReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Name.required":             `"name"不能为空`,
		"Name.checkAgentAndTplName": customvalidator.GenAgentAndTplNameErrMsg(`"Agent名称"`),
		"Name.max":                  `"name"长度不能超过50`,
		"AvatarType.required":       `"avatar_type"不能为空`,
		"Avatar.required":           `"avatar"不能为空`,
		"Config.required":           `"config"不能为空`,
		"Profile.required":          `"profile"不能为空`,
		"Profile.max":               `"profile"长度不能超过500`,
		"ProductKey.required":       `"product_key"不能为空`,
	}
}

func (p *UpdateReq) D2e() (eo *daconfeo.DataAgent, err error) {
	// 1. 生成allowed_file_types
	err = HandleConfig(p.Config)
	if err != nil {
		err = errors.Wrap(err, "[UpdateReq]: HandleConfig failed")
		return
	}

	// 2 . dto to eo
	eo = &daconfeo.DataAgent{}

	err = cutil.CopyStructUseJSON(eo, p)
	if err != nil {
		return
	}

	// 3. d2e后处理
	D2eCommonAfterD2e(eo)

	return
}

func (p *UpdateReq) ReqCheckWithCtx(ctx context.Context) (err error) {
	// 1. 验证avatar_type
	if err = p.AvatarType.EnumCheck(); err != nil {
		err = errors.Wrap(err, "[UpdateReq]: avatar_type is invalid")
		return
	}

	// 2. 验证config
	if p.Config != nil {
		if err = p.Config.ValObjCheckWithCtx(ctx, p.IsInternalAPI); err != nil {
			err = errors.Wrap(err, "[UpdateReq]: config is invalid")
			return
		}
	}

	// 3. 验证product_key
	if p.ProductKey == "" {
		err = errors.New("[UpdateReq]: product_key is required")
		return
	}
	// 4. 验证数据源与产品类型是否相符
	err = p.Config.CheckProductAndDataSource(cdaenum.Product(p.ProductKey))
	if err != nil {
		err = errors.Wrap(err, "[UpdateReq]: data source is invalid")
		return
	}

	return
}

func (p *UpdateReq) CustomCheck() (err error) {
	// 1. 验证is_private_api相关
	if !p.IsInternalAPI {
		if p.UpdatedBy != "" {
			err = errors.New("[UpdateReq]: updated_by is valid when is_private_api is false")
		}
	} else {
		if p.UpdatedBy == "" {
			err = errors.New("[UpdateReq]: updated_by is required when is_private_api is true")
		}
	}

	return
}

func (p *UpdateReq) IsChanged(oldPo *dapo.DataAgentPo) (isChanged bool) {
	newConfigStr, err := cutil.JSON().MarshalToString(p.Config)
	if err != nil {
		return
	}

	isChangedConfig, err := p.IsConfigChanged(oldPo.Config, newConfigStr)
	if err != nil {
		return
	}

	isChanged = oldPo.Name != p.Name || oldPo.GetProfileStr() != p.Profile || oldPo.AvatarType != p.AvatarType || oldPo.Avatar != p.Avatar || oldPo.ProductKey != p.ProductKey || isChangedConfig

	if cenvhelper.IsDebugMode() && isChanged {
		var changedFields []string
		if oldPo.Name != p.Name {
			changedFields = append(changedFields, fmt.Sprintf("Name: [%s] -> [%s]", oldPo.Name, p.Name))
		}

		if oldPo.GetProfileStr() != p.Profile {
			changedFields = append(changedFields, fmt.Sprintf("Profile: [%s] -> [%s]", oldPo.GetProfileStr(), p.Profile))
		}

		if oldPo.AvatarType != p.AvatarType {
			changedFields = append(changedFields, fmt.Sprintf("AvatarType: [%v] -> [%v]", oldPo.AvatarType, p.AvatarType))
		}

		if oldPo.Avatar != p.Avatar {
			changedFields = append(changedFields, fmt.Sprintf("Avatar: [%s] -> [%s]", oldPo.Avatar, p.Avatar))
		}

		if oldPo.ProductKey != p.ProductKey {
			changedFields = append(changedFields, fmt.Sprintf("ProductKey: [%s] -> [%s]", oldPo.ProductKey, p.ProductKey))
		}

		if isChangedConfig {
			changedFields = append(changedFields, fmt.Sprintf("Config: [%s] -> [%s]", oldPo.Config, newConfigStr))
		}

		fmt.Printf("[IsChanged] changed fields: %v\n", changedFields)
	}

	return
}

// Validate 对 UpdateReq 进行参数校验
func (p *UpdateReq) Validate() (err error) {
	// 获取验证器引擎
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		// 如果验证器引擎类型不正确，直接抛出panic
		panic("binding.Validator.Engine() is not *validator.Validate")
	}

	// 使用验证器对结构体进行验证
	err = v.Struct(p)
	if err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[UpdateReq] invalid")
		return
	}

	return
}

func (p *UpdateReq) IsConfigChanged(oldConfigJSONStr, newConfigJSONStr string) (isChanged bool, err error) {
	metadataKey := "metadata"

	oldStr, err := cutil.RemoveKeyFromJSON(oldConfigJSONStr, metadataKey)
	if err != nil {
		return
	}

	newStr, err := cutil.RemoveKeyFromJSON(newConfigJSONStr, metadataKey)
	if err != nil {
		return
	}

	// 移除config_metadata后，再比较
	isChanged = oldStr != newStr

	return
}
