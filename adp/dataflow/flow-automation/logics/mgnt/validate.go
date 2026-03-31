package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
)

// ErrTypeV1 新错误码规范前的错误结构
var ErrTypeV1 = "v1"

// ErrTypeV1 新错误码规范后的错误结构
var ErrTypeV2 = "v2"

type Validate struct {
	Ctx         context.Context
	Steps       []map[string]interface{}
	IsAdminRole bool
	UserInfo    *drivenadapters.UserInfo
	ErrType     string
	ParseFunc   interface{} // common目录定义的解析函数
}

// ValidateError 参数校验错误返回结构体，用于兼容新旧错误码规范
type ValidateError struct {
	// 上下文信息，用于请求链路追踪和日志记录
	Ctx context.Context

	// 错误码类型标识，取值应为 ErrTypeV1 或 ErrTypeV2
	// - ErrTypeV1: 旧版错误码规范
	// - ErrTypeV2: 新版错误码规范
	ErrType string

	// 新版错误码的公共错误类型，从预定义枚举中取值
	// 例如: ierr.PublicErrorType 或 ierr.CustomErrorType
	PublicErrorType string

	// 旧版错误码的主错误码，对应 aerr 包中定义的错误码常量
	// 例如: aerr.InvalidParameter, aerr.NoPermission 等
	MainCode string

	// 新版错误码的主错误码，对应 ierr 包中定义的错误码常量
	// 例如: ierr.PErrorBadRequest, ierr.PErrorForbidden 等
	MainCodeV2 string

	// 扩展错误码，用于更细粒度的错误分类
	ExtCode string

	// 国际化资源键，对应 locales 资源文件中定义的键值
	// 用于前端展示多语言错误信息
	DescriptionKey string

	// 错误详情信息，通常为键值对形式的map
	// 包含具体的错误参数和上下文信息
	// 例如: map[string]interface{}{"operator": "invalid operator"}
	Detail interface{}

	// 可直接使用的预构建错误对象
	// 当此字段不为空时，将直接返回此错误而忽略其他字段
	Error error
}

// BuildError 构建错误
func (v *ValidateError) BuildError() error {
	if v == nil {
		return nil
	}

	if v.Error != nil {
		return v.Error
	}

	switch v.ErrType {
	case ErrTypeV1:
		return aerr.NewIError(v.MainCode, v.ExtCode, v.Detail)
	case ErrTypeV2:
		switch v.PublicErrorType {
		case ierr.PublicErrorType:
			return ierr.NewPublicRestError(v.Ctx, v.MainCodeV2, v.DescriptionKey, v.Detail)
		case ierr.CustomErrorType:
			return ierr.NewCustomRestError(v.Ctx, v.MainCodeV2, v.ExtCode, v.DescriptionKey, v.Detail)
		default:
			panic("UnSupported Rest Error Type")
		}
	default:
		return aerr.NewIError(v.MainCode, v.ExtCode, v.Detail)
	}
}

func (v *Validate) Valid(data []byte, path string) error {
	switch v.ErrType {
	case ErrTypeV1:
		f, _ := v.ParseFunc.(func([]byte, string) error)
		return f(data, path)
	case ErrTypeV2:
		f, _ := v.ParseFunc.(func(context.Context, []byte, string) error)
		return f(v.Ctx, data, path)
	default:
		return common.JSONSchemaValid(data, path)
	}
}

