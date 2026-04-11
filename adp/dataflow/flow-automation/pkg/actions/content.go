package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"regexp"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	libstore "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/store"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils/extractor"
)

type ContentAbstract struct {
	DocID   string `json:"docid"`
	Version string `json:"version"`
}

func (a *ContentAbstract) Name() string {
	return common.OpContentAbstract
}

func (a *ContentAbstract) ParameterNew() interface{} {
	return &ContentAbstract{}
}

func (a *ContentAbstract) Run(ctx entity.ExecuteContext, input interface{}, _ *entity.Token) (interface{}, error) {
	var err error
	var result map[string]interface{}
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)

	params := input.(*ContentAbstract)

	asdoc := ctx.NewASDoc()
	res, err := asdoc.DocSetSubdocAbstract(newCtx, params.DocID, params.Version)

	if err != nil {
		ctx.Trace(newCtx, fmt.Sprintf("DocSetSubdocAbstract error: %v", err))
		result = map[string]interface{}{
			"doc_id":  params.DocID,
			"version": params.Version,
			"status":  "error",
			"url":     "",
			"data":    "",
			"err_msg": err.Error(),
		}
	} else {
		result = map[string]interface{}{
			"doc_id":  res.DocID,
			"version": res.Version,
			"status":  res.Status,
			"url":     res.Url,
			"data":    res.Data,
			"err_msg": "",
		}
	}

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)
	ctx.Trace(newCtx, "run end")
	return result, nil
}

type ContentFullText struct {
	DocID   string `json:"docid"`
	Version string `json:"version"`
}

func (a *ContentFullText) Name() string {
	return common.OpContentFullText
}

func (a *ContentFullText) ParameterNew() interface{} {
	return &ContentFullText{}
}

func (a *ContentFullText) Run(ctx entity.ExecuteContext, input interface{}, _ *entity.Token) (interface{}, error) {
	var err error
	var result map[string]interface{}
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	params := input.(*ContentFullText)

	asdoc := ctx.NewASDoc()
	res, err := asdoc.DocSetSubdocFulltext(newCtx, params.DocID, params.Version)

	if err != nil {
		ctx.Trace(newCtx, fmt.Sprintf("DocSetSubdocFulltext error: %v", err))
		result = map[string]interface{}{
			"doc_id":  params.DocID,
			"version": params.Version,
			"status":  "error",
			"url":     "",
			"err_msg": err.Error(),
		}
	} else {
		result = map[string]interface{}{
			"doc_id":  res.DocID,
			"version": res.Version,
			"status":  res.Status,
			"url":     res.Url,
			"err_msg": res.ErrMsg,
		}
	}

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)
	ctx.Trace(newCtx, "run end")
	return result, nil
}

type EntityTemp struct {
	Name  string   `json:"name" yaml:"name"`
	Attrs []string `json:"attrs" yaml:"attrs"`
}

// RelationTemp 图谱关系模板
type RelationTemp struct {
	Start string   `json:"start" yaml:"start"`
	End   string   `json:"end" yaml:"end"`
	Name  string   `json:"name" yaml:"name"`
	Attrs []string `json:"attrs"  yaml:"attrs"`
}

type ContentEntity struct {
	DocID     string   `json:"docid"`
	Version   string   `json:"version"`
	GraphID   uint64   `json:"graph_id"`
	MinFraq   int      `json:"min_fraq"`
	Priority  int      `json:"priority"`
	EntityIds []string `json:"entity_ids"`
	EdgeIds   []string `json:"edge_ids"`
}

type Relationship struct {
	Start RelationStart `json:"start"`
	End   RelationEnd   `json:"end"`
	Name  string        `json:"name,omitempty"`
}

type RelationStart struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RelationEnd struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (a *ContentEntity) Name() string {
	return common.OpContentEntity
}

func (a *ContentEntity) ParameterNew() interface{} {
	return &ContentEntity{}
}

