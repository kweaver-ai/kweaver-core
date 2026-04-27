package cdaenum

import "github.com/pkg/errors"

// StatusThreeState agent三态状态（未发布、已发布、发布后有修改）
type StatusThreeState string

const (
	// StatusThreeStateUnpublished 未发布
	StatusThreeStateUnpublished StatusThreeState = "unpublished"

	// StatusThreeStatePublished 已发布
	StatusThreeStatePublished StatusThreeState = "published"

	// StatusThreeStatePublishedEdited 发布后有修改
	StatusThreeStatePublishedEdited StatusThreeState = "published_edited"
)

func (t StatusThreeState) EnumCheck() (err error) {
	if t != StatusThreeStateUnpublished && t != StatusThreeStatePublished && t != StatusThreeStatePublishedEdited {
		err = errors.New("[StatusThreeState]: invalid status")
		return
	}

	return
}

func (t StatusThreeState) IsPublished() bool {
	return t == StatusThreeStatePublished || t == StatusThreeStatePublishedEdited
}
