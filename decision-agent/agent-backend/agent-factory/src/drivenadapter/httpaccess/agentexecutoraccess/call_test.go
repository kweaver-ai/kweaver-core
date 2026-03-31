package agentexecutoraccess

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutordto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/httpclient/mock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// ==================== Call ====================

func TestCall_DebugChat(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStream := mock.NewMockHTTPClient(ctrl)

	acc := &agentExecutorHttpAcc{
		logger:         aeTestLogger{},
		streamClient:   mockStream,
		privateAddress: "http://localhost:9999",
	}

	msgCh := make(chan string, 1)
	errCh := make(chan error, 1)
	msgCh <- "data: hello"
	close(msgCh)
	close(errCh)

	mockStream.EXPECT().StreamPost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, url string, headers map[string]string, req interface{}) (chan string, chan error, error) {
			assert.Contains(t, url, "/api/agent-executor/v1/agent/debug")
			return msgCh, errCh, nil
		})

	req := &agentexecutordto.AgentCallReq{
		CallType:          constant.DebugChat,
		UserID:            "user-1",
		Token:             "tk-1",
		XAccountID:        "acc-1",
		XAccountType:      cenum.AccountTypeUser,
		XBusinessDomainID: "bd-1",
		Config: agentexecutordto.Config{
			SessionID: "sess-1",
			AgentID:   "agent-1",
		},
	}

	messages, errs, err := acc.Call(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, messages)
	assert.NotNil(t, errs)
}

func TestCall_NormalChat(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStream := mock.NewMockHTTPClient(ctrl)

	acc := &agentExecutorHttpAcc{
		logger:         aeTestLogger{},
		streamClient:   mockStream,
		privateAddress: "http://localhost:9999",
	}

	msgCh := make(chan string)
	errCh := make(chan error)

	close(msgCh)
	close(errCh)

	mockStream.EXPECT().StreamPost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, url string, headers map[string]string, req interface{}) (chan string, chan error, error) {
			assert.Contains(t, url, "/api/agent-executor/v1/agent/run")
			return msgCh, errCh, nil
		})

	req := &agentexecutordto.AgentCallReq{
		CallType: constant.APIChat,
		Config: agentexecutordto.Config{
			SessionID: "sess-1",
			AgentID:   "agent-1",
		},
	}

	messages, errs, err := acc.Call(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, messages)
	assert.NotNil(t, errs)
}

func TestCall_StreamPostError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStream := mock.NewMockHTTPClient(ctrl)

	acc := &agentExecutorHttpAcc{
		logger:         aeTestLogger{},
		streamClient:   mockStream,
		privateAddress: "http://localhost:9999",
	}

	mockStream.EXPECT().StreamPost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil, assert.AnError)

	req := &agentexecutordto.AgentCallReq{
		CallType: constant.InternalChat,
		Config: agentexecutordto.Config{
			SessionID: "sess-1",
			AgentID:   "agent-1",
		},
	}

	_, _, err := acc.Call(context.Background(), req)
	assert.Error(t, err)
}

func TestCall_WithToken(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStream := mock.NewMockHTTPClient(ctrl)

	acc := &agentExecutorHttpAcc{
		logger:         aeTestLogger{},
		streamClient:   mockStream,
		privateAddress: "http://localhost:9999",
	}

	msgCh := make(chan string)
	errCh := make(chan error)

	close(msgCh)
	close(errCh)

	mockStream.EXPECT().StreamPost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, url string, headers map[string]string, req interface{}) (chan string, chan error, error) {
			assert.Equal(t, "test-token", headers["token"])
			assert.Equal(t, "Bearer test-token", headers["Authorization"])

			return msgCh, errCh, nil
		})

	req := &agentexecutordto.AgentCallReq{
		Token: "test-token",
		Config: agentexecutordto.Config{
			SessionID: "sess-1",
			AgentID:   "agent-1",
		},
	}

	_, _, err := acc.Call(context.Background(), req)
	assert.NoError(t, err)
}
