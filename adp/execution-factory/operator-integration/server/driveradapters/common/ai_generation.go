package common

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/aigeneration"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// AIGenerationHandler AI生成处理接口
type AIGenerationHandler interface {
	FunctionAIGeneration(c *gin.Context)
	// GetPromptTemplate 获取指定类型的提示词模板
	GetPromptTemplate(c *gin.Context)
}

type aiGenerationHandler struct {
	aiGenerationService interfaces.AIGenerationService
	Logger              interfaces.Logger
	Validator           interfaces.Validator
}

var (
	aiGenerationHandlerOnce sync.Once
	aiGenerationH           AIGenerationHandler
)

// NewAIGenerationHandler 创建 AI 生成处理接口实例
func NewAIGenerationHandler() AIGenerationHandler {
	aiGenerationHandlerOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		aiGenerationH = &aiGenerationHandler{
			aiGenerationService: aigeneration.NewAIGenerationService(),
			Logger:              confLoader.GetLogger(),
			Validator:           validator.NewValidator(),
		}
	})
	return aiGenerationH
}

// FunctionAIGeneration 处理函数 AI 生成请求
func (h *aiGenerationHandler) FunctionAIGeneration(c *gin.Context) {
	req := &interfaces.FunctionAIGenerateReq{}
	if err := c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	if err := c.ShouldBindJSON(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	err := h.Validator.ValidatorStruct(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = req.Validate()
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	if !req.Stream {
		resp, err := h.aiGenerationService.FunctionAIGenerate(c.Request.Context(), req)
		if err != nil {
			rest.ReplyError(c, err)
			return
		}
		rest.ReplyOK(c, http.StatusOK, resp)
		return
	}
	messageChan, errorChan, err := h.aiGenerationService.FunctionAIGenerateStream(c.Request.Context(), req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusInternalServerError, err.Error())
		rest.ReplyError(c, err)
		return
	}
	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")
	c.Header("Access-Control-Allow-Credentials", "false")
	// SSE 响应
	var finish bool
	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				return false // 消息通道已关闭，结束流
			}
			// 检查是否为结束标记
			if isEndMarker(msg) {
				// 发送SSE结束标记
				fmt.Fprintf(w, "%s\n\n", msg)
				flushIfSupported(w)
				return false
			}

			// 转发前对data内数据检查，如果和预期格式不符合直接报错结束流
			if strings.HasPrefix(msg, "data:") {
				content := strings.TrimPrefix(msg, "data:") // 移除"data:"前缀
				// 结果预期格式
				result := &interfaces.ChatCompletionResp{}
				err = utils.StringToObject(content, result)
				if err != nil {
					// 提示模型异常，返回错误
					h.Logger.WithContext(c.Request.Context()).Error(fmt.Sprintf("invalid SSE data format: %s, err: %s", content, err.Error()))
					err = errors.NewHTTPError(c.Request.Context(), http.StatusBadRequest, errors.ErrExtFunctionAIGenerateModelFailed, fmt.Sprintf("invalid SSE data format: %s, err: %s", content, err.Error()))
					c.SSEvent("error", utils.ObjectToJSON(err))
					flushIfSupported(w) // 确保错误消息立即发送
					return false
				}
				if len(result.Choices) > 0 && result.Choices[0].FinishReason == "stop" {
					finish = true
				}
				// 检查是否有choices
				if !finish && len(result.Choices) == 0 && result.Model == "" && result.ID == "" && result.Object == "" {
					h.Logger.WithContext(c.Request.Context()).Error(fmt.Sprintf("invalid SSE data format: %s", content))
					err = errors.NewHTTPError(c.Request.Context(), http.StatusBadRequest, errors.ErrExtFunctionAIGenerateModelFailed, fmt.Sprintf("invalid SSE data format: %s", content))
					c.SSEvent("error", utils.ObjectToJSON(err))
					flushIfSupported(w) // 确保错误消息立即发送
					return false
				}
			}
			fmt.Fprintf(w, "%s\n\n", msg)
			flushIfSupported(w)
			return true
		case err, ok := <-errorChan:
			if !ok {
				return false // 错误通道已关闭，结束流
			}
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				c.SSEvent("data", " [DONE]")
				flushIfSupported(w) // 确保最后一条消息立即发送
				return false
			}
			if err != io.ErrUnexpectedEOF && err != io.EOF {
				errMsg := errors.DefaultHTTPError(c.Request.Context(), http.StatusInternalServerError, err.Error())
				c.SSEvent("error", utils.ObjectToJSON(errMsg))

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
			return false // 发生错误，结束流
		case <-c.Request.Context().Done():
			// 客户端断开连接或请求被取消
			h.Logger.WithContext(c.Request.Context()).Info("SSE connection closed by client")
			return false
		}
	})
}

// flushIfSupported 确保数据立即发送
func flushIfSupported(w io.Writer) {
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// isEndMarker 检查是否为结束标记
func isEndMarker(line string) bool {
	// 常见的结束标记模式
	endMarkers := []string{
		"data: [DONE]",
		"data: [END]",
		"data: DONE",
		"data: END",
		"[DONE]",
		"[END]",
	}

	for _, marker := range endMarkers {
		if strings.Contains(line, marker) {
			return true
		}
	}
	return false
}

// GetPromptTemplate 获取指定类型的提示词模板
func (h *aiGenerationHandler) GetPromptTemplate(c *gin.Context) {
	req := &interfaces.GetPromptTemplateReq{}
	if err := c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	err := h.Validator.ValidatorStruct(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	promptTemplate, err := h.aiGenerationService.GetPromptTemplate(c.Request.Context(), req.Type)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, promptTemplate)
}
