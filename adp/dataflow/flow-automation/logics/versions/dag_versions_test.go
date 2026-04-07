package versions

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	i18n "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/i18n"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	traceCommon "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/common"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_logics/mock_perm"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
)

func initARLog() {
	if commonLog.NewLogger() == nil {
		logout := "1"
		logDir := "/var/log/contentAutoMation/ut"
		logName := "contentAutoMation.log"
		commonLog.InitLogger(logout, logDir, logName)
	}
	traceLog.InitARLog(&traceCommon.TelemetryConf{LogLevel: "all"})
}

func initErrorInfo() {
	i18n.InitI18nTranslator("../../" + common.MultiResourcePath)
	ierr.InitServiceName(common.ErrCodeServiceName)
}

type MockDependency struct {
	mongo     *mod.MockStore
	usermgnt  *mock_drivenadapters.MockUserManagement
	permCheck *mock_perm.MockPermCheckerService
}

func NewDependency(t *testing.T) MockDependency {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return MockDependency{
		mongo:     mod.NewMockStore(ctrl),
		usermgnt:  mock_drivenadapters.NewMockUserManagement(ctrl),
		permCheck: mock_perm.NewMockPermCheckerService(ctrl),
	}
}

func NewDagVersionInstance(dep MockDependency) *dagVersion {
	initARLog()
	initErrorInfo()
	return &dagVersion{
		mongo:     dep.mongo,
		usermgnt:  dep.usermgnt,
		permCheck: dep.permCheck,
	}
}

func TestListDagVersions(t *testing.T) {
	dependency := NewDependency(t)
	dagVersion := NewDagVersionInstance(dependency)

	user := &drivenadapters.UserInfo{}

	versions := []entity.DagVersion{
		{
			DagID:     "582444344745686460",
			UserID:    "4fa5fafe-e751-11ef-b014-dac047ec7bab",
			Version:   "v0.0.1",
			VersionID: "582444426165515708",
			ChangeLog: "",
		},
		{
			DagID:     "582444344745686459",
			UserID:    "4fa5fafe-e751-11ef-b014-dac047ec7bab",
			Version:   "v0.0.0",
			VersionID: "582444426165515707",
			ChangeLog: "",
		},
	}

	accessors := map[string]string{"4fa5fafe-e751-11ef-b014-dac047ec7bab": "test"}

	Convey("ListDagVersions", t, func() {
		Convey("No Permission", func() {
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("no permission"))
			_, err := dagVersion.ListDagVersions(context.Background(), "", user)
			assert.NotEqual(t, err, nil)
		})

		Convey("List Dag Versions Error", func() {
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.mongo.EXPECT().ListDagVersions(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test error"))
			_, err := dagVersion.ListDagVersions(context.Background(), "", user)
			assert.NotEqual(t, err, nil)
		})

		Convey("List Dag Versions Success", func() {
			want := []DagVersionSimple{
				{
					ID:        "582444344745686460",
					Version:   "v0.0.1",
					VersionID: "582444426165515708",
					ChangeLog: "",
					UserID:    "4fa5fafe-e751-11ef-b014-dac047ec7bab",
					UserName:  "test",
				},
				{
					ID:        "582444344745686459",
					Version:   "v0.0.0",
					VersionID: "582444426165515707",
					ChangeLog: "",
					UserID:    "4fa5fafe-e751-11ef-b014-dac047ec7bab",
					UserName:  "test",
				},
			}
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.mongo.EXPECT().ListDagVersions(gomock.Any(), gomock.Any()).Return(versions, nil)
			dependency.usermgnt.EXPECT().GetNameByAccessorIDs(gomock.Any()).Return(accessors, nil)
			res, err := dagVersion.ListDagVersions(context.Background(), "", user)
			assert.Equal(t, err, nil)
			So(res, ShouldResemble, want)
		})
	})
}

