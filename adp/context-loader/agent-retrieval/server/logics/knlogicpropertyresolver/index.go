// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knlogicpropertyresolver

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

const (
	// defaultMaxConcurrency 默认最大并发数
	defaultMaxConcurrency = 4
)

// knLogicPropertyResolverService 逻辑属性解析服务实现
type knLogicPropertyResolverService struct {
	logger              interfaces.Logger
	bknBackendAccess    interfaces.BknBackendAccess
	ontologyQueryClient interfaces.DrivenOntologyQuery
	agentApp            interfaces.AgentApp
}

var (
	serviceOnce sync.Once
	service     interfaces.IKnLogicPropertyResolverService
)

// NewKnLogicPropertyResolverService 创建逻辑属性解析服务
func NewKnLogicPropertyResolverService() interfaces.IKnLogicPropertyResolverService {
	serviceOnce.Do(func() {
		conf := config.NewConfigLoader()
		service = &knLogicPropertyResolverService{
			logger:              conf.GetLogger(),
			bknBackendAccess:    drivenadapters.NewBknBackendAccess(),
			ontologyQueryClient: drivenadapters.NewOntologyQueryAccess(),
			agentApp:            drivenadapters.NewAgentAppClient(),
		}
	})
	return service
}

// ResolveLogicProperties 解析逻辑属性
func (s *knLogicPropertyResolverService) ResolveLogicProperties(
	ctx context.Context,
	req *interfaces.ResolveLogicPropertiesRequest,
) (*interfaces.ResolveLogicPropertiesResponse, error) {
	// 简化日志：Handler 层已记录详细请求参数
	s.logger.WithContext(ctx).Debugf("[Service] 开始处理 %d 个逻辑属性", len(req.Properties))

	// 设置默认 Options
	if req.Options == nil {
		req.Options = &interfaces.ResolveOptions{
			ReturnDebug:     false,
			MaxRepairRounds: 1,
			MaxConcurrency:  defaultMaxConcurrency,
		}
	}

	// Step 1: 参数校验
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Step 2: 获取对象类定义
	s.logger.WithContext(ctx).Debugf("[Step 1] 获取对象类定义: kn_id=%s, ot_id=%s", req.KnID, req.OtID)
	objectType, err := s.getObjectTypeDefinition(ctx, req.KnID, req.OtID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[Step 1] ❌ 失败: %v", err)
		return nil, err
	}
	s.logger.WithContext(ctx).Debugf("[Step 1] ✅ 成功")

	// Step 3: 提取逻辑属性定义
	logicPropertiesDef, err := s.extractLogicProperties(ctx, objectType, req.Properties)
	if err != nil {
		return nil, err
	}

	// 初始化 debug 信息收集器
	var debugCollector *DebugCollector
	if req.Options.ReturnDebug {
		debugCollector = NewDebugCollector()
		debugCollector.SetTraceID("")
		debugCollector.SetNowMs(time.Now().UnixMilli())
	}

	// Step 4: 生成 dynamic_params
	s.logger.WithContext(ctx).Debugf("[Step 2] 生成 dynamic_params（Agent 并发调用）")
	startTime := time.Now()
	dynamicParams, missingParams, err := s.generateDynamicParams(ctx, req, logicPropertiesDef, debugCollector)
	generateParamsDuration := time.Since(startTime)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[Step 2] ❌ 失败: %v", err)
		return nil, err
	}

	// 如果有缺参，根据是否开启 debug 决定处理方式
	if len(missingParams) > 0 {
		s.logger.WithContext(ctx).Warnf("[Step 2] ⚠️ 存在缺参: %d 个属性", len(missingParams))

		// 特殊处理：如果开启了 debug，返回正常响应，错误信息放在 debug 中
		if req.Options.ReturnDebug {
			s.logger.WithContext(ctx).Infof("[Step 2] 🔍 Debug模式：缺参场景返回正常响应，错误信息放在 debug 中")

			// 构建正常响应，datas 为空数组
			debugInfo := debugCollector.BuildDebugInfo()
			return &interfaces.ResolveLogicPropertiesResponse{
				Datas: []map[string]any{}, // 空数组，因为没有成功的数据
				Debug: debugInfo,
			}, nil
		}

		// 未开启 debug：保持现有行为，抛出错误
		missingError := s.buildMissingParamsError(ctx, missingParams, nil)
		return nil, missingError
	}
	s.logger.WithContext(ctx).Infof("⏱️ [耗时] 生成动态参数: %dms", generateParamsDuration.Milliseconds())

	// Step 5: 调用 ontology-query 查询逻辑属性值
	s.logger.WithContext(ctx).Debugf("[Step 3] 调用 ontology-query 查询属性值")
	startTime = time.Now()
	result, err := s.queryLogicProperties(ctx, req, dynamicParams)
	queryDuration := time.Since(startTime)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[Step 3] ❌ 失败: %v", err)
		return nil, err
	}
	s.logger.WithContext(ctx).Infof("⏱️ [耗时] 查询属性值: %dms", queryDuration.Milliseconds())

	// Step 6: 构建响应
	resp := &interfaces.ResolveLogicPropertiesResponse{
		Datas: result,
	}

	// 如果需要返回 debug 信息
	if req.Options.ReturnDebug {
		debugCollector.SetNowMs(time.Now().UnixMilli())
		resp.Debug = debugCollector.BuildDebugInfo()
	}

	s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Resolve logic properties successfully")
	return resp, nil
}

