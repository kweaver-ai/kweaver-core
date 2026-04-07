package umarg

// UserInfoField 查询哪一个信息（查询字段）
type UserInfoField string

// 可以参考um的文档根据需要在此处添加
const (
	FieldName       UserInfoField = "name"        // 用户显示名
	FieldParentDeps UserInfoField = "parent_deps" // 父部门信息

	FieldEnabled UserInfoField = "enabled" // 用户状态

	FieldRoles UserInfoField = "roles" // 用户角色

	FieldAccount UserInfoField = "account" // 用户账户名

	FieldGroups UserInfoField = "groups" // 用户及其所属部门的所属用户组信息集
)

type Fields []UserInfoField

type GetUserInfoArgDto struct {
	UserIds []string
	Fields  Fields
}

func (f Fields) ToStrings() (fs []string) {
	fs = make([]string, len(f))
	for i := range f {
		fs[i] = string(f[i])
	}

	return
}

type GetUserEnableStatusArgDto struct {
	UserIds []string
}

type GetUserInfoSingleArgDto struct {
	UserID string
	Fields Fields
}
