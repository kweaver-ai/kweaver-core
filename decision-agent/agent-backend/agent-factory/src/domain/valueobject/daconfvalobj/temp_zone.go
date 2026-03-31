package daconfvalobj

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/pkg/errors"
)

// TempZoneConfig 表示临时区配置
type TempZoneConfig struct {
	Name string `json:"name"` // 区域名称

	TmpFileUseType cdaenum.TmpFileUseType `json:"tmp_file_use_type" binding:"required"` // 临时文件使用类型

	MaxFileCount *int `json:"max_file_count"` // 最大文件数量

	SingleChatMaxSelectFileCount *int `json:"single_chat_max_select_file_count"` // 单次对话中支持选择的文件个数

	SingleFileSizeLimit     int             `json:"single_file_size_limit" binding:"required"`      // 单文件大小限制
	SingleFileSizeLimitUnit cdaenum.BitUnit `json:"single_file_size_limit_unit" binding:"required"` // 单文件大小限制单位

	SupportDataType cdaenum.SupportDataTypes `json:"support_data_type" binding:"required"` // 支持的数据类型

	AllowedFileTypes []string `json:"allowed_file_types"` // 允许的文件类型。前端不需要传，根据allowed_file_categories自动生成

	AllowedFileCategories cdaenum.AllowedFileCategories `json:"allowed_file_categories" binding:"required"` // 允许的文件类别
}

func (p *TempZoneConfig) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		//"Name.required":                    `"name"不能为空`,
		"TmpFileUseType.required": `"tmp_file_use_type"不能为空`,
		//"MaxFileCount.required":            `"max_file_count"不能为空`,
		//"MaxFileCount.max":                 `"max_file_count"最大值为50`,
		//"MaxFileCount.min":                 `"max_file_count"最小值为1`,
		//"SingleChatMaxSelectFileCount.max": `"single_chat_max_select_file_count"最大值为5`,
		//"SingleChatMaxSelectFileCount.min": `"single_chat_max_select_file_count"最小值为1`,
		"SingleFileSizeLimit.required":     `"single_file_size_limit"不能为空`,
		"SingleFileSizeLimitUnit.required": `"single_file_size_limit_unit"不能为空`,
		"SupportDataType.required":         `"support_data_type"不能为空`,
		"AllowedFileCategories.required":   `"allowed_file_categories"不能为空`,
	}
}

// Validate 对 TempZoneConfig 进行参数校验
func (p *TempZoneConfig) Validate() (err error) {
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
		err = errors.Wrap(err, "[TempZoneConfig] invalid")
		return
	}

	return
}

func (p *TempZoneConfig) ValObjCheck() (err error) {
	// 1. 使用验证器进行基本参数校验
	if err = p.Validate(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[TempZoneConfig]: invalid")
		return
	}

	// 2. 校验临时文件使用类型
	if err = p.TmpFileUseType.EnumCheck(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[TempZoneConfig]: tmp_file_use_type is invalid")
		return
	}

	// 3. 校验支持的数据类型枚举值是否有效
	if err = p.SupportDataType.EnumCheck(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[TempZoneConfig]: support_data_type is invalid")
		return
	}

	// 4. 校验允许的文件类别枚举值是否有效
	if err = p.AllowedFileCategories.EnumCheck(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[TempZoneConfig]: allowed_file_categories is invalid")
		return
	}

	// 5. 校验单文件大小限制
	// 单文件最大100M，先根据单位计算实际字节数
	var actualSizeInBytes int64

	switch p.SingleFileSizeLimitUnit {
	case cdaenum.KB:
		// 如果单位是KB，将大小乘以1024转换为字节
		actualSizeInBytes = int64(p.SingleFileSizeLimit) * 1024
	case cdaenum.MB:
		// 如果单位是MB，将大小乘以1024*1024转换为字节
		actualSizeInBytes = int64(p.SingleFileSizeLimit) * 1024 * 1024
	case cdaenum.GB:
		// 如果单位是GB，将大小乘以1024*1024*1024转换为字节
		actualSizeInBytes = int64(p.SingleFileSizeLimit) * 1024 * 1024 * 1024
	default:
		// 如果单位不是KB、MB或GB，则返回错误
		err = errors.New("[TempZoneConfig]: invalid single_file_size_limit_unit")
		return
	}

	// 判断计算出的实际大小是否超出系统允许的最大限制（100MB）
	maxAllowedSizeInBytes := int64(100) * 1024 * 1024
	if actualSizeInBytes > maxAllowedSizeInBytes {
		// 如果超出限制，返回带有详细信息的错误
		err = errors.Errorf("[TempZoneConfig]: single file size limit exceeds maximum allowed (100MB), current: %d bytes", actualSizeInBytes)
		return
	}

	// 6. 验证MaxFileCount和SingleChatMaxSelectFileCount的范围
	if p.MaxFileCount != nil && (*p.MaxFileCount < 1 || *p.MaxFileCount > 50) {
		err = errors.New("[TempZoneConfig]: max_file_count must be between 1 and 50")
		return
	}

	if p.SingleChatMaxSelectFileCount != nil && (*p.SingleChatMaxSelectFileCount < 1 || *p.SingleChatMaxSelectFileCount > 5) {
		err = errors.New("[TempZoneConfig]: single_chat_max_select_file_count must be between 1 and 5")
		return
	}

	return
}

func (p *TempZoneConfig) GenAllowedFileTypes() (err error) {
	if p.AllowedFileCategories == nil {
		// 如果允许的文件类别为空，直接返回nil
		return
	}

	// 根据允许的文件类别生成允许的文件类型
	allowedFileTypes, err := p.AllowedFileCategories.GetAllowedFileTypes()
	if err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[TempZoneConfig]: allowed_file_categories is invalid")
		return
	}

	// 将生成的文件类型赋值给结构体字段
	p.AllowedFileTypes = allowedFileTypes

	// 返回nil表示没有错误
	return
}
