package drivenadapters

import (
	"reflect"
	"sync"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/go-playground/assert/v2"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	. "github.com/smartystreets/goconvey/convey"
)

type Dependency struct {
	hydra HydraPublic
}

func NewDependency() Dependency {
	return Dependency{
		hydra: NewHydraPublic(),
	}
}

func NewMockAppTokenMgnt(deps Dependency) AppTokenMgnt {
	InitLogger()
	return &tokenMgnt{
		token:        "",
		expireTime:   0,
		updatedAt:    0,
		mutex:        sync.RWMutex{},
		hydra:        deps.hydra,
		clientID:     "clientID",
		clientSecret: "clientSecret",
		logger:       commonLog.NewLogger(),
	}
}

func TestGetAppToken(t *testing.T) {
	deps := NewDependency()
	tokenMgnt := NewMockAppTokenMgnt(deps)

	Convey("TestGetAppToken", t, func() {
		Convey("Get App Token Success", func() {
			patch := ApplyMethod(reflect.TypeOf(deps.hydra), "RequestTokenWithCredential", func(*hydraPublic, string, string, []string) (tokenInfo TokenInfo, code int, err error) {
				return TokenInfo{
					Token:        "token",
					ExpiresIn:    3600,
					RefreshToken: "refreshToken",
				}, 200, nil
			})
			defer patch.Reset()
			token, expired, err := tokenMgnt.GetAppToken("")
			assert.Equal(t, err, nil)
			assert.Equal(t, "token", token)
			assert.Equal(t, int64(3600), expired)
		})
	})
}
