package bootstrap

import (
	"context"
	"encoding/json"
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

func TestEmbeddedContextLoaderToolsetContract(t *testing.T) {
	Convey("TestEmbeddedContextLoaderToolsetContract", t, func() {
		var data map[string]interface{}
		err := json.Unmarshal(contextLoaderToolsetData, &data)
		So(err, ShouldBeNil)

		toolbox, ok := data["toolbox"].(map[string]interface{})
		So(ok, ShouldBeTrue)
		configs, ok := toolbox["configs"].([]interface{})
		So(ok, ShouldBeTrue)
		So(len(configs), ShouldEqual, 1)
		toolboxConfig, ok := configs[0].(map[string]interface{})
		So(ok, ShouldBeTrue)
		So(toolboxConfig["box_name"], ShouldEqual, "contextloader工具集")
		So(toolboxConfig["box_desc"], ShouldEqual, "ContextLoader 标准内置工具集；契约版本: 0.8.0")

		tools, ok := toolboxConfig["tools"].([]interface{})
		So(ok, ShouldBeTrue)
		var searchSchemaTool map[string]interface{}
		for _, tool := range tools {
			toolMap, ok := tool.(map[string]interface{})
			So(ok, ShouldBeTrue)
			if toolMap["name"] == "search_schema" {
				searchSchemaTool = toolMap
				break
			}
		}
		So(searchSchemaTool, ShouldNotBeNil)

		metadata, ok := searchSchemaTool["metadata"].(map[string]interface{})
		So(ok, ShouldBeTrue)
		apiSpec, ok := metadata["api_spec"].(map[string]interface{})
		So(ok, ShouldBeTrue)
		components, ok := apiSpec["components"].(map[string]interface{})
		So(ok, ShouldBeTrue)
		schemas, ok := components["schemas"].(map[string]interface{})
		So(ok, ShouldBeTrue)
		searchScope, ok := schemas["SearchScope"].(map[string]interface{})
		So(ok, ShouldBeTrue)
		properties, ok := searchScope["properties"].(map[string]interface{})
		So(ok, ShouldBeTrue)
		conceptGroups, ok := properties["concept_groups"].(map[string]interface{})
		So(ok, ShouldBeTrue)
		So(conceptGroups["type"], ShouldEqual, "array")
	})
}
