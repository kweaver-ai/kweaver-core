package cdaenum

import "github.com/pkg/errors"

// Status agent状态
type Status string

const (
	// StatusUnpublished 未发布
	StatusUnpublished Status = "unpublished"

	// StatusPublished 已发布
	StatusPublished Status = "published"
)

func (t Status) EnumCheck() (err error) {
	if t != StatusUnpublished && t != StatusPublished {
		err = errors.New("[Status]: invalid status")
		return
	}

	return
}

func (t Status) IsPublished() bool {
	return t == StatusPublished
}
