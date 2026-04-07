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

// ==================== DisassociateResource ====================

func TestDisassociateResource_Happy(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := &bizdomainhttpreq.DisassociateResourceReq{}
	err := acc.DisassociateResource(context.Background(), req)
	require.NoError(t, err)
}

func TestDisassociateResource_HTTPError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := &bizdomainhttpreq.DisassociateResourceReq{}
	err := acc.DisassociateResource(context.Background(), req)
	assert.Error(t, err)
}

// ==================== AssociateResourceBatch ====================

func TestAssociateResourceBatch_Happy(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, associateResourceBatchPath, r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := bizdomainhttpreq.AssociateResourceBatchReq{}
	err := acc.AssociateResourceBatch(context.Background(), req)
	require.NoError(t, err)
}

func TestAssociateResourceBatch_HTTPError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := bizdomainhttpreq.AssociateResourceBatchReq{}
	err := acc.AssociateResourceBatch(context.Background(), req)
	assert.Error(t, err)
}

// ==================== HasResourceAssociation ====================

func TestHasResourceAssociation_HasAssoc(t *testing.T) {
	t.Parallel()

	respData := &bizdomainhttpres.QueryResourceAssociationsRes{
		Items: []*bizdomainhttpres.ResourceAssociationItem{{ID: "item-1"}},
	}
	respBytes, _ := json.Marshal(respData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respBytes)
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := &bizdomainhttpreq.QueryResourceAssociationSingleReq{
		BdID: "bd-1",
		ID:   "agent-1",
		Type: cdaenum.ResourceTypeDataAgent,
	}
	has, err := acc.HasResourceAssociation(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, has)
}

func TestHasResourceAssociation_NoAssoc(t *testing.T) {
	t.Parallel()

	respData := &bizdomainhttpres.QueryResourceAssociationsRes{
		Items: []*bizdomainhttpres.ResourceAssociationItem{},
	}
	respBytes, _ := json.Marshal(respData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respBytes)
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	req := &bizdomainhttpreq.QueryResourceAssociationSingleReq{BdID: "bd-1", ID: "agent-1", Type: cdaenum.ResourceTypeDataAgent}
	has, err := acc.HasResourceAssociation(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, has)
}

func TestHasResourceAssociation_Error(t *testing.T) {
	t.Parallel()

	acc := newBizDomainAcc("http://127.0.0.1:19998")
	req := &bizdomainhttpreq.QueryResourceAssociationSingleReq{BdID: "bd-1", ID: "agent-1", Type: cdaenum.ResourceTypeDataAgent}
	_, err := acc.HasResourceAssociation(context.Background(), req)
	assert.Error(t, err)
}

// ==================== GetAllAgentTplIDList ====================

func TestGetAllAgentTplIDList_Happy(t *testing.T) {
	t.Parallel()

	respData := &bizdomainhttpres.QueryResourceAssociationsRes{
		Items: []*bizdomainhttpres.ResourceAssociationItem{{ID: "tpl-1"}},
	}
	respBytes, _ := json.Marshal(respData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respBytes)
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	ids, err := acc.GetAllAgentTplIDList(context.Background(), []string{"bd-1"})
	require.NoError(t, err)
	assert.NotEmpty(t, ids)
}

func TestGetAllAgentTplIDList_EmptyBdIDs(t *testing.T) {
	t.Parallel()

	acc := newBizDomainAcc("http://127.0.0.1:19998")
	ids, err := acc.GetAllAgentTplIDList(context.Background(), []string{})
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestGetAllAgentTplIDList_Error(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	acc := newBizDomainAcc(ts.URL)
	_, err := acc.GetAllAgentTplIDList(context.Background(), []string{"bd-1"})
	assert.Error(t, err)
}
