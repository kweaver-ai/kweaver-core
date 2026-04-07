package operator

import (
	"bytes"
	"mime/multipart"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterInternalOperator 内置算子注册更新
func (op *operatorHandle) RegisterInternalOperator(c *gin.Context) {
	req := &interfaces.RegisterInternalOperatorReq{
		OperatorExecuteControl: &interfaces.OperatorExecuteControl{},
		OpenAPIInput:           &interfaces.OpenAPIInput{},
		Functions:              []*interfaces.FunctionInput{},
	}
	ctx := c.Request.Context()
	err := c.ShouldBindHeader(req)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	switch c.ContentType() {
	case "application/json":
		err = utils.GetBindJSONRaw(c, req)
	case "application/x-www-form-urlencoded":
		err = c.ShouldBind(req)
		if err != nil {
			err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
		}
	case "multipart/form-data":
		err = c.ShouldBindWith(req, binding.Form)
		if err != nil {
			err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
			rest.ReplyError(c, err)
			return
		}
		var file *multipart.FileHeader
		file, err = c.FormFile("data")
		if err != nil {
			err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
			rest.ReplyError(c, err)
			return
		}
		var fileContent multipart.File
		// TODO: 检查文件大小
		fileContent, err = file.Open()
		if err != nil {
			err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
			rest.ReplyError(c, err)
			return
		}
		defer func() {
			_ = fileContent.Close()
		}()
		buf := new(bytes.Buffer)
		if _, err = buf.ReadFrom(fileContent); err != nil {
			err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
			rest.ReplyError(c, err)
			return
		}
		req.Data = buf.Bytes()
	default:
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "content type not supported")
	}
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	req.IsPublic = c.GetHeader(string(interfaces.IsPublic)) == "true"
	if req.IsPublic && req.UserID == "" {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "invalid user_id")
		rest.ReplyError(c, err)
		return
	}
	// 检查名字
	err = op.Validator.ValidatorToolBoxName(ctx, req.Name)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = op.Validator.ValidatorIntCompVersion(ctx, req.ConfigVersion)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	resp, err := op.OperatorManager.RegisterInternalOperator(ctx, req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}
