// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knrerank 知识重排模块
// 对输入的候选概念列表进行相关性排序，过滤掉与问题无关的概念
package knrerank

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// rerankPromptTemplate 重排序提示词模板
// 变量说明：
//
//	{{.Intent}}   - 用户意图（可选）
//	{{.Question}}  - 用户问题
//	{{.Concepts}}  - 概念列表（由代码动态生成）
//
// rerankPromptTemplateStr 重排序提示词模板字符串
const rerankPromptTemplateStr = `{{if .Intent}}用户的意图是：{{.Intent}}
{{end}}问题: {{.Question}}
请仔细分析以下概念列表，只返回与上述问题相关的概念编号。对于与问题无关的概念，请完全忽略，不要在输出中包含它们。请按照与问题相关性从高到低的顺序返回概念编号。
概念列表:{{.Concepts}}
请只返回与问题相关的概念编号，按照相关性排序，以列表形式返回，例如：[3, 1, 5]，都不相关，返回[]`

const (
	promptBufferBaseSize       = 350
	promptBufferPerConceptSize = 100
)

var (
	// 预编译模板，init时执行 panic 确保模板无误
	rerankTemplate = template.Must(template.New("rerank").Parse(rerankPromptTemplateStr))
)

// promptData 模板渲染数据结构
type promptData struct {
	Intent   string // 用户意图（可选）
	Question string // 用户问题
	Concepts string // 概念列表
}

// KnowledgeReranker 知识重排器
type KnowledgeReranker struct {
	mfModelClient interfaces.DrivenMFModelAPIClient
	logger        interfaces.Logger
	config        *config.RerankLLMConfig
}

// 单例支持
var (
	rerankerOnce sync.Once
	rerankerInst *KnowledgeReranker
)

// NewKnowledgeReranker 获取知识重排器单例实例
func NewKnowledgeReranker(
	mfModelClient interfaces.DrivenMFModelAPIClient,
	logger interfaces.Logger,
) *KnowledgeReranker {
	rerankerOnce.Do(func() {
		conf := config.NewConfigLoader()
		rerankerInst = &KnowledgeReranker{
			mfModelClient: mfModelClient,
			logger:        logger,
			config:        &conf.RerankLLM,
		}
	})
	return rerankerInst
}

// Rerank 主入口：对概念进行重排序
func (r *KnowledgeReranker) Rerank(ctx context.Context, req *interfaces.KnowledgeRerankReq) ([]*interfaces.ConceptResult, error) {
	if len(req.KnowledgeConcepts) == 0 {
		return []*interfaces.ConceptResult{}, nil
	}

	// 根据action类型分发
	switch req.Action {
	case interfaces.KnowledgeRerankActionVector:
		return r.rerankByVector(ctx, req)
	case interfaces.KnowledgeRerankActionLLM, interfaces.KnowledgeRerankActionDefault, "":
		return r.rerankByLLM(ctx, req)
	default:
		// 默认使用LLM模式
		return r.rerankByLLM(ctx, req)
	}
}

// rerankByLLM 使用LLM进行重排序
func (r *KnowledgeReranker) rerankByLLM(ctx context.Context, req *interfaces.KnowledgeRerankReq) ([]*interfaces.ConceptResult, error) {
	concepts := req.KnowledgeConcepts
	question := ""
	var intents []interface{}

	if req.QueryUnderstanding != nil {
		question = req.QueryUnderstanding.OriginQuery
		if len(req.QueryUnderstanding.Intent) > 0 {
			for _, intent := range req.QueryUnderstanding.Intent {
				intents = append(intents, intent)
			}
		}
	}

	// 获取账号信息（用于Header透传）
	accountID := ""
	accountType := "user"

	// 分批处理
	batchSize := 128
	allIndices := []int{}
	seen := make(map[int]bool)

	for i := 0; i < len(concepts); i += batchSize {
		end := i + batchSize
		if end > len(concepts) {
			end = len(concepts)
		}
		batchConcepts := concepts[i:end]
		batchStartIndex := i

		// 构建Prompt
		prompt := r.buildPrompt(question, batchConcepts, intents)

		// 调用LLM
		indices, err := r.processLLMBatch(ctx, prompt, accountID, accountType)
		if err != nil {
			r.logger.WithContext(ctx).Warnf("[KnowledgeReranker#rerankByLLM] Batch %d failed: %v", i/batchSize+1, err)
			// 批次失败时，不添加任何索引（降级处理）
			continue
		}

		// 转换为全局索引
		for _, idx := range indices {
			globalIdx := idx + batchStartIndex
			if globalIdx >= 0 && globalIdx < len(concepts) && !seen[globalIdx] {
				allIndices = append(allIndices, globalIdx)
				seen[globalIdx] = true
			}
		}
	}

	// 记录LLM返回的索引集合
	llmReturnedIndices := make(map[int]bool)
	for _, idx := range allIndices {
		llmReturnedIndices[idx] = true
	}

	// 补充未选中的索引（保持原顺序）
	for i := 0; i < len(concepts); i++ {
		if !seen[i] {
			allIndices = append(allIndices, i)
		}
	}

	// 构建结果：二值化打分
	results := make([]*interfaces.ConceptResult, len(concepts))
	for i, idx := range allIndices {
		concept := concepts[idx]
		// 深拷贝概念
		result := *concept
		if llmReturnedIndices[idx] {
			result.RerankScore = 1.0 // LLM选中的概念
		} else {
			result.RerankScore = 0.0 // 未选中的概念
		}
		results[i] = &result
	}

	r.logger.WithContext(ctx).Infof("[KnowledgeReranker#rerankByLLM] Reranked %d concepts, %d selected",
		len(concepts), len(llmReturnedIndices))

	return results, nil
}

