package agentconfigenum

import "errors"

type ConfigTplVersionT string

const (
	ConfigTplVersionV1 ConfigTplVersionT = "v1"
)

func (t ConfigTplVersionT) EnumCheck() (err error) {
	if t != ConfigTplVersionV1 {
		err = errors.New("[ConfigTplVersionT]: invalid config version")
		return
	}

	return
}
