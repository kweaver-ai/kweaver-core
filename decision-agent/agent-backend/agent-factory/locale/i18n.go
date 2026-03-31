package locale

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type I18nKey string

const (
	SystemCreatedBy I18nKey = "SystemCreatedBy" // 系统创建
	UnknownUser     I18nKey = "UnknownUser"     // 未知用户
	Copy            I18nKey = "Copy"            // 未知用户
)

type I18nMap map[rest.Language]string

var AllI18nMap = map[I18nKey]I18nMap{
	SystemCreatedBy: {
		rest.SimplifiedChinese: "系统",
		// langcmp.ZhTW: "系統",
		rest.AmericanEnglish: "System",
	},
	UnknownUser: {
		rest.SimplifiedChinese: "未知用户",
		// langcmp.ZhTW: "未知用戶",
		rest.AmericanEnglish: "Unknown User",
	},
	Copy: {
		rest.SimplifiedChinese: "副本",
		// langcmp.ZhTW: "副本",
		rest.AmericanEnglish: "Duplicate",
	},
}

func GetI18n(key I18nKey, lang rest.Language) string {
	i18nMap, ok := AllI18nMap[key]
	if !ok {
		return ""
	}

	i18n, ok := i18nMap[lang]
	if !ok {
		return ""
	}

	return i18n
}

func GetI18nByCtx(ctx context.Context, key I18nKey) string {
	lang := chelper.GetVisitLanguageCtx(ctx)
	return GetI18n(key, lang)
}