// validateRequest 校验请求参数
func (s *knLogicPropertyResolverService) validateRequest(req *interfaces.ResolveLogicPropertiesRequest) error {
	if req.KnID == "" {
		return fmt.Errorf("kn_id is required")
	}
	if req.OtID == "" {
		return fmt.Errorf("ot_id is required")
	}
	if req.Query == "" {
		return fmt.Errorf("query is required")
	}
	if len(req.InstanceIdentities) == 0 {
		return fmt.Errorf("_instance_identities is required and cannot be empty")
	}
	if len(req.Properties) == 0 {
		return fmt.Errorf("properties is required and cannot be empty")
	}
	return nil
}

// getObjectTypeDefinition 获取对象类定义
func (s *knLogicPropertyResolverService) getObjectTypeDefinition(
	ctx context.Context,
	knID string,
	otID string,
) (*interfaces.ObjectType, error) {
	s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Getting object type definition: kn_id=%s, ot_id=%s", knID, otID)

	// 调用 BKN backend 获取对象类定义（include_detail=true 以获取 logic_properties）
	objectTypes, err := s.bknBackendAccess.GetObjectTypeDetail(ctx, knID, []string{otID}, true)
	if err != nil {
		return nil, err
	}

	// 检查返回结果
	if len(objectTypes) == 0 {
		return nil, errors.DefaultHTTPError(ctx, http.StatusNotFound,
			fmt.Sprintf("object type %s not found in knowledge network %s", otID, knID))
	}

	// 返回第一个对象类定义（我们只请求了一个 otID）
	return objectTypes[0], nil
}

// extractLogicProperties 从对象类定义中提取逻辑属性定义
func (s *knLogicPropertyResolverService) extractLogicProperties(
	ctx context.Context,
	objectType *interfaces.ObjectType,
	properties []string,
) (map[string]*interfaces.LogicPropertyDef, error) {
	s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Extracting logic properties: %v", properties)

	// 检查 objectType.LogicProperties 是否为空
	if len(objectType.LogicProperties) == 0 {
		s.logger.WithContext(ctx).Warnf("[KnLogicPropertyResolver] Object type %s has no logic properties", objectType.ID)
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest,
			fmt.Sprintf("object type %s has no logic properties defined", objectType.ID))
	}

	// 1. 构建请求属性的 set，便于查找和验证
	requestedProps := make(map[string]bool, len(properties))
	for _, prop := range properties {
		requestedProps[prop] = true
	}

	// 2. 遍历 objectType.LogicProperties，筛选出请求的属性
	logicPropertiesDef := make(map[string]*interfaces.LogicPropertyDef, len(properties))
	for _, logicProp := range objectType.LogicProperties {
		if requestedProps[logicProp.Name] {
			logicPropertiesDef[logicProp.Name] = logicProp
			s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Found logic property: %s (type: %s)",
				logicProp.Name, logicProp.Type)
		}
	}

	// 3. 检查是否所有请求的属性都找到了
	notFoundProps := []string{}
	for _, prop := range properties {
		if _, found := logicPropertiesDef[prop]; !found {
			notFoundProps = append(notFoundProps, prop)
		}
	}

	// 4. 如果有属性不存在，返回 INVALID_PROPERTY 错误
	if len(notFoundProps) > 0 {
		s.logger.WithContext(ctx).Errorf("[KnLogicPropertyResolver] Properties not found: %v", notFoundProps)

		// 构建可用的逻辑属性列表（用于错误提示）
		availableProps := make([]string, 0, len(objectType.LogicProperties))
		for _, logicProp := range objectType.LogicProperties {
			availableProps = append(availableProps, logicProp.Name)
		}

		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest,
			fmt.Sprintf("properties not found or not logic properties: %v (available logic properties: %v)",
				notFoundProps, availableProps))
	}
	return logicPropertiesDef, nil
}

