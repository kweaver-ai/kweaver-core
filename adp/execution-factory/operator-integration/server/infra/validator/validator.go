// Package validator 定义接口
// @file validator.go
// @description: 初始化验证器
package validator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/asaskevich/govalidator"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	myErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

const (
	defaultNameMaxLength = 50 // 算子名称最大长度(字符)
)

var (
	// 验证类型
	validParameterTypes = map[interfaces.ParameterType]bool{
		interfaces.ParameterTypeString:  true,
		interfaces.ParameterTypeNumber:  true,
		interfaces.ParameterTypeBoolean: true,
		interfaces.ParameterTypeArray:   true,
		interfaces.ParameterTypeObject:  true,
	}
)

// Validator 验证器接口
type validator struct {
	Validator           *validatorv10.Validate
	ImportMaxCount      int64 // 算子导入限制(单次导入最大算子数量)
	NameLimit           int64 // 算子名称限制
	DescLimit           int64 // 算子描述限制
	ImportFileSizeLimit int64 // 算子导入限制(单次导入最大文件大小)
}

var (
	vOnce sync.Once
	v     interfaces.Validator

	// 仅支持中文、字母、数字、下划线
	commonNameReg = `^[[:word:]\p{Han}]+$`
)

func NewValidator() interfaces.Validator {
	vOnce.Do(func() {
		conf := config.NewConfigLoader()
		v = &validator{
			Validator:           validatorv10.New(),
			ImportMaxCount:      conf.OperatorConfig.ImportOperatorMaxCount,
			NameLimit:           defaultNameMaxLength,
			DescLimit:           conf.OperatorConfig.DescLengthLimit,
			ImportFileSizeLimit: conf.OperatorConfig.ImportFileSizeLimit,
		}
	})
	return v
}

// init 初始化验证器
func init() {
	validator := validatorv10.New()
	// 自定义验证器使用的字段名称标签
	validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// 从结构体字段的json标签中获取第一个值（忽略其他选项）
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0] //nolint:mnd

		// 如果标签设置为"-"则跳过该字段
		if name == "-" {
			return ""
		}
		// 返回json标签定义的字段名
		return name
	})
	_ = validator.RegisterValidation("uuid4", func(fl validatorv10.FieldLevel) bool {
		return govalidator.IsUUIDv4(fl.Field().String())
	})
}

// ValidateOperatorName 验证算子名称是否合法
// 仅支持中英文、数字和键盘上的特殊字符
func (v *validator) ValidateOperatorName(ctx context.Context, name string) (err error) {
	if name == "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtOperatorNameEmpty, "operator name cannot be empty")
		return
	}

	// 校验长度（按字符数计算）
	if utf8.RuneCountInString(name) > int(v.NameLimit) {
		err = fmt.Errorf("operator name %s length exceeds limit [%d]", name, v.NameLimit)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtOperatorNameTooLong, err.Error(),
			v.NameLimit)
		return
	}

	matched, err := regexp.MatchString(commonNameReg, name)
	if err != nil {
		err = fmt.Errorf("operator name %s contains invalid characters %v", name, err)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtCommonNameInvalid, err.Error())
		return
	}
	if !matched {
		err = fmt.Errorf("operator name %s contains invalid characters", name)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtCommonNameInvalid, err.Error())
	}
	return
}

// ValidateOperatorDesc 验证算子描述是否合法
func (v *validator) ValidateOperatorDesc(ctx context.Context, desc string) (err error) {
	// 算子描述不允许为空
	if desc == "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtOperatorDescEmpty, "operator description cannot be empty")
		return
	}
	// 校验长度（按字符数计算）
	if utf8.RuneCountInString(desc) > int(v.DescLimit) {
		err = fmt.Errorf("operator description length exceeds limit [%d]", v.DescLimit)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtOperatorDescTooLong, err.Error(), v.DescLimit)
	}
	return
}

// 校验算子导入数量是否超过限制
func (v *validator) ValidateOperatorImportCount(ctx context.Context, count int64) (err error) {
	if count == 0 {
		err = fmt.Errorf("operator import count %d is zero", count)
		err = myErr.NewHTTPError(ctx, http.StatusNotFound, myErr.ErrExtOperatorUnparsed, err.Error())
		return
	}
	if count > v.ImportMaxCount {
		err = fmt.Errorf("operator import count %d exceeds limit [%d]", count, v.ImportMaxCount)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtOperatorImportLimit, err.Error(),
			v.ImportMaxCount)
	}
	return
}

