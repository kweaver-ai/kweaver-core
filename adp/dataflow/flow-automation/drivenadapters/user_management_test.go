package drivenadapters

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func InitLogger() {
	logout := os.Getenv("LOGOUT")
	logDir := "/var/log/contentAutoMation/"
	logName := "contentAutoMation.log"
	commonLog.InitLogger(logout, logDir, logName)
}

func NewMockUserManagement(clients *HttpClientMock) UserManagement {
	InitLogger()
	return &userManagement{
		adminAddress: "http://localhost:8080",
		log:          commonLog.NewLogger(),
		httpClient:   clients.httpClient1,
	}
}

func TestGetGroupUserList(t *testing.T) {
	httpClients := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClients)

	Convey("TestGetGroupUserList", t, func() {
		Convey("Get Group List Error", func() {
			httpClients.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := userManagement.GetGroupUserList([]string{"group_id"})
			assert.NotEqual(t, err, nil)
		})

		Convey("Get Group List Success", func() {
			mockResp := map[string]interface{}{
				"user_ids": []interface{}{"1", "2"},
			}
			httpClients.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			users, err := userManagement.GetGroupUserList([]string{"group_id"})
			assert.Equal(t, err, nil)
			assert.Equal(t, len(users), 2)
		})
	})
}

func TestGetUserMailList(t *testing.T) {
	httpClients := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClients)

	Convey("TestGetUserMailList", t, func() {
		Convey("Get User Mail List Error", func() {
			httpClients.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			_, err := userManagement.GetUserMailList([]string{"user_id"})
			assert.NotEqual(t, err, nil)
		})
		Convey("Get User Mail List Success", func() {
			mockResp := map[string]interface{}{
				"user_emails": []interface{}{
					map[string]interface{}{
						"email": "zhangsan@aishu.cn",
					},
					map[string]interface{}{
						"email": "lisi@aishu.cn",
					},
				},
			}
			httpClients.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			emails, err := userManagement.GetUserMailList([]string{"user_id1", "user_id2"})
			assert.Equal(t, err, nil)
			assert.Equal(t, len(emails), 2)
		})
	})
}

func TestGetDeptMailList(t *testing.T) {
	httpClients := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClients)

	Convey("TestGetDeptMailList", t, func() {
		Convey("Get Dept Mail List Error", func() {
			httpClients.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			_, err := userManagement.GetDeptMailList([]string{"dept_id"})
			assert.NotEqual(t, err, nil)
		})
		Convey("Get Dept Mail List Success", func() {
			mockResp := map[string]interface{}{
				"department_emails": []interface{}{
					map[string]interface{}{
						"email": "dept1@aishu.cn",
					},
					map[string]interface{}{
						"email": "dept2@aishu.cn",
					},
				},
			}
			httpClients.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			emails, err := userManagement.GetDeptMailList([]string{"dept_id1", "dept_id2"})
			assert.Equal(t, err, nil)
			assert.Equal(t, len(emails), 2)
		})
	})
}

func TestBatchGetUserInfo(t *testing.T) {
	httpClients := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClients)

	Convey("TestBatchGetUserInfo", t, func() {
		Convey("Batch Get User Info Error", func() {
			httpClients.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := userManagement.BatchGetUserInfo([]string{"user_id"})
			assert.NotEqual(t, err, nil)
		})
		Convey("Batch Get User Info Success", func() {
			mockResp := map[string]interface{}{
				"user_names": []interface{}{
					map[string]interface{}{
						"id":   "user_id1",
						"name": "zhangsan",
					},
					map[string]interface{}{
						"id":   "user_id2",
						"name": "lisi",
					},
				},
			}
			httpClients.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			users, err := userManagement.BatchGetUserInfo([]string{"user_id1", "user_id2"})
			assert.Equal(t, err, nil)
			assert.Equal(t, len(users), 2)
		})
	})
}

func TestGetUserInfo(t *testing.T) {
	httpClients := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClients)

	Convey("TestGetUserInfo", t, func() {
		Convey("Get User Info Error", func() {
			httpClients.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			_, err := userManagement.GetUserInfo("user_id")
			assert.NotEqual(t, err, nil)
		})
		Convey("Get User Info Success", func() {
			mockResp := []interface{}{
				map[string]interface{}{
					"id":        "user_id",
					"name":      "zhangsan",
					"csf_level": float64(5),
					"telephone": "12345678901",
					"email":     "zhangsan@aishu.cn",
					"enabled":   true,
					"parent_deps": []interface{}{
						[]interface{}{
							map[string]interface{}{
								"id":   "dept_id",
								"name": "dept_name",
							},
						},
					},
					"roles": []interface{}{
						"role1",
					},
					"custom_attr": map[string]interface{}{
						"is_knowledge": "1",
					},
				},
			}
			httpClients.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			user, err := userManagement.GetUserInfo("user_id")
			assert.Equal(t, err, nil)
			assert.Equal(t, user.UserName, "zhangsan")
		})
	})
}

