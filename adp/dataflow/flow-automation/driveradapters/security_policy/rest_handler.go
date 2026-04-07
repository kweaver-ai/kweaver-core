package security_policy

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	normalizeutil "github.com/kweaver-ai/adp/autoflow/flow-automation/utils/normalize"
)

type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
	RegisterPrivateAPI(engine *gin.RouterGroup)
}

var (
	once sync.Once
	rh   RESTHandler
)

var (
	createFlowParamsSchema = "security-policy/create-flow-params.json"
	flowStepsSchema        = "security-policy/flow-steps.json"
	procParamsSchema       = "security-policy/proc-params.json"
	updateProcStatusSchema = "security-policy/update-proc-status.json"
)

type SecurityPolicyHandler struct {
	config     *common.Config
	hydra      drivenadapters.HydraPublic
	hydraAdmin drivenadapters.HydraAdmin
	userMgnt   drivenadapters.UserManagement
	mgnt       mgnt.MgntHandler
}

func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &SecurityPolicyHandler{
			config:     common.NewConfig(),
			hydra:      drivenadapters.NewHydraPublic(),
			hydraAdmin: drivenadapters.NewHydraAdmin(),
			userMgnt:   drivenadapters.NewUserManagement(),
			mgnt:       mgnt.NewMgnt(),
		}
	})

	return rh
}

func (h *SecurityPolicyHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.POST("/flows", middleware.TokenAuth(), middleware.CheckAdminOrKnowledgeManager(), h.create)
	engine.PUT("/flows/:id/steps", middleware.TokenAuth(), middleware.CheckAdminOrKnowledgeManager(), h.update)
	engine.DELETE("/flows/:id", middleware.TokenAuth(), middleware.CheckAdminOrKnowledgeManager(), h.delete)
	engine.GET("/flows/:id", middleware.TokenAuth(), middleware.CheckAdminOrKnowledgeManager(), h.getFlowById)
	engine.PUT(("/procs/:pid/status"), middleware.TokenAuth(), h.stopProcByClient)
}

func (h *SecurityPolicyHandler) RegisterPrivateAPI(engine *gin.RouterGroup) {
	engine.DELETE("/flows/:id", h.deleteInternal)
	engine.GET("/flows/:id/form", h.getFlowForm)
	engine.POST("/procs", h.createProc)
	engine.PUT(("/procs/:pid/status"), h.stopProc)
}

func (h *SecurityPolicyHandler) create(c *gin.Context) {
	user, isexist := c.Get("user")
	if !isexist {
		c.JSON(http.StatusUnauthorized, errors.NewIError(errors.UnAuthorization, "", map[string]string{"auth": "user info not found"}))
		return
	}
	userInfo := user.(*drivenadapters.UserInfo)
	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, createFlowParamsSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var createFlowParams mgnt.CreateFlowParams
	err = json.Unmarshal(data, &createFlowParams)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	var flowID string
	flowID, err = h.mgnt.CreateSecurityPolicyFlow(c.Request.Context(), &createFlowParams, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Writer.Header().Set("Location", "/api/automation/v1/security-policy/flows/"+flowID)
	c.JSON(http.StatusCreated, gin.H{"id": flowID})
}

func (h *SecurityPolicyHandler) update(c *gin.Context) {
	flowID := c.Param("id")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, flowStepsSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var steps []entity.Step
	err = json.Unmarshal(data, &steps)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	err = h.mgnt.UpdateSecurityPolicyFlow(c.Request.Context(), flowID, steps, userInfo)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SecurityPolicyHandler) deleteInternal(c *gin.Context) {
	flowID := c.Param("id")

	err := h.mgnt.DeleteSecurityPolicyFlow(c.Request.Context(), flowID, nil)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SecurityPolicyHandler) delete(c *gin.Context) {
	flowID := c.Param("id")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	err := h.mgnt.DeleteSecurityPolicyFlow(c.Request.Context(), flowID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SecurityPolicyHandler) getFlowById(c *gin.Context) {
	flowID := c.Param("id")
	flow, err := h.mgnt.GetSecurityPolicyFlowByID(c.Request.Context(), flowID)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, flow)
}

func (h *SecurityPolicyHandler) getFlowForm(c *gin.Context) {
	flowID := c.Param("id")

	flow, err := h.mgnt.GetSecurityPolicyFlowByID(c.Request.Context(), flowID)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	if len(flow.Steps) > 0 && flow.Steps[0].Parameters != nil && flow.Steps[0].Parameters["fields"] != nil {
		fields, ok := normalizeutil.AsSlice(flow.Steps[0].Parameters["fields"])

		if ok {

			formFields := make([]mgnt.FormField, len(fields))

			for i, v := range fields {
				formFields[i].Key = v.(map[string]interface{})["key"].(string)
				formFields[i].Name = v.(map[string]interface{})["name"].(string)
				formFields[i].Type = v.(map[string]interface{})["type"].(string)

				if v.(map[string]interface{})["required"] != nil {
					formFields[i].Required = v.(map[string]interface{})["required"].(bool)
				}
			}
			c.JSON(http.StatusOK, map[string]interface{}{"fields": formFields})
			return
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{"fields": []mgnt.FormField{}})
}

func (h *SecurityPolicyHandler) createProc(c *gin.Context) {

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, procParamsSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var procParams mgnt.ProcParams
	err = json.Unmarshal(data, &procParams)

	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	pid, err := h.mgnt.StartSecurityPolicyFlowProc(c.Request.Context(), procParams)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusCreated, map[string]interface{}{"id": pid, "status": "init"})
}

func (h *SecurityPolicyHandler) stopProc(c *gin.Context) {

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, updateProcStatusSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	pid := c.Param("pid")
	err = h.mgnt.StopSecurityPolicyFlowProc(c.Request.Context(), pid, nil)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// 客户端调用停止流程
func (h *SecurityPolicyHandler) stopProcByClient(c *gin.Context) {
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, updateProcStatusSchema)

	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	pid := c.Param("pid")
	err = h.mgnt.StopSecurityPolicyFlowProc(c.Request.Context(), pid, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