// 校验导入数据的大小是否超过限制
func (v *validator) ValidateOperatorImportSize(ctx context.Context, size int64) (err error) {
	if size == 0 {
		err = myErr.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("operator import size %d is zero", size))
		return
	}
	// 将文件大小转换为MB
	if size < v.ImportFileSizeLimit {
		return
	}
	// 返回提示信息中，将当前限制转化为 B、KB、MB、GB、TB为单位的字符串
	sizeStr := utils.ConvertToBytes(v.ImportFileSizeLimit)
	err = fmt.Errorf("file size %d exceeds limit [%s]", size, sizeStr)
	err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtOperatorImportDataLimit,
		err.Error(), sizeStr)
	return
}

func (v *validator) ValidatorToolBoxName(ctx context.Context, name string) (err error) {
	if name == "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtToolBoxNameEmpty, "toolbox name cannot be empty")
		return
	}

	// 校验长度（按字符数计算）
	if utf8.RuneCountInString(name) > int(v.NameLimit) {
		err = fmt.Errorf("toolbox name %s length exceeds limit [%d]", name, v.NameLimit)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtToolBoxNameLimit, err.Error(),
			v.NameLimit)
		return
	}
	matched, _ := regexp.MatchString(commonNameReg, name)
	if !matched {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtCommonNameInvalid,
			fmt.Sprintf("toolbox name %s format is invalid", name))
	}
	return
}

func (v *validator) ValidatorToolBoxDesc(ctx context.Context, desc string) (err error) {
	// 工具箱描述不允许为空
	if desc == "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtToolBoxDescEmpty, "toolbox description cannot be empty")
		return
	}
	// 校验长度（按字符数计算）
	if utf8.RuneCountInString(desc) > int(v.DescLimit) {
		err = fmt.Errorf("toolbox description length exceeds limit [%d]", v.DescLimit)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtToolBoxDescLimit, err.Error(),
			v.DescLimit)
	}
	return
}
func (v *validator) ValidatorToolName(ctx context.Context, name string) (err error) {
	if name == "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtToolNameEmpty, "tool name cannot be empty")
		return
	}
	if utf8.RuneCountInString(name) > int(v.NameLimit) {
		err = fmt.Errorf("tool name length exceeds limit [%d]", v.NameLimit)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtToolNameLimit, err.Error(),
			v.NameLimit)
	}
	matched, _ := regexp.MatchString(commonNameReg, name)
	if !matched {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtCommonNameInvalid,
			fmt.Sprintf("tool name %s format is invalid", name))
	}
	return
}
func (v *validator) ValidatorToolDesc(ctx context.Context, desc string) (err error) {
	// 工具描述不允许为空
	if desc == "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtToolDescEmpty, "tool description cannot be empty")
		return
	}
	if utf8.RuneCountInString(desc) > int(v.DescLimit) {
		err = fmt.Errorf("tool description length exceeds limit [%d]", v.DescLimit)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtToolDescLimit, err.Error(),
			v.DescLimit)
	}
	return
}

// ValidatorIntCompVersion 验证内置组件版本
func (v *validator) ValidatorIntCompVersion(ctx context.Context, version string) (err error) {
	pattern := `^[0-9]+(\.[0-9]+){2,}$` // 允许x.y.z或更长格式（如x.y.z.w）
	matched, err := regexp.MatchString(pattern, version)
	if !matched || err != nil {
		err = fmt.Errorf("internal component version:%s, format is invalid", version)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtInternalToolBoxVersion, err.Error(),
			version)
	}
	return
}

func (v *validator) ValidatorMCPName(ctx context.Context, name string) (err error) {
	if name == "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtMCPNameEmpty, "mcp name cannot be empty")
		return
	}

	// 校验长度（按字符数计算）
	if utf8.RuneCountInString(name) > int(v.NameLimit) {
		err = fmt.Errorf("mcp name %s length exceeds limit [%d]", name, v.NameLimit)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtMCPNameLimit, err.Error(),
			v.NameLimit)
		return
	}
	matched, _ := regexp.MatchString(commonNameReg, name)
	if !matched {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtCommonNameInvalid,
			fmt.Sprintf("mcp name %s format is invalid", name))
	}
	return
}

