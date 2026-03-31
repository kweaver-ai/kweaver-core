package actions

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	liberrors "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

const maxRetrySecond = 5
const baseNum = 2
const maxretryTimes = 100
const DataTransConcurrentLimit = "intelliinfo.DataTrans.ConcurrentLimit"

// IntelliinfoTranfer 智能数据转换节点
type IntelliinfoTranfer struct {
	RuleID string      `json:"rule_id"`
	Data   interface{} `json:"data"`
}

// Name 操作名称
func (a *IntelliinfoTranfer) Name() string {
	return common.IntelliinfoTranfer
}

// Run 操作方法
func (a *IntelliinfoTranfer) Run(ctx entity.ExecuteContext, params interface{}, _ *entity.Token) (interface{}, error) {
	var status = entity.TaskInstanceStatusBlocked

	var err error
	taskIns := ctx.GetTaskInstance()
	if taskIns == nil {
		return nil, fmt.Errorf("get taskinstance failed")
	}
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	intelliinfoAdapters := drivenadapters.NewIntelliinfo()
	log := traceLog.WithContext(ctx.Context())
	applyID := taskIns.ID

	input := params.(*IntelliinfoTranfer)
	config := common.NewConfig()

	if ctx.IsDebug() {
		applyID = fmt.Sprintf("DEBUG:%v", applyID)
	}
	callback := fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/continue/%s", config.ContentAutomation.PrivateHost, config.ContentAutomation.PrivatePort, applyID)

	var payload interface{}
	var apiVers int
	if input.RuleID != "" {
		req, err := parseTransDataV1(input)
		if err != nil {
			return nil, err
		}
		req.CallBack = callback
		apiVers = drivenadapters.V1
		payload = req
	} else {
		req := &drivenadapters.DataTransferReqV2{}

		v, ok := input.Data.(string)
		if ok {
			err = json.Unmarshal([]byte(v), &req.Datas)
			if err != nil {
				return nil, err
			}
		}

		// 最后一个实体操作才给定callback
		req.CallBack = callback

		apiVers = drivenadapters.V2
		payload = req
	}

	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	retrytimesKey := fmt.Sprintf("__retrytimes_%s", taskIns.ID)
	userID, _ := taskIns.RelatedDagInstance.VarsGetter()("operator_id")
	userType, _ := taskIns.RelatedDagInstance.VarsGetter()("operator_type")
	_, err = intelliinfoAdapters.TransferData(ctx.Context(), payload, apiVers, fmt.Sprintf("%v", userID), fmt.Sprintf("%v", userType))
	if err != nil {
		parseError, err0 := errors.ExHTTPErrorParser(err)
		if err0 == nil {
			if parseError["code"] == DataTransConcurrentLimit {
				log.Debugf("[TransferData] TransferData failed, taskID %v, err: %v", applyID, err.Error())
			} else {
				log.Warnf("[TransferData] TransferData failed, taskID %v, err: %v", applyID, err.Error())
			}
		} else {
			log.Warnf("[TransferData] TransferData failed, taskID %v, err: %v", applyID, err.Error())
		}

		if os.IsTimeout(err) {
			status = entity.TaskInstanceStatusRetrying
			ctx.ShareData().Set(statusKey, status)
			retryTime, ok := ctx.ShareData().Get(retrytimesKey)
			if ok {
				retryTimesInt := ParseInt(retryTime)
				if retryTimesInt > 0 {
					ctx.ShareData().Set(retrytimesKey, retryTimesInt+1)
					return nil, nil
				}
			}
			ctx.ShareData().Set(retrytimesKey, 1)
			return nil, nil
		}
		parsedError, _err := liberrors.ExHTTPErrorParser(err)
		if _err != nil {
			return nil, err
		}
		if parsedError.Body["code"] == "intelliinfo.DataTrans.ConcurrentLimit" {
			status = entity.TaskInstanceStatusRetrying
			ctx.ShareData().Set(statusKey, status)
			retryTime, ok := ctx.ShareData().Get(retrytimesKey)
			if ok {
				retryTimesInt := ParseInt(retryTime)
				if retryTimesInt > maxretryTimes {
					return nil, err
				}
				if retryTimesInt > 0 {
					ctx.ShareData().Set(retrytimesKey, retryTimesInt+1)
					return nil, nil
				}
			}
			ctx.ShareData().Set(retrytimesKey, 1)
			return nil, nil
		}
		return nil, err
	}
	ctx.ShareData().Set(statusKey, status)
	if config.Debug == "true" {
		log.Infof("[TransferData] request transferData success, taskID %v", applyID)
	}
	ctx.Trace(ctx.Context(), fmt.Sprintf("%v", payload))
	ctx.Trace(ctx.Context(), "run end")
	return map[string]interface{}{}, nil
}

