package entity

import (
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/render"
)

type ParamsRender struct {
	paramRender *render.TplRender
}

// NewParamsRender
func NewParamsRender() *ParamsRender {
	return &ParamsRender{
		paramRender: render.NewTplRender(),
	}
}

func (e *ParamsRender) RenderParam(dagIns *DagInstance, key interface{}) (interface{}, error) {
	data := map[string]interface{}{}

	dagInstance := dagIns
	if dagInstance != nil {
		data["vars"] = dagInstance.Vars
		if dagInstance.ShareData != nil {
			data["shareData"] = dagInstance.ShareData.GetAll()
		}
	}

	ks, ok := key.(string)
	if !ok {
		return key, nil
	}

	if strings.Contains(ks, "{{__") && strings.Contains(ks, "}}") {
		// ks = strings.Replace(ks, "{{", "{{.shareData.", 1)
		ks = strings.ReplaceAll(ks, "{{", "{{.shareData.")
		result, err := e.paramRender.Render(ks, data)
		if err != nil {
			return ks, err
		}
		return result, nil
	}
	return ks, nil
}
