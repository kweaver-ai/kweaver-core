package agentconfigreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListForBenchmarkReq_StructFields(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{
		Name:                "Test Agent",
		AgentIDsByBizDomain: []string{"agent-1", "agent-2"},
	}
	req.Size = 10
	req.Page = 1

	assert.Equal(t, "Test Agent", req.Name)
	assert.Len(t, req.AgentIDsByBizDomain, 2)
	assert.Equal(t, "agent-1", req.AgentIDsByBizDomain[0])
	assert.Equal(t, "agent-2", req.AgentIDsByBizDomain[1])
	assert.Equal(t, 10, req.Size)
	assert.Equal(t, 1, req.Page)
}

func TestListForBenchmarkReq_Empty(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{}

	assert.Empty(t, req.Name)
	assert.Nil(t, req.AgentIDsByBizDomain)
	assert.Equal(t, 0, req.Size)
	assert.Equal(t, 0, req.Page)
}

func TestListForBenchmarkReq_WithPagination(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{}
	req.Size = 20
	req.Page = 2

	offset := req.GetOffset()
	assert.Equal(t, 20, offset)
}

func TestListForBenchmarkReq_WithDefaultPagination(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{}
	// PageSize has default values when Size is 0
	req.Size = 0
	req.Page = 0

	offset := req.GetOffset()
	assert.Equal(t, 0, offset)
}

func TestListForBenchmarkReq_WithName(t *testing.T) {
	t.Parallel()

	names := []string{
		"Test Agent",
		"中文智能体",
		"Agent with numbers 123",
		"",
	}

	for _, name := range names {
		req := ListForBenchmarkReq{
			Name: name,
		}
		assert.Equal(t, name, req.Name)
	}
}

func TestListForBenchmarkReq_WithAgentIDsByBizDomain(t *testing.T) {
	t.Parallel()

	agentIDs := []string{
		"agent-001",
		"agent-002",
		"agent-003",
	}

	req := ListForBenchmarkReq{
		AgentIDsByBizDomain: agentIDs,
	}

	assert.Len(t, req.AgentIDsByBizDomain, 3)
	assert.Equal(t, "agent-001", req.AgentIDsByBizDomain[0])
	assert.Equal(t, "agent-002", req.AgentIDsByBizDomain[1])
	assert.Equal(t, "agent-003", req.AgentIDsByBizDomain[2])
}

func TestListForBenchmarkReq_WithEmptyAgentIDsByBizDomain(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{
		AgentIDsByBizDomain: []string{},
	}

	assert.NotNil(t, req.AgentIDsByBizDomain)
	assert.Len(t, req.AgentIDsByBizDomain, 0)
}

func TestListForBenchmarkReq_WithNilAgentIDsByBizDomain(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{
		AgentIDsByBizDomain: nil,
	}

	assert.Nil(t, req.AgentIDsByBizDomain)
}

func TestListForBenchmarkReq_PaginationEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     int
		size     int
		expected int
	}{
		{
			name:     "first page",
			page:     1,
			size:     10,
			expected: 0,
		},
		{
			name:     "second page",
			page:     2,
			size:     10,
			expected: 10,
		},
		{
			name:     "large page number",
			page:     100,
			size:     20,
			expected: 1980,
		},
		{
			name:     "zero page",
			page:     0,
			size:     10,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := ListForBenchmarkReq{}
			req.Page = tt.page
			req.Size = tt.size

			offset := req.GetOffset()
			assert.Equal(t, tt.expected, offset)
		})
	}
}

func TestListForBenchmarkReq_EmbeddedPageSize(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{}

	// Verify that PageSize is embedded
	assert.IsType(t, req.Size, 0)
	assert.IsType(t, req.Page, 0)

	// Set and verify pagination values
	req.Size = 15
	req.Page = 3

	assert.Equal(t, 15, req.Size)
	assert.Equal(t, 3, req.Page)
	assert.Equal(t, 30, req.GetOffset())
}

func TestListForBenchmarkReq_WithNameFilter(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{
		Name: "My Agent",
	}
	req.Size = 5
	req.Page = 1

	assert.Equal(t, "My Agent", req.Name)
	assert.Equal(t, 5, req.Size)
	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 0, req.GetOffset())
}

func TestListForBenchmarkReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{
		Name:                "Complete Agent",
		AgentIDsByBizDomain: []string{"agent-1", "agent-2", "agent-3"},
	}
	req.Size = 25
	req.Page = 2

	assert.Equal(t, "Complete Agent", req.Name)
	assert.Len(t, req.AgentIDsByBizDomain, 3)
	assert.Equal(t, 25, req.Size)
	assert.Equal(t, 2, req.Page)
	assert.Equal(t, 25, req.GetOffset())
}

func TestListForBenchmarkReq_ModifyAgentIDsByBizDomain(t *testing.T) {
	t.Parallel()

	req := ListForBenchmarkReq{
		AgentIDsByBizDomain: []string{"agent-1"},
	}

	assert.Len(t, req.AgentIDsByBizDomain, 1)

	// Append new agent IDs
	req.AgentIDsByBizDomain = append(req.AgentIDsByBizDomain, "agent-2", "agent-3")

	assert.Len(t, req.AgentIDsByBizDomain, 3)
	assert.Equal(t, "agent-2", req.AgentIDsByBizDomain[1])
	assert.Equal(t, "agent-3", req.AgentIDsByBizDomain[2])
}
