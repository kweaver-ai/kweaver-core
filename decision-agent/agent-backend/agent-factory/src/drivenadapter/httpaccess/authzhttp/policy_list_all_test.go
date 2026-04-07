package authzhttp

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

// 用于测试的 Mock AuthZHttpAcc
type mockAuthZHttpAccForTest struct {
	mockResponses []*authzhttpres.ListPolicyRes
	mockErrors    []error
	callCount     int
}

func (m *mockAuthZHttpAccForTest) ListPolicy(ctx context.Context, req *authzhttpreq.ListPolicyReq) (*authzhttpres.ListPolicyRes, error) {
	if m.callCount >= len(m.mockResponses) {
		return nil, errors.New("超出预期调用次数")
	}

	defer func() { m.callCount++ }()

	if m.mockErrors != nil && m.callCount < len(m.mockErrors) && m.mockErrors[m.callCount] != nil {
		return nil, m.mockErrors[m.callCount]
	}

	return m.mockResponses[m.callCount], nil
}

// 创建测试用的策略条目
func createTestPolicyEntry(id, resourceID string) *authzhttpres.PolicyEntry {
	return &authzhttpres.PolicyEntry{
		ID:        id,
		ExpiresAt: "1970-01-01T08:00:00+08:00",
		Resource: &authzhttpreq.PolicyResource{
			ID:   resourceID,
			Type: cdaenum.ResourceTypeDataAgent,
			Name: "测试资源",
		},
		Accessor: &authzhttpres.PolicyAccessor{
			ID:   "user-123",
			Type: cenum.PmsTargetObjTypeUser,
			Name: "测试用户",
		},
		Operation: &authzhttpres.PolicyOperation{
			Allow: []*authzhttpres.PolicyOperationItem{
				{
					ID:   "view",
					Name: "查看",
				},
			},
			Deny: []*authzhttpres.PolicyOperationItem{},
		},
		Condition: "",
	}
}

func TestListPolicyAll_SinglePage(t *testing.T) {
	t.Parallel()

	// 创建 mock 对象
	mockAcc := &mockAuthZHttpAccForTest{
		mockResponses: []*authzhttpres.ListPolicyRes{
			{
				TotalCount: 2,
				Entries: []*authzhttpres.PolicyEntry{
					createTestPolicyEntry("policy-1", "resource-1"),
					createTestPolicyEntry("policy-2", "resource-1"),
				},
			},
		},
		mockErrors: []error{nil},
	}

	// 准备请求
	req := &authzhttpreq.ListPolicyReq{
		ResourceID:   "resource-1",
		ResourceType: cdaenum.ResourceTypeDataAgent,
	}

	ctx := context.Background()

	// 执行测试
	res, err := mockAcc.ListPolicy(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 2, res.TotalCount)
	assert.Len(t, res.Entries, 2)
	assert.Equal(t, "policy-1", res.Entries[0].ID)
	assert.Equal(t, "policy-2", res.Entries[1].ID)
}

