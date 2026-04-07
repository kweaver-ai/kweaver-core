// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package localize 语言资源
package localize

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	// trMap   = map[language.Tag]*I18nTranslator{}
	langMap = map[string]language.Tag{
		"zh_CN": language.SimplifiedChinese,
		"zh_TW": language.TraditionalChinese,
		"en_US": language.AmericanEnglish,
	}
	matcher = language.NewMatcher([]language.Tag{
		language.SimplifiedChinese,
		language.TraditionalChinese,
		language.AmericanEnglish,
	})

	defaultLang = "zh_CN"
	//go:embed locales/*.json
	locales embed.FS
)

// I18nTranslator 翻译器
type I18nTranslator struct {
	current language.Tag
	loc     *i18n.Localizer
}

// NewI18nTranslator 新建翻译器
func NewI18nTranslator(lang string) *I18nTranslator {
	if lang == "" {
		lang = defaultLang
	}
	lt, ok := langMap[lang]
	if !ok {
		lt = language.SimplifiedChinese
	}
	tr := &I18nTranslator{
		current: lt,
	}
	bundle := i18n.NewBundle(tr.current)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	_, err := bundle.LoadMessageFileFS(locales, fmt.Sprintf("locales/%s.json", lt.String()))
	if err != nil {
		panic(err)
	}
	tr.loc = i18n.NewLocalizer(bundle, tr.current.String())
	return tr
}

// Trans 翻译
func (tr *I18nTranslator) Trans(msg string, params ...interface{}) string {
	l := len(params)
	localizeConf := &i18n.LocalizeConfig{
		MessageID: msg,
	}
	if l > 0 {
		localizeConf.TemplateData = params[0]
	}
	if l > 1 {
		localizeConf.PluralCount = params[1]
	}
	str, err := tr.loc.Localize(localizeConf)
	if err != nil {
		str = msg
		if l > 0 {
			str = strings.TrimRight(fmt.Sprintln(msg, params[0]), "\n")
		}
	}
	return str
}

func getLang(lang string) (lt language.Tag, l string, err error) {
	tag, _ := language.MatchStrings(matcher, lang)
	b, _ := tag.Base()
	r, _ := tag.Region()
	l = fmt.Sprintf("%s_%s", b, r)
	lt, ok := langMap[l]
	if !ok {
		err = fmt.Errorf("not support lang %s", lang)
	}
	return
}

func SetDefaultLang(lang string) (err error) {
	_, l, err := getLang(lang)
	if err != nil {
		return
	}
	defaultLang = l
	return
}

// GetI18nTranslator 获取翻译器
// func GetI18nTranslator(lang string) *I18nTranslator {
// 	lt, l, err := getLang(lang)
// 	if err != nil {
// 		lt = langMap[defaultLang]
// 	}
// 	tr, ok := trMap[lt]
// 	if !ok {
// 		tr = NewI18nTranslator(l)
// 		trMap[lt] = tr
// 	}
// 	return tr
// }
