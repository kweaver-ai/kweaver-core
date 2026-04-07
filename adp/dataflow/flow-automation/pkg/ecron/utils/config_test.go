package utils

import (
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewLang(t *testing.T) {
	Convey("newLang", t, func() {
		Convey("lang is null", func() {
			lang := newLang(common.ConfigLoader{Lang: ""})
			assert.Equal(t, lang["IDS_INTERNAL_ERROR"], "内部错误。")
		})

		Convey("lang is zh_CN", func() {
			lang := newLang(common.ConfigLoader{Lang: "zh_CN"})
			assert.Equal(t, lang["IDS_INTERNAL_ERROR"], "内部错误。")
		})

		Convey("lang is zh_TW", func() {
			lang := newLang(common.ConfigLoader{Lang: "zh_TW"})
			assert.Equal(t, lang["IDS_INTERNAL_ERROR"], "內部錯誤。")
		})

		Convey("lang is en_US", func() {
			lang := newLang(common.ConfigLoader{Lang: "en_US"})
			assert.Equal(t, lang["IDS_INTERNAL_ERROR"], "Internal error.")
		})
	})
}
