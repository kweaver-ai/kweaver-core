package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils/ptr"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/mock/gomock"
)

func TestCreateComboOperator(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)
	userInfo := &drivenadapters.UserInfo{
		UserID:     "UserID",
		UserName:   "UserName",
		ParentDeps: nil,
		CsfLevel:   0,
		UdID:       "",
		LoginIP:    "",
		TokenID:    "",
	}

	req := &ComboOperatorReq{
		Title:       "测试Title",
		Description: "测试描述",
		Steps: []entity.Step{
			{
				ID:       "0",
				Title:    "",
				Operator: "@trigger/form",
				Parameters: map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"key":  "fbgcgRkxnPycGoeb",
							"type": "string",
							"name": "abc",
						},
					},
				},
			},
		},
		Category: "data_split",
		OutPuts:  []*entity.OutPut{},
	}
	Convey("CreateComboOperator", t, func() {
		Convey("GetUserInfo Failed", func() {
			dependency.usermgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(drivenadapters.UserInfo{}, fmt.Errorf("get user info err"))
			_, _, err := mockMgnt.CreateComboOperator(context.Background(), &ComboOperatorReq{}, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("CheckAndBuildDag Failed", func() {
			dependency.usermgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(*userInfo, nil)
			patch := ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return fmt.Errorf("CheckAndBuildDag Err")
			})
			defer patch.Reset()
			_, _, err := mockMgnt.CreateComboOperator(context.Background(), req, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("CreateDag Failed", func() {
			dependency.usermgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(*userInfo, nil)
			patch := ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return nil
			})
			defer patch.Reset()
			dependency.mongo.EXPECT().CreateDag(gomock.Any(), gomock.Any()).Times(1).Return("", fmt.Errorf("CreateDag Err"))
			_, _, err := mockMgnt.CreateComboOperator(context.Background(), req, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("RegisterOperator Failed", func() {
			dependency.usermgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(*userInfo, nil)
			patch := ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return nil
			})
			defer patch.Reset()
			dependency.mongo.EXPECT().CreateDag(gomock.Any(), gomock.Any()).Times(1).Return("DagID", nil)
			dependency.op.EXPECT().RegisterOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, fmt.Errorf("RegisterOperator Err"))
			dependency.mongo.EXPECT().DeleteDag(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			_, _, err := mockMgnt.CreateComboOperator(context.Background(), req, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("Success", func() {
			dependency.usermgnt.EXPECT().GetUserInfo(gomock.Any()).Times(1).Return(*userInfo, nil)
			patch := ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return nil
			})
			defer patch.Reset()
			res := []*drivenadapters.OperatorModifyResp{
				{
					Status:     RegisterSuccessStatus,
					OperatorID: "",
					Version:    "",
					Error:      map[string]any{},
				},
			}
			dependency.mongo.EXPECT().CreateDag(gomock.Any(), gomock.Any()).Times(1).Return("DagID", nil)
			dependency.op.EXPECT().RegisterOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(res, nil)
			dependency.mongo.EXPECT().UpdateDag(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			_, _, err := mockMgnt.CreateComboOperator(context.Background(), req, userInfo)
			assert.Equal(t, err, nil)
		})
	})
}

func TestUpdateComboOperator(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)
	userInfo := &drivenadapters.UserInfo{
		UserID:     "UserID",
		UserName:   "UserName",
		ParentDeps: nil,
		CsfLevel:   0,
		UdID:       "",
		LoginIP:    "",
		TokenID:    "",
	}

	dagID := "563502292490286524"
	newDagID := "563502292490286525"

	req := &OptionalComboOperatorReq{
		Title:       ptr.String("测试Title"),
		Description: ptr.String("测试描述"),
		Steps: &[]entity.Step{
			{
				ID:       "0",
				Title:    "",
				Operator: "@trigger/form",
				Parameters: map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"key":  "fbgcgRkxnPycGoeb",
							"type": "string",
							"name": "abc",
						},
					},
				},
			},
		},
		Category: ptr.String("data_split"),
	}

	dag := &entity.Dag{
		BaseInfo: entity.BaseInfo{
			ID:        dagID,
			CreatedAt: 1678838400,
			UpdatedAt: 1678838400,
		},
		Name:        "测试Title",
		Description: "测试描述",
		Steps: []entity.Step{
			{
				ID:       "0",
				Title:    "",
				Operator: "@trigger/form",
				Parameters: map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"key":  "doc_id",
							"type": "string",
							"name": "abc",
						},
					},
				},
			},
		},
		ExecMode: "sync",
		Category: "data_split",
		OutPuts: []*entity.OutPut{
			{
				Key:  "key",
				Name: "abc",
				Type: "string",
			},
		},
	}

	op := &drivenadapters.OperatorResponse{
		Status: PublishedStatus,
	}

	Convey("UpdateComboOperator", t, func() {
		Convey("Get Operator Info InternalErr", func() {
			patch := ApplyMethod(reflect.TypeOf(mockMgnt.permCheck), "CheckDagAndPerm", func(perm.PermCheckerService, context.Context, string, *drivenadapters.UserInfo, perm.OperationProvider) (bool, error) {
				return true, nil
			})
			defer patch.Reset()
			errBytes, _ := json.Marshal(map[string]interface{}{"code": "InternalError"})
			dependency.op.EXPECT().LatestOperatorInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, ierr.ExHTTPError{Body: string(errBytes), Status: 500})
			err := mockMgnt.UpdateComboOperator(context.Background(), req, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("Operator Not Found", func() {
			patch := ApplyMethod(reflect.TypeOf(mockMgnt.permCheck), "CheckDagAndPerm", func(perm.PermCheckerService, context.Context, string, *drivenadapters.UserInfo, perm.OperationProvider) (bool, error) {
				return true, nil
			})
			defer patch.Reset()
			errBytes, _ := json.Marshal(map[string]interface{}{"code": ""})
			dependency.op.EXPECT().LatestOperatorInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, ierr.ExHTTPError{Body: string(errBytes), Status: 404})
			err := mockMgnt.UpdateComboOperator(context.Background(), req, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("Dag Not Found", func() {
			patch := ApplyMethod(reflect.TypeOf(mockMgnt.permCheck), "CheckDagAndPerm", func(perm.PermCheckerService, context.Context, string, *drivenadapters.UserInfo, perm.OperationProvider) (bool, error) {
				return true, nil
			})
			defer patch.Reset()
			dependency.op.EXPECT().LatestOperatorInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(op, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Times(1).Return(nil, mongo.ErrNoDocuments)
			err := mockMgnt.UpdateComboOperator(context.Background(), req, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("CheckAndBuildDag Failed", func() {
			patch := ApplyMethod(reflect.TypeOf(mockMgnt.permCheck), "CheckDagAndPerm", func(perm.PermCheckerService, context.Context, string, *drivenadapters.UserInfo, perm.OperationProvider) (bool, error) {
				return true, nil
			})
			dependency.op.EXPECT().LatestOperatorInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(op, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Times(1).Return(dag, nil)
			patch.ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return fmt.Errorf("CheckAndBuildDag Err")
			})
			defer patch.Reset()
			err := mockMgnt.UpdateComboOperator(context.Background(), req, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("New Dag Version, Update Operator Error", func() {
			patch := ApplyMethod(reflect.TypeOf(mockMgnt.permCheck), "CheckDagAndPerm", func(perm.PermCheckerService, context.Context, string, *drivenadapters.UserInfo, perm.OperationProvider) (bool, error) {
				return true, nil
			})
			dependency.op.EXPECT().LatestOperatorInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(op, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Times(1).Return(dag, nil)
			patch.ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return nil
			})
			defer patch.Reset()
			dependency.mongo.EXPECT().CreateDag(gomock.Any(), gomock.Any()).Times(1).Return(newDagID, nil)
			dependency.op.EXPECT().UpdateOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, fmt.Errorf("UpdateOperator Err"))
			dependency.mongo.EXPECT().DeleteDag(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			err := mockMgnt.UpdateComboOperator(context.Background(), req, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("New Dag Version, Update Operator Success", func() {
			patch := ApplyMethod(reflect.TypeOf(mockMgnt.permCheck), "CheckDagAndPerm", func(perm.PermCheckerService, context.Context, string, *drivenadapters.UserInfo, perm.OperationProvider) (bool, error) {
				return true, nil
			})
			dependency.op.EXPECT().LatestOperatorInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(op, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Times(1).Return(dag, nil)
			patch.ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return nil
			})
			defer patch.Reset()
			dependency.mongo.EXPECT().CreateDag(gomock.Any(), gomock.Any()).Times(1).Return(newDagID, nil)
			dependency.op.EXPECT().UpdateOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return([]*drivenadapters.OperatorModifyResp{}, nil)
			err := mockMgnt.UpdateComboOperator(context.Background(), req, userInfo)
			assert.Equal(t, err, nil)
		})

		Convey("Dag Version Merge, Update Operator Failed", func() {
			req := &OptionalComboOperatorReq{
				Title:       ptr.String("New测试Title"),
				Description: ptr.String("New测试描述"),
				Category:    ptr.String("data_split"),
			}
			patch := ApplyMethod(reflect.TypeOf(mockMgnt.permCheck), "CheckDagAndPerm", func(perm.PermCheckerService, context.Context, string, *drivenadapters.UserInfo, perm.OperationProvider) (bool, error) {
				return true, nil
			})
			dependency.op.EXPECT().LatestOperatorInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(op, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Times(1).Return(dag, nil)
			patch.ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return nil
			})
			defer patch.Reset()
			dependency.mongo.EXPECT().UpdateDag(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dependency.op.EXPECT().UpdateOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, fmt.Errorf("UpdateOperator Err"))
			err := mockMgnt.UpdateComboOperator(context.Background(), req, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("Dag Version Merge, Update Operator Sucess", func() {
			req := &OptionalComboOperatorReq{
				Title:       ptr.String("New测试Title"),
				Description: ptr.String("New测试描述"),
				Category:    ptr.String("data_split"),
			}
			patch := ApplyMethod(reflect.TypeOf(mockMgnt.permCheck), "CheckDagAndPerm", func(perm.PermCheckerService, context.Context, string, *drivenadapters.UserInfo, perm.OperationProvider) (bool, error) {
				return true, nil
			})
			dependency.op.EXPECT().LatestOperatorInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(op, nil)
			dependency.mongo.EXPECT().GetDagByFields(gomock.Any(), gomock.Any()).Times(1).Return(dag, nil)
			patch.ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return nil
			})
			defer patch.Reset()
			dependency.mongo.EXPECT().UpdateDag(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			dependency.op.EXPECT().UpdateOperator(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return([]*drivenadapters.OperatorModifyResp{}, nil)
			err := mockMgnt.UpdateComboOperator(context.Background(), req, userInfo)
			assert.Equal(t, err, nil)
		})

	})
}

func TestListComboOperator(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)
	userInfo := &drivenadapters.UserInfo{}

	params := map[string]interface{}{
		"page":        int64(0),
		"limit":       int64(20),
		"sortby":      "name",
		"order":       "asc",
		"operator_id": "id",
		"name":        "name",
		"version":     "v",
		"status":      "published",
		"category":    "data_split",
	}

	Convey("ListComboOperator", t, func() {
		Convey("List Operator Err", func() {
			dependency.op.EXPECT().OperatorList(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, fmt.Errorf("List Operator Err"))
			_, err := mockMgnt.ListComboOperator(context.Background(), params, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("Success", func() {
			now := time.Now().Unix()
			dependency.op.EXPECT().OperatorList(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&drivenadapters.OperatorList{
				Total:    1,
				Page:     0,
				PageSize: 20,
				Data: []*drivenadapters.OperatorResponse{
					{
						OperatorID: "id",
						Name:       "name",
						Version:    "v",
						Status:     "published",
						Metadata: drivenadapters.OperatorMetadata{
							Description: "desc",
						},
						CreateTime: now,
						UpdateTime: now,
						CreateUser: "user",
						UpdateUser: "user",
						OperatorInfo: drivenadapters.OperatorInfo{
							Category:      "data_split",
							ExecutionMode: "sync",
							OperatorType:  "composite",
							Source:        "unknown",
						},
						ExtendInfo: drivenadapters.ExtendInfo{
							"dag_id": "dag_id",
						},
					},
				},
			}, nil)
			res, err := mockMgnt.ListComboOperator(context.Background(), params, userInfo)
			assert.Equal(t, err, nil)
			assert.Equal(t, res.Ops[0].OperatorID, "id")
		})
	})
}

func TestCheckAndBuildDag(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	userInfo := &drivenadapters.UserInfo{
		UserID:     "UserID",
		UserName:   "UserName",
		ParentDeps: nil,
		CsfLevel:   0,
		UdID:       "",
		LoginIP:    "",
		TokenID:    "",
	}

	dagID := "563502292490286524"

	req := &OptionalComboOperatorReq{
		Title:       ptr.String("测试Title"),
		Description: ptr.String("测试描述"),
		Steps: &[]entity.Step{
			{
				ID:       "0",
				Title:    "",
				Operator: "@trigger/form",
				Parameters: map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"key":  "fbgcgRkxnPycGoeb",
							"type": "string",
							"name": "abc",
						},
					},
				},
			},
		},
		Category: ptr.String("data_split"),
	}

	dag := &entity.Dag{
		BaseInfo: entity.BaseInfo{
			ID:        dagID,
			CreatedAt: 1678838400,
			UpdatedAt: 1678838400,
		},
		Name:        "测试Title",
		Description: "测试描述",
		Steps: []entity.Step{
			{
				ID:       "0",
				Title:    "",
				Operator: "@trigger/form",
				Parameters: map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"key":  "doc_id",
							"type": "string",
							"name": "abc",
						},
					},
				},
			},
		},
		ExecMode: "sync",
		Category: "data_split",
		OutPuts: []*entity.OutPut{
			{
				Key:  "key",
				Name: "abc",
				Type: "string",
			},
		},
	}

	Convey("CheckAndBuildDag", t, func() {
		Convey("Valid Step Err", func() {
			patch := ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "validSteps", func(*mgnt, *Validate) *ValidateError {
				return &ValidateError{
					Ctx:             context.Background(),
					ErrType:         "v2",
					PublicErrorType: ierr.PublicErrorType,
					MainCode:        "",
					MainCodeV2:      ierr.PErrorBadRequest,
					ExtCode:         "",
					DescriptionKey:  ierr.PErrorBadRequest,
					Detail:          nil,
					Error:           nil,
				}
			})
			defer patch.Reset()
			err := mockMgnt.CheckAndBuildDag(context.Background(), req, dag, userInfo)
			assert.NotEqual(t, err, nil)
		})

		// 创建或更新组合算子内，流程配置的算子是否合法
		Convey("Valid Operator In Step Err", func() {
			patch := ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "validSteps", func(*mgnt, *Validate) *ValidateError {
				return nil
			})
			patch.ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "validOperatorInSteps", func(context.Context, []map[string]interface{}) (*ValidOperatorsResult, error) {
				return nil, fmt.Errorf("validOperatorInSteps Err")
			})
			defer patch.Reset()
			err := mockMgnt.CheckAndBuildDag(context.Background(), req, dag, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("Not Form Trigger", func() {
			req := &OptionalComboOperatorReq{
				Title:       ptr.String("测试Title"),
				Description: ptr.String("测试描述"),
				Steps: &[]entity.Step{
					{
						ID:         "0",
						Title:      "",
						Operator:   "@trigger/manual",
						Parameters: map[string]interface{}{},
					},
				},
				Category: ptr.String("data_split"),
			}
			patch := ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "validSteps", func(*mgnt, *Validate) *ValidateError {
				return nil
			})
			patch.ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "validOperatorInSteps", func(context.Context, []map[string]interface{}) (*ValidOperatorsResult, error) {
				return &ValidOperatorsResult{
					RefDagIDs: []string{},
					OpInfoMap: map[string]*OperatorInfo{},
				}, nil
			})
			defer patch.Reset()
			err := mockMgnt.CheckAndBuildDag(context.Background(), req, dag, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("Composite Operator Has Cycle", func() {
			patch := ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "validSteps", func(*mgnt, *Validate) *ValidateError {
				return nil
			})
			patch.ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "validOperatorInSteps", func(context.Context, []map[string]interface{}) (*ValidOperatorsResult, error) {
				return &ValidOperatorsResult{
					RefDagIDs: []string{},
					OpInfoMap: map[string]*OperatorInfo{},
				}, nil
			})
			patch.ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "hasCycle", func(context.Context, string, []string) (*CycleError, error) {
				return &CycleError{
					Cycle:      true,
					CurrID:     "id",
					ReferDagID: "refer_dag_id",
					ReferName:  "refer_name",
				}, nil
			})
			defer patch.Reset()
			err := mockMgnt.CheckAndBuildDag(context.Background(), req, dag, userInfo)
			assert.NotEqual(t, err, nil)
		})

		Convey("Check And Build Success", func() {
			patch := ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "validSteps", func(*mgnt, *Validate) *ValidateError {
				return nil
			})
			patch.ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "validOperatorInSteps", func(context.Context, []map[string]interface{}) (*ValidOperatorsResult, error) {
				return &ValidOperatorsResult{
					RefDagIDs: []string{},
					OpInfoMap: map[string]*OperatorInfo{},
				}, nil
			})
			patch.ApplyPrivateMethod(reflect.TypeOf(mockMgnt), "hasCycle", func(context.Context, string, []string) (*CycleError, error) {
				return &CycleError{
					Cycle: false,
				}, nil
			})
			defer patch.Reset()
			err := mockMgnt.CheckAndBuildDag(context.Background(), req, dag, userInfo)
			assert.Equal(t, err, nil)
		})

	})
}

func TestHasCycle(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	dagID := "563502390787995068"
	referIDs := []string{"563502390787995069"}

	Convey("HasCycle", t, func() {
		Convey("List Dag Err", func() {
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, fmt.Errorf("List Dag Err"))
			_, err := mockMgnt.hasCycle(context.Background(), dagID, referIDs, []string{common.DagTypeComboOperator})
			assert.NotEqual(t, err, nil)
		})

		Convey("Has Cycle", func() {
			dags1 := []*entity.Dag{
				{
					BaseInfo: entity.BaseInfo{
						ID:        "563502390787995069",
						CreatedAt: 1678838400,
						UpdatedAt: 1678838400,
					},
					Type:   common.DagTypeComboOperator,
					SubIDs: []string{"563502390787995070"},
				},
			}
			dags2 := []*entity.Dag{
				{
					BaseInfo: entity.BaseInfo{
						ID:        "563502390787995070",
						CreatedAt: 1678838400,
						UpdatedAt: 1678838400,
					},
					Type:   common.DagTypeComboOperator,
					SubIDs: []string{"563502390787995069"},
				},
			}
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags1, nil)
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags2, nil)
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags1, nil)
			cycle, err := mockMgnt.hasCycle(context.Background(), dagID, referIDs, []string{common.DagTypeComboOperator})
			assert.Equal(t, err, nil)
			assert.Equal(t, cycle.Cycle, true)
		})

		Convey("Do Not Has Cycle", func() {
			dags1 := []*entity.Dag{
				{
					BaseInfo: entity.BaseInfo{
						ID:        "563502390787995069",
						CreatedAt: 1678838400,
						UpdatedAt: 1678838400,
					},
					Type:   common.DagTypeComboOperator,
					SubIDs: []string{"563502390787995070"},
				},
			}
			dags2 := []*entity.Dag{
				{
					BaseInfo: entity.BaseInfo{
						ID:        "563502390787995070",
						CreatedAt: 1678838400,
						UpdatedAt: 1678838400,
					},
					Type: common.DagTypeComboOperator,
				},
			}
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags1, nil)
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(dags2, nil)
			cycle, err := mockMgnt.hasCycle(context.Background(), dagID, referIDs, []string{common.DagTypeComboOperator})
			assert.Equal(t, err, nil)
			assert.Equal(t, cycle.Cycle, false)
		})
	})
}

