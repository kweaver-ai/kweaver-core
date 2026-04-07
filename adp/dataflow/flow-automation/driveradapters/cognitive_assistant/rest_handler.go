package cognitiveassistant

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
)

type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once sync.Once
	rh   RESTHandler
)

var (
	CustomPromptParamsSchema = "cognitive-assistant/custom-prompt-params.json"
)

type CognitiveAssistantHandler struct {
	config             *common.Config
	hydra              drivenadapters.HydraPublic
	hydraAdmin         drivenadapters.HydraAdmin
	userMgnt           drivenadapters.UserManagement
	cognitiveAssistant drivenadapters.CognitiveAssistant
	efast              drivenadapters.Efast
}

func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &CognitiveAssistantHandler{
			config:             common.NewConfig(),
			hydra:              drivenadapters.NewHydraPublic(),
			hydraAdmin:         drivenadapters.NewHydraAdmin(),
			userMgnt:           drivenadapters.NewUserManagement(),
			cognitiveAssistant: drivenadapters.NewCognitiveAssistant(),
			efast:              drivenadapters.NewEfast(),
		}
	})
	return rh
}

func (h *CognitiveAssistantHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.GET("/custom-prompt", middleware.TokenAuth(), h.getCustomPrompts)
	engine.GET("/custom-prompt-workcenter", middleware.TokenAuth(), h.getWcCustomPrompts)
	engine.POST("/custom-prompt/:serviceID", middleware.TokenAuth(), h.customPrompt)
}

func (h *CognitiveAssistantHandler) getCustomPrompts(c *gin.Context) {
	var err error
	ctx, span := trace.StartInternalSpan(c.Request.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	result, err := h.cognitiveAssistant.GetCustomPrompts(ctx)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewIError(errors.InternalError, "", err))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CognitiveAssistantHandler) getWcCustomPrompts(c *gin.Context) {
	var err error
	ctx, span := trace.StartInternalSpan(c.Request.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	result, err := h.cognitiveAssistant.GetCustomPrompts(ctx)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewIError(errors.InternalError, "", err))
		return
	}

	wcPromptList := make([]map[string]interface{}, 0)

	if promptList, ok := result.([]interface{}); ok {
		for _, prompt := range promptList {
			if promptMap, ok := prompt.(map[string]interface{}); ok {
				if promptMap["class_name"] != "WorkCenter" {
					continue
				}
				if _, ok := promptMap["prompt"].([]interface{}); !ok {
					continue
				}

				for _, prompt := range promptMap["prompt"].([]interface{}) {
					if promptDetailMap, ok := prompt.(map[string]interface{}); ok {
						wcPromptList = append(wcPromptList, promptDetailMap)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, wcPromptList)
}

type CustomPromptParams struct {
	DocID string `json:"docid"`
}

func (h *CognitiveAssistantHandler) customPrompt(c *gin.Context) {

	var err error
	ctx, span := trace.StartInternalSpan(c.Request.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	user, isexist := c.Get("user")

	if !isexist {
		c.JSON(http.StatusUnauthorized, errors.NewIError(errors.UnAuthorization, "", map[string]string{"auth": "user info not found"}))
		return
	}
	userInfo := user.(*drivenadapters.UserInfo)

	serviceID, ok := c.Params.Get("serviceID")

	if !ok {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", map[string]string{"err": "invalid serviceID"}))
		return
	}

	data, _ := io.ReadAll(c.Request.Body)

	err = common.JSONSchemaValid(data, CustomPromptParamsSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var params CustomPromptParams
	err = json.Unmarshal(data, &params)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	token := userInfo.TokenID

	token = strings.TrimPrefix(token, "Bearer ")

	res, err := h.efast.CheckPerm(ctx, params.DocID, "read", token, userInfo.LoginIP)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	if res != 0 {
		c.JSON(http.StatusForbidden, errors.NewIError(errors.NoPermission, "", map[string]interface{}{
			"info": "has no perm to get doc metadata",
			"doc": map[string]string{
				"docid": params.DocID,
			},
		}))
		return
	}

	result, err := h.cognitiveAssistant.CustomPromptWithFile(ctx, serviceID, params.DocID)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]string{"result": result})
}
