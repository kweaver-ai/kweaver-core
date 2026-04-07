package versions

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/versions"
	"github.com/gin-gonic/gin"
)

// RESTHandler 公共RESTful api Handler接口
type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once                  sync.Once
	rh                    RESTHandler
	revertToVersionSchema = "base/reverttoversion.json"
)

// DagVersionHandler dag版本
type DagVersionHandler struct {
	dagVersions versions.DagVersionService
}

// NewRESTHandler 创建RESTHandler
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &DagVersionHandler{
			dagVersions: versions.NewDagVersionService(),
		}
	})

	return rh
}

// RegisterAPI 注册API
func (h *DagVersionHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.GET("/dags/:dagId/versions", middleware.TokenAuth(), h.list)
	engine.GET("/dags/:dagId/versions/next", middleware.TokenAuth(), h.getNextVersion)
	engine.POST("/dags/:dagId/versions/:versionID/rollback", middleware.TokenAuth(), h.revert)
}

func (h *DagVersionHandler) list(c *gin.Context) {
	dagID := c.Param("dagId")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	versions, err := h.dagVersions.ListDagVersions(c.Request.Context(), dagID, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, versions)

}

func (h *DagVersionHandler) getNextVersion(c *gin.Context) {
	dagID := c.Param("dagId")

	version, err := h.dagVersions.GetNextVersion(c.Request.Context(), dagID)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"version": version,
	})
}

func (g *DagVersionHandler) revert(c *gin.Context) {
	dagID := c.Param("dagId")
	versionID := c.Param("versionID")
	user, _ := c.Get("user")
	userInfo := user.(*drivenadapters.UserInfo)

	data, _ := io.ReadAll(c.Request.Body)

	err := common.JSONSchemaValid(data, revertToVersionSchema)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	var params versions.RevertDagReq
	err = json.Unmarshal(data, &params)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewIError(errors.InvalidParameter, "", []interface{}{err.Error()}))
		return
	}

	params.DagID = dagID
	params.VersionID = versionID

	versionID, err = g.dagVersions.RevertToVersion(c.Request.Context(), params, userInfo)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version_id": versionID,
	})
}
