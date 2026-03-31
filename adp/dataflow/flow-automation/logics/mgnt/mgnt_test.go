package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
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
	"go.uber.org/mock/gomock"
)

const (
	// AnyshareManualTrigger 手动触发器schema路径
	AnyshareManualTrigger = "../../schema/trigger/manualtrigger.json"
	// AnyshareCopyOrMove 文件、文件夹复制移动操作schema路径
	AnyshareCopyOrMove = "../../schema/anyshare/copyormove.json"
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
	ad        *mock_drivenadapters.MockAnyData
	op        *mock_drivenadapters.MockAgentOperatorIntegration
	permCheck *mock_perm.MockPermCheckerService
}

func NewDependency(t *testing.T) MockDependency {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return MockDependency{
		mongo:     mod.NewMockStore(ctrl),
		usermgnt:  mock_drivenadapters.NewMockUserManagement(ctrl),
		ad:        mock_drivenadapters.NewMockAnyData(ctrl),
		op:        mock_drivenadapters.NewMockAgentOperatorIntegration(ctrl),
		permCheck: mock_perm.NewMockPermCheckerService(ctrl),
	}
}

// NewMgntInstance new mgnt instance
func NewMgntInstance(dep MockDependency) *mgnt {
	initARLog()
	initErrorInfo()
	return &mgnt{
		mongo:             dep.mongo,
		usermgnt:          dep.usermgnt,
		operator:          dep.op,
		permCheck:         dep.permCheck,
		taskTimeoutConfig: common.NewTimeoutConfig(),
	}
}

func TestCreateDag(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	var param = &CreateDagReq{
		Title:       "Title",
		Description: "Description",
		Status:      "Status",
		Steps: []entity.Step{
			{
				ID:         "0",
				Title:      "",
				Operator:   "@trigger/manual",
				Parameters: map[string]interface{}{},
				Branches:   []entity.Branch{},
				Steps:      []entity.Step{},
			},
			{
				ID:       "2",
				Title:    "",
				Operator: "@control/flow/branches",
				Branches: []entity.Branch{
					{
						Conditions: [][]entity.TaskCondition{
							{
								{
									ID:     "3",
									Source: "",
									Parameter: entity.TaskConditionParameter{
										A: "111111",
										B: "111111",
									},
									Op:           "string/eq",
									ParamsRender: entity.ParamsRender{},
								}},
						},
						Steps: []entity.Step{
							{
								ID:       "4",
								Title:    "",
								Operator: "@internal/string/split",
								Parameters: map[string]interface{}{
									"text":      "a-b",
									"separator": "-",
								},
								Branches: []entity.Branch{},
								Steps:    []entity.Step{},
							},
						},
					},
				},
				Steps: []entity.Step{},
			},
			{
				ID:       "1",
				Title:    "",
				Operator: "@anyshare/file/copy",
				Parameters: map[string]interface{}{
					"docid":      "gns://9A8C9277947D4898A350427C768A194C/33CD537CD6914AA69649A9EA01917B06",
					"destparent": "gns://9A8C9277947D4898A350427C768A194C/63B5E4667AD643188D67E65AF5570E10",
					"ondup":      2},
				Branches: []entity.Branch{},
				Steps:    []entity.Step{},
			},
		},
	}
	var userInfo = &drivenadapters.UserInfo{
		UserID:     "UserID",
		UserName:   "UserName",
		ParentDeps: nil,
		CsfLevel:   0,
		UdID:       "",
		LoginIP:    "",
		TokenID:    "",
	}
	Convey("CreateDag", t, func() {
		internalErr := aerr.NewIError(aerr.InternalError, "", nil)
		patch := ApplyGlobalVar(&common.ActionMap, map[string]string{common.MannualTriggerOpt: AnyshareManualTrigger,
			common.AnyshareFolderCopyOpt: AnyshareCopyOrMove})
		defer patch.Reset()

		// ListDagCount internal err
		dependency.mongo.EXPECT().ListDagCount(gomock.Any(), gomock.Any()).Times(1).Return(int64(0), internalErr)
		dagID, err := mockMgnt.CreateDag(context.Background(), param, userInfo)
		assert.Equal(t, dagID, "")
		assert.Equal(t, err, internalErr)

		// total大于50
		dependency.mongo.EXPECT().ListDagCount(gomock.Any(), gomock.Any()).Times(1).Return(int64(50), nil)
		dagID, err = mockMgnt.CreateDag(context.Background(), param, userInfo)
		assert.Equal(t, "", dagID)
		assert.NotEqual(t, err, nil)

		// total 小于50
		dependency.mongo.EXPECT().ListDagCount(gomock.Any(), gomock.Any()).AnyTimes().Return(int64(10), nil)

		dependency.usermgnt.EXPECT().GetUserInfo(gomock.Any()).AnyTimes().Return(*userInfo, nil)
		// trigger type invalid
		var param1 = &CreateDagReq{}
		data, _ := json.Marshal(*param)
		json.Unmarshal(data, param1) //nolint
		param1.Steps[0].Operator = "@trigger/manual1"
		dagID, err = mockMgnt.CreateDag(context.Background(), param1, userInfo)
		assert.Equal(t, "", dagID)
		assert.NotEqual(t, err, nil)

		// start node not trigger type invalid
		param1.Steps[0].Operator = "@anyshare/file/copy"
		dagID, err = mockMgnt.CreateDag(context.Background(), param1, userInfo)
		assert.Equal(t, "", dagID)
		assert.NotEqual(t, err, nil)

		// other node operator is trigger
		var param2 = &CreateDagReq{
			Title:       "Title",
			Description: "Description",
			Status:      "Status",
			Steps: []entity.Step{
				{
					ID:         "0",
					Title:      "",
					Operator:   "@trigger/manual",
					Parameters: map[string]interface{}{},
					Branches:   []entity.Branch{},
					Steps:      []entity.Step{},
				},
				{
					ID:       "1",
					Title:    "",
					Operator: "@trigger/manual",
					Parameters: map[string]interface{}{
						"docid":      "gns://9A8C9277947D4898A350427C768A194C/33CD537CD6914AA69649A9EA01917B06",
						"destparent": "gns://9A8C9277947D4898A350427C768A194C/63B5E4667AD643188D67E65AF5570E10",
						"ondup":      2},
					Branches: []entity.Branch{},
					Steps:    []entity.Step{},
				},
			},
		}
		dagID, err = mockMgnt.CreateDag(context.Background(), param2, userInfo)
		assert.Equal(t, "", dagID)
		assert.NotEqual(t, err, nil)
	})
}