func (v *validator) ValidatorMCPDesc(ctx context.Context, desc string) (err error) {
	if utf8.RuneCountInString(desc) > int(v.DescLimit) {
		err = fmt.Errorf("mcp description length exceeds limit [%d]", v.DescLimit)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtMCPDescLimit, err.Error(),
			v.DescLimit)
	}
	return
}

func (v *validator) ValidatorCategoryName(ctx context.Context, name string) (err error) {
	if name == "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtCategoryNameEmpty, "category name cannot be empty")
		return
	}

	// 校验长度（按字符数计算）
	if utf8.RuneCountInString(name) > int(v.NameLimit) {
		err = fmt.Errorf("category name %s length exceeds limit [%d]", name, v.NameLimit)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtCategoryNameLimit, err.Error(),
			v.NameLimit)
		return
	}
	matched, _ := regexp.MatchString(commonNameReg, name)
	if !matched {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtCommonNameInvalid,
			fmt.Sprintf("category name %s format is invalid", name))
	}
	return
}

// ValidatorStruct 验证结构体
func (v *validator) ValidatorStruct(ctx context.Context, obj interface{}) (err error) {
	err = v.Validator.Struct(obj)
	if err == nil {
		return
	}
	vErr := make(validatorv10.ValidationErrors, 0)
	if !errors.As(err, &vErr) {
		return
	}
	extCode := TagToErrorType[vErr[0].Tag()]
	if extCode != "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, extCode, vErr[0].Error())
	}
	return
}

// 验证URL是否符合格式
func (v *validator) ValidatorURL(ctx context.Context, url string) (err error) {
	if url == "" {
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtOpenAPIInvalidURLFormat, "URL cannot be empty")
		return
	}

	if !govalidator.IsURL(url) {
		err = fmt.Errorf("URL %s format is invalid", url)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtOpenAPIInvalidURLFormat, err.Error())
	}
	return
}

// VisitorParameterDef 访问参数定义
func (v *validator) VisitorParameterDef(ctx context.Context, paramDef *interfaces.ParameterDef) (err error) {
	if paramDef == nil {
		err = myErr.DefaultHTTPError(ctx, http.StatusBadRequest, "parameter def cannot be nil")
		return
	}

	if paramDef.Type != "" && !validParameterTypes[paramDef.Type] {
		err = fmt.Errorf("parameter %s type %s is invalid, must be string, number, boolean, array, object", paramDef.Name, paramDef.Type)
		err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtFunctionInvalidParameterType, err.Error(), paramDef.Name, paramDef.Type)
		return
	}

	// 验证 SubParameters 只能用于 array 和 object 类型
	if len(paramDef.SubParameters) > 0 {
		if paramDef.Type != "array" && paramDef.Type != "object" {
			err = fmt.Errorf("parameter %s type %s is invalid, must be array or object", paramDef.Name, paramDef.Type)
			err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtFunctionInvalidParameterSubParameters, err.Error(), paramDef.Name, paramDef.Type)
			return
		}

		// 对于 array 类型,SubParameters 应该只有一个元素
		if paramDef.Type == "array" && len(paramDef.SubParameters) != 1 {
			err = fmt.Errorf("parameter %s is array type, sub_parameters must only contain one element to define the structure of array items, current has %d elements",
				paramDef.Name, len(paramDef.SubParameters))
			err = myErr.NewHTTPError(ctx, http.StatusBadRequest, myErr.ErrExtFunctionInvalidParameterSubParametersCount, err.Error(), paramDef.Name, len(paramDef.SubParameters))
			return
		}

		// 递归验证所有子参数
		for _, subParam := range paramDef.SubParameters {
			if err = v.VisitorParameterDef(ctx, subParam); err != nil {
				return
			}
		}
	}
	if paramDef.Type == "array" && len(paramDef.SubParameters) == 0 {
		// 为 array 类型添加默认子参数
		paramDef.SubParameters = []*interfaces.ParameterDef{
			{
				Type:        "string",
				Description: paramDef.Description,
				Required:    false,
			},
		}
	}
	return nil
}