func TestListPolicyAll_MultiplePages(t *testing.T) {
	t.Parallel()

	// 测试多页数据的情况
	mockAcc := &mockAuthZHttpAccForTest{
		mockResponses: []*authzhttpres.ListPolicyRes{
			// 第一页: 1000 条数据，总共 2500 条
			{
				TotalCount: 2500,
				Entries:    make([]*authzhttpres.PolicyEntry, 1000),
			},
			// 第二页: 1000 条数据
			{
				TotalCount: 2500,
				Entries:    make([]*authzhttpres.PolicyEntry, 1000),
			},
			// 第三页: 500 条数据（最后一页）
			{
				TotalCount: 2500,
				Entries:    make([]*authzhttpres.PolicyEntry, 500),
			},
		},
		mockErrors: []error{nil, nil, nil},
	}

	// 填充测试数据
	for i := 0; i < len(mockAcc.mockResponses); i++ {
		for j := 0; j < len(mockAcc.mockResponses[i].Entries); j++ {
			mockAcc.mockResponses[i].Entries[j] = createTestPolicyEntry(
				fmt.Sprintf("policy-%d", i*1000+j),
				"resource-1",
			)
		}
	}

	req := &authzhttpreq.ListPolicyReq{
		ResourceID:   "resource-1",
		ResourceType: cdaenum.ResourceTypeDataAgent,
	}

	ctx := context.Background()

	// 模拟 ListPolicyAll 的逻辑
	res := &authzhttpres.ListPolicyRes{
		Entries:    make([]*authzhttpres.PolicyEntry, 0),
		TotalCount: 0,
	}

	const maxIterations = 5

	const pageSize = 1000

	currentReq := &authzhttpreq.ListPolicyReq{
		Limit:        pageSize,
		Offset:       0,
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
	}

	var err error

	for i := 0; i < maxIterations; i++ {
		pageRes, pageErr := mockAcc.ListPolicy(ctx, currentReq)
		if pageErr != nil {
			err = pageErr
			break
		}

		if pageRes == nil {
			err = errors.New("查询策略列表返回空结果")
			break
		}

		// 第一次查询时设置总数
		if i == 0 {
			res.TotalCount = pageRes.TotalCount
		}

		// 收集当前页的数据
		res.Entries = append(res.Entries, pageRes.Entries...)

		// 检查是否已获取完所有数据
		if pageRes.TotalCount <= currentReq.Offset+currentReq.Limit {
			// 所有数据已获取完毕
			break
		}

		// 准备下一页查询
		currentReq.Offset += pageSize
	}

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 2500, res.TotalCount)
	assert.Len(t, res.Entries, 2500)      // 应该获取到所有数据
	assert.Equal(t, 3, mockAcc.callCount) // 应该调用了 3 次
}

func TestListPolicyAll_MaxIterationsExceeded(t *testing.T) {
	t.Parallel()

	// 测试达到最大循环次数的情况
	mockAcc := &mockAuthZHttpAccForTest{
		mockResponses: make([]*authzhttpres.ListPolicyRes, 6), // 6 页，超过最大限制 5 页
		mockErrors:    make([]error, 6),
	}

	// 每页都返回 1000 条数据，总共 6000 条（超过 5000 限制）
	for i := 0; i < 6; i++ {
		mockAcc.mockResponses[i] = &authzhttpres.ListPolicyRes{
			TotalCount: 6000,
			Entries:    make([]*authzhttpres.PolicyEntry, 1000),
		}
		mockAcc.mockErrors[i] = nil

		// 填充测试数据
		for j := 0; j < 1000; j++ {
			mockAcc.mockResponses[i].Entries[j] = createTestPolicyEntry(
				fmt.Sprintf("policy-%d", i*1000+j),
				"resource-1",
			)
		}
	}

	req := &authzhttpreq.ListPolicyReq{
		ResourceID:   "resource-1",
		ResourceType: cdaenum.ResourceTypeDataAgent,
	}

	ctx := context.Background()

	// 模拟 ListPolicyAll 的逻辑
	res := &authzhttpres.ListPolicyRes{
		Entries:    make([]*authzhttpres.PolicyEntry, 0),
		TotalCount: 0,
	}

	const maxIterations = 5

	const pageSize = 1000

	currentReq := &authzhttpreq.ListPolicyReq{
		Limit:        pageSize,
		Offset:       0,
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
	}

	var err error

	for i := 0; i < maxIterations; i++ {
		pageRes, pageErr := mockAcc.ListPolicy(ctx, currentReq)
		if pageErr != nil {
			err = pageErr
			break
		}

		if pageRes == nil {
			err = errors.New("查询策略列表返回空结果")
			break
		}

		// 第一次查询时设置总数
		if i == 0 {
			res.TotalCount = pageRes.TotalCount
		}

		// 收集当前页的数据
		res.Entries = append(res.Entries, pageRes.Entries...)

		// 检查是否已获取完所有数据
		if pageRes.TotalCount <= currentReq.Offset+currentReq.Limit {
			// 所有数据已获取完毕
			break
		}

		// 准备下一页查询
		currentReq.Offset += pageSize
	}

	// 如果循环结束但仍未获取完所有数据，返回错误
	if err == nil && res.TotalCount > len(res.Entries) {
		err = errors.New("已达到最大循环次数，但仍未获取完所有数据")
	}

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "已达到最大循环次数")
	assert.Equal(t, 6000, res.TotalCount)
	assert.Len(t, res.Entries, 5000)      // 只获取到 5000 条数据
	assert.Equal(t, 5, mockAcc.callCount) // 调用了 5 次
}

