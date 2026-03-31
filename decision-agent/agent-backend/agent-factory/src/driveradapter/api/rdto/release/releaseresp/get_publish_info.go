package releaseresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// PublishInfoResp 发布信息响应
type PublishInfoResp struct {
	// CategoryID     string                `json:"category_id"`      // 分类ID
	// CategoryName   string                `json:"category_name"`    // 分类名称
	Categories []*CategoryInfo `json:"categories"` // 分类列表

	Description    string                  `json:"description"`      // 发布描述
	PublishToWhere []daenum.PublishToWhere `json:"publish_to_where"` // 发布到的目标
	// CustomSpaces   []CustomSpaceInfo     `json:"custom_spaces"`    // 自定义空间列表
	PmsControl   *PmsControlResp       `json:"pms_control"`    // 权限控制信息
	PublishToBes []cdaenum.PublishToBe `json:"publish_to_bes"` // 发布为什么
}

func (r *PublishInfoResp) SetPublishedToBes(po *dapo.ReleasePO) {
	r.PublishToBes = make([]cdaenum.PublishToBe, 0)

	if po.IsAPIAgentBool() {
		r.PublishToBes = append(r.PublishToBes, cdaenum.PublishToBeAPIAgent)
	}

	if po.IsWebSDKAgentBool() {
		r.PublishToBes = append(r.PublishToBes, cdaenum.PublishToBeWebSDKAgent)
	}

	if po.IsSkillAgentBool() {
		r.PublishToBes = append(r.PublishToBes, cdaenum.PublishToBeSkillAgent)
	}

	if po.IsDataFlowAgentBool() {
		r.PublishToBes = append(r.PublishToBes, cdaenum.PublishToBeDataFlowAgent)
	}
}

func (r *PublishInfoResp) SetPublishToWhere(po *dapo.ReleasePO) {
	r.PublishToWhere = make([]daenum.PublishToWhere, 0)

	if po.IsToCustomSpaceBool() {
		r.PublishToWhere = append(r.PublishToWhere, daenum.PublishToWhereCustomSpace)
	}

	if po.IsToSquareBool() {
		r.PublishToWhere = append(r.PublishToWhere, daenum.PublishToWhereSquare)
	}
}

type CategoryInfo struct {
	ID   string `json:"id"`   // 分类ID
	Name string `json:"name"` // 分类名称
}

// CustomSpaceInfo 自定义空间信息
type CustomSpaceInfo struct {
	SpaceID   string `json:"space_id"`   // 空间ID
	SpaceName string `json:"space_name"` // 空间名称
}

// PmsControlResp 权限控制响应信息
type PmsControlResp struct {
	Roles       []comvalobj.RoleInfo       `json:"roles"`       // 角色列表
	Users       []comvalobj.UserInfo       `json:"user"`        // 用户列表
	UserGroups  []comvalobj.UserGroupInfo  `json:"user_group"`  // 用户组列表
	Departments []comvalobj.DepartmentInfo `json:"department"`  // 部门列表
	AppAccounts []comvalobj.AppAccountInfo `json:"app_account"` // 应用账号列表
}

func NewPmsControlResp() *PmsControlResp {
	return &PmsControlResp{
		Roles:       make([]comvalobj.RoleInfo, 0),
		Users:       make([]comvalobj.UserInfo, 0),
		UserGroups:  make([]comvalobj.UserGroupInfo, 0),
		Departments: make([]comvalobj.DepartmentInfo, 0),
		AppAccounts: make([]comvalobj.AppAccountInfo, 0),
	}
}

// NewPublishInfoResp 创建发布信息响应
func NewPublishInfoResp() *PublishInfoResp {
	return &PublishInfoResp{
		Categories:     make([]*CategoryInfo, 0),
		PublishToWhere: make([]daenum.PublishToWhere, 0),
		// CustomSpaces:   make([]CustomSpaceInfo, 0),
		PublishToBes: make([]cdaenum.PublishToBe, 0),
	}
}
