package agentexecutoraccess

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/decision-agent/agent-factory/cconf"
	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/ctype"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type aeTestLogger struct{}

func (aeTestLogger) Infof(string, ...interface{})  {}
func (aeTestLogger) Infoln(...interface{})         {}
func (aeTestLogger) Debugf(string, ...interface{}) {}
func (aeTestLogger) Debugln(...interface{})        {}
func (aeTestLogger) Errorf(string, ...interface{}) {}
func (aeTestLogger) Errorln(...interface{})        {}
func (aeTestLogger) Warnf(string, ...interface{})  {}
func (aeTestLogger) Warnln(...interface{})         {}
func (aeTestLogger) Panicf(string, ...interface{}) {}
func (aeTestLogger) Panicln(...interface{})        {}
func (aeTestLogger) Fatalf(string, ...interface{}) {}
func (aeTestLogger) Fatalln(...interface{})        {}

func newTestAgentExecutorAcc(serverURL string) *agentExecutorHttpAcc {
	return &agentExecutorHttpAcc{
		logger:            aeTestLogger{},
		agentExecutorConf: &conf.AgentExecutorConf{},
		restClient:        rest.NewHTTPClient(),
		privateAddress:    serverURL,
	}
}

func TestAgentCacheManage_Happy(t *testing.T) {
	t.Parallel()

	respData := agentexecutoraccres.AgentCacheManageResp{
		CacheID: "cache-1",
		TTL:     3600,
	}
	respBytes, _ := json.Marshal(respData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/agent-executor/v1/agent/cache/manage", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respBytes)
	}))
	defer ts.Close()

	acc := newTestAgentExecutorAcc(ts.URL)
	req := &agentexecutoraccreq.AgentCacheManageReq{
		AgentID:      "agent-1",
		AgentVersion: "v1",
		Action:       agentexecutoraccreq.AgentCacheActionUpsert,
	}
	visitorInfo := &ctype.VisitorInfo{
		XAccountID:        "user-1",
		XAccountType:      cenum.AccountTypeUser,
		XBusinessDomainID: "bd-1",
	}

	result, err := acc.AgentCacheManage(context.Background(), req, visitorInfo)
	require.NoError(t, err)
	assert.Equal(t, "cache-1", result.CacheID)
	assert.Equal(t, 3600, result.TTL)
}

func TestAgentCacheManage_NonOKStatus(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer ts.Close()

	acc := newTestAgentExecutorAcc(ts.URL)
	req := &agentexecutoraccreq.AgentCacheManageReq{
		AgentID: "agent-1",
		Action:  agentexecutoraccreq.AgentCacheActionUpsert,
	}
	visitorInfo := &ctype.VisitorInfo{}

	_, err := acc.AgentCacheManage(context.Background(), req, visitorInfo)
	assert.Error(t, err)
}

func TestAgentCacheManage_BadJSON(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not-valid-json`))
	}))
	defer ts.Close()

	acc := newTestAgentExecutorAcc(ts.URL)
	req := &agentexecutoraccreq.AgentCacheManageReq{
		AgentID: "agent-1",
		Action:  agentexecutoraccreq.AgentCacheActionGetInfo,
	}
	visitorInfo := &ctype.VisitorInfo{}

	_, err := acc.AgentCacheManage(context.Background(), req, visitorInfo)
	assert.Error(t, err)
}

func TestAgentCacheManage_RequestFailed(t *testing.T) {
	t.Parallel()

	acc := newTestAgentExecutorAcc("http://127.0.0.1:19999")
	req := &agentexecutoraccreq.AgentCacheManageReq{AgentID: "agent-1"}
	visitorInfo := &ctype.VisitorInfo{}

	_, err := acc.AgentCacheManage(context.Background(), req, visitorInfo)
	assert.Error(t, err)
}

func TestNewAgentExecutorHttpAcc(t *testing.T) {
	t.Parallel()

	agentConf := &conf.AgentExecutorConf{
		PrivateSvc: cconf.SvcConf{
			Host:     "localhost",
			Port:     8080,
			Protocol: "http",
		},
	}

	acc := NewAgentExecutorHttpAcc(aeTestLogger{}, agentConf, nil, nil, rest.NewHTTPClient())
	assert.NotNil(t, acc)
}