func TestValidOperatorInSteps(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	op := &drivenadapters.OperatorResponse{
		Status: PublishedStatus,
		ExtendInfo: drivenadapters.ExtendInfo{
			"dag_id": "dag_id",
		},
	}

	userInfo := &drivenadapters.UserInfo{}

	Convey("ValidOperatorInSteps", t, func() {
		Convey("Without Return", func() {
			stepList := []map[string]interface{}{
				{
					"id":         "0",
					"title":      "",
					"operator":   "@trigger/form",
					"parameters": map[string]interface{}{},
				},
			}
			_, err := mockMgnt.validOperatorInSteps(context.Background(), stepList, userInfo, "bd_public")
			assert.NotEqual(t, err, nil)
		})

		Convey("Valid Sucess", func() {
			stepList := []map[string]interface{}{
				{
					"id":         "0",
					"title":      "",
					"operator":   "@trigger/form",
					"parameters": map[string]interface{}{},
				},
				{
					"id":       "1",
					"title":    "",
					"operator": "@operator/operator-id",
					"parameters": map[string]interface{}{
						"version": "v",
					},
				},
				{
					"id":         "2",
					"title":      "",
					"operator":   "@internal/return",
					"parameters": map[string]interface{}{},
				},
			}
			dependency.op.EXPECT().LatestOperatorInfo(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(op, nil)
			vRes, err := mockMgnt.validOperatorInSteps(context.Background(), stepList, userInfo, "bd_public")
			assert.Equal(t, err, nil)
			assert.Equal(t, len(vRes.RefDagIDs), 1)
		})
	})
}

func TestAddOperatorAfter(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	Convey("AddOperatorAfter", t, func() {
		Convey("Create Or Update Operator BadParams", func() {
			errBytes, _ := json.Marshal(map[string]interface{}{"code": ""})
			err := mockMgnt.addOperatorAfter(context.Background(), []*drivenadapters.OperatorModifyResp{}, ierr.ExHTTPError{Body: string(errBytes), Status: 400})
			assert.NotEqual(t, err, nil)
		})

		Convey("Create Or Update Operator InterNal Err", func() {
			errBytes, _ := json.Marshal(map[string]interface{}{"code": "InternalError"})
			err := mockMgnt.addOperatorAfter(context.Background(), []*drivenadapters.OperatorModifyResp{}, ierr.ExHTTPError{Body: string(errBytes)})
			assert.NotEqual(t, err, nil)
		})

		Convey("Create Or Update Operator Name Conflict", func() {
			results := []*drivenadapters.OperatorModifyResp{
				{
					Status:     "failed",
					OperatorID: "id",
					Version:    "v",
					Error: map[string]interface{}{
						"code": OperatorErrorConflict,
					},
				},
			}
			err := mockMgnt.addOperatorAfter(context.Background(), results, nil)
			assert.NotEqual(t, err, nil)
		})

		Convey("Create Or Update Operator Err", func() {
			results := []*drivenadapters.OperatorModifyResp{
				{
					Status:     "failed",
					OperatorID: "id",
					Version:    "v",
					Error: map[string]interface{}{
						"code": "InternalError",
					},
				},
			}
			err := mockMgnt.addOperatorAfter(context.Background(), results, nil)
			assert.NotEqual(t, err, nil)
		})

		Convey("Create Or Update Operator Success", func() {
			results := []*drivenadapters.OperatorModifyResp{
				{
					Status:     RegisterSuccessStatus,
					OperatorID: "id",
					Version:    "v",
				},
			}
			err := mockMgnt.addOperatorAfter(context.Background(), results, nil)
			assert.Equal(t, err, nil)
		})
	})
}

func TestGetOperatorExecutionMode(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	Convey("GetOperatorExecutionMode", t, func() {
		Convey("Should return async mode for python operator", func() {
			stepList := []map[string]interface{}{
				{"operator": common.InternalToolPy3Opt, "parameters": map[string]interface{}{}},
			}
			mode, err := mockMgnt.getOperatorExecutionMode(nil, stepList)
			assert.Equal(t, err, nil)
			assert.Equal(t, mode, common.ExecutionModeAsync)
		})

		Convey("Should return async mode for approval operator", func() {
			stepList := []map[string]interface{}{
				{"operator": common.WorkflowApproval, "parameters": map[string]interface{}{}},
			}
			mode, err := mockMgnt.getOperatorExecutionMode(nil, stepList)
			assert.Equal(t, err, nil)
			assert.Equal(t, mode, common.ExecutionModeAsync)
		})

		Convey("Should return async mode for custom operator", func() {
			stepList := []map[string]interface{}{
				{"operator": common.CustomOperatorPrefix + "test", "parameters": map[string]interface{}{}},
			}
			mode, err := mockMgnt.getOperatorExecutionMode(nil, stepList)
			assert.Equal(t, err, nil)
			assert.Equal(t, mode, common.ExecutionModeAsync)
		})

		Convey("Should return async mode when any operator in map is async", func() {
			opInfoMap := map[string]*OperatorInfo{
				"op1": {Status: PublishedStatus, OperatorInfo: drivenadapters.OperatorInfo{ExecutionMode: common.ExecutionModeSync}},
				"op2": {Status: PublishedStatus, OperatorInfo: drivenadapters.OperatorInfo{ExecutionMode: common.ExecutionModeAsync}},
			}
			stepList := []map[string]interface{}{
				{"operator": "other_operator", "parameters": map[string]interface{}{}},
			}
			mode, err := mockMgnt.getOperatorExecutionMode(opInfoMap, stepList)
			assert.Equal(t, err, nil)
			assert.Equal(t, mode, common.ExecutionModeAsync)
		})

		Convey("Should return sync mode when no async operators found", func() {
			opInfoMap := map[string]*OperatorInfo{
				"op1": {Status: PublishedStatus, OperatorInfo: drivenadapters.OperatorInfo{ExecutionMode: common.ExecutionModeSync}},
				"op2": {Status: PublishedStatus, OperatorInfo: drivenadapters.OperatorInfo{ExecutionMode: common.ExecutionModeSync}},
			}
			stepList := []map[string]interface{}{
				{"operator": "other_operator", "parameters": map[string]interface{}{}},
			}
			mode, err := mockMgnt.getOperatorExecutionMode(opInfoMap, stepList)
			assert.Equal(t, err, nil)
			assert.Equal(t, mode, common.ExecutionModeSync)
		})

		Convey("Should prioritize step list operators over operator map", func() {
			opInfoMap := map[string]*OperatorInfo{
				"op1": {Status: PublishedStatus, OperatorInfo: drivenadapters.OperatorInfo{ExecutionMode: common.ExecutionModeSync}},
			}
			stepList := []map[string]interface{}{
				{"operator": common.InternalToolPy3Opt, "parameters": map[string]interface{}{"mode": "async"}},
			}
			mode, err := mockMgnt.getOperatorExecutionMode(opInfoMap, stepList)
			assert.Equal(t, err, nil)
			assert.Equal(t, mode, common.ExecutionModeAsync)
		})
	})
}

func TestExportOperator(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	dags := []*entity.Dag{
		{
			BaseInfo: entity.BaseInfo{
				ID: "583491894695543018",
			},
			UserID: "4fa5fafe-e751-11ef-b014-dac047ec7bab",
			Name:   "test",
			Steps: []entity.Step{
				{
					ID:       "0",
					Title:    "",
					Operator: "@trigger/form",
					Parameters: map[string]interface{}{
						"fields": []interface{}{
							map[string]interface{}{
								"key":  "doc_id",
								"type": "string",
								"name": "abc",
							},
						},
					},
				},
				{
					ID:         "1",
					Operator:   "@operator/0a2d4bf8-6386-4b62-bedb-2a0b0ac973ab",
					DataSource: &entity.DataSource{},
					Parameters: map[string]interface{}{},
					Cron:       "",
					Branches:   []entity.Branch{},
					Steps:      []entity.Step{},
				},
			},
			Description:   "",
			Type:          common.DagTypeComboOperator,
			TriggerConfig: &entity.TriggerConfig{},
			Category:      "data_split",
			OutPuts: []*entity.OutPut{
				{
					Key:  "key",
					Name: "abc",
					Type: "string",
				},
			},
			OperatorID: "eb853234-9722-4bf8-8f16-c596a8c38e14",
		},
	}

	Convey("TestExportOperator", t, func() {
		Convey("GetDag Error - Not Found", func() {
			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Return(nil, mongo.ErrNoDocuments)
			_, err := mockMgnt.ExportOperator(context.Background(), []string{"583491894695543018"})
			assert.NotEqual(t, err, nil)
		})

		Convey("GetDag Error - Internal Error", func() {
			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("ListDag Error"))
			_, err := mockMgnt.ExportOperator(context.Background(), []string{"583491894695543018"})
			assert.NotEqual(t, err, nil)
		})

		Convey("ID Empty", func() {
			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Return(dags, nil)
			res, err := mockMgnt.ExportOperator(context.Background(), []string{})
			assert.Equal(t, err, nil)
			assert.Equal(t, len(res.Configs), 0)
		})

		Convey("Dag Not Found", func() {
			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Return(dags, nil)
			_, err := mockMgnt.ExportOperator(context.Background(), []string{"583491894695543018", "583491894695543019"})
			assert.NotEqual(t, err, nil)
		})

		Convey("Export Operator Success - No Refer Combo Operator", func() {
			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Return(dags, nil)
			res, err := mockMgnt.ExportOperator(context.Background(), []string{"583491894695543018"})
			assert.Equal(t, err, nil)
			assert.Equal(t, len(res.Configs), 1)
		})

		Convey("Export Operator Success - Refer Combo Operator", func() {
			dags[0].SubIDs = []string{"583489305725415658", "579986223192387818"}
			subDags := []*entity.Dag{
				{
					BaseInfo: entity.BaseInfo{
						ID: "583489305725415658",
					},
					UserID: "4fa5fafe-e751-11ef-b014-dac047ec7bab",
					Name:   "test1",
					Steps: []entity.Step{
						{
							ID:       "0",
							Title:    "",
							Operator: "@trigger/form",
							Parameters: map[string]interface{}{
								"fields": []interface{}{
									map[string]interface{}{
										"key":  "doc_id",
										"type": "string",
										"name": "abc",
									},
								},
							},
						},
					},
					Description:   "",
					Type:          common.DagTypeComboOperator,
					TriggerConfig: &entity.TriggerConfig{},
					Category:      "data_split",
					OutPuts: []*entity.OutPut{
						{
							Key:  "key",
							Name: "abc",
							Type: "string",
						},
					},
					OperatorID: "0a2d4bf8-6386-4b62-bedb-2a0b0ac973ab",
				},
				{
					BaseInfo: entity.BaseInfo{
						ID: "579986223192387818",
					},
					UserID: "4fa5fafe-e751-11ef-b014-dac047ec7bab",
					Name:   "test2",
					Steps: []entity.Step{
						{
							ID:       "0",
							Title:    "",
							Operator: "@trigger/form",
							Parameters: map[string]interface{}{
								"fields": []interface{}{
									map[string]interface{}{
										"key":  "doc_id",
										"type": "string",
										"name": "abc",
									},
								},
							},
						},
					},
					Description:   "",
					Type:          common.DagTypeComboOperator,
					TriggerConfig: &entity.TriggerConfig{},
					Category:      "data_split",
					OutPuts: []*entity.OutPut{
						{
							Key:  "key",
							Name: "abc",
							Type: "string",
						},
					},
					OperatorID: "0a2d4bf8-6386-4b62-bedb-2a0b0ac973ao",
				},
			}
			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Times(1).Return(dags, nil)
			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Times(1).Return(subDags, nil)
			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Return([]*entity.Dag{}, nil)
			res, err := mockMgnt.ExportOperator(context.Background(), []string{"583491894695543018"})
			assert.Equal(t, err, nil)
			assert.Equal(t, len(res.Configs), 3)
		})
	})
}

