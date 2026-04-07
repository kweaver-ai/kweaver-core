package umret

type UserGroupInfo struct {
	ID   string `json:"id"`   // 组唯一标识
	Name string `json:"name"` // 组名，唯一
}

type UserGroupListResDto struct {
	Entries    []*UserGroupInfo `json:"entries"`
	TotalCount int64            `json:"total_count"`
}

func NewUserGroupListResDto() *UserGroupListResDto {
	return &UserGroupListResDto{
		Entries: make([]*UserGroupInfo, 0),
	}
}
