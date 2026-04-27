// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package common 定义通用类型
// @file language.go
// @description: 定义语言类型
package common

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

// Language 语言类型
type Language = string

// 语言类型
const (
	SimplifiedChinese Language = "zh-CN" // 简体中文
	// TraditionalChinese          // 繁体中文
	AmericanEnglish Language = "en-US" // 美国英语
)

type LanguageKey string

const XLangKey LanguageKey = "X-Language"

var (
	// Langs 支持的语言
	Languages = map[Language]Language{
		SimplifiedChinese: SimplifiedChinese,
		AmericanEnglish:   AmericanEnglish,
	}
	DefaultLanguage = SimplifiedChinese
)

var (
	langMatcher = language.NewMatcher([]language.Tag{
		language.SimplifiedChinese,
		language.AmericanEnglish,
	})

	langTagMap = map[language.Tag]Language{
		language.SimplifiedChinese: SimplifiedChinese,
		language.AmericanEnglish:   AmericanEnglish,
	}
)

// SetLang 设置语言
func SetLang(langStr string) {
	lang := GetBCP47(langStr)
	DefaultLanguage = lang
}

// GetXLang 解析获取 Header x-language
func GetXLang(c *gin.Context) Language {
	lang := GetBCP47(c.GetHeader(string(XLangKey)))
	langTag, _ := language.MatchStrings(langMatcher, lang)
	return langTagMap[langTag]
}

// GetBCP47 将约定的语言标签转换为符合BCP47标准的语言标签
// 默认值为 zh-CN, 中国大陆简体中文
// https://www.rfc-editor.org/info/bcp47
func GetBCP47(langStr string) Language {
	switch strings.ToLower(langStr) {
	case "zh_cn", "zh-cn":
		return SimplifiedChinese
	case "en_us", "en-us":
		return AmericanEnglish
	default:
		return DefaultLanguage
	}
}

func GetLanguageInfo(c *gin.Context) Language {
	return GetBCP47(c.GetHeader(string(XLangKey)))
}

func GetLanguageByCtx(ctx context.Context) Language {
	lang := DefaultLanguage
	langV := ctx.Value(XLangKey)
	if langV != nil {
		lang = langV.(Language)
	}
	if _, ok := Languages[lang]; !ok {
		lang = DefaultLanguage
	}
	return lang
}

func SetLanguageByCtx(ctx context.Context, lang Language) context.Context {
	return context.WithValue(ctx, XLangKey, lang)
}
