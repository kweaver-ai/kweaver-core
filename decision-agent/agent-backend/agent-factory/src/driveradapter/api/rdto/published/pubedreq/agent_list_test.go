package pubedreq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubedAgentListReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &PubedAgentListReq{}
	errMap := req.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Empty(t, errMap)
}

func TestPubedAgentListReq_CustomCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     *PubedAgentListReq
		wantErr bool
	}{
		{
			name: "valid request",
			req: &PubedAgentListReq{
				IDs:               make([]string, 10),
				AgentKeys:         make([]string, 5),
				ExcludeAgentKeys:  make([]string, 3),
				BusinessDomainIDs: make([]string, 1),
			},
			wantErr: false,
		},
		{
			name: "ids too long",
			req: &PubedAgentListReq{
				IDs: make([]string, 1001),
			},
			wantErr: true,
		},
		{
			name: "agent_keys too long",
			req: &PubedAgentListReq{
				AgentKeys: make([]string, 1001),
			},
			wantErr: true,
		},
		{
			name: "exclude_agent_keys too long",
			req: &PubedAgentListReq{
				ExcludeAgentKeys: make([]string, 1001),
			},
			wantErr: true,
		},
		{
			name: "business_domain_ids too long",
			req: &PubedAgentListReq{
				BusinessDomainIDs: make([]string, 3),
			},
			wantErr: true,
		},
		{
			name: "business_domain_ids at limit",
			req: &PubedAgentListReq{
				BusinessDomainIDs: make([]string, 2),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.req.CustomCheck()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPubedAgentListReq_LoadMarkerStr(t *testing.T) {
	t.Parallel()

	t.Run("empty marker string", func(t *testing.T) {
		t.Parallel()

		req := &PubedAgentListReq{
			PaginationMarkerStr: "",
		}

		err := req.LoadMarkerStr()
		assert.NoError(t, err)
		assert.Nil(t, req.Marker)
	})

	t.Run("valid marker string", func(t *testing.T) {
		t.Parallel()
		// Create a valid marker with proper field names: published_at and last_release_id
		validMarker := "eyJwdWJsaXNoZWRfYXQiOjEyMzQ1Njc4OTAsImxhc3RfcmVsZWFzZV9pZCI6ImFnZW50LTEyMyJ9"
		req := &PubedAgentListReq{
			PaginationMarkerStr: validMarker,
		}

		err := req.LoadMarkerStr()
		assert.NoError(t, err)
		assert.NotNil(t, req.Marker)
		assert.Equal(t, int64(1234567890), req.Marker.PublishedAt)
		assert.Equal(t, "agent-123", req.Marker.LastReleaseID)
	})

	t.Run("invalid base64 marker string", func(t *testing.T) {
		t.Parallel()

		req := &PubedAgentListReq{
			PaginationMarkerStr: "not-valid-base64!!!",
		}

		err := req.LoadMarkerStr()
		// Should return error because the marker is not valid base64
		assert.Error(t, err)
	})

	t.Run("invalid JSON in marker string", func(t *testing.T) {
		t.Parallel()
		// Valid base64 but invalid JSON
		invalidJSONMarker := "eyJpbnZhbGlkIGpzb24ifQ==" // base64 of "{invalid json}"
		req := &PubedAgentListReq{
			PaginationMarkerStr: invalidJSONMarker,
		}

		err := req.LoadMarkerStr()
		// Should return error because the marker contains invalid JSON
		assert.Error(t, err)
	})

	t.Run("nil marker after load", func(t *testing.T) {
		t.Parallel()

		req := &PubedAgentListReq{
			PaginationMarkerStr: "",
		}

		err := req.LoadMarkerStr()
		require.NoError(t, err)
		assert.Nil(t, req.Marker)
	})
}

func TestPubedAgentListReq_Fields(t *testing.T) {
	t.Parallel()

	req := &PubedAgentListReq{
		Name:                "test agent",
		CategoryID:          "cat-1",
		ToBeFlag:            cdaenum.PublishToBeAPIAgent,
		CustomSpaceID:       "space-1",
		IsToCustomSpace:     1,
		IsToSquare:          0,
		Size:                20,
		PaginationMarkerStr: "marker-123",
		BusinessDomainIDs:   []string{"domain-1"},
		AgentKeys:           []string{"agent-1"},
		ExcludeAgentKeys:    []string{"agent-2"},
	}

	assert.Equal(t, "test agent", req.Name)
	assert.Equal(t, "cat-1", req.CategoryID)
	assert.Equal(t, cdaenum.PublishToBeAPIAgent, req.ToBeFlag)
	assert.Equal(t, "space-1", req.CustomSpaceID)
	assert.Equal(t, 1, req.IsToCustomSpace)
	assert.Equal(t, 0, req.IsToSquare)
	assert.Equal(t, 20, req.Size)
	assert.Equal(t, "marker-123", req.PaginationMarkerStr)
	assert.Equal(t, []string{"domain-1"}, req.BusinessDomainIDs)
	assert.Equal(t, []string{"agent-1"}, req.AgentKeys)
	assert.Equal(t, []string{"agent-2"}, req.ExcludeAgentKeys)
}

func TestPubedAgentListReq_Empty(t *testing.T) {
	t.Parallel()

	req := &PubedAgentListReq{}

	assert.Empty(t, req.Name)
	assert.Empty(t, req.CategoryID)
	assert.Empty(t, req.CustomSpaceID)
	assert.Empty(t, req.IDs)
	assert.Empty(t, req.AgentKeys)
	assert.Empty(t, req.ExcludeAgentKeys)
	assert.Empty(t, req.BusinessDomainIDs)
	assert.Equal(t, 0, req.Size)
	assert.Empty(t, req.PaginationMarkerStr)
}

func TestPubedAgentListReq_WithAllFieldsSet(t *testing.T) {
	t.Parallel()

	ids := make([]string, 500)
	agentKeys := make([]string, 200)
	excludeKeys := make([]string, 100)

	req := &PubedAgentListReq{
		Name:                "Full Name",
		IDs:                 ids,
		AgentKeys:           agentKeys,
		ExcludeAgentKeys:    excludeKeys,
		CategoryID:          "category-full",
		ToBeFlag:            cdaenum.PublishToBeWebSDKAgent,
		CustomSpaceID:       "custom-space-full",
		IsToCustomSpace:     1,
		IsToSquare:          1,
		BusinessDomainIDs:   []string{"domain-1", "domain-2"},
		Size:                100,
		PaginationMarkerStr: "marker-full",
	}

	assert.Equal(t, "Full Name", req.Name)
	assert.Len(t, req.IDs, 500)
	assert.Len(t, req.AgentKeys, 200)
	assert.Len(t, req.ExcludeAgentKeys, 100)
	assert.Equal(t, "category-full", req.CategoryID)
	assert.Equal(t, cdaenum.PublishToBeWebSDKAgent, req.ToBeFlag)
	assert.Equal(t, "custom-space-full", req.CustomSpaceID)
	assert.Equal(t, 1, req.IsToCustomSpace)
	assert.Equal(t, 1, req.IsToSquare)
	assert.Len(t, req.BusinessDomainIDs, 2)
	assert.Equal(t, 100, req.Size)
	assert.Equal(t, "marker-full", req.PaginationMarkerStr)
}

func TestPubedAgentListReq_CustomCheck_EmptySlices(t *testing.T) {
	t.Parallel()

	req := &PubedAgentListReq{
		IDs:               []string{},
		AgentKeys:         []string{},
		ExcludeAgentKeys:  []string{},
		BusinessDomainIDs: []string{},
	}

	err := req.CustomCheck()
	assert.NoError(t, err)
}

func TestPubedAgentListReq_CustomCheck_AllBoundaryConditions(t *testing.T) {
	t.Parallel()

	t.Run("ids at boundary (1000)", func(t *testing.T) {
		t.Parallel()

		req := &PubedAgentListReq{
			IDs: make([]string, 1000),
		}
		err := req.CustomCheck()
		assert.NoError(t, err)
	})

	t.Run("agent_keys at boundary (1000)", func(t *testing.T) {
		t.Parallel()

		req := &PubedAgentListReq{
			AgentKeys: make([]string, 1000),
		}
		err := req.CustomCheck()
		assert.NoError(t, err)
	})

	t.Run("exclude_agent_keys at boundary (1000)", func(t *testing.T) {
		t.Parallel()

		req := &PubedAgentListReq{
			ExcludeAgentKeys: make([]string, 1000),
		}
		err := req.CustomCheck()
		assert.NoError(t, err)
	})
}
