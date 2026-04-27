package operator

import (
	"net/http"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

func (op *operatorHandle) OperatorEdit(c *gin.Context) {
	req := &interfaces.OperatorEditReq{
		OperatorInfoEdit:       &interfaces.OperatorInfoEdit{},
		OperatorExecuteControl: &interfaces.OperatorExecuteControl{},
		OpenAPIInput:           &interfaces.OpenAPIInput{},
		FunctionInputEdit: &interfaces.FunctionInputEdit{
			Dependencies: []*interfaces.DependencyInfo{},
		},
	}
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	switch c.ContentType() {
	case "application/json":
		err = utils.GetBindJSONRaw(c, req)
	case "application/x-www-form-urlencoded":
		err = utils.GetBindFormRaw(c, req)
	case "multipart/form-data":
		req.Data, err = utils.GetBindMultipartFormRaw(c, req, "data", 0)
	default:
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "unsupported content type")
	}
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	result, err := op.OperatorManager.EditOperator(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	rest.ReplyOK(c, http.StatusOK, result)
}

// OperatorDelete 删除算子
func (op *operatorHandle) OperatorDelete(c *gin.Context) {
	// 定义一个切片来接收请求体
	var req interfaces.OperatorDeleteReq
	var err error

	if err = utils.GetBindJSONRaw(c, &req); err != nil {
		rest.ReplyError(c, err)
		return
	}

	var userID string
	userID, err = op.getUserID(c, "")
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	ctx := c.Request.Context()
	err = op.OperatorManager.DeleteOperator(ctx, req, userID)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, nil)
}