// validOperator validate operator
func validOperator(index int, operator string, isDataSource *bool, isAdminRole bool) (string, *ValidateError) {
	vErr := &ValidateError{
		PublicErrorType: ierr.PublicErrorType,
		MainCode:        aerr.InvalidParameter,
		MainCodeV2:      ierr.PErrorBadRequest,
		DescriptionKey:  ierr.PErrorBadRequest,
	}
	// step0: check action permission
	if !isAdminRole && (operator == common.IntelliinfoTranfer || operator == common.OpEcoconfigReindex) {
		vErr.MainCode = aerr.NoPermission
		vErr.MainCodeV2 = ierr.PErrorForbidden
		vErr.Detail = map[string]interface{}{"operator": fmt.Sprintf("%s should have admin role", operator)}
		return "", vErr
	}
	// step1: validate oprator format
	if !strings.HasPrefix(operator, "@") {
		vErr.Detail = map[string]interface{}{"operator": fmt.Sprintf("%s should follow the format @xxxx/xxxx/xxx", operator)}
		return "", vErr
	}
	// step2: case1 operator not datasource return error
	// step2: case2 except first node, other node operator's type should not trigger
	if *isDataSource {
		if !strings.HasPrefix(operator, "@anyshare-data") {
			vErr.ExtCode = aerr.ErrorIncorretTrigger
			vErr.DescriptionKey = aerr.DescKeyUnSupportedTrigger
			vErr.Detail = map[string]interface{}{"operator": fmt.Sprintf("datasource unsupported operation type: %s", operator)}
			return "", vErr
		}
		*isDataSource = false
	} else {
		if index == 0 && !strings.HasPrefix(operator, "@trigger") && !strings.HasPrefix(operator, "@anyshare-trigger") {
			vErr.ExtCode = aerr.ErrorIncorretTrigger
			vErr.DescriptionKey = aerr.DescKeyUnSupportedTrigger
			vErr.Detail = map[string]interface{}{"operator": fmt.Sprintf("%s, start node operator type should be trigger", operator)}
			return "", vErr
		}
		if index != 0 && (strings.HasPrefix(operator, "@trigger") || strings.HasPrefix(operator, "@anyshare-trigger")) {
			vErr.ExtCode = aerr.ErrorIncorretOperator
			vErr.DescriptionKey = aerr.DescKeyUnSupportedTrigger
			vErr.Detail = map[string]interface{}{
				"operator": fmt.Sprintf("%s, except start node, operator type should be exec or branch", operator)}
			return "", vErr
		}
	}
	// 组合算子signature只使用前缀来作为key进行参数校验
	if strings.HasPrefix(operator, "@operator") {
		operator = "@operator/"
	}

	if strings.HasPrefix(operator, common.TriggerOperatorPrefix) {
		operator = common.TriggerOperatorPrefix
	}

	// step3: operator whether exist
	path, ok := common.ActionMap[operator]
	if !ok {
		vErr.Detail = map[string]interface{}{"operator": fmt.Sprintf("unsupported operation type: %s", operator)}
		return "", vErr
	}
	return path, nil
}

