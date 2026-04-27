package i18n

import (
	"context"
	"fmt"

	"golang.org/x/text/language"
)

const XLangKey = "X-Language"

// Language 语言类型
type Language string

func (l *Language) String() string {
	return string(*l)
}

// 语言类型
const (
	SimplifiedChinese  Language = "zh-CN" // 简体中文
	TraditionalChinese Language = "zh-TW" // 繁体中文
	AmericanEnglish    Language = "en-US" // 美国英语
	DefaultLanguage    Language = "zh-CN" // 默认语言
)

var langMatcher = language.NewMatcher([]language.Tag{
	language.SimplifiedChinese,
	language.TraditionalChinese,
	language.AmericanEnglish,
})

var langTagMap = map[interface{}]Language{
	language.SimplifiedChinese:  SimplifiedChinese,
	language.TraditionalChinese: TraditionalChinese,
	language.AmericanEnglish:    AmericanEnglish,
	language.BritishEnglish:     AmericanEnglish,
}

var langTagStrMap = map[interface{}]Language{
	language.SimplifiedChinese.String():  SimplifiedChinese,
	language.TraditionalChinese.String(): TraditionalChinese,
}

// GetLangType 获取符合BCP47标准的语言标签
func (l *Language) GetLangType() *Language {
	inputTag, err := language.Parse(l.String())
	if err != nil {
		return nil
	}

	lt, ok := langTagMap[inputTag]
	if ok {
		return &lt
	}

	langTag, _ := language.MatchStrings(langMatcher, inputTag.String())

	lt, ok = langTagMap[langTag]
	if ok {
		return &lt
	}

	b, s, _ := langTag.Raw()
	key := fmt.Sprintf("%v-%v", b.String(), s.String())

	lt, ok = langTagStrMap[key]
	if !ok {
		return nil
	}
	return &lt
}

// GetLangFromCTX 从ctx上下文中解析language, 如果不存在则默认为zh-CN
func GetLangFromCTX(ctx context.Context) Language {
	var lang Language
	_lang := ctx.Value(XLangKey)
	if _lang != nil {
		lang = _lang.(Language)
	}

	lt := lang.GetLangType()
	if lt == nil {
		return DefaultLanguage
	}

	return *lt
}
