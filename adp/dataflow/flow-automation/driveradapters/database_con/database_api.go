package database_con

import (
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/dependency"
	"github.com/gin-gonic/gin"
)

var (
	once sync.Once
	rh   RESTHandler
)


type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
}

type restHandler struct {
	databaseService dependency.DatabaseTableService
}

func NewRestHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			databaseService: dependency.NewDatabaseTableService(),
		}
	})

	return rh
}

func (rh *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.GET("/database/tables", middleware.TokenAuth(), rh.listTables)
	engine.GET("/database/table/columns", middleware.TokenAuth(), rh.listTableColumns)

}

func (rh *restHandler) listTables(c *gin.Context) {
	tables, err := rh.databaseService.ListTables(c.Request.Context(), c.Query("data_source_id"), c.GetHeader("Authorization"), c.ClientIP())
	if err != nil {
		errors.ReplyError(c, err)
		return
	}
	c.JSON(http.StatusOK, tables)
}

// listTableColumns 查询指定表的字段信息
func (rh *restHandler) listTableColumns(c *gin.Context) {
	dataSourceID := c.Query("data_source_id")
	tableName := c.Query("table_name")
	schema := c.Query("schema") // 可选参数

	if dataSourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_source_id is required"})
		return
	}

	if tableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table_name is required"})
		return
	}

	columns, err := rh.databaseService.ListTableColumns(c.Request.Context(), dataSourceID, tableName, schema, c.GetHeader("Authorization"), c.ClientIP())
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.JSON(http.StatusOK, columns)
}