// rerankByVector 使用向量服务进行重排序
func (r *KnowledgeReranker) rerankByVector(ctx context.Context, req *interfaces.KnowledgeRerankReq) ([]*interfaces.ConceptResult, error) {
	concepts := req.KnowledgeConcepts
	question := ""
	if req.QueryUnderstanding != nil {
		question = req.QueryUnderstanding.OriginQuery
	}

	// 格式化概念为文本切片
	documents := make([]string, len(concepts))
	for i, concept := range concepts {
		text, err := r.formatConceptText(concept)
		if err != nil {
			r.logger.WithContext(ctx).Warnf("[KnowledgeReranker#rerankByVector] Format concept failed: %v", err)
			documents[i] = concept.ConceptName // 降级使用名称
		} else {
			documents[i] = text
		}
	}

	// 调用向量重排服务
	resp, err := r.mfModelClient.Rerank(ctx, question, documents)
	if err != nil {
		r.logger.WithContext(ctx).Errorf("[KnowledgeReranker#rerankByVector] Rerank service failed: %v", err)
		// 降级：返回原顺序，分数都为0
		results := make([]*interfaces.ConceptResult, len(concepts))
		for i, concept := range concepts {
			result := *concept
			result.RerankScore = 0.0
			results[i] = &result
		}
		return results, nil
	}

	// 创建索引到分数的映射
	scoreMap := make(map[int]float64)
	for _, result := range resp.Results {
		scoreMap[result.Index] = result.RelevanceScore
	}

	// 构建结果并按分数排序
	type scoredConcept struct {
		concept *interfaces.ConceptResult
		score   float64
	}
	scored := make([]scoredConcept, len(concepts))
	for i, concept := range concepts {
		result := *concept
		if score, ok := scoreMap[i]; ok {
			result.RerankScore = score
		} else {
			result.RerankScore = 0.0
		}
		scored[i] = scoredConcept{concept: &result, score: result.RerankScore}
	}

	// 按分数降序排序
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// 提取结果
	results := make([]*interfaces.ConceptResult, len(scored))
	for i, s := range scored {
		results[i] = s.concept
	}

	r.logger.WithContext(ctx).Infof("[KnowledgeReranker#rerankByVector] Reranked %d concepts", len(concepts))

	return results, nil
}

// buildPrompt 构建重排序Prompt
func (r *KnowledgeReranker) buildPrompt(question string, concepts []*interfaces.ConceptResult, intents []interface{}) string {
	// 1. 构建概念列表
	conceptsList := r.buildConceptsList(concepts)

	// 2. 格式化意图文本
	intentText := ""
	if len(intents) > 0 {
		intentText = r.formatIntentsText(intents)
	}

	// 3. 渲染模板
	data := promptData{
		Intent:   intentText,
		Question: question,
		Concepts: conceptsList,
	}

	var buf bytes.Buffer
	buf.Grow(promptBufferBaseSize + len(concepts)*promptBufferPerConceptSize)

	if err := rerankTemplate.Execute(&buf, data); err != nil {
		r.logger.Errorf("[KnowledgeReranker#buildPrompt] Template execution failed: %v", err)
		// 降级：模板已预编译且数据为纯字符串，极少出错，这里返回简化版本
		return fmt.Sprintf("问题: %s\n概念列表:%s", question, conceptsList)
	}

	return buf.String()
}

// buildConceptsList 构建概念列表字符串
func (r *KnowledgeReranker) buildConceptsList(concepts []*interfaces.ConceptResult) string {
	var builder strings.Builder
	builder.Grow(len(concepts) * promptBufferPerConceptSize)

	for i, concept := range concepts {
		text, err := r.formatConceptText(concept)
		if err != nil {
			text = concept.ConceptName // 降级使用名称
		}
		fmt.Fprintf(&builder, "\n[%d] %s", i+1, text)
	}

	return builder.String()
}