func TestGetUserAccessorIDs(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestGetUserAccessorIDs", t, func() {
		Convey("Get User Accessor IDs Error", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			_, err := userManagement.GetUserAccessorIDs("user_id")
			assert.NotEqual(t, err, nil)
		})
		Convey("Get User Accessor IDs Success", func() {
			mockResp := []interface{}{
				"accessor_id1",
				"accessor_id2",
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			accessorIDs, err := userManagement.GetUserAccessorIDs("user_id")
			assert.Equal(t, err, nil)
			assert.Equal(t, len(accessorIDs), 2)
		})
	})

}

func TestGetNameByAccessorIDs(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	accessorIDs := map[string]string{
		"accessor_id1": common.User.ToString(),
		"accessor_id2": common.Group.ToString(),
		"accessor_id3": common.Contactor.ToString(),
		"accessor_id4": common.Department.ToString(),
		"accessor_id5": common.User.ToString(),
	}

	Convey("TestGetNameByAccessorIDs", t, func() {
		Convey("Get Name By Accessor IDs Error", func() {
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := userManagement.GetNameByAccessorIDs(accessorIDs)
			assert.NotEqual(t, err, nil)
		})

		Convey("Get Name By Accessor IDs Sucess", func() {
			body := `{"code": 400019001, "detail": {"ids": ["accessor_id5"]}}`
			NotFoundEror := errors.ExHTTPError{
				Body:   body,
				Status: 400,
				Err:    nil,
			}
			mockResp := map[string]interface{}{
				"user_names": []interface{}{
					map[string]interface{}{
						"id":   "accessor_id1",
						"name": "zhangsan",
					},
				},
				"group_names": []interface{}{
					map[string]interface{}{
						"id":   "accessor_id2",
						"name": "lisi",
					},
				},
				"contactor_names": []interface{}{
					map[string]interface{}{
						"id":   "accessor_id3",
						"name": "wangwu",
					},
				},
				"department_names": []interface{}{
					map[string]interface{}{
						"id":   "accessor_id4",
						"name": "zhaoliu",
					},
				},
			}
			mockRes := map[string]string{
				"accessor_id1": "zhangsan",
				"accessor_id2": "lisi",
				"accessor_id3": "wangwu",
				"accessor_id4": "zhaoliu",
			}
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(400, nil, NotFoundEror)
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			names, err := userManagement.GetNameByAccessorIDs(accessorIDs)
			assert.Equal(t, err, nil)
			assert.Equal(t, names, mockRes)
		})
	})
}

func TestRegisterInternalAccount(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestRegisterInternalAccount", t, func() {
		Convey("Register Internal Account Error", func() {
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := userManagement.RegisterInternalAccount("name", "pwd")
			assert.NotEqual(t, err, nil)
		})

		Convey("Register Internal Account Duplicate", func() {
			body := `{"code": 409000000, "detail": {"id": "id"}}`
			NotFoundEror := errors.ExHTTPError{
				Body:   body,
				Status: 400,
				Err:    nil,
			}
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(409, nil, NotFoundEror)
			id, err := userManagement.RegisterInternalAccount("name", "pwd")
			assert.Equal(t, err, nil)
			assert.Equal(t, id, "id")
		})

		Convey("Register Internal Account Success", func() {
			mockResp := map[string]interface{}{
				"id": "id",
			}
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			id, err := userManagement.RegisterInternalAccount("name", "pwd")
			assert.Equal(t, err, nil)
			assert.Equal(t, id, "id")
		})
	})
}

func TestQueryInternalAccount(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestQueryInternalAccount", t, func() {
		Convey("Query Internal Account Error", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			_, err := userManagement.QueryInternalAccount("id")
			assert.NotEqual(t, err, nil)
		})
		Convey("Query Internal Account Not Found", func() {
			body := `{"code": 404000000, "detail": {"id": "id"}}`
			NotFoundEror := errors.ExHTTPError{
				Body:   body,
				Status: 404,
				Err:    nil,
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, NotFoundEror)
			name, err := userManagement.QueryInternalAccount("id")
			assert.Equal(t, err, nil)
			assert.Equal(t, name, "")
		})
		Convey("Query Internal Account Success", func() {
			mockResp := map[string]interface{}{
				"name": "name",
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			name, err := userManagement.QueryInternalAccount("id")
			assert.Equal(t, err, nil)
			assert.Equal(t, name, "name")
		})
	})
}

func TestCreateInternalGroup(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestCreateInternalGroup", t, func() {
		Convey("Create Internal Group Error", func() {
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := userManagement.CreateInternalGroup()
			assert.NotEqual(t, err, nil)
		})

		Convey("Create Internal Group Success", func() {
			mockResp := map[string]interface{}{
				"id": "id",
			}
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			id, err := userManagement.CreateInternalGroup()
			assert.Equal(t, err, nil)
			assert.Equal(t, id, "id")
		})
	})
}

