package crest

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

func GetRestHttpErr(err error) (httpError *rest.HTTPError, b bool) {
	if err == nil {
		return
	}

	if errors.As(err, &httpError) {
		b = true
		return
	}

	return
}
