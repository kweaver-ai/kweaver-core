package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var translator *I18nTranslator
var defaultLang = language.SimplifiedChinese

type I18nTranslator struct {
	lang      string
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
}

// InitI18nTranslator 服务调用时初始化国际化资源
func InitI18nTranslator(resource string) {
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	translator = &I18nTranslator{
		bundle: bundle,
		lang:   defaultLang.String(),
	}
	translator.LoadMessageFiles(resource)
}

// NewI18nTranslator 创建一个i18n资源解析器
func NewI18nTranslator(lang string) *I18nTranslator {
	return &I18nTranslator{
		localizer: i18n.NewLocalizer(translator.bundle, lang),
	}
}

// LoadMessageFiles 加载翻译文件
func (i *I18nTranslator) LoadMessageFiles(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		fullPath := filepath.Join(dir, file.Name())
		if _, err := i.bundle.LoadMessageFile(fullPath); err != nil {
			panic(err)
		}
	}
}

/*
Trans 翻译消息键并提供可选参数

@param msg - 需要翻译的消息键
@param params - 可变参数，其中:
  - params[0]: 用于消息插值的模板数据 (map[string]interface{})
  - params[1]: 用于复数规则的计数 (int)

示例:

	// 简单翻译
	t.Trans("welcome")

	// 带模板数据
	t.Trans("greeting", map[string]interface{}{"Name": "John"})

	// 带复数形式
	t.Trans("unread_messages", nil, 5)

	// 同时带模板和复数
	t.Trans("new_messages", map[string]interface{}{"App": "Mail"}, 3)
*/
func (i *I18nTranslator) Trans(keyType, key string, params ...interface{}) string {
	if i.localizer == nil {
		return key
	}

	config := &i18n.LocalizeConfig{
		MessageID: fmt.Sprintf("%s.%s", keyType, key),
	}

	l := len(params)

	if l > 0 {
		config.TemplateData = params[0]
	}

	if l > 1 {
		config.PluralCount = params[1]
	}

	msg, err := i.localizer.Localize(config)
	if err != nil {
		return key
	}

	return msg
}