// RunAfter 操作方法
func (a *IntelliinfoTranfer) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	taskIns := ctx.GetTaskInstance()
	retrytimesKey := fmt.Sprintf("__retrytimes_%s", taskIns.ID)
	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	status, ok := ctx.ShareData().Get(statusKey)
	if !ok {
		return entity.TaskInstanceStatusBlocked, nil
	}
	if status == entity.TaskInstanceStatusRetrying {
		retryTime, ok := ctx.ShareData().Get(retrytimesKey)
		if ok {
			retryTimesInt := ParseInt(retryTime)
			if retryTimesInt > maxRetrySecond {
				time.Sleep(time.Duration(math.Pow(baseNum, maxRetrySecond)) * time.Second)
			} else {
				time.Sleep(time.Duration(math.Pow(baseNum, float64(retryTimesInt))) * time.Second)
			}
		} else {
			time.Sleep(baseNum * time.Second)
		}
		return entity.TaskInstanceStatusRetrying, nil
	}

	return entity.TaskInstanceStatusBlocked, nil
}

// ParameterNew 初始化参数
func (a *IntelliinfoTranfer) ParameterNew() interface{} {
	return &IntelliinfoTranfer{}
}

func parseTransDataV1(input *IntelliinfoTranfer) (*drivenadapters.DataTransferReqV1, error) {
	var parsedData interface{}
	switch data := input.Data.(type) {
	case string:
		jerr := json.Unmarshal([]byte(data), &parsedData)
		if jerr != nil {
			parsedData = data
		}
	case map[string]interface{}:
		if data["id"] == "" {
			return nil, fmt.Errorf("[TransferData] TransferData empty")
		}
		for item, val := range data {
			if item == "modify_time" || item == "create_time" {
				switch t := val.(type) {
				case string:
					v, err := time.Parse(time.RFC3339, t)
					if err != nil {
						var newNum float64
						_, err := fmt.Sscanf(t, "%e", &newNum)
						if err != nil {
							data[item] = 0
							break
						}
						data[item] = int64(newNum)
						break
					}
					data[item] = v.UnixNano() / 1e3 // 微秒
				default:
					data[item] = t
				}

				continue
			}

			if item == "size" || item == "csflevel" || item == "version" {
				var newNum float64
				_, err := fmt.Sscanf(val.(string), "%e", &newNum)
				if err != nil {
					data[item] = 0
					continue
				}
				data[item] = int64(newNum)
				continue
			}
			if item == "is_expert" {
				parsed, _ := strconv.ParseBool(val.(string))
				data[item] = parsed
				continue
			}
			if item == "parent_ids" || item == "university" || item == "tags" || item == "verification_info" || item == "professional" {
				var parsed = make([]string, 0)
				_ = json.Unmarshal([]byte(val.(string)), &parsed)
				data[item] = parsed
				if item == "parent_ids" {
					if _, ok := data["parent_id"]; !ok {
						// 适配parent_ids和parent_id表示相同的意义
						data["parent_id"] = parsed
					}
				}
				continue
			}
			if item == "work_at" {
				var targetLen = 12
				var parsed = make([]string, 0)
				err := json.Unmarshal([]byte(val.(string)), &parsed)
				if err != nil {
					continue
				}
				if len(parsed) > 0 {
					code := parsed[len(parsed)-1]
					paddingString := fmt.Sprintf("%s%0*d", code, targetLen-len(code), 0)
					data[item] = paddingString
				}
				continue
			}
		}
		parsedData = data
	}

	var req = &drivenadapters.DataTransferReqV1{
		RuleID: input.RuleID,
		Data:   parsedData,
	}
	return req, nil
}
