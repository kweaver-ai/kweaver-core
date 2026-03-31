package umhttpaccess

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

// GetUserInfo 获取用户信息
func (u *umHttpAcc) GetUserInfo(ctx context.Context, dto *umarg.GetUserInfoArgDto) (uim umcmp.UserInfoMap, err error) {
	//	if helpers.IsLocalDev() {
	//		// 模拟数据
	//		uim = um.UserInfoMap{
	//			// e2398a78-b33e-11ed-a382-cec886b9f898
	//			// d0121e78-b640-11ed-85a7-8a41c44025b9
	//			"d0121e78-b640-11ed-85a7-8a41c44025b9": &um.UserInfo{
	//				Id:   "d0121e78-b640-11ed-85a7-8a41c44025b9",
	//				Name: "sap01",
	//				ParentDeps: [][]um.ObjectBaseInfo{
	//					{
	//						{
	//							ID:   "151bcb65-48ce-4b62-973f-0bb6685f9cb8",
	//							Name: "组织结构",
	//							Type: "department",
	//						},
	//						{
	//							ID:   "12baa3d8-f941-11ed-aa89-264a25126d02",
	//							Name: "SAP",
	//							Type: "department",
	//						},
	//					},
	//				},
	//			},
	//			"e2398a78-b33e-11ed-a382-cec886b9f898": &um.UserInfo{
	//				Id:   "e2398a78-b33e-11ed-a382-cec886b9f898",
	//				Name: "aaron",
	//				ParentDeps: [][]um.ObjectBaseInfo{
	//					{
	//						{
	//							ID:   "151bcb65-48ce-4b62-973f-0bb6685f9cb8",
	//							Name: "组织结构",
	//							Type: "department",
	//						},
	//						{
	//							ID:   "12baa3d8-f941-11ed-aa89-264a25126d02",
	//							Name: "SAP",
	//							Type: "department",
	//						},
	//					},
	//					{
	//						{
	//							ID:   "151bcb65-48ce-4b62-973f-0bb6685f9cb8",
	//							Name: "组织结构",
	//							Type: "department",
	//						},
	//						{
	//							ID:   "12baa3d8-f941-11ed-aa89-264a25126d22",
	//							Name: "SAP02",
	//							Type: "department",
	//						},
	//					},
	//				},
	//			},
	//		}
	//
	//		return
	//	}
	uim, err = u.um.GetUserInfo(ctx, dto)
	if err != nil {
		chelper.RecordErrLogWithPos(u.logger, err, "umHttpAcc.GetUserInfo")
		return nil, errors.Wrap(err, "获取用户信息失败")
	}

	return
}

// GetAppIDNameKv 获取应用账号ID=>Name键值对
func (u *umHttpAcc) GetAppIDNameKv(ctx context.Context, appIDs []string) (idNameKvMap map[string]string, err error) {
	getNamesDto := &umarg.GetOsnArgDto{
		AppIDs: appIDs,
	}

	ret, err := u.GetOsnNames(ctx, getNamesDto)
	if err != nil {
		chelper.RecordErrLogWithPos(u.logger, err, "umHttpAcc.GetAppIDNameKv")
		return nil, errors.Wrap(err, "获取应用账号ID=>Name键值对失败")
	}

	idNameKvMap = ret.AppNameMap

	return
}
