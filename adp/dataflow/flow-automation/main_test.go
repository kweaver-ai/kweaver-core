package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	liberrors "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	i18n "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/i18n"
)

func TestMain(t *testing.T) {

	i18n.InitI18nTranslator(common.MultiResourcePath)

	err := liberrors.NewPublicRestError(context.Background(), liberrors.PErrorInternalServerError,
		liberrors.PErrorInternalServerError,
		nil)
	fmt.Println(err)
}
