package perm

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_drivenadapters"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

type MockPolicyDependency struct {
	auth *mock_drivenadapters.MockAuthorizationDriven
}

func NewPolicyDependency(t *testing.T) MockPolicyDependency {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return MockPolicyDependency{
		auth: mock_drivenadapters.NewMockAuthorizationDriven(ctrl),
	}
}

func NewMockPermPolicy(dep MockPolicyDependency) *permPolicy {
	initARLog()
	initErrorInfo()
	return &permPolicy{
		auth: dep.auth,
		publish: func(topic string, message []byte) error {
			return nil
		},
	}
}

func TestResourceList(t *testing.T) {
	resources := ResourceList{
		"*",
		"582617262561715743:data-flow",
		"582617262561715743:combo-operator",
	}
	rl := resources.ToMap("data-flow")
	assert.Equal(t, len(rl), 1)
}

func TestIsDataAdmin(t *testing.T) {
	dependency := NewPolicyDependency(t)
	permPolicy := NewMockPermPolicy(dependency)

	Convey("IsDataAdmin", t, func() {
		Convey("Check DataFlow Permission Error", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("permission check error"))
			_, err := permPolicy.IsDataAdmin(context.Background(), "", "")
			assert.NotEqual(t, err, nil)
		})

		Convey("Check O11y Permission Success", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("permission check error"))
			_, err := permPolicy.IsDataAdmin(context.Background(), "", "")
			assert.NotEqual(t, err, nil)
		})

		Convey("isDataAdmin", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			isDataAdmin, err := permPolicy.IsDataAdmin(context.Background(), "", "")
			assert.Equal(t, err, nil)
			assert.Equal(t, isDataAdmin, true)
		})
	})
}

func TestCheckPerm(t *testing.T) {
	dependency := NewPolicyDependency(t)
	permPolicy := NewMockPermPolicy(dependency)

	filterRes := []drivenadapters.Resource{
		{
			ID: "582617262561715743:data-flow",
		},
		{
			ID: "582617262561715745:data-flow",
		},
	}

	Convey("CheckPerm", t, func() {
		Convey("Check IsDataAdmin Error", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("permission check error"))
			_, err := permPolicy.CheckPerm(context.Background(), "", "", []string{"resource_id"})
			assert.NotEqual(t, err, nil)
		})

		Convey("IsDataAdmin", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			hasPerm, err := permPolicy.CheckPerm(context.Background(), "", "", []string{"resource_id"})
			assert.Equal(t, err, nil)
			assert.Equal(t, hasPerm, true)
		})

		Convey("Is Not DataAdmin, Check Resource Filter Error", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, nil)
			dependency.auth.EXPECT().ResourceFilter(gomock.Any(), gomock.Any()).Return([]drivenadapters.Resource{}, fmt.Errorf("resource filter error"))
			_, err := permPolicy.CheckPerm(context.Background(), "", "", []string{"resource_id"})
			assert.NotEqual(t, err, nil)
		})

		Convey("Is Not DataAdmin, And No Permssion", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, nil)
			dependency.auth.EXPECT().ResourceFilter(gomock.Any(), gomock.Any()).Return(filterRes, nil)
			_, err := permPolicy.CheckPerm(context.Background(), "", "", []string{"resource_id"})
			assert.Equal(t, ierr.Is(err, ierr.PublicErrorType, ierr.PErrorForbidden), true)
		})

		Convey("Is Not DataAdmin, And Has Permssion", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, nil)
			dependency.auth.EXPECT().ResourceFilter(gomock.Any(), gomock.Any()).Return(filterRes, nil)
			hasPerm, err := permPolicy.CheckPerm(context.Background(), "", "", []string{"582617262561715743:data-flow", "582617262561715745:data-flow"})
			assert.Equal(t, err, nil)
			assert.Equal(t, hasPerm, true)
		})
	})
}

func TestOperationCheckWithResType(t *testing.T) {
	dependency := NewPolicyDependency(t)
	permPolicy := NewMockPermPolicy(dependency)

	Convey("OperationCheckWithResType", t, func() {
		Convey("OperationPermCheck Error", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("permission check error"))
			err := permPolicy.OperationCheckWithResType(context.Background(), "accessorID", "accessorType", "resourceID", "resourceType", "view")
			assert.NotEqual(t, err, nil)
		})

		Convey("No Permission", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, nil)
			err := permPolicy.OperationCheckWithResType(context.Background(), "accessorID", "accessorType", "resourceID", "resourceType", "view")
			assert.Equal(t, ierr.Is(err, ierr.PublicErrorType, ierr.PErrorForbidden), true)
		})

		Convey("Has Permission", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			err := permPolicy.OperationCheckWithResType(context.Background(), "accessorID", "accessorType", "resourceID", "resourceType", "view")
			assert.Equal(t, err, nil)
		})
	})
}

func TestCreatePolicy(t *testing.T) {
	dependency := NewPolicyDependency(t)
	permPolicy := NewMockPermPolicy(dependency)

	Convey("CreatePolicy", t, func() {
		Convey("CreatePolicy Error", func() {
			dependency.auth.EXPECT().CreatePolicy(gomock.Any(), gomock.Any()).Return(fmt.Errorf("create policy error"))
			err := permPolicy.CreatePolicy(context.Background(), "userID", "userType", "userName", "resourceID", "resourceName", []string{}, []string{})
			assert.NotEqual(t, err, nil)
		})
		Convey("CreatePolicy Success", func() {
			dependency.auth.EXPECT().CreatePolicy(gomock.Any(), gomock.Any()).Return(nil)
			err := permPolicy.CreatePolicy(context.Background(), "userID", "userType", "userName", "resourceID", "resourceName", []string{"view"}, []string{"display"})
			assert.Equal(t, err, nil)
		})
	})
}

