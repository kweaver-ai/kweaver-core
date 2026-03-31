package operator

import (
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// QueryOperatorHistory 查询操作符历史
func (op *operatorHandle) QueryOperatorHistoryDetail(c *gin.Context) {
	req := &interfaces.OperatorHistoryDetailReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		err = errors.DefaultHTTPError(c, http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	if err := c.ShouldBindUri(req); err != nil {
		err = errors.DefaultHTTPError(c, http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	if err := c.ShouldBindQuery(req); err != nil {
		err = errors.DefaultHTTPError(c, http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	if err := validator.New().Struct(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	result, err := op.OperatorManager.QueryOperatorHistoryDetail(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, result)
}

// QueryOperatorHistoryList 查询算子历史版本列表
func (op *operatorHandle) QueryOperatorHistoryList(c *gin.Context) {
	req := &interfaces.OperatorHistoryListReq{}
	if err := c.ShouldBindHeader(req); err != nil {
		err = errors.DefaultHTTPError(c, http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	if err := c.ShouldBindUri(req); err != nil {
		err = errors.DefaultHTTPError(c, http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	if err := validator.New().Struct(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	result, err := op.OperatorManager.QueryOperatorHistoryList(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, result)
}