func TestImportOperator(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	Convey("TestImportOperator", t, func() {
		Convey("List Dag Error", func() {
			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Times(1).Return(nil, fmt.Errorf("List Dag Err"))
			err := mockMgnt.ImportOperator(context.Background(), &ImportOperatorReq{}, &drivenadapters.UserInfo{})
			assert.NotEqual(t, err, nil)
		})

		Convey("Dag Conflict", func() {
			dag := &entity.Dag{
				BaseInfo: entity.BaseInfo{
					ID: "583491894695543018",
				},
			}

			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Return([]*entity.Dag{dag}, nil)
			err := mockMgnt.ImportOperator(context.Background(), &ImportOperatorReq{}, &drivenadapters.UserInfo{})
			assert.Equal(t, ierr.Is(err, ierr.PublicErrorType, ierr.PErrorConflict), true)
		})

		Convey("Import Operator Success", func() {
			params := &ImportOperatorReq{
				Mode: "upsert",
				Configs: []OperatorImportExportItem{
					{
						ID:          "583490894695543018",
						Title:       "test",
						Description: "test",
						Steps: []entity.Step{
							{
								ID:       "0",
								Title:    "",
								Operator: "@trigger/form",
								Parameters: map[string]interface{}{
									"fields": []interface{}{
										map[string]interface{}{
											"key":  "doc_id",
											"type": "string",
											"name": "abc",
										},
									},
								},
							},
						},
						Category: "default",
						OutPuts: []*entity.OutPut{
							{
								Key:  "key",
								Name: "abc",
								Type: "string",
							},
						},
						OperatorID: "eb853234-9722-4bf8-8f16-c596a8c38e14",
						IsRoot:     true,
					},
					{
						ID:          "583490894695543019",
						Title:       "test1",
						Description: "test1",
						Steps: []entity.Step{
							{
								ID:       "0",
								Title:    "",
								Operator: "@trigger/form",
								Parameters: map[string]interface{}{
									"fields": []interface{}{
										map[string]interface{}{
											"key":  "doc_id",
											"type": "string",
											"name": "abc",
										},
									},
								},
							},
						},
						Category: "default",
						OutPuts: []*entity.OutPut{
							{
								Key:  "key",
								Name: "abc",
								Type: "string",
							},
						},
						OperatorID: "eb853234-9722-4bf8-8f16-c596a8c38e15",
					},
				},
			}

			dags := []*entity.Dag{
				{
					BaseInfo: entity.BaseInfo{
						ID: "583490894695543018",
					},
				},
				{
					BaseInfo: entity.BaseInfo{
						ID: "583490894695543019",
					},
				},
			}

			dependency.mongo.EXPECT().ListDag(gomock.Any(), gomock.Any()).Return(dags, nil)
			patch := ApplyMethod(reflect.TypeOf(mockMgnt), "CheckAndBuildDag", func(*mgnt, context.Context, *OptionalComboOperatorReq, *entity.Dag, *drivenadapters.UserInfo) error {
				return nil
			})
			defer patch.Reset()
			dependency.mongo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).Return(nil)
			dependency.mongo.EXPECT().DeleteDag(gomock.Any(), gomock.Any()).Return(nil)
			dependency.mongo.EXPECT().BatchCreateDag(gomock.Any(), gomock.Any()).Return([]*entity.Dag{}, nil)
			err := mockMgnt.ImportOperator(context.Background(), params, &drivenadapters.UserInfo{})
			assert.Equal(t, err, nil)
		})

	})
}

