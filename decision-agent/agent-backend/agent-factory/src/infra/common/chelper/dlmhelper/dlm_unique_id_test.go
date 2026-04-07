package dlmhelper

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/dbmock"
	"github.com/stretchr/testify/assert"
)

func TestGenRedisDlmUniqueValue_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := dbmock.NewMockUlidRepo(ctrl)
	mockRepo.EXPECT().GenUniqID(gomock.Any(), cconstant.UniqueIDFlagRedisDlm).
		Return("unique-id-123", nil)

	// 注入 mock
	uniqueIDRepo = mockRepo

	value, err := genRedisDlmUniqueValue()
	assert.NoError(t, err)
	assert.Equal(t, "unique-id-123", value)

	// 清理
	uniqueIDRepo = nil
}

func TestGenRedisDlmUniqueValue_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := dbmock.NewMockUlidRepo(ctrl)
	mockRepo.EXPECT().GenUniqID(gomock.Any(), cconstant.UniqueIDFlagRedisDlm).
		Return("", errors.New("db error"))

	uniqueIDRepo = mockRepo

	value, err := genRedisDlmUniqueValue()
	assert.Error(t, err)
	assert.Empty(t, value)

	uniqueIDRepo = nil
}

func TestGenRedisDlmUniqueValue_NilRepo(t *testing.T) {
	// 当 uniqueIDRepo 为 nil 时，函数会创建真实的 repo
	// 由于需要 DB 连接，这里跳过
	t.Skip("需要数据库连接来测试 nil repo 初始化路径")
}

func TestDelRedisDlmUniqueValue_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := dbmock.NewMockUlidRepo(ctrl)
	mockRepo.EXPECT().DelUniqID(gomock.Any(), cconstant.UniqueIDFlagRedisDlm, "unique-id-123").
		Return(nil)

	uniqueIDRepo = mockRepo

	err := delRedisDlmUniqueValue("unique-id-123")
	assert.NoError(t, err)

	uniqueIDRepo = nil
}

func TestDelRedisDlmUniqueValue_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := dbmock.NewMockUlidRepo(ctrl)
	mockRepo.EXPECT().DelUniqID(gomock.Any(), cconstant.UniqueIDFlagRedisDlm, "nonexistent").
		Return(errors.New("not found"))

	uniqueIDRepo = mockRepo

	err := delRedisDlmUniqueValue("nonexistent")
	assert.Error(t, err)

	uniqueIDRepo = nil
}