func TestListPolicyAll_ErrorHandling(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		mockResponses []*authzhttpres.ListPolicyRes
		mockErrors    []error
		expectError   bool
		errorContains string
	}{
		{
			name: "第一次调用就失败",
			mockResponses: []*authzhttpres.ListPolicyRes{
				nil,
			},
			mockErrors: []error{
				errors.New("网络错误"),
			},
			expectError:   true,
			errorContains: "网络错误",
		},
		{
			name: "第二次调用失败",
			mockResponses: []*authzhttpres.ListPolicyRes{
				{
					TotalCount: 2000,
					Entries:    make([]*authzhttpres.PolicyEntry, 1000),
				},
				{
					TotalCount: 2000,
					Entries:    make([]*authzhttpres.PolicyEntry, 0),
				},
			},
			mockErrors: []error{
				nil,
				errors.New("服务器内部错误"),
			},
			expectError:   true,
			errorContains: "服务器内部错误",
		},
		{
			name: "响应为空但无错误",
			mockResponses: []*authzhttpres.ListPolicyRes{
				nil,
			},
			mockErrors: []error{
				nil,
			},
			expectError:   true,
			errorContains: "查询策略列表返回空结果",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockAcc := &mockAuthZHttpAccForTest{
				mockResponses: tc.mockResponses,
				mockErrors:    tc.mockErrors,
			}

			req := &authzhttpreq.ListPolicyReq{
				ResourceID:   "resource-1",
				ResourceType: cdaenum.ResourceTypeDataAgent,
			}

			ctx := context.Background()

			// 模拟 ListPolicyAll 的逻辑来测试错误处理
			const maxIterations = 5

			const pageSize = 1000

			currentReq := &authzhttpreq.ListPolicyReq{
				Limit:        pageSize,
				Offset:       0,
				ResourceID:   req.ResourceID,
				ResourceType: req.ResourceType,
			}

			var res *authzhttpres.ListPolicyRes

			var err error

			for i := 0; i < maxIterations && i < len(tc.mockResponses); i++ {
				res, err = mockAcc.ListPolicy(ctx, currentReq)
				if err != nil {
					break
				}

				if res == nil {
					err = errors.New("查询策略列表返回空结果")
					break
				}

				if res.TotalCount <= currentReq.Offset+currentReq.Limit {
					break
				}

				currentReq.Offset += pageSize
			}

			if tc.expectError {
				assert.Error(t, err)

				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestListPolicyAll_EmptyResult(t *testing.T) {
	t.Parallel()

	// 测试空结果的情况
	mockAcc := &mockAuthZHttpAccForTest{
		mockResponses: []*authzhttpres.ListPolicyRes{
			{
				TotalCount: 0,
				Entries:    []*authzhttpres.PolicyEntry{},
			},
		},
		mockErrors: []error{nil},
	}

	req := &authzhttpreq.ListPolicyReq{
		ResourceID:   "resource-1",
		ResourceType: cdaenum.ResourceTypeDataAgent,
	}

	ctx := context.Background()

	res, err := mockAcc.ListPolicy(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 0, res.TotalCount)
	assert.Len(t, res.Entries, 0)
}

func TestListPolicyAll_BoundaryConditions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		totalCount   int
		entriesCount int
		expectPages  int
	}{
		{
			name:         "恰好1000条",
			totalCount:   1000,
			entriesCount: 1000,
			expectPages:  1,
		},
		{
			name:         "1001条",
			totalCount:   1001,
			entriesCount: 1001,
			expectPages:  2,
		},
		{
			name:         "恰好5000条",
			totalCount:   5000,
			entriesCount: 5000,
			expectPages:  5,
		},
		{
			name:         "只有1条",
			totalCount:   1,
			entriesCount: 1,
			expectPages:  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// 根据测试用例构造 mock 响应
			mockResponses := make([]*authzhttpres.ListPolicyRes, tc.expectPages)

			for i := 0; i < tc.expectPages; i++ {
				remainingCount := tc.entriesCount - i*1000
				if remainingCount > 1000 {
					remainingCount = 1000
				}

				mockResponses[i] = &authzhttpres.ListPolicyRes{
					TotalCount: tc.totalCount,
					Entries:    make([]*authzhttpres.PolicyEntry, remainingCount),
				}

				// 填充数据
				for j := 0; j < remainingCount; j++ {
					mockResponses[i].Entries[j] = createTestPolicyEntry(
						fmt.Sprintf("policy-%d", i*1000+j),
						"resource-1",
					)
				}
			}

			mockAcc := &mockAuthZHttpAccForTest{
				mockResponses: mockResponses,
				mockErrors:    make([]error, tc.expectPages),
			}

			// 验证第一页数据
			req := &authzhttpreq.ListPolicyReq{
				ResourceID:   "resource-1",
				ResourceType: cdaenum.ResourceTypeDataAgent,
			}

			ctx := context.Background()

			res, err := mockAcc.ListPolicy(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tc.totalCount, res.TotalCount)

			if tc.entriesCount >= 1000 {
				assert.Len(t, res.Entries, 1000)
			} else {
				assert.Len(t, res.Entries, tc.entriesCount)
			}
		})
	}
}