func (a *ContentEntity) Run(ctx entity.ExecuteContext, input interface{}, token *entity.Token) (interface{}, error) {
	var err error
	var result map[string]interface{}
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)

	taskIns := ctx.GetTaskInstance()
	taskID := ctx.GetTaskID()

	params := input.(*ContentEntity)
	asdoc := ctx.NewASDoc()

	docID := params.DocID

	if len(docID) >= 32 {
		docID = docID[len(docID)-32:]
	}

	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	ad := drivenadapters.NewAnyData()
	graphInfo, err := ad.GetGraphInfo(newCtx, params.GraphID, token.Token)

	if err != nil {
		result = map[string]interface{}{
			"doc_id":  params.DocID,
			"rev":     params.Version,
			"status":  "error",
			"err_msg": err.Error(),
			"data":    "{}",
			"url":     "",
		}
		ctx.ShareData().Set(statusKey, entity.TaskInstanceStatusSuccess)
		ctx.ShareData().Set(taskID, result)
		ctx.Trace(newCtx, fmt.Sprintf("run error: %v", err))
		return result, nil
	}

	entityIdMap := make(map[string]interface{})
	for _, id := range params.EntityIds {
		entityIdMap[id] = true
	}

	edgeIdMap := make(map[string]interface{})
	for _, id := range params.EdgeIds {
		edgeIdMap[id] = true
	}

	entities := make([]*EntityTemp, 0)
	relationships := make([]*RelationTemp, 0)

	for _, entity := range graphInfo.Entity {
		if _, ok := entityIdMap[entity.EntityID]; ok {
			attrs := make([]string, 0)
			for _, property := range entity.Properties {
				attrs = append(attrs, property.Name)
			}
			entities = append(entities, &EntityTemp{
				Name:  entity.Name,
				Attrs: attrs,
			})
		}
	}

	for _, edge := range graphInfo.Edge {
		if _, ok := edgeIdMap[edge.EdgeID]; ok {

			if len(edge.Relation) != 3 {
				continue
			}

			start := edge.Relation[0]
			end := edge.Relation[2]
			name := strings.TrimPrefix(edge.Relation[1], start+"_2_"+end+"_")

			if start == "document" {
				start = "$this"
			}

			if end == "document" {
				end = "$this"
			}

			attrs := make([]string, 0)
			for _, property := range edge.Properties {
				attrs = append(attrs, property.Name)
			}
			relationships = append(relationships, &RelationTemp{
				Name:  name,
				Start: start,
				End:   end,
				Attrs: attrs,
			})
		}
	}

	var res drivenadapters.DocSetSubdocRes

	subdocParams := drivenadapters.DocSetSubdocParams{
		DocID:    docID,
		Version:  params.Version,
		Type:     "graph_info",
		Format:   drivenadapters.DocSetSubdocFormatRaw,
		Priority: 1,
		Custom: map[string]interface{}{
			"graph_id":      params.GraphID,
			"doc_id":        docID,
			"version":       params.Version,
			"min_freq":      3,
			"entities":      entities,
			"relationships": relationships,
			"priority":      1,
		},
	}

	res, err = asdoc.DocSetSubdoc(newCtx, subdocParams, -1)

	if err != nil {
		ctx.Trace(newCtx, fmt.Sprintf("run error: %v", err))
		ctx.ShareData().Set(statusKey, entity.TaskInstanceStatusSuccess)
		result = map[string]interface{}{
			"doc_id":  params.DocID,
			"rev":     params.Version,
			"status":  "error",
			"err_msg": err.Error(),
			"data":    "{}",
			"url":     "",
		}
	} else {

		data := res.Data

		if data == nil || data == "" {
			data = "{}"
		}

		result = map[string]interface{}{
			"doc_id":  res.DocID,
			"rev":     res.Rev,
			"status":  res.Status,
			"err_msg": res.ErrMsg,
			"data":    data,
			"url":     res.Url,
		}

		relationshipMap, dok := data.(map[string]interface{})
		if dok {
			relations := []*Relationship{}
			resultMap := make(map[string]interface{}, 0)

			relationsBytes, _ := json.Marshal(relationshipMap["relationships"])
			_ = json.Unmarshal(relationsBytes, &relations)

			for _, val := range relations {
				if v, ok := resultMap[val.End.Name]; ok {
					vMap := v.(map[string]interface{})
					switch vv := vMap["_vid"].(type) {
					case []interface{}:
						vv = append(vv, val.End.ID)
						vMap["_vid"] = vv
					case string:
						vInterface := []interface{}{vv, val.End.ID}
						vMap["_vid"] = vInterface
					default:
						continue
					}
					resultMap[val.End.Name] = vMap
				} else {
					resultMap[val.End.Name] = map[string]interface{}{"_vid": val.End.ID}
				}
			}
			result["relation_map"] = resultMap
		}

		if res.Status == "processing" {
			taskBlockKey := fmt.Sprintf("%s%s", entity.ContentEntityKeyPrefix, docID)

			redis := libstore.NewRedis()
			client := redis.GetClient()
			err = client.HSet(ctx.Context(), taskBlockKey, taskIns.ID, "").Err()
			bytes, _ := json.Marshal(subdocParams)
			result["__subdocParams"] = string(bytes)
			if err != nil {
				ctx.ShareData().Set(statusKey, entity.TaskInstanceStatusSuccess)
			} else {
				_ = client.Expire(ctx.Context(), taskBlockKey, time.Hour*24).Err()
				ctx.ShareData().Set(statusKey, entity.TaskInstanceStatusBlocked)
			}
		} else {
			ctx.ShareData().Set(statusKey, entity.TaskInstanceStatusSuccess)
		}
	}
	ctx.ShareData().Set(taskID, result)
	ctx.Trace(newCtx, "run end")

	return result, nil
}