func TestListHistoryData(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	Convey("ListHistoryData", t, func() {
		Convey("ListDagCountByFields Failed", func() {
			dependency.mongo.EXPECT().ListDagCountByFields(gomock.Any(), gomock.Any()).Times(1).Return(int64(0), fmt.Errorf("ListDagCountByFields Err"))

			_, err := mockMgnt.ListHistoryData(context.Background(), 1, 10)
			assert.NotEqual(t, err, nil)
		})

		Convey("ListDagByFields Failed", func() {
			dependency.mongo.EXPECT().ListDagCountByFields(gomock.Any(), gomock.Any()).Times(1).Return(int64(10), nil)
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, fmt.Errorf("ListDagByFields Err"))

			_, err := mockMgnt.ListHistoryData(context.Background(), 1, 10)
			assert.NotEqual(t, err, nil)
		})

		Convey("Success with empty result", func() {
			dependency.mongo.EXPECT().ListDagCountByFields(gomock.Any(), gomock.Any()).Times(1).Return(int64(0), nil)

			res, err := mockMgnt.ListHistoryData(context.Background(), 1, 10)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Total, int64(0))
			assert.Equal(t, len(res.Items), 0)
		})

		Convey("Success with data", func() {
			dags := []*entity.Dag{
				{
					BaseInfo: entity.BaseInfo{
						ID: "dag1"},
					Type: "test_type",
				},
				{
					BaseInfo: entity.BaseInfo{
						ID: "dag2"},
					Type: "",
				},
			}

			dependency.mongo.EXPECT().ListDagCountByFields(gomock.Any(), gomock.Any()).Times(1).Return(int64(2), nil)
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags, nil)

			res, err := mockMgnt.ListHistoryData(context.Background(), 0, 10)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Total, int64(2))
			assert.Equal(t, len(res.Items), 2)
			assert.Equal(t, res.Items[0].ID, "dag1")
			assert.Equal(t, res.Items[0].Type, "test_type")
			assert.Equal(t, res.Items[1].ID, "dag2")
			assert.Equal(t, res.Items[1].Type, common.DagTypeDefault)
		})

		Convey("Limit exceeds maximum", func() {
			dependency.mongo.EXPECT().ListDagCountByFields(gomock.Any(), gomock.Any()).Times(1).Return(int64(0), nil)

			res, err := mockMgnt.ListHistoryData(context.Background(), 1, 2000)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Limit, int64(1000))
		})
	})
}
