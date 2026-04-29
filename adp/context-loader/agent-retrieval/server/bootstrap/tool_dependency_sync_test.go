package bootstrap

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestToolDependencySyncSyncOnce(t *testing.T) {
	Convey("TestToolDependencySyncSyncOnce", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		syncer := &ToolDependencySync{
			logger:              mockLogger,
			operatorIntegration: mockOperatorIntegration,
			config: config.ToolDependencySyncConfig{
				Enabled: true,
			},
		}
		syncer.loadPackages = func() []embeddedToolDependencyPackage {
			return []embeddedToolDependencyPackage{
				{name: "execution_factory_tools.adp", data: []byte(`{"toolbox":{"configs":[]}}`)},
				{name: "context_loader_toolset.adp", data: []byte(`{"toolbox":{"configs":[{"box_id":"context-loader"}]}}`)},
			}
		}
		syncer.wait = func(context.Context, time.Duration) bool { return true }

		expectedPackageData := []string{
			`{"toolbox":{"configs":[]}}`,
			`{"toolbox":{"configs":[{"box_id":"context-loader"}]}}`,
		}
		callIndex := 0
		mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockOperatorIntegration.EXPECT().SyncToolDependencyPackage(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, syncReq *interfaces.SyncToolDependencyPackageRequest) error {
				So(syncReq.Mode, ShouldEqual, "upsert")
				So(string(syncReq.PackageData), ShouldEqual, expectedPackageData[callIndex])
				callIndex++
				return nil
			},
		).Times(2)

		err := syncer.syncOnce(context.Background())
		So(err, ShouldBeNil)
		So(callIndex, ShouldEqual, 2)
	})
}

func TestToolDependencySyncStartRetry(t *testing.T) {
	Convey("TestToolDependencySyncStartRetry", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		attempt := 0
		waitCount := 0
		syncer := &ToolDependencySync{
			logger:              mockLogger,
			operatorIntegration: mockOperatorIntegration,
			config: config.ToolDependencySyncConfig{
				Enabled:                     true,
				InitialRetryIntervalSeconds: 1,
				MaxRetryIntervalSeconds:     2,
			},
		}
		syncer.loadPackages = func() []embeddedToolDependencyPackage {
			return []embeddedToolDependencyPackage{
				{name: "execution_factory_tools.adp", data: []byte(`{"toolbox":{"configs":[]}}`)},
			}
		}
		syncer.wait = func(context.Context, time.Duration) bool {
			waitCount++
			return true
		}

		mockOperatorIntegration.EXPECT().SyncToolDependencyPackage(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, _ *interfaces.SyncToolDependencyPackageRequest) error {
				attempt++
				if attempt == 1 {
					return errors.New("temporary error")
				}
				return nil
			},
		).Times(2)

		syncer.Start(context.Background())
		So(attempt, ShouldEqual, 2)
		So(waitCount, ShouldEqual, 1)
	})
}

func TestNextRetryDelay(t *testing.T) {
	Convey("TestNextRetryDelay", t, func() {
		syncer := &ToolDependencySync{
			config: config.ToolDependencySyncConfig{
				MaxRetryIntervalSeconds: 10,
			},
		}
		So(syncer.nextRetryDelay(3*time.Second), ShouldEqual, 6*time.Second)
		So(syncer.nextRetryDelay(6*time.Second), ShouldEqual, 10*time.Second)
	})
}
