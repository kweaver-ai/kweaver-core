package umarg

// DeptInfoField 查询哪一个信息（查询字段）
type DeptInfoField string

// 可以参考um的文档根据需要在此处添加
const (
	DeptFieldName       DeptInfoField = "name"        // 用户显示名
	DeptFieldParentDeps DeptInfoField = "parent_deps" // 父部门信息

	DeptFieldManagers DeptInfoField = "managers" // 管理员信息
)

type DeptFields []DeptInfoField

type GetDeptInfoArgDto struct {
	DeptIds []string
	Fields  DeptFields
}

func (f DeptFields) ToStrings() (fs []string) {
	fs = make([]string, len(f))
	for i := range f {
		fs[i] = string(f[i])
	}

	return
}
