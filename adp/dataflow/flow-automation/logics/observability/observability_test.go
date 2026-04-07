package observability

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	i18n "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/i18n"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	cstore "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/store"
	traceCommon "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/common"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_logics/mock_perm"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils/ptr"
	. "github.com/smartystreets/goconvey/convey"
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
	mongo       *mod.MockStore
	uniquery    *mock_drivenadapters.MockUniqueryDriven
	usermgnt    *mock_drivenadapters.MockUserManagement
	permChecker *mock_perm.MockPermCheckerService
}

func NewDependency(t *testing.T) MockDependency {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return MockDependency{
		mongo:       mod.NewMockStore(ctrl),
		uniquery:    mock_drivenadapters.NewMockUniqueryDriven(ctrl),
		usermgnt:    mock_drivenadapters.NewMockUserManagement(ctrl),
		permChecker: mock_perm.NewMockPermCheckerService(ctrl),
	}
}

func NewObservabilityInstance(dep MockDependency) *observability {
	initARLog()
	initErrorInfo()
	return &observability{
		mongo:       dep.mongo,
		uniquery:    dep.uniquery,
		usermgnt:    dep.usermgnt,
		permChecker: dep.permChecker,
		memoryCache: cstore.NewLocalCache(&cstore.Option{
			Expiration:      30 * time.Second,
			CleanUpInterval: 5 * time.Minute,
		}),
	}
}

func mockListDagWithFilters(dags []*entity.Dag, total int64, err error) *Patches {
	return ApplyFunc(mgnt.ListDagWithFilters, func(context.Context, mgnt.QueryParams, ...mgnt.ListDagOption) ([]*entity.Dag, int64, error) {
		return dags, total, err
	})
}