func (a *ContentEntity) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	taskIns := ctx.GetTaskInstance()
	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	status, ok := ctx.ShareData().Get(statusKey)
	if ok && status == entity.TaskInstanceStatusBlocked {
		return entity.TaskInstanceStatusBlocked, nil
	}

	return entity.TaskInstanceStatusSuccess, nil
}

type SliceVector string

const (
	SliceVectorNone  = "none"
	SliceVectorSlice = "slice"
	SliceVectorBoth  = "slice_vector"
)

type ContentFileParse struct {
	SourceType  SourceType  `json:"source_type"`
	DocID       string      `json:"docid"`
	Version     string      `json:"version"`
	Filename    string      `json:"filename"`
	Url         string      `json:"url"`
	SliceVector SliceVector `json:"slice_vector"`
	Model       string      `json:"model"`

	// 多模态配置(平铺展示)
	// 是否启用多模态图片描述
	MultimodalEnabled bool `json:"multimodal_enabled,omitempty"`

	// 多模态模型标识(如 "gpt-4-vision-preview", "claude-3-opus", "gemini-pro-vision")
	MultimodalModelName string `json:"multimodal_model_name,omitempty"`

	// 提示词模板(由前端传递默认值)
	MultimodalPromptTemplate string `json:"multimodal_prompt_template,omitempty"`

	// 多模态模型参数配置
	Temperature      float64 `json:"temperature,omitempty"`
	TopP             float64 `json:"top_p,omitempty"`
	MaxTokens        int     `json:"max_tokens,omitempty"`
	TopK             int     `json:"top_k,omitempty"`
	PresencePenalty  float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
}

func (a *ContentFileParse) Name() string {
	return common.OpContentFileParse
}

func (a *ContentFileParse) ParameterNew() interface{} {
	return &ContentFileParse{}
}

func (a *ContentFileParse) Validate(ctx entity.ExecuteContext) error {
	switch a.SourceType {
	case SourceTypeUrl:
		if a.Filename == "" {
			return fmt.Errorf("filename is required")
		}
		return nil
	default:
		if a.DocID == "" {
			return fmt.Errorf("docid is required")
		}

		downloadInfo, err := GetFileDownloadInfo(ctx.Context(), ctx, a.DocID, a.Version)
		if err != nil {
			return err
		}

		a.Url = downloadInfo.URL
		a.Filename = downloadInfo.Filename
		return nil
	}
}

func (a *ContentFileParse) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)

	input := params.(*ContentFileParse)

	err = input.Validate(ctx)
	if err != nil {
		traceLog.WithContext(ctx.Context()).Warnf("[ContentFileParse] validate err: %s", err.Error())
		return nil, err
	}

	// 创建执行器包装器
	executor := &contentFileParseExecutor{
		action:  a,
		input:   input,
		token:   token,
		context: ctx,
	}

	// 使用通用的异步任务缓存管理器（使用统一的 topic）
	manager := NewAsyncTaskManager(ctx.NewExecuteMethods()).
		WithLockPrefix("automation:file_parse")

	return manager.Run(ctx, executor)
}

func (a *ContentFileParse) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	manager := NewAsyncTaskManager(ctx.NewExecuteMethods())
	return manager.RunAfter(ctx)
}