// generateDynamicParams 生成 dynamic_params（按 property 并发）
//
//nolint:unparam // 保持接口一致性，error 返回用于后续扩展
func (s *knLogicPropertyResolverService) generateDynamicParams(
	ctx context.Context,
	req *interfaces.ResolveLogicPropertiesRequest,
	logicPropertiesDef map[string]*interfaces.LogicPropertyDef,
	debugCollector *DebugCollector,
) (dynamicParams map[string]interface{}, missingParams []interfaces.MissingPropertyParams, err error) {
	s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Generating dynamic params for %d properties", len(logicPropertiesDef))

	// 获取并发配置
	maxConcurrency := req.Options.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 4 // 默认并发数
	}

	// Step 1: 准备阶段 - 构建 property 列表
	type PropertyTask struct {
		Name     string
		Property *interfaces.LogicPropertyDef
	}

	tasks := make([]PropertyTask, 0, len(logicPropertiesDef))
	for name, prop := range logicPropertiesDef {
		tasks = append(tasks, PropertyTask{Name: name, Property: prop})
	}

	// Step 2: 并发调用 LLM（统一控制 max_concurrency）
	type PropertyResult struct {
		Name          string
		DynamicParams map[string]interface{}
		MissingParams *interfaces.MissingPropertyParams
		Error         error
	}

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, maxConcurrency)
	results := make(chan PropertyResult, len(tasks))

	// 并发处理每个 property
	for _, task := range tasks {
		go func(t PropertyTask) {
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 收集 property 类型信息
			if debugCollector != nil {
				debugCollector.AddPropertyType(t.Name, string(t.Property.Type))
			}

			// 生成单个 property 的 dynamic_params
			params, missingParams, err := s.generateSinglePropertyParams(ctx, req, t.Name, t.Property, debugCollector)
			results <- PropertyResult{
				Name:          t.Name,
				DynamicParams: params,
				MissingParams: missingParams,
				Error:         err,
			}
		}(task)
	}

	// Step 3: 收集结果
	dynamicParams = make(map[string]interface{})
	missingParams = []interfaces.MissingPropertyParams{}

	for range len(tasks) {
		result := <-results

		// 如果有错误，记录但继续处理其他 property
		if result.Error != nil {
			s.logger.WithContext(ctx).Errorf("[KnLogicPropertyResolver] Generate params for property %s failed: %v",
				result.Name, result.Error)
			// 记录错误到 debug 信息
			if debugCollector != nil {
				debugCollector.RecordAgentResponseError(result.Name, result.Error.Error())
			}
			// 将错误转换为缺参（让上游知道哪个 property 失败了）
			missingParams = append(missingParams, interfaces.MissingPropertyParams{
				Property: result.Name,
				ErrorMsg: fmt.Sprintf("generate params failed: %v", result.Error),
			})
			continue
		}

		// 如果有缺参，收集缺参信息
		if result.MissingParams != nil {
			missingParams = append(missingParams, *result.MissingParams)
			continue
		}

		// 收集成功的结果
		// 关键修复：需要将参数对象放在 property name 的 key 下
		// ontology-query 期望的格式：{"property_name": {"param1": value1, ...}}
		if result.DynamicParams != nil {
			dynamicParams[result.Name] = result.DynamicParams
			s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Collected params for %s: %+v",
				result.Name, result.DynamicParams)
		}
	}

	s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Generated dynamic params for %d properties, %d missing",
		len(dynamicParams), len(missingParams))

	return dynamicParams, missingParams, nil
}