func TestDeleteComboOperator(t *testing.T) {
	dependency := NewDependency(t)
	mockMgnt := NewMgntInstance(dependency)

	Convey("TestDeleteComboOperator", t, func() {
		Convey("Delete Operator Success", func() {
			dags := []*entity.Dag{
				{
					BaseInfo: entity.BaseInfo{
						ID: "583490894695543018",
					},
				},
			}
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Return(dags, nil)
			dependency.mongo.EXPECT().DeleteDag(gomock.Any(), gomock.Any()).Return(nil)
			err := mockMgnt.DeleteComboOperator(context.Background(), "583490894695543018")
			assert.Equal(t, err, nil)
		})

		Convey("List Dag Error", func() {
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("err"))
			err := mockMgnt.DeleteComboOperator(context.Background(), "583490894695543018")
			assert.NotEqual(t, err, nil)
		})

		Convey("Delete Operator Error", func() {
			dags := []*entity.Dag{
				{
					BaseInfo: entity.BaseInfo{
						ID: "583490894695543018",
					},
				},
			}
			dependency.mongo.EXPECT().ListDagByFields(gomock.Any(), gomock.Any(), gomock.Any()).Return(dags, nil)
			dependency.mongo.EXPECT().DeleteDag(gomock.Any(), gomock.Any()).Return(fmt.Errorf("err"))
			err := mockMgnt.DeleteComboOperator(context.Background(), "583490894695543018")
			assert.NotEqual(t, err, nil)
		})
	})
}