func TestDeleteInternalGroup(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestDeleteInternalGroup", t, func() {
		Convey("Delete Internal Group Error", func() {
			httpClient.httpClient1.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			err := userManagement.DeleteInternalGroup([]string{"id"})
			assert.NotEqual(t, err, nil)
		})
		Convey("Delete Internal Group Success", func() {
			httpClient.httpClient1.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil, nil)
			err := userManagement.DeleteInternalGroup([]string{"id"})
			assert.Equal(t, err, nil)
		})
	})
}

func TestUpdateInternalGroupMember(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestUpdateInternalGroupMember", t, func() {
		Convey("Update Internal Group Member Error", func() {
			httpClient.httpClient1.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			err := userManagement.UpdateInternalGroupMember("id", []string{"id"})
			assert.NotEqual(t, err, nil)
		})
		Convey("Update Internal Group Member Success", func() {
			httpClient.httpClient1.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil, nil)
			err := userManagement.UpdateInternalGroupMember("id", []string{"id"})
			assert.Equal(t, err, nil)
		})
	})
}

func TestGetInternalGroupMembers(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestGetInternalGroupMembers", t, func() {
		Convey("Get Internal Group Members Error", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			_, err := userManagement.GetInternalGroupMembers("id")
			assert.NotEqual(t, err, nil)
		})
		Convey("Get Internal Group Members Success", func() {
			mockResp := []interface{}{
				map[string]interface{}{
					"id": "id",
				},
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			users, err := userManagement.GetInternalGroupMembers("id")
			assert.Equal(t, err, nil)
			assert.Equal(t, users, []string{"id"})
		})
	})
}

func TestGetDepartmentInfo(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestGetDepartmentInfo", t, func() {
		Convey("Get Department Info Error", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			_, err := userManagement.GetDepartmentInfo("id")
			assert.NotEqual(t, err, nil)
		})
		Convey("Get Department Info Success", func() {
			mockResp := []interface{}{
				map[string]interface{}{
					"department_id": "id",
					"name":          "name",
					"parent_deps":   []interface{}{},
				},
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			dept, err := userManagement.GetDepartmentInfo("id")
			assert.Equal(t, err, nil)
			assert.Equal(t, dept.Name, "name")
			assert.Equal(t, dept.DepartmentID, "id")
		})
	})
}

func TestGetDepartments(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestGetDepartments", t, func() {
		Convey("Get Departments Error", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			_, err := userManagement.GetDepartments(1)
			assert.NotEqual(t, err, nil)
		})
		Convey("Get Departments Success", func() {
			mockResp := []interface{}{
				map[string]interface{}{
					"id":   "id",
					"name": "name",
					"type": "type",
				},
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			depts, err := userManagement.GetDepartments(1)
			assert.Equal(t, err, nil)
			assert.Equal(t, len(*depts), 1)
		})
	})

}

func TestGetDepartmentMemberIDs(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	Convey("TestGetDepartmentMemberIDs", t, func() {
		Convey("Get Department Member IDs Error", func() {
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
			_, err := userManagement.GetDepartmentMemberIDs("id")
			assert.NotEqual(t, err, nil)
		})
		Convey("Get Department Member IDs Success", func() {
			mockResp := map[string]interface{}{
				"user_ids":       []interface{}{"id1", "id2"},
				"department_ids": []interface{}{"dep1", "dep2"},
			}
			httpClient.httpClient1.EXPECT().Get(gomock.Any(), gomock.Any()).Return(mockResp, nil)
			depmembers, err := userManagement.GetDepartmentMemberIDs("id")
			assert.Equal(t, err, nil)
			assert.Equal(t, len(depmembers.UserIDs), 2)
			assert.Equal(t, len(depmembers.DepartmentIDs), 2)
		})
	})
}

func TestBatchGetNames(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	userManagement := NewMockUserManagement(httpClient)

	params := map[string][]string{
		"key1": {"val1"},
		"key2": {"val2"},
	}
	Convey("TestBatchGetNames", t, func() {
		Convey("Batch Get Names Error", func() {
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := userManagement.BatchGetNames(params)
			assert.NotEqual(t, err, nil)
		})
		Convey("Batch Get Names Success", func() {
			mockResp := map[string]interface{}{
				"user_names": []interface{}{
					map[string]string{
						"id":   "id1",
						"name": "name1",
					},
				},
				"group_names": []interface{}{
					map[string]string{
						"id":   "id2",
						"name": "name2",
					},
				},
				"department_names": []interface{}{
					map[string]string{
						"id":   "id3",
						"name": "name3",
					},
				},
				"contactor_names": []interface{}{
					map[string]string{
						"id":   "id4",
						"name": "name4",
					},
				},
				"app_names": []interface{}{
					map[string]string{
						"id":   "id5",
						"name": "name5",
					},
				},
			}
			httpClient.httpClient1.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			names, err := userManagement.BatchGetNames(params)
			assert.Equal(t, err, nil)
			assert.Equal(t, len(names.UserNames), 1)
			assert.Equal(t, len(names.GroupNames), 1)
			assert.Equal(t, len(names.DepartmentNames), 1)
			assert.Equal(t, len(names.ContactorNames), 1)
			assert.Equal(t, len(names.AppNames), 1)
		})
	})

}
