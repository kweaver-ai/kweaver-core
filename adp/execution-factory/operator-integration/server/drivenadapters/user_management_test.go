package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestGetUsersInfo(t *testing.T) {
	Convey("GetUsersInfo", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		logger := logger.DefaultLogger()
		httpClient := mocks.NewMockHTTPClient(ctrl)
		mockClient := &userManagementClient{
			baseURL: fmt.Sprintf("%s://%s:%d/api/user-management", "http",
				"127.0.0.1", 30980),
			logger:     logger,
			httpClient: httpClient,
		}

		Convey("Get请求报错", func() {
			httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(0, nil, fmt.Errorf("http get error"))
			_, err := mockClient.GetUsersInfo(context.Background(), []string{"test"}, []string{"name"})
			So(err, ShouldNotBeNil)
		})

		Convey("Get请求成功,返回结果错误,Unmarshal报错", func() {
			httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, 0, nil)
			_, err := mockClient.GetUsersInfo(context.Background(), []string{"test"}, []string{"name"})
			So(err, ShouldNotBeNil)
		})

		Convey("404错误-成功解析不存在的用户ID", func() {
			// 模拟404错误响应
			errorResponse := map[string]interface{}{
				"cause": "those users are not existing",
				"detail": map[string]interface{}{
					"ids": []interface{}{"user123", "user456"},
				},
				"code":    404019001,
				"message": "数据不存在",
			}
			httpError := &rest.ExHTTPError{
				HTTPCode: http.StatusNotFound,
				Body:     utils.ObjectToByte(errorResponse),
			}

			httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(http.StatusNotFound, nil, httpError)

			infos, err := mockClient.GetUsersInfo(context.Background(), []string{"user123", "user456"}, []string{"name"})
			So(err, ShouldBeError)
			So(len(infos), ShouldEqual, 2)
			So(infos[0].UserID, ShouldEqual, "user123")
			So(infos[0].DisplayName, ShouldEqual, interfaces.UnknownUser)
			So(infos[1].UserID, ShouldEqual, "user456")
			So(infos[1].DisplayName, ShouldEqual, interfaces.UnknownUser)
		})

		Convey("404错误-解析失败", func() {
			// 模拟无效的404错误响应
			invalidErrorResponse := "invalid json"
			httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(404, invalidErrorResponse, fmt.Errorf("not found"))

			_, err := mockClient.GetUsersInfo(context.Background(), []string{"test"}, []string{"name"})
			So(err, ShouldNotBeNil)
		})

		Convey("成功", func() {
			httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, []*interfaces.UserInfo{}, nil)
			_, err := mockClient.GetUsersInfo(context.Background(), []string{"test"}, []string{"name"})
			So(err, ShouldBeNil)
		})
	})
}

func TestNewUserManagementClient_WhenAuthDisabled_ReturnsNoop(t *testing.T) {
	Convey("NewUserManagementClient auth disabled returns noop implementation", t, func() {
		t.Setenv("AUTH_ENABLED", "false")
		syncOnce = sync.Once{}
		um = nil
		defer func() {
			syncOnce = sync.Once{}
			um = nil
		}()

		client := NewUserManagementClient()
		userMap, err := client.GetUsersName(context.Background(), []string{interfaces.SystemUser, "user1"})

		So(err, ShouldBeNil)
		So(userMap[interfaces.SystemUser], ShouldEqual, interfaces.SystemUser)
		So(userMap["user1"], ShouldEqual, "user1")

		userMap, err = client.GetUsersName(context.Background(), []string{""})
		So(err, ShouldBeNil)
		So(userMap[""], ShouldEqual, "")
	})
}

func TestParseNotFoundUserIDs(t *testing.T) {
	Convey("parseNotFoundUserIDs", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		logger := mocks.NewMockLogger(ctrl)
		logger.EXPECT().WithContext(gomock.Any()).Return(logger).AnyTimes()

		mockClient := &userManagementClient{
			logger: logger,
		}
		ctx := context.Background()

		Convey("正确解析404错误响应", func() {
			errorResponse := map[string]interface{}{
				"cause": "those users are not existing",
				"detail": map[string]interface{}{
					"ids": []interface{}{"user1", "user2", "user3"},
				},
				"code":    404019001,
				"message": "数据不存在",
			}

			userIDs, err := mockClient.parseNotFoundUserIDs(ctx, utils.ObjectToByte(errorResponse))
			So(err, ShouldBeNil)
			So(len(userIDs), ShouldEqual, 3)
			So(userIDs, ShouldContain, "user1")
			So(userIDs, ShouldContain, "user2")
			So(userIDs, ShouldContain, "user3")
		})

		Convey("解析空的IDs数组", func() {
			errorResponse := map[string]interface{}{
				"detail": map[string]interface{}{
					"ids": []interface{}{},
				},
			}

			userIDs, err := mockClient.parseNotFoundUserIDs(ctx, utils.ObjectToByte(errorResponse))
			So(err, ShouldBeNil)
			So(len(userIDs), ShouldEqual, 0)
		})

		Convey("JSON反序列化失败", func() {
			invalidResponse := make(chan int) // 无法序列化的类型

			logger.EXPECT().Warnf(gomock.Any()).Return()
			userIDs, err := mockClient.parseNotFoundUserIDs(ctx, utils.ObjectToByte(invalidResponse))
			So(err, ShouldNotBeNil)
			So(len(userIDs), ShouldEqual, 0)
		})
	})
}

