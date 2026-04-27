package cdaenum

import "github.com/pkg/errors"

type OutputDefaultFormat string

const (
	OutputDefaultFormatJson     OutputDefaultFormat = "json"
	OutputDefaultFormatMarkdown OutputDefaultFormat = "markdown"
)

func (o OutputDefaultFormat) EnumCheck() (err error) {
	if o != OutputDefaultFormatJson && o != OutputDefaultFormatMarkdown {
		err = errors.New("[OutputDefaultFormat]: invalid output default format")
		return
	}

	return
}