// formatConceptText 将概念格式化为自然语言文本
func (r *KnowledgeReranker) formatConceptText(concept *interfaces.ConceptResult) (string, error) {
	conceptType := string(concept.ConceptType)
	conceptName := concept.ConceptName

	// 校验必填字段
	if conceptType == "" || conceptName == "" {
		return "", fmt.Errorf("概念缺少必要的类型或名称信息")
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("我们有一个'%s'的概念", conceptName))

	// 添加概念详情 - 需要类型断言
	if concept.ConceptDetail != nil {
		conceptDetail, ok := concept.ConceptDetail.(map[string]interface{})
		if ok {
			// 处理comment字段 - 直接拼接"描述为xxx"
			if comment, ok := conceptDetail["comment"].(string); ok && comment != "" {
				text.WriteString(fmt.Sprintf("，描述为%s", comment))
			}

			// 处理data_properties下的comment字段 - 放入属性描述数组
			var attrDescriptions []string
			if dataProps, ok := conceptDetail["data_properties"].([]interface{}); ok {
				for _, prop := range dataProps {
					if propMap, ok := prop.(map[string]interface{}); ok {
						if propComment, ok := propMap["comment"].(string); ok && propComment != "" {
							attrDescriptions = append(attrDescriptions, propComment)
						}
					}
				}
			}

			if len(attrDescriptions) > 0 {
				text.WriteString("，具有")
				text.WriteString(strings.Join(attrDescriptions, "，"))
			}
		}
	}

	text.WriteString("。")
	return text.String(), nil
}

// formatIntentsText 将意图列表格式化为文本
func (r *KnowledgeReranker) formatIntentsText(intents []interface{}) string {
	if len(intents) == 0 {
		return ""
	}

	var intentTexts []string
	for i, intent := range intents {
		intentMap, ok := intent.(map[string]interface{})
		if !ok {
			continue
		}

		querySegment, _ := intentMap["query_segment"].(string)
		if querySegment == "" {
			continue
		}

		var parts []string
		parts = append(parts, fmt.Sprintf("意图%d是关于'%s'", i+1, querySegment))

		if confidence, ok := intentMap["confidence"].(float64); ok && confidence > 0 {
			parts = append(parts, fmt.Sprintf("置信度为%v", confidence))
		}

		if reasoning, ok := intentMap["reasoning"].(string); ok && reasoning != "" {
			parts = append(parts, fmt.Sprintf("推理说明：%s", reasoning))
		}

		// 处理相关概念 - 回退逻辑
		if relatedConcepts, ok := intentMap["related_concepts"].([]interface{}); ok && len(relatedConcepts) > 0 {
			var names []string
			for _, rc := range relatedConcepts {
				rcMap, ok := rc.(map[string]interface{})
				if !ok {
					continue
				}
				// 先取concept_name，不存在则取concept_id
				name := ""
				if n, ok := rcMap["concept_name"].(string); ok && n != "" {
					name = n
				} else if id, ok := rcMap["concept_id"].(string); ok && id != "" {
					name = id
				}
				if name != "" {
					names = append(names, name)
				}
			}
			if len(names) > 0 {
				parts = append(parts, fmt.Sprintf("相关概念包括：%s", strings.Join(names, ", ")))
			}
		}

		intentTexts = append(intentTexts, strings.Join(parts, "，"))
	}

	if len(intentTexts) == 0 {
		return ""
	}

	return strings.Join(intentTexts, "；") + "。"
}

// processLLMBatch 处理单个LLM批次
func (r *KnowledgeReranker) processLLMBatch(ctx context.Context, prompt, accountID, accountType string) ([]int, error) {
	// 构建消息
	messages := []interfaces.LLMMessage{
		{
			Role:    "system",
			Content: "你是一个智能概念筛选和排序助手。你能够根据用户提出的问题，从概念列表中筛选出相关的概念，并按照相关性进行排序，只返回相关概念的编号。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// 构建请求
	req := &interfaces.LLMChatReq{
		Model:            r.config.Model,
		Messages:         messages,
		Temperature:      r.config.Temperature,
		TopK:             r.config.TopK,
		TopP:             r.config.TopP,
		FrequencyPenalty: r.config.FrequencyPenalty,
		PresencePenalty:  r.config.PresencePenalty,
		MaxTokens:        r.config.MaxTokens,
		Stream:           true,
		AccountID:        accountID,
		AccountType:      accountType,
	}

	// 调用LLM
	content, err := r.mfModelClient.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	r.logger.WithContext(ctx).Debugf("[KnowledgeReranker#processLLMBatch] LLM response: %s", content)

	// 解析响应
	return r.parseIndices(content)
}

// parseIndices 从LLM响应中解析索引列表
func (r *KnowledgeReranker) parseIndices(content string) ([]int, error) {
	// 优先匹配JSON数组格式 [1, 2, 3]
	arrayRegex := regexp.MustCompile(`\[([0-9,\s]+)\]`)
	match := arrayRegex.FindStringSubmatch(content)
	if len(match) > 1 {
		indicesStr := match[1]
		return r.parseNumberList(indicesStr)
	}

	// 降级：匹配所有数字
	numberRegex := regexp.MustCompile(`\d+`)
	matches := numberRegex.FindAllString(content, -1)
	if len(matches) == 0 {
		return []int{}, nil
	}

	var indices []int
	for _, m := range matches {
		n, err := strconv.Atoi(m)
		if err == nil && n > 0 {
			indices = append(indices, n-1) // 转为0-based
		}
	}

	return indices, nil
}

// parseNumberList 解析逗号分隔的数字列表
func (r *KnowledgeReranker) parseNumberList(s string) ([]int, error) {
	var indices []int
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		if n > 0 {
			indices = append(indices, n-1) // 转为0-based
		}
	}
	return indices, nil
}