func TestFullView(t *testing.T) {
	dependency := NewDependency(t)
	obs := NewObservabilityInstance(dependency)

	params := ObservabilityQueryParams{
		Page:  ptr.Int64(0),
		Limit: ptr.Int64(10),
		Type:  ptr.String(common.DagTypeDataFlow),
	}

	dags := []*entity.Dag{
		{
			BaseInfo: entity.BaseInfo{
				ID: "1",
			},
			Name: "test",
		},
	}

	dataViews := drivenadapters.MetricQueryRes{
		Datas: []drivenadapters.DataEntry{
			{
				Labels: drivenadapters.LabelData{
					DagID:  "1",
					Status: "success",
				},
				Times:  []int64{},
				Values: []interface{}{1.0},
			},
			{
				Labels: drivenadapters.LabelData{
					DagID:  "1",
					Status: "failed",
				},
				Times:  []int64{},
				Values: []interface{}{1.0},
			},
		},
	}

	inss := []*entity.DagInstanceGroup{
		{
			Total: 1,
			DagIns: &entity.DagInstance{
				Status: entity.DagInstanceStatusInit,
			},
		},
		{
			Total: 1,
			DagIns: &entity.DagInstance{
				Status: entity.DagInstanceStatusRunning,
			},
		},
	}

	user := &drivenadapters.UserInfo{}
	Convey("FullView", t, func() {
		Convey("Perm Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("check perm error"))
			_, err := obs.FullView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("Query Time Error", func() {
			params := ObservabilityQueryParams{
				StartTime: ptr.Int64(1747902762001),
				EndTime:   ptr.Int64(1747902762000),
			}
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			_, err := obs.FullView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("ListDag Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(nil, 0, fmt.Errorf("list dag error"))
			defer patch.Reset()
			_, err := obs.FullView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("Query Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(drivenadapters.MetricQueryRes{}, fmt.Errorf("query error"))
			_, err := obs.FullView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("Success", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dataViews, nil)
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(inss, nil)
			res, err := obs.FullView(context.Background(), params, user)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Basic.DagCnt, int64(1))
			assert.Equal(t, res.Run.SuccessCnt, int64(1))
			assert.Equal(t, res.Run.FailedCnt, int64(1))
			assert.Equal(t, res.Run.Running, int64(1))
			assert.Equal(t, res.Run.Scheduled, int64(1))
			assert.Equal(t, res.Run.TotalCnt, int64(4))
		})
	})
}

func TestRuntimeView(t *testing.T) {
	dependency := NewDependency(t)
	obs := NewObservabilityInstance(dependency)

	params := ObservabilityQueryParams{
		Page:  ptr.Int64(0),
		Limit: ptr.Int64(10),
		Type:  ptr.String(common.DagTypeDataFlow),
	}

	dags := []*entity.Dag{
		{
			BaseInfo: entity.BaseInfo{
				ID: "1",
			},
			Name: "test",
		},
	}

	dataViews := drivenadapters.MetricQueryRes{
		Datas: []drivenadapters.DataEntry{
			{
				Labels: drivenadapters.LabelData{
					DagID:  "1",
					Status: "success",
				},
				Times:  []int64{},
				Values: []interface{}{1.0},
			},
			{
				Labels: drivenadapters.LabelData{
					DagID:  "1",
					Status: "failed",
				},
				Times:  []int64{},
				Values: []interface{}{1.0},
			},
			{
				Labels: drivenadapters.LabelData{
					DagID:  "1",
					Status: "canceled",
				},
				Times:  []int64{},
				Values: []interface{}{1.0},
			},
		},
	}

	timeAvgDataViews := drivenadapters.MetricQueryRes{
		Datas: []drivenadapters.DataEntry{
			{
				Labels: drivenadapters.LabelData{
					DagID: "1",
				},
				Values: []interface{}{1.0},
			},
		},
	}

	dagInss := []*entity.DagInstanceGroup{
		{
			DagIns: &entity.DagInstance{
				DagID:  "1",
				Status: entity.DagInstanceStatusInit,
			},
			Total: 1,
		},
		{
			DagIns: &entity.DagInstance{
				DagID:  "1",
				Status: entity.DagInstanceStatusRunning,
			},
			Total: 1,
		},
		{
			DagIns: &entity.DagInstance{
				DagID:  "1",
				Status: entity.DagInstanceStatusScheduled,
			},
			Total: 1,
		},
		{
			DagIns: &entity.DagInstance{
				DagID:  "1",
				Status: entity.DagInstanceStatusBlocked,
			},
			Total: 1,
		},
	}

	user := &drivenadapters.UserInfo{}
	Convey("RuntimeView", t, func() {
		Convey("Perm Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("check perm error"))
			_, err := obs.RuntimeView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("List Dag Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(nil, 0, fmt.Errorf("list dag error"))
			defer patch.Reset()
			_, err := obs.RuntimeView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("CollectDagMetrics Error - Query Dag Count Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(drivenadapters.MetricQueryRes{}, fmt.Errorf("query error"))
			_, err := obs.RuntimeView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("CollectDagMetrics Error - Query Dag Run Time Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dataViews, nil)
			dependency.uniquery.EXPECT().QueryDagRunTimeAvg(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(drivenadapters.MetricQueryRes{}, fmt.Errorf("query error"))
			_, err := obs.RuntimeView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("CollectDagMetrics Error - Group Dag Instance Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dataViews, nil)
			dependency.uniquery.EXPECT().QueryDagRunTimeAvg(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(timeAvgDataViews, nil)
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("group dag instance error"))
			_, err := obs.RuntimeView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("Success", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dataViews, nil)
			dependency.uniquery.EXPECT().QueryDagRunTimeAvg(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(timeAvgDataViews, nil)
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(dagInss, nil)
			dependency.usermgnt.EXPECT().GetNameByAccessorIDs(gomock.Any()).Return(map[string]string{"1": "test"}, nil)
			res, err := obs.RuntimeView(context.Background(), params, user)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Datas[0].ID, "1")
			assert.Equal(t, res.Datas[0].Name, "test")
			assert.Equal(t, res.Datas[0].StatusSummary.Total, int64(7))
			assert.Equal(t, res.Datas[0].StatusSummary.Success, int64(1))
			assert.Equal(t, res.Datas[0].StatusSummary.Failed, int64(1))
			assert.Equal(t, res.Datas[0].StatusSummary.Canceled, int64(1))
			assert.Equal(t, res.Datas[0].StatusSummary.Blocked, int64(1))
			assert.Equal(t, res.Datas[0].StatusSummary.Running, int64(1))
			assert.Equal(t, res.Datas[0].StatusSummary.Init, int64(2))
		})
	})
}

func TestRecentRunView(t *testing.T) {
	dependency := NewDependency(t)
	obs := NewObservabilityInstance(dependency)

	dags := []*entity.Dag{
		{
			BaseInfo: entity.BaseInfo{
				ID: "1",
			},
			Name: "test",
		},
	}

	dagInss := []*entity.DagInstanceGroup{
		{
			DagIns: &entity.DagInstance{
				DagID:  "1",
				Status: entity.DagInstanceStatusInit,
			},
			Total: 1,
		},
		{
			DagIns: &entity.DagInstance{
				DagID:  "1",
				Status: entity.DagInstanceStatusRunning,
			},
			Total: 1,
		},
		{
			DagIns: &entity.DagInstance{
				DagID:  "1",
				Status: entity.DagInstanceStatusScheduled,
			},
			Total: 1,
		},
		{
			DagIns: &entity.DagInstance{
				DagID:  "1",
				Status: entity.DagInstanceStatusBlocked,
			},
			Total: 1,
		},
	}

	dataViews := drivenadapters.MetricQueryRes{
		Datas: []drivenadapters.DataEntry{
			{
				Labels: drivenadapters.LabelData{
					DagID:  "1",
					Status: "success",
				},
				Times:  []int64{},
				Values: []interface{}{1.0},
			},
			{
				Labels: drivenadapters.LabelData{
					DagID:  "1",
					Status: "failed",
				},
				Times:  []int64{},
				Values: []interface{}{1.0},
			},
			{
				Labels: drivenadapters.LabelData{
					DagID:  "1",
					Status: "canceled",
				},
				Times:  []int64{},
				Values: []interface{}{1.0},
			},
		},
	}

	timeAvgDataViews := drivenadapters.MetricQueryRes{
		Datas: []drivenadapters.DataEntry{
			{
				Labels: drivenadapters.LabelData{
					DagID: "1",
				},
				Values: []interface{}{1.0},
			},
		},
	}

	user := &drivenadapters.UserInfo{}
	params := ObservabilityQueryParams{}

	Convey("TestRecentRunView", t, func() {
		Convey("Perm Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("check perm error"))
			_, err := obs.RecentRunView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("List Dag Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(nil, 0, fmt.Errorf("list dag error"))
			defer patch.Reset()
			_, err := obs.RecentRunView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("Group Dag Instance Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("group dag instance error"))
			_, err := obs.RecentRunView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("CollectDagMetrics Error - Query Dag Count Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(dagInss, nil)
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(drivenadapters.MetricQueryRes{}, fmt.Errorf("query error"))
			_, err := obs.RecentRunView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("CollectDagMetrics Error - Query Dag Run Time Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(dagInss, nil)
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dataViews, nil)
			dependency.uniquery.EXPECT().QueryDagRunTimeAvg(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(drivenadapters.MetricQueryRes{}, fmt.Errorf("query error"))
			_, err := obs.RecentRunView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("CollectDagMetrics Error - Group Dag Instance Error", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(dagInss, nil)
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dataViews, nil)
			dependency.uniquery.EXPECT().QueryDagRunTimeAvg(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(timeAvgDataViews, nil)
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("group dag instance error"))
			_, err := obs.RecentRunView(context.Background(), params, user)
			assert.NotEqual(t, err, nil)
		})

		Convey("Success", func() {
			dependency.permChecker.EXPECT().CheckPerm(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			patch := mockListDagWithFilters(dags, 1, nil)
			defer patch.Reset()
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(dagInss, nil)
			dependency.uniquery.EXPECT().QueryDagStatusCount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dataViews, nil)
			dependency.uniquery.EXPECT().QueryDagRunTimeAvg(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(timeAvgDataViews, nil)
			dependency.mongo.EXPECT().GroupDagInstance(gomock.Any(), gomock.Any()).Return(dagInss, nil)
			dependency.usermgnt.EXPECT().GetNameByAccessorIDs(gomock.Any()).Return(map[string]string{"1": "test"}, nil)
			res, err := obs.RecentRunView(context.Background(), params, user)
			assert.Equal(t, err, nil)
			assert.Equal(t, res[0].ID, "1")
			assert.Equal(t, res[0].Name, "test")
			assert.Equal(t, res[0].StatusSummary.Total, int64(7))
			assert.Equal(t, res[0].StatusSummary.Success, int64(1))
			assert.Equal(t, res[0].StatusSummary.Failed, int64(1))
			assert.Equal(t, res[0].StatusSummary.Canceled, int64(1))
			assert.Equal(t, res[0].StatusSummary.Blocked, int64(1))
			assert.Equal(t, res[0].StatusSummary.Running, int64(1))
			assert.Equal(t, res[0].StatusSummary.Init, int64(2))
		})
	})
}