func TestGetUsersName(t *testing.T) {
	Convey("GetUsersName", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		logger := logger.DefaultLogger()
		httpClient := mocks.NewMockHTTPClient(ctrl)

		mockClient := &userManagementClient{
			baseURL: fmt.Sprintf("%s://%s:%d/api/user-management", "http",
				"127.0.0.1", 30980),
			logger:     logger,
			httpClient: httpClient,
		}
		ctx := context.Background()

		Convey("处理SystemUser", func() {
			userMap, err := mockClient.GetUsersName(ctx, []string{interfaces.SystemUser})
			So(err, ShouldBeNil)
			So(userMap[interfaces.SystemUser], ShouldEqual, interfaces.SystemUser)
		})

		Convey("空用户列表", func() {
			userMap, err := mockClient.GetUsersName(ctx, []string{})
			So(err, ShouldBeNil)
			So(len(userMap), ShouldEqual, 0)
		})

		Convey("成功获取用户信息", func() {
			mockUsers := []*interfaces.UserInfo{
				{UserID: "user1", DisplayName: "张三"},
				{UserID: "user2", DisplayName: "李四"},
			}

			httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockUsers, nil)

			userMap, err := mockClient.GetUsersName(ctx, []string{"user1", "user2"})
			So(err, ShouldBeNil)
			So(len(userMap), ShouldEqual, 2)
			So(userMap["user1"], ShouldEqual, "张三")
			So(userMap["user2"], ShouldEqual, "李四")
		})

		Convey("处理404错误循环逻辑", func() {
			// 第二次调用返回剩余用户信息
			remainingUsers := []*interfaces.UserInfo{
				{UserID: "user2", DisplayName: "李四"},
			}

			// 模拟404错误响应包含不存在的用户
			errorResponse := map[string]interface{}{
				"detail": map[string]interface{}{
					"ids": []interface{}{"user1"},
				},
			}
			httpErr := &rest.ExHTTPError{
				HTTPCode: http.StatusNotFound,
				Body:     utils.ObjectToByte(errorResponse),
			}

			// 第一次调用返回404错误
			gomock.InOrder(
				httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(404, nil, httpErr),
				httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, remainingUsers, nil),
			)
			userMap, err := mockClient.GetUsersName(ctx, []string{"user1", "user2"})
			So(err, ShouldBeNil)
			So(len(userMap), ShouldEqual, 2)
			So(userMap["user1"], ShouldEqual, interfaces.UnknownUser)
			So(userMap["user2"], ShouldEqual, "李四")
		})

		Convey("GetUsersInfo返回其他错误", func() {
			httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("internal server error"))
			_, err := mockClient.GetUsersName(ctx, []string{"user1"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "internal server error")
		})

		Convey("混合场景-SystemUser和普通用户", func() {
			mockUsers := []*interfaces.UserInfo{
				{UserID: "user1", DisplayName: "张三"},
			}

			httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockUsers, nil)

			userMap, err := mockClient.GetUsersName(ctx, []string{interfaces.SystemUser, "user1"})
			So(err, ShouldBeNil)
			So(len(userMap), ShouldEqual, 2)
			So(userMap[interfaces.SystemUser], ShouldEqual, interfaces.SystemUser)
			So(userMap["user1"], ShouldEqual, "张三")
		})
	})
}

func TestRemoveUserIDs(t *testing.T) {
	Convey("removeUserIDs", t, func() {
		mockClient := &userManagementClient{}

		Convey("正常移除用户ID", func() {
			source := []string{"user1", "user2", "user3", "user4"}
			toRemove := []string{"user1", "user3"}

			result := mockClient.removeUserIDs(source, toRemove)
			So(len(result), ShouldEqual, 2)
			So(result, ShouldContain, "user2")
			So(result, ShouldContain, "user4")
			So(result, ShouldNotContain, "user1")
			So(result, ShouldNotContain, "user3")
		})

		Convey("移除不存在的用户ID", func() {
			source := []string{"user1", "user2"}
			toRemove := []string{"user3", "user4"}

			result := mockClient.removeUserIDs(source, toRemove)
			So(len(result), ShouldEqual, 2)
			So(result, ShouldContain, "user1")
			So(result, ShouldContain, "user2")
		})

		Convey("空数组", func() {
			result := mockClient.removeUserIDs([]string{}, []string{"user1"})
			So(len(result), ShouldEqual, 0)

			result = mockClient.removeUserIDs([]string{"user1"}, []string{})
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, "user1")
		})
	})
}

// func TestGetUsersInfo(t *testing.T) {
// 	um = &userManagementClient{
// 		baseURL: fmt.Sprintf("%s://%s:%d/api/user-management", "http",
// 			"10.4.111.251", 30980),
// 		logger:     interfaces.DefaultLogger(),
// 		httpClient: rest.NewHTTPClient(),
// 	}
// 	resp, err := um.GetUsersInfo(context.Background(), []string{"test"}, []string{"name"})
// 	if err != nil {
// 		t.Errorf("get users info failed, err: %v", err)
// 	}
// 	fmt.Println(utils.ObjectToJSON(resp))
// }