// validSteps validate step params
func (m *mgnt) validSteps(validate *Validate) *ValidateError {
	var (
		exist             bool
		id, operator      string
		idMap             = map[string]struct{}{}
		customActionSteps = make([]map[string]interface{}, 0)
		log               = traceLog.WithContext(validate.Ctx)
	)
	vErr := &ValidateError{
		Ctx:             validate.Ctx,
		ErrType:         validate.ErrType,
		PublicErrorType: ierr.PublicErrorType,
		MainCode:        aerr.InvalidParameter,
		MainCodeV2:      ierr.PErrorBadRequest,
		DescriptionKey:  ierr.PErrorBadRequest,
	}

	handleErr := func(err error) {
		detail := map[string]interface{}{}
		switch iErr := err.(type) {
		case *aerr.IError:
			detail = iErr.ErrorDetails.(map[string]interface{})
		case *ierr.RestError:
			detail = iErr.ErrorDetails.(map[string]interface{})
		}
		vErr.Detail = map[string]interface{}{"id": id, "operator": operator, "invalidparams": detail["params"]}
	}

	for index, step := range validate.Steps {
		var isDataSource = false
	REDO:
		// check node id is unique
		if id, exist = step["id"].(string); !exist {
			vErr.Detail = map[string]interface{}{"id": "String length must be greater than or equal to 1"}
			return vErr
		}
		if _, ok := idMap[id]; ok {
			vErr.Detail = map[string]interface{}{"id": fmt.Sprintf("task id is repeat, id: %v", id)}
			return vErr
		}
		idMap[id] = struct{}{}
		if operator, exist = step["operator"].(string); !exist {
			vErr.Detail = map[string]interface{}{"operator": "should follow the format @xxxx/xxxx/xxx"}
			return vErr
		}

		if operator == common.InternalReturnOpt {
			continue
		}

		if strings.HasPrefix(operator, "@custom/") {
			customActionSteps = append(customActionSteps, step)
			continue
		}

		path, voErr := validOperator(index, operator, &isDataSource, validate.IsAdminRole)
		if voErr != nil {
			voErr.Ctx = validate.Ctx
			voErr.ErrType = validate.ErrType
			return voErr
		}
		if path == "" {
			log.Infof("[logic.validOperator] id: %s, operator: %s, skip validate\n", id, operator)

			// 因为有的节点未定义jsonschema，通过接口创建会存在非法参数问题
			// 因此统一校验高级配置，setting字段非必填，如果存在则校验参数合法性
			taskByte, _ := json.Marshal(step)
			err := validate.Valid(taskByte, "base/settings.json")
			if err != nil {
				handleErr(err)
				return vErr
			}

			continue
		}
		// according to trigger type, choose different values dataSource or parameter
		var dataSource map[string]interface{}
		if step["dataSource"] != nil {
			dataSource = step["dataSource"].(map[string]interface{})
			emptyDataSource := map[string]interface{}{}
			bt, _ := json.Marshal(entity.DataSource{})
			json.Unmarshal(bt, &emptyDataSource) //nolint
			if operator != common.MannualTriggerOpt && operator != common.CronTrigger && operator != common.CronWeekTrigger && operator != common.CronMonthTrigger && operator != common.CronCustomTrigger {
				if !reflect.DeepEqual(dataSource, emptyDataSource) {
					vErr.Detail = map[string]interface{}{"dataSource": fmt.Sprintf("operator: %s, only supporte manual and cron trigger set datasource", operator)}
					return vErr
				}
				delete(step, "dataSource")
			} else if !reflect.DeepEqual(dataSource, emptyDataSource) {
				isDataSource = true
			}
		}

		taskByte, _ := json.Marshal(step)
		err := validate.Valid(taskByte, path)
		if err != nil {
			handleErr(err)
			return vErr
		}

		// set datasource
		if isDataSource {
			step = dataSource
			goto REDO
		}
	}

	if len(customActionSteps) > 0 {
		accessorIDs, err := m.usermgnt.GetUserAccessorIDs(validate.UserInfo.UserID)

		if err != nil {
			vErr.MainCode = aerr.InternalError
			vErr.MainCodeV2 = ierr.PErrorInternalServerError
			vErr.DescriptionKey = ierr.PErrorInternalServerError
			vErr.Detail = err.Error()
			return vErr
		}

		executors, err := m.executor.GetAccessableExecutors(validate.Ctx, validate.UserInfo.UserID, accessorIDs)
		if err != nil {
			log.Infof("[logic.validOperator] GetAccessableExecutors err, detail: %s", err.Error())
			vErr.MainCode = aerr.InternalError
			vErr.MainCodeV2 = ierr.PErrorInternalServerError
			vErr.DescriptionKey = ierr.PErrorInternalServerError
			return vErr
		}

		actionMap := make(map[string]*rds.ExecutorActionModel, 0)

		if len(executors) > 0 {
			for _, executor := range executors {
				for _, action := range executor.Actions {
					actionMap[*action.Operator] = action
				}
			}
		}

		for _, step := range customActionSteps {
			if _, ok := actionMap[step["operator"].(string)]; !ok {
				vErr.Detail = map[string]interface{}{"operator": fmt.Sprintf("unsupported operation type: %s", operator)}
				return vErr
			}

			// TODO 验证自定义节点输入值是否符合定义
		}
	}

	return nil
}
