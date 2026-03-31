package render

import (
	"encoding/json"
	"fmt"
	"strings"
)

var (
	// CacheSize tpl cache size
	CacheSize = 1000
)

type TplRender struct {
	tplProvider *TplProvider
}

func NewTplRender() *TplRender {
	return &TplRender{
		tplProvider: NewCachedTplProvider(CacheSize),
	}
}

// Render 模板渲染方法
func (t *TplRender) Render(tplText string, data interface{}) (interface{}, error) {
	tpl, err := t.tplProvider.GetTpl(tplText)
	if err != nil {
		return "", fmt.Errorf("get tpl failed: %w", err)
	}

	var buf strings.Builder
	err = tpl.Execute(&buf, data)
	if err != nil {
		return "", nil
	}

	var res interface{}
	err = json.Unmarshal([]byte(buf.String()), &res)
	if err != nil {
		return buf.String(), nil
	}

	return res, nil

}
