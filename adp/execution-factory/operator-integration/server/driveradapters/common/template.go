package common

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/localize"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
)

var (
	// CodeTemplate 代码模板
	CodeTemplate = "template.%s" // 代码模板
)

// TemplateHandler 模板处理接口
type TemplateHandler interface {
	GetTemplate(c *gin.Context)
}

type templateHandler struct {
	I18nTranslator map[string]*localize.I18nTranslator
	Validator      interfaces.Validator
}

var (
	tOnce sync.Once
	th    TemplateHandler
)

func NewTemplateHandler() TemplateHandler {
	tOnce.Do(func() {
		th = &templateHandler{
			Validator: validator.NewValidator(),
			I18nTranslator: map[string]*localize.I18nTranslator{
				common.SimplifiedChinese: localize.NewI18nTranslator(common.SimplifiedChinese),
				common.AmericanEnglish:   localize.NewI18nTranslator(common.AmericanEnglish),
			},
		}
	})
	return th
}

// TemplateParams 模板参数
type TemplateParams struct {
	TemplateType string `uri:"template_type" default:"python" validate:"required,oneof=python"`
}

// TemplateResponse 模板响应参数
type TemplateResponse struct {
	TemplateType string `json:"template_type"`
	CodeTemplate string `json:"code_template"`
}

func (t *templateHandler) GetTemplate(c *gin.Context) {
	req := &TemplateParams{}
	if err := c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, err)
		return
	}

	if err := defaults.Set(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	if err := t.Validator.ValidatorStruct(c.Request.Context(), req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	language := common.GetLanguageByCtx(c.Request.Context())
	tr, ok := t.I18nTranslator[language]
	if !ok {
		tr = t.I18nTranslator[common.DefaultLanguage]
	}
	codeTemplate := tr.Trans(fmt.Sprintf(CodeTemplate, req.TemplateType))
	rest.ReplyOK(c, http.StatusOK, &TemplateResponse{
		TemplateType: req.TemplateType,
		CodeTemplate: codeTemplate,
	})
}