// generateSinglePropertyParams 生成单个 property 的 dynamic_params
func (s *knLogicPropertyResolverService) generateSinglePropertyParams(
	ctx context.Context,
	req *interfaces.ResolveLogicPropertiesRequest,
	propertyName string,
	property *interfaces.LogicPropertyDef,
	debugCollector *DebugCollector,
) (dynamicParams map[string]interface{}, missingParams *interfaces.MissingPropertyParams, err error) {
	s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Generating params for property: %s (type: %s)",
		propertyName, property.Type)

	// 根据属性类型，调用对应的参数生成方法
	// 注：当前使用 Agent 平台实现，后续可扩展支持直接调用 LLM
	switch property.Type {
	case interfaces.LogicPropertyTypeMetric:
		dynamicParams, missingParams, err = s.generateMetricParams(ctx, req, property, propertyName, debugCollector)
	case interfaces.LogicPropertyTypeOperator:
		dynamicParams, missingParams, err = s.generateOperatorParams(ctx, req, property, propertyName, debugCollector)
	default:
		return nil, nil, fmt.Errorf("unknown property type: %s", property.Type)
	}

	if err != nil {
		// 记录 Agent 错误响应
		if debugCollector != nil {
			debugCollector.RecordAgentResponseError(propertyName, err.Error())
		}
		return nil, nil, fmt.Errorf("generate params failed: %w", err)
	}

	// 记录 Agent 响应信息
	if debugCollector != nil {
		if missingParams != nil {
			debugCollector.RecordAgentResponseMissingParams(propertyName, missingParams)
		} else if dynamicParams != nil {
			debugCollector.RecordAgentResponseSuccess(propertyName, dynamicParams)
		}
	}

	// 如果有返回的 dynamic_params，进行类型校验
	if dynamicParams != nil {
		// 详细日志：校验前查看参数内容
		s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Validating params for %s (type: %s), params: %+v",
			propertyName, property.Type, dynamicParams)

		switch property.Type {
		case interfaces.LogicPropertyTypeMetric:
			err = s.validateMetricParams(ctx, property, dynamicParams)
		case interfaces.LogicPropertyTypeOperator:
			err = s.validateOperatorParams(ctx, property, dynamicParams)
		}

		if err != nil {
			s.logger.WithContext(ctx).Errorf("[KnLogicPropertyResolver] Validation failed for %s: %v", propertyName, err)
			// 校验失败时，返回校验错误（不返回 missingParams，因为这是校验失败，不是缺参）
			return nil, nil, fmt.Errorf("validate params failed for %s: %w", propertyName, err)
		}

		s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Validation passed for %s", propertyName)
	}

	return dynamicParams, missingParams, nil
}

// generateMetricParams 通过 Agent 生成 metric 类型的动态参数
// 注：此方法封装了 Agent 调用，后续可扩展支持直接调用 LLM
func (s *knLogicPropertyResolverService) generateMetricParams(
	ctx context.Context,
	req *interfaces.ResolveLogicPropertiesRequest,
	property *interfaces.LogicPropertyDef,
	propertyName string,
	debugCollector *DebugCollector,
) (dynamicParams map[string]any, missingParams *interfaces.MissingPropertyParams, err error) {
	s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Generating metric params via Agent for: %s", property.Name)

	// 生成 now_ms（如果调用方未在 additional_context 中提供）
	nowMs := time.Now().UnixMilli()

	// 构建 Agent 请求
	agentReq := &interfaces.MetricDynamicParamsGeneratorReq{
		LogicProperty:     property,
		Query:             req.Query,
		UniqueIdentities:  req.InstanceIdentities,
		AdditionalContext: req.AdditionalContext,
		NowMs:             nowMs,
		Timezone:          "", // 暂时不考虑 timezone
	}

	// 记录 Agent 请求信息
	if debugCollector != nil {
		debugCollector.RecordMetricAgentRequest(propertyName, agentReq)
	}

	// 调用 Metric Agent
	agentResult, missingParams, err := s.agentApp.MetricDynamicParamsGeneratorAgent(ctx, agentReq)
	if err != nil {
		return nil, nil, err
	}

	// 如果有缺参，直接返回
	if missingParams != nil {
		return nil, missingParams, nil
	}

	// 从 Agent 返回的结果中提取对应 property 的参数对象
	// Agent 返回格式：{"approved_drug_count": {"instant": false, "start": xxx, ...}}
	// 我们需要提取：{"instant": false, "start": xxx, ...}
	if agentResult != nil {
		if propertyParams, ok := agentResult[property.Name]; ok {
			if paramsMap, ok := propertyParams.(map[string]any); ok {
				return paramsMap, nil, nil
			}
		}
		// 如果提取失败，返回错误
		return nil, nil, fmt.Errorf("failed to extract params for property %s from agent result: %+v", property.Name, agentResult)
	}

	return nil, nil, nil
}