func TestGetNextVersion(t *testing.T) {
	dependency := NewDependency(t)
	dagVersion := NewDagVersionInstance(dependency)

	versions := []entity.DagVersion{
		{
			DagID:     "582444344745686460",
			UserID:    "4fa5fafe-e751-11ef-b014-dac047ec7bab",
			Version:   "v0.0.1",
			VersionID: "582444426165515708",
			ChangeLog: "",
		},
	}

	Convey("TestGetNextVersion", t, func() {
		Convey("Dag Not Found", func() {
			dependency.mongo.EXPECT().GetDag(gomock.Any(), gomock.Any()).Return(nil, mongo.ErrNoDocuments)
			_, err := dagVersion.GetNextVersion(context.Background(), "")
			assert.NotEqual(t, err, nil)
		})

		Convey("Get Dag Internal Error", func() {
			dependency.mongo.EXPECT().GetDag(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test error"))
			_, err := dagVersion.GetNextVersion(context.Background(), "")
			assert.NotEqual(t, err, nil)
		})

		Convey("GetLatestDagVersion Error", func() {
			dependency.mongo.EXPECT().GetDag(gomock.Any(), gomock.Any()).Return(&entity.Dag{}, nil)
			dependency.mongo.EXPECT().ListDagVersions(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test error"))
			_, err := dagVersion.GetNextVersion(context.Background(), "")
			assert.NotEqual(t, err, nil)
		})

		Convey("No Historical Version", func() {
			dependency.mongo.EXPECT().GetDag(gomock.Any(), gomock.Any()).Return(&entity.Dag{}, nil)
			dependency.mongo.EXPECT().ListDagVersions(gomock.Any(), gomock.Any()).Return([]entity.DagVersion{}, nil)
			version, err := dagVersion.GetNextVersion(context.Background(), "")
			assert.Equal(t, err, nil)
			assert.Equal(t, version, "v0.0.1")
		})

		Convey("Get Latest Version", func() {
			dependency.mongo.EXPECT().GetDag(gomock.Any(), gomock.Any()).Return(&entity.Dag{}, nil)
			dependency.mongo.EXPECT().ListDagVersions(gomock.Any(), gomock.Any()).Return(versions, nil)
			version, err := dagVersion.GetNextVersion(context.Background(), "")
			assert.Equal(t, err, nil)
			assert.Equal(t, version, "v0.0.2")
		})
	})
}

func TestRevertToVersion(t *testing.T) {
	dependency := NewDependency(t)
	dagVersion := NewDagVersionInstance(dependency)

	user := &drivenadapters.UserInfo{}

	historyDag := &entity.DagVersion{
		DagID:     "582444344745686460",
		UserID:    "fa5fafe-e751-11ef-b014-dac047ec7bab",
		Version:   "v0.0.1",
		VersionID: "581751096591036551",
		ChangeLog: "",
	}

	dagInfo := &entity.Dag{
		BaseInfo: entity.BaseInfo{
			ID: "582444344745686460",
		},
		Name: "test",
	}

	versions := []entity.DagVersion{
		{
			DagID:     "582444344745686460",
			UserID:    "4fa5fafe-e751-11ef-b014-dac047ec7bab",
			Version:   "v1.0.1",
			VersionID: "582444426165515708",
			ChangeLog: "",
		},
	}

	Convey("TestRevertToVersion", t, func() {
		Convey("No Permission", func() {
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("no permission"))
			_, err := dagVersion.ListDagVersions(context.Background(), "", user)
			assert.NotEqual(t, err, nil)
		})

		Convey("GetHistoryDagByVersionID Error - Version Not Found", func() {
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.mongo.EXPECT().GetHistoryDagByVersionID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mongo.ErrNoDocuments)
			_, err := dagVersion.RevertToVersion(context.Background(), RevertDagReq{}, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("GetHistoryDagByVersionID Error - Internal Err", func() {
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.mongo.EXPECT().GetHistoryDagByVersionID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test error"))
			_, err := dagVersion.RevertToVersion(context.Background(), RevertDagReq{}, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("HistoryDag Config ParseToDag Error", func() {
			historyDag.Config = "[{}]"
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.mongo.EXPECT().GetHistoryDagByVersionID(gomock.Any(), gomock.Any(), gomock.Any()).Return(historyDag, nil)
			_, err := dagVersion.RevertToVersion(context.Background(), RevertDagReq{}, user)
			assert.NotEqual(t, err, nil)
		})

		historyDag.Config = `{"id":"582444344745686460","createdAt":1756693482,"updatedAt":1756693482,"userid":"4fa5fafe-e751-11ef-b014-dac047ec7bab","name":"测试多触发方式55","trigger":"manually","vars":{"docid":{},"userid":{"defaultValue":"4fa5fafe-e751-11ef-b014-dac047ec7bab"}},"status":"normal","tasks":[{"id":"0","actionName":"@trigger/dataflow-user"},{"id":"1","dependOn":["0"],"actionName":"@intelliinfo/transfer","timeoutSecs":86400,"params":{"data":"[{\"graph_id\":1,\"entities\":[{\"name\":\"people\",\"action\":\"upsert\",\"fields\":[{\"key\":\"people\",\"type\":\"string\",\"value\":\"{{__0.id}}\"}]}],\"edges\":[]}]"}}],"steps":[{"id":"0","title":"","operator":"@trigger/dataflow-user"},{"id":"1","title":"","operator":"@intelliinfo/transfer","parameters":{"data":"[{\"graph_id\":1,\"entities\":[{\"name\":\"people\",\"action\":\"upsert\",\"fields\":[{\"key\":\"people\",\"type\":\"string\",\"value\":\"{{__0.id}}\"}]}],\"edges\":[]}]"}}],"type":"data-flow","appinfo":{"enable":true},"priority":"high","trigger_config":{"parameters":{"accessorid":"00000000-0000-0000-0000-000000000000"}},"version":"v0.0.1","version_id":"582444339561526716","modify_by":"4fa5fafe-e751-11ef-b014-dac047ec7bab"}`
		Convey("GetDagByFields Error", func() {
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.mongo.EXPECT().GetHistoryDagByVersionID(gomock.Any(), gomock.Any(), gomock.Any()).Return(historyDag, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("test error"))
			_, err := dagVersion.RevertToVersion(context.Background(), RevertDagReq{}, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("Duplicate Title", func() {
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.mongo.EXPECT().GetHistoryDagByVersionID(gomock.Any(), gomock.Any(), gomock.Any()).Return(historyDag, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Return(dagInfo, nil)
			_, err := dagVersion.RevertToVersion(context.Background(), RevertDagReq{DagID: "581751096591036651", Title: "test"}, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("Invalid Semver Version", func() {
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.mongo.EXPECT().GetHistoryDagByVersionID(gomock.Any(), gomock.Any(), gomock.Any()).Return(historyDag, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Return(dagInfo, nil)
			dependency.mongo.EXPECT().ListDagVersions(gomock.Any(), gomock.Any()).Return(versions, nil)
			_, err := dagVersion.RevertToVersion(context.Background(), RevertDagReq{DagID: "582444344745686460", Title: "test", Version: "v1.0.0"}, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("Rollback To Version Success", func() {
			dependency.permCheck.EXPECT().CheckDagAndPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.mongo.EXPECT().GetHistoryDagByVersionID(gomock.Any(), gomock.Any(), gomock.Any()).Return(historyDag, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Return(dagInfo, nil)
			dependency.mongo.EXPECT().ListDagVersions(gomock.Any(), gomock.Any()).Return(versions, nil)
			dependency.mongo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).Return(nil)
			dependency.mongo.EXPECT().UpdateDag(gomock.Any(), gomock.Any()).Return(nil)
			dependency.mongo.EXPECT().CreateDagVersion(gomock.Any(), gomock.Any()).Return("", nil)
			_, err := dagVersion.RevertToVersion(context.Background(), RevertDagReq{DagID: "582444344745686460", Title: "test"}, user)
			assert.Equal(t, err, nil)
		})
	})

}
