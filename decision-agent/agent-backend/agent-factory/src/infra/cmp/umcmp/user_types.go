package umcmp

// ObjectBaseInfo 对象基本信息
// 可用于部门信息
type ObjectBaseInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type GroupInfo struct {
	ID    string `json:"id" `
	Name  string `json:"name" `
	Notes string `json:"notes" `
}

// UserInfo 用户信息，包含用户名称和用户所属部门路径
// 根据需要可添加其他信息，可参考：UserManagement driveradapters/user_rest_handler.go:502
type UserInfo struct {
	//nolint:stylecheck
	Id   string `json:"id"`   // 用户id
	Name string `json:"name"` // 用户显示名

	// 父部门信息，描述多个父部门的层级关系信息，每个父部门层级数组内第一个对象是根部门，最后一个对象是直接父部门
	ParentDeps [][]ObjectBaseInfo `json:"parent_deps"`

	Enabled bool `json:"enabled"` // 用户禁用状态，true:未禁用 false:禁用

	// 用户角色
	Roles []string `json:"roles"`

	// 用户账户名
	Account string `json:"account"`

	// 用户及其所属部门的所属用户组信息集合
	Groups []*GroupInfo `json:"groups"`
}

func (ui *UserInfo) GroupIDs() (groupIDs []string) {
	groupIDs = make([]string, 0, len(ui.Groups))
	for _, g := range ui.Groups {
		groupIDs = append(groupIDs, g.ID)
	}

	return
}

type UserInfos []*UserInfo

// UserInfoMap 用户信息map
// key: 用户id
// value: 用户信息
type UserInfoMap map[string]*UserInfo
