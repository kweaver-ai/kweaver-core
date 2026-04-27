package sandbox

import (
	"context"
	"testing"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestGetDependenciesCachesFromSessionQuery(t *testing.T) {
	Convey("GetDependencies caches dependencies from QuerySession", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockSandBoxControlPlane(ctrl)
		pool := &sessionPoolImpl{
			client: mockClient,
			sessions: map[string]*sessionItem{
				"sess_aoi_0": {
					ID:           "sess_aoi_0",
					RunningTasks: 0,
					LastUsedAt:   time.Now(),
				},
			},
			maxConcurrentTasks: 10,
			logger:             logger.DefaultLogger(),
			stopCh:             make(chan struct{}),
		}

		firstDetail := &interfaces.SessionDetail{
			ID:     "sess_aoi_0",
			Status: interfaces.SessionStatusRunning,
			InstalledDependencies: []*interfaces.DependencyInfo{
				{Name: "requests", Version: "2.28.1"},
			},
		}
		secondDetail := &interfaces.SessionDetail{
			ID:     "sess_aoi_0",
			Status: interfaces.SessionStatusRunning,
		}

		gomock.InOrder(
			mockClient.EXPECT().
				QuerySession(gomock.Any(), "sess_aoi_0").
				Return(true, firstDetail, nil),
			mockClient.EXPECT().
				QuerySession(gomock.Any(), "sess_aoi_0").
				Return(true, secondDetail, nil),
		)

		resp, err := pool.GetDependencies(context.Background())
		So(err, ShouldBeNil)
		So(resp.SessionID, ShouldEqual, "sess_aoi_0")
		So(resp.Dependencies, ShouldResemble, []*interfaces.DependencyInfo{
			{Name: "requests", Version: "2.28.1"},
		})

		item, ok := pool.getSessionItem("sess_aoi_0")
		So(ok, ShouldBeTrue)
		So(item.Dependencies, ShouldResemble, firstDetail.InstalledDependencies)
	})
}