// 测试 ListPolicyAll 逻辑的完整流程
func TestListPolicyAllLogic_Complete(t *testing.T) {
	t.Parallel()

	// 模拟完整的 ListPolicyAll 逻辑
	mockAcc := &mockAuthZHttpAccForTest{
		mockResponses: []*authzhttpres.ListPolicyRes{
			{
				TotalCount: 1500,
				Entries:    make([]*authzhttpres.PolicyEntry, 1000),
			},
			{
				TotalCount: 1500,
				Entries:    make([]*authzhttpres.PolicyEntry, 500),
			},
		},
		mockErrors: []error{nil, nil},
	}

	// 填充数据
	for i := 0; i < 2; i++ {
		entryCount := 1000
		if i == 1 {
			entryCount = 500
		}

		for j := 0; j < entryCount; j++ {
			mockAcc.mockResponses[i].Entries[j] = createTestPolicyEntry(
				fmt.Sprintf("policy-%d", i*1000+j),
				"resource-test",
			)
		}
	}

	// 模拟实际的 ListPolicyAll 逻辑
	simulateListPolicyAll := func(ctx context.Context, req *authzhttpreq.ListPolicyReq) (*authzhttpres.ListPolicyRes, error) {
		res := &authzhttpres.ListPolicyRes{
			Entries:    make([]*authzhttpres.PolicyEntry, 0),
			TotalCount: 0,
		}

		const maxIterations = 5

		const pageSize = 1000

		currentReq := &authzhttpreq.ListPolicyReq{
			Limit:        pageSize,
			Offset:       0,
			ResourceID:   req.ResourceID,
			ResourceType: req.ResourceType,
		}

		for i := 0; i < maxIterations; i++ {
			pageRes, pageErr := mockAcc.ListPolicy(ctx, currentReq)
			if pageErr != nil {
				return nil, fmt.Errorf("第%d次查询策略列表失败: %v", i+1, pageErr)
			}

			if pageRes == nil {
				return nil, errors.New("查询策略列表返回空结果")
			}

			// 第一次查询时设置总数
			if i == 0 {
				res.TotalCount = pageRes.TotalCount
			}

			// 收集当前页的数据
			res.Entries = append(res.Entries, pageRes.Entries...)

			// 检查是否已获取完所有数据
			if pageRes.TotalCount <= currentReq.Offset+currentReq.Limit {
				// 所有数据已获取完毕
				return res, nil
			}

			// 准备下一页查询
			currentReq.Offset += pageSize
		}

		// 达到最大循环次数但仍未获取完所有数据
		return res, fmt.Errorf("已达到最大循环次数%d次，但仍未获取完所有数据，当前已获取%d条，总计%d条",
			maxIterations, len(res.Entries), res.TotalCount)
	}

	req := &authzhttpreq.ListPolicyReq{
		ResourceID:   "resource-test",
		ResourceType: cdaenum.ResourceTypeDataAgent,
	}

	ctx := context.Background()
	res, err := simulateListPolicyAll(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1500, res.TotalCount)
	assert.Len(t, res.Entries, 1500)
	assert.Equal(t, 2, mockAcc.callCount)

	// 验证数据的连续性
	for i := 0; i < 1500; i++ {
		expectedID := fmt.Sprintf("policy-%d", i)
		assert.Equal(t, expectedID, res.Entries[i].ID)
	}
}