// generateOperatorParams 通过 Agent 生成 operator 类型的动态参数
// 注：此方法封装了 Agent 调用，后续可扩展支持直接调用 LLM
func (s *knLogicPropertyResolverService) generateOperatorParams(
	ctx context.Context,
	req *interfaces.ResolveLogicPropertiesRequest,
	property *interfaces.LogicPropertyDef,
	propertyName string,
	debugCollector *DebugCollector,
) (dynamicParams map[string]any, missingParams *interfaces.MissingPropertyParams, err error) {
	s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Generating operator params via Agent for: %s", property.Name)

	// 从 data_source 中提取 operator_id
	var operatorID string
	if property.DataSource != nil {
		if id, ok := property.DataSource["id"].(string); ok {
			operatorID = id
		}
	}

	// 构建 Agent 请求
	agentReq := &interfaces.OperatorDynamicParamsGeneratorReq{
		OperatorID:        operatorID,
		LogicProperty:     property,
		Query:             req.Query,
		UniqueIdentities:  req.InstanceIdentities,
		AdditionalContext: req.AdditionalContext,
	}

	// 记录 Agent 请求信息
	if debugCollector != nil {
		debugCollector.RecordOperatorAgentRequest(propertyName, agentReq)
	}

	// 调用 Operator Agent
	agentResult, missingParams, err := s.agentApp.OperatorDynamicParamsGeneratorAgent(ctx, agentReq)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[KnLogicPropertyResolver] OperatorDynamicParamsGeneratorAgent failed: %v", err)
		return nil, nil, err
	}

	// 如果有缺参，直接返回
	if missingParams != nil {
		return nil, missingParams, nil
	}

	// 从 Agent 返回的结果中提取对应 property 的参数对象
	if agentResult != nil {
		if propertyParams, ok := agentResult[property.Name]; ok {
			if paramsMap, ok := propertyParams.(map[string]any); ok {
				return paramsMap, nil, nil
			}
		}
		// 如果提取失败，返回错误
		return nil, nil, fmt.Errorf("failed to extract params for property %s from agent result: %+v", property.Name, agentResult)
	}

	return nil, nil, nil
}

// validateMetricParams 校验 metric 类型的参数
func (s *knLogicPropertyResolverService) validateMetricParams(
	ctx context.Context,
	property *interfaces.LogicPropertyDef,
	params map[string]any,
) error {
	// 1. 检查 instant 字段（必需）
	instantVal, hasInstant := params["instant"]
	if !hasInstant {
		// 🔧 临时方案：如果缺少 instant，根据是否有 step 自动推断
		_, hasStep := params["step"]
		if hasStep {
			// 有 step 说明是趋势查询
			params["instant"] = false
			s.logger.WithContext(ctx).Warnf("[KnLogicPropertyResolver] Auto-inferred instant=false for metric property: %s (has step field)", property.Name)
			instantVal = false
		} else {
			// 没有 step 说明是即时查询
			params["instant"] = true
			s.logger.WithContext(ctx).Warnf("[KnLogicPropertyResolver] Auto-inferred instant=true for metric property: %s (no step field)", property.Name)
			instantVal = true
		}
	}

	instant, ok := instantVal.(bool)
	if !ok {
		return fmt.Errorf("param 'instant' must be boolean for metric property: %s", property.Name)
	}

	// 2. 检查 start 和 end（通常必需）
	if _, hasStart := params["start"]; !hasStart {
		return fmt.Errorf("missing required param 'start' for metric property: %s", property.Name)
	}
	if _, hasEnd := params["end"]; !hasEnd {
		return fmt.Errorf("missing required param 'end' for metric property: %s", property.Name)
	}

	// 3. 检查 step 字段
	stepVal, hasStep := params["step"]

	// instant=true 时，不应该有 step
	if instant && hasStep {
		return fmt.Errorf("metric property %s: instant=true cannot have 'step' field", property.Name)
	}

	// instant=false 时，必须有 step
	if !instant && !hasStep {
		return fmt.Errorf("metric property %s: instant=false must have 'step' field", property.Name)
	}

	// 4. 如果有 step，校验枚举值
	if hasStep {
		step, ok := stepVal.(string)
		if !ok {
			return fmt.Errorf("param 'step' must be string for metric property: %s", property.Name)
		}

		validSteps := []string{"day", "week", "month", "quarter", "year"}
		isValid := false
		for _, validStep := range validSteps {
			if step == validStep {
				isValid = true
				break
			}
		}

		if !isValid {
			return fmt.Errorf("metric property %s: invalid step value '%s', must be one of: day, week, month, quarter, year",
				property.Name, step)
		}
	}

	// 5. 校验 start 和 end 是数字类型（时间戳）
	if err := s.validateTimestamp(ctx, params["start"], "start", property.Name); err != nil {
		return err
	}
	if err := s.validateTimestamp(ctx, params["end"], "end", property.Name); err != nil {
		return err
	}

	s.logger.WithContext(ctx).Debugf("[KnLogicPropertyResolver] Metric params validation passed for: %s", property.Name)
	return nil
}