// contentFileParseExecutor 实现 AsyncTaskExecutor 接口
type contentFileParseExecutor struct {
	action  *ContentFileParse
	input   *ContentFileParse
	token   *entity.Token
	context entity.ExecuteContext
}

func (e *contentFileParseExecutor) GetTaskType() string {
	return e.action.Name()
}

func (e *contentFileParseExecutor) GetHashContent() string {
	if e.input.SourceType == SourceTypeDocID {
		return fmt.Sprintf("%s:%s:%s:%s:%s:%s", e.action.Name(), e.input.SourceType, e.input.DocID, e.input.Version, e.input.Model, e.input.SliceVector)
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s", e.action.Name(), e.input.SourceType, e.input.Url, e.input.Model, e.input.SliceVector)
}

func (e *contentFileParseExecutor) GetExpireSeconds() int64 {
	config := common.NewConfig()
	return config.ActionConfig.FileParse.ExpireSec
}

func (e *contentFileParseExecutor) GetResultFileExt() string {
	return ".json"
}

func (e *contentFileParseExecutor) Execute(ctx context.Context) (map[string]interface{}, error) {
	log := traceLog.WithContext(ctx)

	structureExtractor := drivenadapters.NewStructureExtractor()
	result := make(map[string]any)

	parseResult, contentList, err := structureExtractor.FileParse(ctx, e.input.Url, e.input.Filename)

	if err != nil {
		log.Warnf("[contentFileParseExecutor] FileParse err: %s, url %s", err.Error(), e.input.Url)
		return nil, err
	}

	if parseResult != nil {

		contentSlice := make([]any, 0)
		chunks := make([]any, 0)

		// 先计算 docMD5（用于生成去重标识）
		docMD5 := extractor.GenerateMD5(parseResult.MdContent)

		// 如果启用了多模态图片描述，先处理图片元素
		var imageMapping map[string]*ImagePathMapping
		if e.input.MultimodalEnabled && e.input.MultimodalModelName != "" && len(contentList) > 0 {
			// 判断输入文件是否是图片文件
			var originalFileURL string
			fileExt := strings.ToLower(filepath.Ext(e.input.Filename))
			isImageFile := fileExt == ".jpg" || fileExt == ".jpeg" || fileExt == ".png" ||
				fileExt == ".bmp" || fileExt == ".tif" || fileExt == ".tiff" ||
				fileExt == ".gif" || fileExt == ".webp"

			if isImageFile && len(contentList) == 1 && contentList[0].Type == "image" {
				// 输入文件本身就是图片，使用原始文件URL
				originalFileURL = e.input.Url
			}

			var err2 error
			imageMapping, err2 = generateMultimodalImageDescriptions(ctx, contentList, e.input.MultimodalModelName, e.input.MultimodalPromptTemplate, e.input.Filename, originalFileURL, e.token, e.input.Temperature, e.input.TopP, e.input.MaxTokens, e.input.TopK, e.input.PresencePenalty, e.input.FrequencyPenalty)
			if err2 != nil {
				// 多模态处理失败不影响整体流程，只记录警告
				log.Warnf("[ContentFileParse] generateMultimodalImageDescriptions err: %s", err2.Error())
			}
		}

		// 将 contentList 转换为 Element 格式的 content_list
		var elements []*extractor.Element
		if len(contentList) > 0 {
			documentID := e.input.DocID
			if documentID == "" {
				// 如果 DocID 为空，使用文件名生成一个临时 ID
				documentID = fmt.Sprintf("doc_%s", strings.ReplaceAll(e.input.Filename, ".", "_"))
			}
			// 获取文档名称（去掉后缀）
			docName := e.input.Filename
			if docName == "" {
				docName = documentID
			}

			// 判断输入文件是否是图片文件
			// 如果是图片文件，且contentList中只有一个元素且类型是image，使用原始文件URL作为img_path
			var originalFileURL string
			fileExt := strings.ToLower(filepath.Ext(e.input.Filename))
			isImageFile := fileExt == ".jpg" || fileExt == ".jpeg" || fileExt == ".png" ||
				fileExt == ".bmp" || fileExt == ".tif" || fileExt == ".tiff" ||
				fileExt == ".gif" || fileExt == ".webp"

			if isImageFile && len(contentList) == 1 && contentList[0].Type == "image" {
				// 输入文件本身就是图片，使用原始文件URL
				originalFileURL = e.input.Url
			}

			elements = extractor.ConvertContentItemsToElements(ctx, contentList, documentID, docName, docMD5, originalFileURL)
			// 将 Element 数组序列化为 JSON
			bytes, err := json.Marshal(elements)
			if err != nil {
				log.Warnf("[contentFileParseExecutor] Marshal elements err: %v", err)
				// 如果序列化失败，回退到原始 contentList
				bytes, _ := json.Marshal(contentList)
				_ = json.Unmarshal(bytes, &contentSlice)
			} else {
				_ = json.Unmarshal(bytes, &contentSlice)
			}
		}

		// 处理 chunks - 直接使用 contentList
		// 注意：Chunk 的父子关系通过 SegmentID 关联到 Element 层获取
		if e.input.SliceVector == SliceVectorSlice || e.input.SliceVector == SliceVectorBoth {
			if len(contentList) > 0 {
				// 传入 elements 列表以建立 Chunk 与 Element 的关联关系
				customChunks := extractor.ProcessCustomChunk(ctx, e.input.Filename, contentList, docMD5, e.input.SliceVector == SliceVectorBoth, e.input.Model, e.token.Token, elements)

				bytes, _ := json.Marshal(customChunks)
				_ = json.Unmarshal(bytes, &chunks)
			}
		}

		// 更新md_content：替换图片路径为OSS地址，并添加多模态描述
		mdContent := parseResult.MdContent
		if imageMapping != nil && len(imageMapping) > 0 {
			mdContent = updateMdContentWithImageMapping(mdContent, imageMapping)
		}

		result["md_content"] = mdContent
		result["content_list"] = contentSlice
		result["chunks"] = chunks
	}

	return result, nil
}

// ImagePathMapping 图片路径到OSS URL和描述的映射
type ImagePathMapping struct {
	OSSURL      string // OSS URL
	Description string // 多模态描述
}

// generateMultimodalImageDescriptions 为图片元素生成多模态描述
// 使用 goroutine 池并发处理，默认最多5个并发
// originalFileURL: 当输入文件本身就是图片时，传入原始文件URL
// 返回图片路径到OSS URL和描述的映射，用于更新md_content
func generateMultimodalImageDescriptions(ctx context.Context, contentList []*drivenadapters.ContentItem, modelName string, promptTemplate string, docName string, originalFileURL string, token *entity.Token, temperature float64, topP float64, maxTokens int, topK int, presencePenalty float64, frequencyPenalty float64) (map[string]*ImagePathMapping, error) {
	log := traceLog.WithContext(ctx)

	// 如果没有提供提示词模板，使用默认值
	if promptTemplate == "" {
		promptTemplate = "请详细描述这张图片的内容,包括主要对象、场景、颜色、布局等信息。请用简洁、准确的语言描述,不超过200字。"
	}

	// 收集所有图片元素
	var imageItems []*drivenadapters.ContentItem
	for _, item := range contentList {
		if item.Type == "image" {
			imageItems = append(imageItems, item)
		}
	}

	if len(imageItems) == 0 {
		return nil, nil
	}

	log.Infof("[generateMultimodalImageDescriptions] found %d image items, starting multimodal description generation", len(imageItems))

	// 创建 goroutine 池，默认最多5个并发
	const maxConcurrency = 5
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	// 图片路径到OSS URL和描述的映射
	imageMapping := make(map[string]*ImagePathMapping)

	ad := drivenadapters.NewAnyData()

	// 并发处理每个图片
	for i, item := range imageItems {
		wg.Add(1)
		go func(imgItem *drivenadapters.ContentItem, index int) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 获取图片URL并转换为Data URL（base64编码）
			// 流程：本地路径/原始URL → 上传到OSS → 获取OSS URL → 转换为Data URL
			// 转换为Data URL可以避免OSS内网地址导致大模型服务无法访问的问题
			var imageURL string
			var err error

			var ossURL string
			var imagePath string // 用于映射的图片路径

			// 如果提供了原始文件URL（输入文件本身就是图片），且这是第一个也是唯一的图片元素，优先使用
			if originalFileURL != "" && len(imageItems) == 1 && index == 0 {
				// 先上传原始图片到OSS（用于持久化存储）
				var err error
				ossURL, err = extractor.UploadOriginalImageToOSS(ctx, originalFileURL, docName)
				if err != nil {
					log.Warnf("[generateMultimodalImageDescriptions] failed to upload original image to OSS: %s, error: %v", originalFileURL, err)
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to upload original image %s: %w", originalFileURL, err))
					mu.Unlock()
					return
				}
				log.Infof("[generateMultimodalImageDescriptions] uploaded original image to OSS, OSS URL: %s", ossURL)
				// 使用原始文件URL作为映射key（如果imgItem有ImgPath则使用ImgPath）
				if imgItem.ImgPath != "" {
					imagePath = imgItem.ImgPath
				} else {
					// 如果没有ImgPath，使用文件名作为key
					imagePath = filepath.Base(originalFileURL)
				}
				// 将OSS URL转换为Data URL（避免内网地址问题）
				imageURL, err = ToDataURL(ossURL, docName)
				if err != nil {
					log.Warnf("[generateMultimodalImageDescriptions] failed to convert OSS URL to Data URL: %s, error: %v", ossURL, err)
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to convert OSS URL %s to Data URL: %w", ossURL, err))
					mu.Unlock()
					return
				}
			} else if imgItem.ImgPath != "" {
				// 先上传图片到OSS（注意：imgItem.ImgPath 是本地路径，如 "images/xxx.jpg"）
				var err error
				ossURL, err = extractor.UploadImageToOSS(ctx, imgItem.ImgPath, docName)
				if err != nil {
					log.Warnf("[generateMultimodalImageDescriptions] failed to upload image to OSS: %s, error: %v", imgItem.ImgPath, err)
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to upload image %s: %w", imgItem.ImgPath, err))
					mu.Unlock()
					return
				}
				log.Infof("[generateMultimodalImageDescriptions] uploaded image to OSS, local path: %s, OSS URL: %s", imgItem.ImgPath, ossURL)
				imagePath = imgItem.ImgPath
				// 将OSS URL转换为Data URL（避免内网地址问题）
				imageURL, err = ToDataURL(ossURL, docName)
				if err != nil {
					log.Warnf("[generateMultimodalImageDescriptions] failed to convert OSS URL to Data URL: %s, error: %v", ossURL, err)
					mu.Lock()
					errors = append(errors, fmt.Errorf("failed to convert OSS URL %s to Data URL: %w", ossURL, err))
					mu.Unlock()
					return
				}
			} else {
				log.Warnf("[generateMultimodalImageDescriptions] image item has no img_path and no original file URL, skipping")
				return
			}

			// 构建多模态消息
			messages := []*drivenadapters.ChatMessage{
				{
					Role: "user",
					Content: []map[string]any{
						{"type": "text", "text": promptTemplate},
						{"type": "image_url", "image_url": map[string]string{"url": imageURL}},
					},
				},
			}

			// 调用模型生成描述
			req := &drivenadapters.ChatCompletionRequest{
				Model:     modelName,
				Messages:  messages,
				MaxTokens: maxTokens,
			}
			// 如果 maxTokens 未设置，使用默认值 500
			if req.MaxTokens == 0 {
				req.MaxTokens = 500
			}
			// 设置其他可选参数
			if temperature != 0 {
				req.Temperature = temperature
			}
			if topP != 0 {
				req.TopP = topP
			}
			if topK != 0 {
				req.TopK = topK
			}
			if presencePenalty != 0 {
				req.PresencePenalty = presencePenalty
			}
			if frequencyPenalty != 0 {
				req.FrequencyPenalty = frequencyPenalty
			}
			res, err := ad.ChatCompletion(ctx, req, token.Token)

			// 无论成功与否，都要保存OSS URL映射（用于替换md_content中的图片路径）
			mu.Lock()
			mapping := &ImagePathMapping{
				OSSURL:      ossURL,
				Description: "",
			}
			mu.Unlock()

			if err != nil {
				log.Warnf("[generateMultimodalImageDescriptions] ChatCompletion failed for image %s: %v", imgItem.ImgPath, err)
				mu.Lock()
				errors = append(errors, fmt.Errorf("ChatCompletion failed for image %s: %w", imgItem.ImgPath, err))
				// 即使失败，也保存OSS URL映射
				imageMapping[imagePath] = mapping
				mu.Unlock()
				return
			}

			// 提取描述文本
			if len(res.Choices) > 0 && res.Choices[0].Message.Content != nil {
				var description string
				switch content := res.Choices[0].Message.Content.(type) {
				case string:
					description = content
				case []map[string]any:
					// 如果是数组格式，提取文本内容
					for _, item := range content {
						if itemType, ok := item["type"].(string); ok && itemType == "text" {
							if text, ok := item["text"].(string); ok {
								description = text
								break
							}
						}
					}
				default:
					// 尝试转换为字符串
					description = fmt.Sprintf("%v", content)
				}

				if description != "" {
					// 将描述写入 ContentItem.Text
					mu.Lock()
					// 如果已有文本，追加描述（用换行分隔）
					if imgItem.Text != "" {
						imgItem.Text = imgItem.Text + "\n" + description
					} else {
						imgItem.Text = description
					}
					// 更新映射中的描述
					mapping.Description = description
					imageMapping[imagePath] = mapping
					mu.Unlock()
					if imgItem.ImgPath != "" {
						log.Infof("[generateMultimodalImageDescriptions] successfully generated description for image %s", imgItem.ImgPath)
					} else {
						log.Infof("[generateMultimodalImageDescriptions] successfully generated description for original image")
					}
				} else {
					// 即使没有描述，也保存OSS URL映射（用于替换md_content中的图片路径）
					mu.Lock()
					imageMapping[imagePath] = mapping
					mu.Unlock()
				}
			} else {
				// 没有返回内容，也保存OSS URL映射
				mu.Lock()
				imageMapping[imagePath] = mapping
				mu.Unlock()
			}
		}(item, i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	// 如果有错误，返回第一个错误（但不会中断流程）
	if len(errors) > 0 {
		log.Warnf("[generateMultimodalImageDescriptions] completed with %d errors out of %d images", len(errors), len(imageItems))
		return imageMapping, errors[0]
	}

	log.Infof("[generateMultimodalImageDescriptions] successfully processed %d images", len(imageItems))
	return imageMapping, nil
}

// truncateDescription 截断描述文本，使其足够简短（用于alt text）
// 限制在 maxLength 个字符以内，如果超过则截断并添加省略号
func truncateDescription(description string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 50 // 默认最大长度50字符（alt text应该简短）
	}
	if len(description) <= maxLength {
		return description
	}
	// 截断并添加省略号
	return description[:maxLength] + "..."
}

// updateMdContentWithImageMapping 更新md_content，将图片路径替换为OSS地址，并在图片下方添加多模态描述
// 匹配格式：![](images/xxx.jpg) 或 ![alt](images/xxx.jpg)
// 替换为：![简短描述](OSS地址)\n完整描述 或 ![](OSS地址)（如果没有描述）
func updateMdContentWithImageMapping(mdContent string, imageMapping map[string]*ImagePathMapping) string {
	if mdContent == "" || len(imageMapping) == 0 {
		return mdContent
	}

	// 匹配 markdown 图片格式：![alt](path) 或 ![](path)
	// 支持可选的 alt text
	imagePattern := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	result := imagePattern.ReplaceAllStringFunc(mdContent, func(match string) string {
		// 提取 alt text 和路径
		matches := imagePattern.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match // 如果匹配失败，返回原字符串
		}

		imagePath := matches[2]

		// 查找对应的映射
		mapping, exists := imageMapping[imagePath]
		if !exists {
			// 尝试只匹配文件名（去掉路径前缀）
			fileName := filepath.Base(imagePath)
			mapping, exists = imageMapping[fileName]
			if !exists {
				// 尝试匹配完整路径（包括 images/ 前缀）
				for key, value := range imageMapping {
					if strings.HasSuffix(imagePath, key) || strings.HasSuffix(key, imagePath) {
						mapping = value
						exists = true
						break
					}
				}
			}
		}

		if !exists || mapping == nil {
			return match // 如果没有找到映射，返回原字符串
		}

		// 如果有描述
		if mapping.Description != "" {
			// 截断后的简短描述作为 alt text（限制在50字符以内）
			shortDescription := truncateDescription(mapping.Description, 50)
			// 完整描述作为图片下方的正文（不截断）
			// 返回格式：![简短描述](OSS地址)\n完整描述
			return fmt.Sprintf("![%s](%s)\n%s", shortDescription, mapping.OSSURL, mapping.Description)
		}

		// 如果没有描述，只返回图片
		return fmt.Sprintf("![](%s)", mapping.OSSURL)
	})

	return result
}
