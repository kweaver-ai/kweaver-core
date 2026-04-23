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
		syncer.readFile = func(path string) ([]byte, error) {
			switch path {
			case toolDependencyPackagePath:
				return []byte(`{"toolbox":{"configs":[]}}`), nil
			case toolDependencyVersionPath:
				return []byte("0.6.0\n"), nil
			default:
				return nil, errors.New("unexpected path")
			}
		}
		syncer.wait = func(context.Context, time.Duration) bool { return true }

		mockOperatorIntegration.EXPECT().SyncToolDependencyPackage(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, syncReq *interfaces.SyncToolDependencyPackageRequest) (*interfaces.SyncToolDependencyPackageResponse, error) {
				So(syncReq.Mode, ShouldEqual, "upsert")
				So(syncReq.PackageVersion, ShouldEqual, "0.6.0")
				So(string(syncReq.PackageData), ShouldEqual, `{"toolbox":{"configs":[]}}`)
				return &interfaces.SyncToolDependencyPackageResponse{Status: "imported"}, nil
			},
		)

		resp, err := syncer.syncOnce(context.Background())
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(resp.Status, ShouldEqual, "imported")
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
		syncer.readFile = func(path string) ([]byte, error) {
			if path == toolDependencyPackagePath {
				return []byte(`{"toolbox":{"configs":[]}}`), nil
			}
			if path == toolDependencyVersionPath {
				return []byte("0.6.0"), nil
			}
			return nil, errors.New("unexpected path")
		}
		syncer.wait = func(context.Context, time.Duration) bool {
			waitCount++
			return true
		}

		mockOperatorIntegration.EXPECT().SyncToolDependencyPackage(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, _ *interfaces.SyncToolDependencyPackageRequest) (*interfaces.SyncToolDependencyPackageResponse, error) {
				attempt++
				if attempt == 1 {
					return nil, errors.New("temporary error")
				}
				return &interfaces.SyncToolDependencyPackageResponse{Status: "updated", Message: "ok"}, nil
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
