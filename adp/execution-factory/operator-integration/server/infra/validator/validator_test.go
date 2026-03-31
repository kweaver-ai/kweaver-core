package validator

import (
	"context"
	"strings"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	validatorv10 "github.com/go-playground/validator/v10"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateName(t *testing.T) {
	v := &validator{
		NameLimit: 50,
		Validator: validatorv10.New(),
	}
	ctx := context.Background()
	ctx = common.SetLanguageToCtx(ctx, common.SimplifiedChinese)
	var err error
	Convey("测试名称为空", t, func() {
		err = v.ValidateOperatorName(ctx, "")
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolName(ctx, "")
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolBoxName(ctx, "")
		So(err, ShouldNotBeNil)
	})
	Convey("测试名称长度超过限制", t, func() {
		err = v.ValidateOperatorName(ctx, strings.Repeat("a", 51))
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolName(ctx, strings.Repeat("a", 51))
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolBoxName(ctx, strings.Repeat("a", 51))
		So(err, ShouldNotBeNil)
	})
	Convey("ValidateOperatorName:测试名称合法性", t, func() {
		err = v.ValidateOperatorName(ctx, "aaa")
		So(err, ShouldBeNil)
		err = v.ValidateOperatorName(ctx, "中文")
		So(err, ShouldBeNil)
		err = v.ValidateOperatorName(ctx, "中文aa")
		So(err, ShouldBeNil)
		err = v.ValidateOperatorName(ctx, "中文 aa")
		So(err, ShouldNotBeNil)
		err = v.ValidateOperatorName(ctx, "中文_aa")
		So(err, ShouldBeNil)
		err = v.ValidateOperatorName(ctx, "中文$aa")
		So(err, ShouldNotBeNil)
		err = v.ValidateOperatorName(ctx, "中文@aa")
		So(err, ShouldNotBeNil)
		err = v.ValidateOperatorName(ctx, "中文^#aa")
		So(err, ShouldNotBeNil)
	})
	Convey("ValidatorToolBoxName:测试名称合法性", t, func() {
		err = v.ValidatorToolBoxName(ctx, "aaa")
		So(err, ShouldBeNil)
		err = v.ValidatorToolBoxName(ctx, "中文")
		So(err, ShouldBeNil)
		err = v.ValidatorToolBoxName(ctx, "中文aa")
		So(err, ShouldBeNil)
		err = v.ValidatorToolBoxName(ctx, "中文 aa")
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolBoxName(ctx, "中文_aa")
		So(err, ShouldBeNil)
		err = v.ValidatorToolBoxName(ctx, "中文$aa")
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolBoxName(ctx, "中文@aa")
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolBoxName(ctx, "中文^#aa")
		So(err, ShouldNotBeNil)
	})
	Convey("ValidatorToolName:测试名称合法性", t, func() {
		err = v.ValidatorToolName(ctx, "aaa")
		So(err, ShouldBeNil)
		err = v.ValidatorToolName(ctx, "中文")
		So(err, ShouldBeNil)
		err = v.ValidatorToolName(ctx, "中文aa")
		So(err, ShouldBeNil)
		err = v.ValidatorToolName(ctx, "中文 aa")
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolName(ctx, "中文_aa")
		So(err, ShouldBeNil)
		err = v.ValidatorToolName(ctx, "中文$aa")
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolName(ctx, "中文@aa")
		So(err, ShouldNotBeNil)
		err = v.ValidatorToolName(ctx, "中文^#aa")
		So(err, ShouldNotBeNil)
	})
}

func TestValidatorInternalToolBoxVersion(t *testing.T) {
	v := &validator{
		NameLimit: 50,
		Validator: validatorv10.New(),
	}
	ctx := context.Background()
	ctx = common.SetLanguageToCtx(ctx, common.SimplifiedChinese)
	var err error
	Convey("TestValidatorInternalToolBoxVersion:检查内置工具版本", t, func() {
		err = v.ValidatorIntCompVersion(ctx, "1.0.0")
		So(err, ShouldBeNil)
		err = v.ValidatorIntCompVersion(ctx, "1.0.aaa")
		So(err, ShouldNotBeNil)
		err = v.ValidatorIntCompVersion(ctx, "10.0.0")
		So(err, ShouldBeNil)
	})
}
