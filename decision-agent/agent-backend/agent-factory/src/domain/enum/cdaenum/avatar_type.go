package cdaenum

import "github.com/pkg/errors"

// AvatarType 头像类型
type AvatarType int

func (t AvatarType) EnumCheck() (err error) {
	if t < AvatarTypeBuiltIn || t > AvatarTypeAIGenerated {
		err = errors.New("头像类型不合法")
		return
	}

	return
}

const (
	// AvatarTypeBuiltIn 内置头像
	AvatarTypeBuiltIn AvatarType = 1

	// AvatarTypeUserUploaded 用户上传头像
	AvatarTypeUserUploaded AvatarType = 2

	// AvatarTypeAIGenerated AI生成头像
	AvatarTypeAIGenerated AvatarType = 3
)