func TestUpdatePolicy(t *testing.T) {
	dependency := NewPolicyDependency(t)
	permPolicy := NewMockPermPolicy(dependency)

	Convey("UpdatePolicy", t, func() {
		Convey("UpdatePolicy Error", func() {
			dependency.auth.EXPECT().UpdatePolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("update policy error"))
			err := permPolicy.UpdatePolicy(context.Background(), []string{}, []string{"view"}, []string{"display"})
			assert.NotEqual(t, err, nil)
		})
		Convey("UpdatePolicy Success", func() {
			dependency.auth.EXPECT().UpdatePolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			err := permPolicy.UpdatePolicy(context.Background(), []string{}, []string{"view"}, []string{"display"})
			assert.Equal(t, err, nil)
		})
	})
}

func TestDeletePolicy(t *testing.T) {
	dependency := NewPolicyDependency(t)
	permPolicy := NewMockPermPolicy(dependency)

	Convey("DeletePolicy", t, func() {
		Convey("DeletePolicy Error", func() {
			dependency.auth.EXPECT().DeletePolicy(gomock.Any(), gomock.Any()).Return(fmt.Errorf("delete policy error"))
			err := permPolicy.DeletePolicy(context.Background(), "id")
			assert.NotEqual(t, err, nil)
		})
		Convey("DeletePolicy Success", func() {
			dependency.auth.EXPECT().DeletePolicy(gomock.Any(), gomock.Any()).Return(nil)
			err := permPolicy.DeletePolicy(context.Background(), "id")
			assert.Equal(t, err, nil)
		})
	})
}

func TestMinPermList(t *testing.T) {
	dependency := NewPolicyDependency(t)
	permPolicy := NewMockPermPolicy(dependency)

	resIDs := []string{"r1:data_flow", "r2:data_flow"}

	Convey("MinPermList", t, func() {
		Convey("IsDataAdmin error", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("permission check error"))
			_, err := permPolicy.MinPermList(context.Background(), "u1", "user", resIDs)
			assert.NotEqual(t, err, nil)
		})

		Convey("IsDataAdmin true returns all perms", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)

			perms, err := permPolicy.MinPermList(context.Background(), "u1", "user", resIDs)
			assert.Equal(t, err, nil)
			assert.Equal(t, len(perms) >= 1, true)
		})

		Convey("Non-admin, ListResourceOperation error", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, nil)
			dependency.auth.EXPECT().ListResourceOperation(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("ListResourceOperation error"))

			_, err := permPolicy.MinPermList(context.Background(), "u1", "user", resIDs)
			assert.Equal(t, ierr.Is(err, ierr.PublicErrorType, ierr.PErrorInternalServerError), true)
		})

		Convey("Non-admin, empty result -> empty perms", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, nil)
			dependency.auth.EXPECT().ListResourceOperation(gomock.Any(), gomock.Any()).Return([]drivenadapters.ListResourceOperationRes{}, nil)

			perms, err := permPolicy.MinPermList(context.Background(), "u1", "user", resIDs)
			assert.Equal(t, err, nil)
			assert.Equal(t, len(perms), 0)
		})

		Convey("Non-admin, intersection of operations", func() {
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(true, nil)
			dependency.auth.EXPECT().OperationPermCheck(gomock.Any(), gomock.Any()).Return(false, nil)
			dependency.auth.EXPECT().ListResourceOperation(gomock.Any(), gomock.Any()).Return([]drivenadapters.ListResourceOperationRes{
				{ID: "r1:data_flow", Operation: []string{"view", "modify", "list"}},
				{ID: "r2:data_flow", Operation: []string{"view", "list"}},
			}, nil)

			perms, err := permPolicy.MinPermList(context.Background(), "u1", "user", resIDs)
			assert.Equal(t, err, nil)
			assert.Equal(t, true, (len(perms) == 2 && ((perms[0] == "view" && perms[1] == "list") || (perms[0] == "list" && perms[1] == "view"))))
		})
	})
}

func TestListResource(t *testing.T) {
	dependency := NewPolicyDependency(t)
	permPolicy := NewMockPermPolicy(dependency)

	Convey("ListResource", t, func() {
		Convey("ListResource error", func() {
			dependency.auth.EXPECT().ListResource(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("list resource error"))
			_, err := permPolicy.ListResource(context.Background(), "u1", "data_flow", "view")
			assert.Equal(t, ierr.Is(err, ierr.PublicErrorType, ierr.PErrorInternalServerError), true)
		})

		Convey("ListResource success returns ids", func() {
			dependency.auth.EXPECT().ListResource(gomock.Any(), gomock.Any()).Return([]drivenadapters.Resource{
				{ID: "a:data_flow"},
				{ID: "b:data_flow"},
			}, nil)
			ids, err := permPolicy.ListResource(context.Background(), "u1", "data_flow", "view")
			assert.Equal(t, err, nil)
			assert.Equal(t, len(*ids), 2)
		})
	})
}

func TestHandlePolicyNameChange_DirectFuncMock(t *testing.T) {
	dependency := NewPolicyDependency(t)
	permPolicy := NewMockPermPolicy(dependency)
	Convey("HandlePolicyNameChange direct func mock", t, func() {
		id := "rid-1"
		name := "rname"
		rType := "data_flow"

		permPolicy.HandlePolicyNameChange(id, name, rType)
	})
}
