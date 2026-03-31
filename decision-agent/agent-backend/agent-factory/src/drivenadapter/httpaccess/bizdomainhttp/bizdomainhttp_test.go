package bizdomainhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpres"
)

type bdTestLogger struct{}

func (bdTestLogger) Infof(string, ...interface{})  {}
func (bdTestLogger) Infoln(...interface{})         {}
func (bdTestLogger) Debugf(string, ...interface{}) {}
func (bdTestLogger) Debugln(...interface{})        {}
func (bdTestLogger) Errorf(string, ...interface{}) {}
func (bdTestLogger) Errorln(...interface{})        {}
func (bdTestLogger) Warnf(string, ...interface{})  {}
func (bdTestLogger) Warnln(...interface{})         {}
func (bdTestLogger) Panicf(string, ...interface{}) {}
func (bdTestLogger) Panicln(...interface{})        {}
func (bdTestLogger) Fatalf(string, ...interface{}) {}
func (bdTestLogger) Fatalln(...interface{})        {}

func newBizDomainAcc(serverURL string) *bizDomainHttpAcc {
	return &bizDomainHttpAcc{
		logger:         bdTestLogger{},
		privateBaseURL: serverURL,
	}
}

func TestAssociateResource_Happy(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, associateResourcePath, r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":0,"data":{}}`))
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := &bizdomainhttpreq.AssociateResourceReq{}
	err := acc.AssociateResource(context.Background(), req)
	require.NoError(t, err)
}

func TestAssociateResource_HTTPError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":500}`))
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := &bizdomainhttpreq.AssociateResourceReq{}
	err := acc.AssociateResource(context.Background(), req)
	assert.Error(t, err)
}

func TestAssociateResource_RequestFailed(t *testing.T) {
	t.Parallel()

	acc := newBizDomainAcc("http://127.0.0.1:19998")
	req := &bizdomainhttpreq.AssociateResourceReq{}
	err := acc.AssociateResource(context.Background(), req)
	assert.Error(t, err)
}

func TestQueryResourceAssociations_Happy(t *testing.T) {
	t.Parallel()

	respData := &bizdomainhttpres.QueryResourceAssociationsRes{}
	respBytes, _ := json.Marshal(respData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, associateResourcePath, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respBytes)
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := &bizdomainhttpreq.QueryResourceAssociationsReq{
		BdID:  "bd-1",
		Type:  cdaenum.ResourceTypeDataAgent,
		Limit: 10,
	}
	res, err := acc.QueryResourceAssociations(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestQueryResourceAssociations_BadJSON(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not-valid-json`))
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := &bizdomainhttpreq.QueryResourceAssociationsReq{BdID: "bd-1"}
	_, err := acc.QueryResourceAssociations(context.Background(), req)
	assert.Error(t, err)
}

func TestQueryResourceAssociations_RequestFailed(t *testing.T) {
	t.Parallel()

	acc := newBizDomainAcc("http://127.0.0.1:19998")
	req := &bizdomainhttpreq.QueryResourceAssociationsReq{BdID: "bd-1"}
	_, err := acc.QueryResourceAssociations(context.Background(), req)
	assert.Error(t, err)
}

func TestGetAllAgentIDList_Happy(t *testing.T) {
	t.Parallel()

	respData := &bizdomainhttpres.QueryResourceAssociationsRes{}
	respBytes, _ := json.Marshal(respData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respBytes)
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	agentIDs, agentID2BdIDMap, err := acc.GetAllAgentIDList(context.Background(), []string{"bd-1"})
	require.NoError(t, err)
	assert.NotNil(t, agentIDs)
	assert.NotNil(t, agentID2BdIDMap)
}

func TestGetAllAgentIDList_EmptyBdIDs(t *testing.T) {
	t.Parallel()

	acc := newBizDomainAcc("http://127.0.0.1:19998")
	agentIDs, agentID2BdIDMap, err := acc.GetAllAgentIDList(context.Background(), []string{})
	require.NoError(t, err)
	assert.Empty(t, agentIDs)
	assert.Empty(t, agentID2BdIDMap)
}

func TestGetAllAgentIDList_HTTPError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	_, _, err := acc.GetAllAgentIDList(context.Background(), []string{"bd-1"})
	assert.Error(t, err)
}