// validateTimestamp 校验时间戳参数
func (s *knLogicPropertyResolverService) validateTimestamp(
	_ context.Context,
	value interface{},
	paramName, propertyName string,
) error {
	switch v := value.(type) {
	case int64:
		// 校验时间戳范围（毫秒级，大致在 2000-2100 年之间）
		if v < 946684800000 || v > 4102444800000 {
			return fmt.Errorf("metric property %s: param '%s' timestamp %d is out of reasonable range",
				propertyName, paramName, v)
		}
		return nil
	case float64:
		// JSON 解析可能将数字解析为 float64
		timestamp := int64(v)
		if timestamp < 946684800000 || timestamp > 4102444800000 {
			return fmt.Errorf("metric property %s: param '%s' timestamp %d is out of reasonable range",
				propertyName, paramName, timestamp)
		}
		return nil
	case int:
		timestamp := int64(v)
		if timestamp < 946684800000 || timestamp > 4102444800000 {
			return fmt.Errorf("metric property %s: param '%s' timestamp %d is out of reasonable range",
				propertyName, paramName, timestamp)
		}
		return nil
	default:
		return fmt.Errorf("metric property %s: param '%s' must be a number (int64 timestamp), got %T",
			propertyName, paramName, value)
	}
}

// validateOperatorParams 校验 operator 类型的参数
//
//nolint:unparam // 保持接口一致性，error 返回用于后续扩展
func (s *knLogicPropertyResolverService) validateOperatorParams(
	ctx context.Context,
	_ *interfaces.LogicPropertyDef,
	_ map[string]interface{},
) error {
	// TODO: 实现 operator 参数校验
	// 1. 检查所有 value_from="input" 的参数是否都存在
	// 2. 参数类型校验（基础类型/对象/数组）
	// 3. 必填参数检查

	s.logger.WithContext(ctx).Warnf("[KnLogicPropertyResolver] validateOperatorParams: TODO - not implemented yet")
	return nil
}

// queryLogicProperties 调用 ontology-query 查询逻辑属性值
func (s *knLogicPropertyResolverService) queryLogicProperties(
	ctx context.Context,
	req *interfaces.ResolveLogicPropertiesRequest,
	dynamicParams map[string]interface{},
) ([]map[string]interface{}, error) {
	// 构建查询请求
	queryReq := &interfaces.QueryLogicPropertiesReq{
		KnID:               req.KnID,
		OtID:               req.OtID,
		InstanceIdentities: req.InstanceIdentities,
		Properties:         req.Properties,
		DynamicParams:      dynamicParams,
	}

	// 调用 ontology-query 服务
	resp, err := s.ontologyQueryClient.QueryLogicProperties(ctx, queryReq)
	if err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("query logic properties failed: %v", err))
	}

	return resp.Datas, nil
}

// buildMissingParamsError 构建缺参错误
func (s *knLogicPropertyResolverService) buildMissingParamsError(
	ctx context.Context,
	missingParams []interfaces.MissingPropertyParams,
	debugInfo *interfaces.ResolveDebugInfo,
) error {
	// 构建错误消息（用于 ErrorMsg 字段）
	errorMsg := ""
	for i, mp := range missingParams {
		if i > 0 {
			errorMsg += "; "
		}
		if mp.ErrorMsg != "" {
			errorMsg += fmt.Sprintf("missing %s: %s", mp.Property, mp.ErrorMsg)
		} else {
			errorMsg += fmt.Sprintf("missing %s", mp.Property)
		}
	}

	missingError := &interfaces.MissingParamsError{
		ErrorCode: "MISSING_INPUT_PARAMS",
		Message:   "dynamic_params 缺少必需的 input 参数",
		ErrorMsg:  errorMsg,
		Debug:     debugInfo,
		TraceID:   "",
		Missing:   missingParams,
	}

	// 返回为 HTTPError
	return errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("%+v", missingError))
}
