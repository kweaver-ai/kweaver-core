package authzhttp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/stretchr/testify/assert"
)

func TestListPolicyReq_ToReqQuery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		req  *authzhttpreq.ListPolicyReq
		want string
	}{
		{
			name: "正常参数",
			req: &authzhttpreq.ListPolicyReq{
				Limit:        50,
				Offset:       0,
				ResourceID:   "5d494a31-f42e-451c-a132-47adb0b15410",
				ResourceType: "operator", // 使用字符串类型，因为 request example 中是 "operator"
			},
			want: "limit=50&offset=0&resource_id=5d494a31-f42e-451c-a132-47adb0b15410&resource_type=operator",
		},
		{
			name: "分页参数",
			req: &authzhttpreq.ListPolicyReq{
				Limit:        100,
				Offset:       200,
				ResourceID:   "test-resource-id",
				ResourceType: cdaenum.ResourceTypeDataAgent,
			},
			want: "limit=100&offset=200&resource_id=test-resource-id&resource_type=agent",
		},
		{
			name: "边界值测试",
			req: &authzhttpreq.ListPolicyReq{
				Limit:        0,
				Offset:       0,
				ResourceID:   "",
				ResourceType: "",
			},
			want: "limit=0&offset=0&resource_id=&resource_type=",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.req.ToReqQuery()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewListPolicyReq(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		resourceID   string
		resourceType cdaenum.ResourceType
		wantLimit    int
		wantOffset   int
	}{
		{
			name:         "数据智能体类型",
			resourceID:   "test-id",
			resourceType: cdaenum.ResourceTypeDataAgent,
			wantLimit:    1000,
			wantOffset:   0,
		},
		{
			name:         "智能体模板类型",
			resourceID:   "tpl-789",
			resourceType: cdaenum.ResourceTypeDataAgentTpl,
			wantLimit:    1000,
			wantOffset:   0,
		},
		{
			name:         "空资源ID",
			resourceID:   "",
			resourceType: cdaenum.ResourceTypeDataAgent,
			wantLimit:    1000,
			wantOffset:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := authzhttpreq.NewListPolicyReq(tc.resourceID, tc.resourceType)

			assert.Equal(t, tc.wantLimit, req.Limit)
			assert.Equal(t, tc.wantOffset, req.Offset)
			assert.Equal(t, tc.resourceID, req.ResourceID)
			assert.Equal(t, tc.resourceType, req.ResourceType)
		})
	}
}

// 测试响应结构的正确性
func TestPolicyResponseStructure(t *testing.T) {
	t.Parallel()

	mockResponse := &authzhttpres.ListPolicyRes{
		TotalCount: 2,
		Entries: []*authzhttpres.PolicyEntry{
			{
				ID:        "c425f82b-0b70-4406-8b47-9baa34ffa27c",
				ExpiresAt: "1970-01-01T08:00:00+08:00",
				Resource: &authzhttpreq.PolicyResource{
					ID:   "5d494a31-f42e-451c-a132-47adb0b15410",
					Type: cdaenum.ResourceTypeDataAgent,
					Name: "文档格式转换sss",
				},
				Condition: "",
			},
		},
	}

	assert.Equal(t, 2, mockResponse.TotalCount)
	assert.Len(t, mockResponse.Entries, 1)
	assert.Equal(t, "c425f82b-0b70-4406-8b47-9baa34ffa27c", mockResponse.Entries[0].ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgent, mockResponse.Entries[0].Resource.Type)
}

// 测试 URL 构造的正确性
func TestListPolicyURLConstruction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		baseURL     string
		req         *authzhttpreq.ListPolicyReq
		expectedURL string
	}{
		{
			name:    "标准 URL 构造",
			baseURL: "http://test-server",
			req: &authzhttpreq.ListPolicyReq{
				Limit:        50,
				Offset:       0,
				ResourceID:   "resource-123",
				ResourceType: cdaenum.ResourceTypeDataAgent,
			},
			expectedURL: "http://test-server/api/authorization/v1/policy?limit=50&offset=0&resource_id=resource-123&resource_type=agent",
		},
		{
			name:    "带分页的 URL",
			baseURL: "https://prod-server",
			req: &authzhttpreq.ListPolicyReq{
				Limit:        100,
				Offset:       200,
				ResourceID:   "tpl-456",
				ResourceType: cdaenum.ResourceTypeDataAgentTpl,
			},
			expectedURL: "https://prod-server/api/authorization/v1/policy?limit=100&offset=200&resource_id=tpl-456&resource_type=agent_tpl",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// 构造预期的 URL
			actualURL := tc.baseURL + "/api/authorization/v1/policy?" + tc.req.ToReqQuery()
			assert.Equal(t, tc.expectedURL, actualURL)
		})
	}
}

// 测试边界条件
func TestListPolicyReq_BoundaryConditions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		req  *authzhttpreq.ListPolicyReq
	}{
		{
			name: "最大限制值",
			req: &authzhttpreq.ListPolicyReq{
				Limit:        5000,
				Offset:       0,
				ResourceID:   "max-test",
				ResourceType: cdaenum.ResourceTypeDataAgent,
			},
		},
		{
			name: "大偏移量",
			req: &authzhttpreq.ListPolicyReq{
				Limit:        1000,
				Offset:       10000,
				ResourceID:   "offset-test",
				ResourceType: cdaenum.ResourceTypeDataAgentTpl,
			},
		},
		{
			name: "最小值",
			req: &authzhttpreq.ListPolicyReq{
				Limit:        1,
				Offset:       0,
				ResourceID:   "min-test",
				ResourceType: cdaenum.ResourceTypeDataAgentTpl,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// 验证请求对象可以正常序列化
			queryString := tc.req.ToReqQuery()
			assert.NotEmpty(t, queryString)
			assert.Contains(t, queryString, "limit=")
			assert.Contains(t, queryString, "offset=")
			assert.Contains(t, queryString, "resource_id=")
			assert.Contains(t, queryString, "resource_type=")
		})
	}
}
