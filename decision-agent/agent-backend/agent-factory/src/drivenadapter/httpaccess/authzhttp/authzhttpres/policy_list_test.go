package authzhttpres

import (
	"testing"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/stretchr/testify/assert"
)

func TestListPolicyRes_FilterByExpiresAt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		listPolicyRes      *ListPolicyRes
		expectedCount      int
		expectedTotalCount int
		expectError        bool
	}{
		{
			name: "包含永不过期和未过期的策略",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 3,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00", // 永不过期
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
					},
					{
						ID:        "policy-2",
						ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339), // 未来24小时
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-2",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源2",
						},
					},
					{
						ID:        "policy-3",
						ExpiresAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339), // 过去24小时（已过期）
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-3",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源3",
						},
					},
				},
			},
			expectedCount:      2, // 只有前两个有效
			expectedTotalCount: 2,
			expectError:        false,
		},
		{
			name: "所有策略都已过期",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 2,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: time.Now().Add(-2 * time.Hour).Format(time.RFC3339), // 过去2小时
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
					},
					{
						ID:        "policy-2",
						ExpiresAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339), // 过去1小时
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-2",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源2",
						},
					},
				},
			},
			expectedCount:      0, // 全部过期
			expectedTotalCount: 0,
			expectError:        false,
		},
		{
			name: "所有策略都永不过期",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 2,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00", // 永不过期
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
					},
					{
						ID:        "policy-2",
						ExpiresAt: "1970-01-01T08:00:00+08:00", // 永不过期
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-2",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源2",
						},
					},
				},
			},
			expectedCount:      2, // 全部有效
			expectedTotalCount: 2,
			expectError:        false,
		},
		{
			name: "包含无效的expires_at格式",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 3,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00", // 永不过期
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
					},
					{
						ID:        "policy-2",
						ExpiresAt: "invalid-time-format", // 无效格式
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-2",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源2",
						},
					},
					{
						ID:        "policy-3",
						ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339), // 未来1小时
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-3",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源3",
						},
					},
				},
			},
			expectedCount:      3,    // 由于遇到错误会直接返回，条目数量保持原值
			expectedTotalCount: 3,    // TotalCount保持原值
			expectError:        true, // 期望返回错误
		},
		{
			name: "只包含有效时间格式的混合策略",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 3,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00", // 永不过期
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
					},
					{
						ID:        "policy-2",
						ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339), // 未来1小时
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-2",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源2",
						},
					},
					{
						ID:        "policy-3",
						ExpiresAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339), // 过去1小时（已过期）
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-3",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源3",
						},
					},
				},
			},
			expectedCount:      2, // 只有前两个有效（永不过期和未来过期）
			expectedTotalCount: 2,
			expectError:        false,
		},
		{
			name: "空列表",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 0,
				Entries:    []*PolicyEntry{},
			},
			expectedCount:      0,
			expectedTotalCount: 0,
			expectError:        false,
		},
		{
			name:               "nil指针",
			listPolicyRes:      nil,
			expectedCount:      0,
			expectedTotalCount: 0,
			expectError:        false,
		},
		{
			name: "包含nil条目",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 2,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00", // 永不过期
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
					},
					nil, // nil条目
				},
			},
			expectedCount:      1, // 只有第1个有效
			expectedTotalCount: 1,
			expectError:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var err error
			if tc.listPolicyRes != nil {
				err = tc.listPolicyRes.FilterByExpiresAt()
			}

			if tc.listPolicyRes == nil {
				// 对于nil指针，不会有错误，直接返回
				assert.NoError(t, err)
				return
			}

			// 检查是否符合错误预期
			if tc.expectError {
				assert.Error(t, err, "期望返回错误但实际未返回错误")
				// 当有错误时，验证数据未被修改（保持原状）
				assert.Len(t, tc.listPolicyRes.Entries, tc.expectedCount)
				assert.Equal(t, tc.expectedTotalCount, tc.listPolicyRes.TotalCount)
			} else {
				assert.NoError(t, err, "不期望返回错误但实际返回了错误: %v", err)
				assert.Len(t, tc.listPolicyRes.Entries, tc.expectedCount)
				assert.Equal(t, tc.expectedTotalCount, tc.listPolicyRes.TotalCount)

				// 只有在没有错误的情况下才验证条目的有效性
				now := time.Now().Unix()

				for _, entry := range tc.listPolicyRes.Entries {
					assert.NotNil(t, entry)
					expiresAt, err := time.Parse(time.RFC3339, entry.ExpiresAt)
					assert.NoError(t, err, "剩余条目的expires_at应该是有效格式")

					expiresAtUnix := expiresAt.Unix()
					assert.True(t, expiresAtUnix == 0 || expiresAtUnix > now,
						"剩余条目应该永不过期或未来过期, expires_at: %s, unix: %d, now: %d",
						entry.ExpiresAt, expiresAtUnix, now)
				}
			}
		})
	}
}

func TestListPolicyRes_Filter_EdgeCases(t *testing.T) {
	t.Parallel()

	// 测试边界情况：刚好到过期时间的策略
	now := time.Now()

	testCases := []struct {
		name        string
		expiresAt   string
		shouldKeep  bool
		description string
	}{
		{
			name:        "刚好过期（等于当前时间）",
			expiresAt:   now.Format(time.RFC3339),
			shouldKeep:  false, // 等于当前时间算作过期
			description: "策略刚好在当前时间过期",
		},
		{
			name:        "差1秒过期",
			expiresAt:   now.Add(1 * time.Second).Format(time.RFC3339),
			shouldKeep:  true, // 未来1秒，应该保留
			description: "策略在未来1秒过期",
		},
		{
			name:        "1秒前过期",
			expiresAt:   now.Add(-1 * time.Second).Format(time.RFC3339),
			shouldKeep:  false, // 过去1秒，应该过滤
			description: "策略在1秒前就过期了",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res := &ListPolicyRes{
				TotalCount: 1,
				Entries: []*PolicyEntry{
					{
						ID:        "test-policy",
						ExpiresAt: tc.expiresAt,
						Resource: &authzhttpreq.PolicyResource{
							ID:   "test-resource",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源",
						},
					},
				},
			}

			err := res.FilterByExpiresAt()
			assert.NoError(t, err)

			if tc.shouldKeep {
				assert.Len(t, res.Entries, 1, tc.description)
				assert.Equal(t, 1, res.TotalCount)
			} else {
				assert.Len(t, res.Entries, 0, tc.description)
				assert.Equal(t, 0, res.TotalCount)
			}
		})
	}
}

func TestListPolicyRes_Filter_TimeZones(t *testing.T) {
	t.Parallel()

	// 测试不同时区的时间格式
	now := time.Now()

	// 创建不同时区的相同时间点
	utcTime := now.UTC().Format("2006-01-02T15:04:05Z")
	localTime := now.Format("2006-01-02T15:04:05+08:00") // 假设是东八区

	res := &ListPolicyRes{
		TotalCount: 3,
		Entries: []*PolicyEntry{
			{
				ID:        "policy-utc",
				ExpiresAt: utcTime,
				Resource: &authzhttpreq.PolicyResource{
					ID:   "resource-1",
					Type: cdaenum.ResourceTypeDataAgent,
					Name: "UTC时区测试",
				},
			},
			{
				ID:        "policy-local",
				ExpiresAt: localTime,
				Resource: &authzhttpreq.PolicyResource{
					ID:   "resource-2",
					Type: cdaenum.ResourceTypeDataAgent,
					Name: "本地时区测试",
				},
			},
			{
				ID:        "policy-never",
				ExpiresAt: "1970-01-01T08:00:00+08:00", // 永不过期
				Resource: &authzhttpreq.PolicyResource{
					ID:   "resource-3",
					Type: cdaenum.ResourceTypeDataAgent,
					Name: "永不过期测试",
				},
			},
		},
	}

	err := res.FilterByExpiresAt()
	assert.NoError(t, err)

	// 由于UTC和本地时间都是当前时间（会被过滤），只有永不过期的会保留
	assert.Len(t, res.Entries, 1)
	assert.Equal(t, 1, res.TotalCount)
	assert.Equal(t, "policy-never", res.Entries[0].ID)
}

func TestListPolicyRes_FilterByOperation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		listPolicyRes      *ListPolicyRes
		operation          cdapmsenum.Operator
		expectedCount      int
		expectedTotalCount int
		expectedIDs        []string
	}{
		{
			name: "允许操作匹配",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 3,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
						Operation: &PolicyOperation{
							Allow: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentUse,
									Name: "使用",
								},
								{
									ID:   cdapmsenum.AgentPublish,
									Name: "发布",
								},
							},
							Deny: []*PolicyOperationItem{},
						},
					},
					{
						ID:        "policy-2",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-2",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源2",
						},
						Operation: &PolicyOperation{
							Allow: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentUnpublish,
									Name: "取消发布",
								},
							},
							Deny: []*PolicyOperationItem{},
						},
					},
					{
						ID:        "policy-3",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-3",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源3",
						},
						Operation: &PolicyOperation{
							Allow: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentUse,
									Name: "使用",
								},
							},
							Deny: []*PolicyOperationItem{},
						},
					},
				},
			},
			operation:          cdapmsenum.AgentUse,
			expectedCount:      2, // policy-1 和 policy-3
			expectedTotalCount: 2,
			expectedIDs:        []string{"policy-1", "policy-3"},
		},
		{
			name: "拒绝操作优先级测试",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 2,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
						Operation: &PolicyOperation{
							Allow: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentPublish,
									Name: "发布",
								},
							},
							Deny: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentPublish,
									Name: "发布",
								},
							},
						},
					},
					{
						ID:        "policy-2",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-2",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源2",
						},
						Operation: &PolicyOperation{
							Allow: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentPublish,
									Name: "发布",
								},
							},
							Deny: []*PolicyOperationItem{},
						},
					},
				},
			},
			operation:          cdapmsenum.AgentPublish,
			expectedCount:      1, // 只有policy-2，policy-1被deny规则过滤
			expectedTotalCount: 1,
			expectedIDs:        []string{"policy-2"},
		},
		{
			name: "没有匹配的操作",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 2,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
						Operation: &PolicyOperation{
							Allow: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentUse,
									Name: "使用",
								},
							},
							Deny: []*PolicyOperationItem{},
						},
					},
					{
						ID:        "policy-2",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-2",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源2",
						},
						Operation: &PolicyOperation{
							Allow: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentPublish,
									Name: "发布",
								},
							},
							Deny: []*PolicyOperationItem{},
						},
					},
				},
			},
			operation:          cdapmsenum.AgentUnpublish, // 没有匹配的操作
			expectedCount:      0,
			expectedTotalCount: 0,
			expectedIDs:        []string{},
		},
		{
			name: "operation为nil的条目",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 3,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
						Operation: nil, // nil Operation
					},
					{
						ID:        "policy-2",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-2",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源2",
						},
						Operation: &PolicyOperation{
							Allow: nil, // nil Allow
							Deny:  []*PolicyOperationItem{},
						},
					},
					{
						ID:        "policy-3",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-3",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源3",
						},
						Operation: &PolicyOperation{
							Allow: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentUse,
									Name: "使用",
								},
							},
							Deny: []*PolicyOperationItem{},
						},
					},
				},
			},
			operation:          cdapmsenum.AgentUse,
			expectedCount:      1, // 只有policy-3有效
			expectedTotalCount: 1,
			expectedIDs:        []string{"policy-3"},
		},
		{
			name: "空列表",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 0,
				Entries:    []*PolicyEntry{},
			},
			operation:          cdapmsenum.AgentUse,
			expectedCount:      0,
			expectedTotalCount: 0,
			expectedIDs:        []string{},
		},
		{
			name:               "nil指针",
			listPolicyRes:      nil,
			operation:          cdapmsenum.AgentUse,
			expectedCount:      0,
			expectedTotalCount: 0,
			expectedIDs:        []string{},
		},
		{
			name: "包含nil条目",
			listPolicyRes: &ListPolicyRes{
				TotalCount: 2,
				Entries: []*PolicyEntry{
					{
						ID:        "policy-1",
						ExpiresAt: "1970-01-01T08:00:00+08:00",
						Resource: &authzhttpreq.PolicyResource{
							ID:   "resource-1",
							Type: cdaenum.ResourceTypeDataAgent,
							Name: "测试资源1",
						},
						Operation: &PolicyOperation{
							Allow: []*PolicyOperationItem{
								{
									ID:   cdapmsenum.AgentUse,
									Name: "使用",
								},
							},
							Deny: []*PolicyOperationItem{},
						},
					},
					nil, // nil条目
				},
			},
			operation:          cdapmsenum.AgentUse,
			expectedCount:      1, // 只有policy-1有效
			expectedTotalCount: 1,
			expectedIDs:        []string{"policy-1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var err error
			if tc.listPolicyRes != nil {
				err = tc.listPolicyRes.FilterByOperation(tc.operation)
			}

			if tc.listPolicyRes == nil {
				// 对于nil指针，不会有错误，直接返回
				assert.NoError(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, tc.listPolicyRes.Entries, tc.expectedCount)
			assert.Equal(t, tc.expectedTotalCount, tc.listPolicyRes.TotalCount)

			// 验证返回的条目ID是否符合预期
			actualIDs := make([]string, len(tc.listPolicyRes.Entries))
			for i, entry := range tc.listPolicyRes.Entries {
				actualIDs[i] = entry.ID
			}

			assert.Equal(t, tc.expectedIDs, actualIDs)

			// 验证剩余的条目都有匹配的操作
			for _, entry := range tc.listPolicyRes.Entries {
				assert.NotNil(t, entry)
				assert.NotNil(t, entry.Operation)
				assert.NotNil(t, entry.Operation.Allow)

				// 检查deny中不包含该操作
				hasDenyOperation := false

				if entry.Operation.Deny != nil {
					for _, op := range entry.Operation.Deny {
						if op.ID == tc.operation {
							hasDenyOperation = true
							break
						}
					}
				}

				assert.False(t, hasDenyOperation, "剩余条目在deny中不应包含操作: %s", tc.operation)

				// 检查allow中包含该操作
				hasAllowOperation := false

				for _, op := range entry.Operation.Allow {
					if op.ID == tc.operation {
						hasAllowOperation = true
						break
					}
				}

				assert.True(t, hasAllowOperation, "剩余条目在allow中应包含操作: %s", tc.operation)
			}
		})
	}
}

func TestListPolicyRes_FilterByOperation_EdgeCases(t *testing.T) {
	t.Parallel()

	// 测试边界情况
	testCases := []struct {
		name      string
		operation cdapmsenum.Operator
		entries   []*PolicyEntry
		expected  int
	}{
		{
			name:      "Agent模板发布操作",
			operation: cdapmsenum.AgentTplPublish,
			entries: []*PolicyEntry{
				{
					ID:        "policy-1",
					ExpiresAt: "1970-01-01T08:00:00+08:00",
					Resource: &authzhttpreq.PolicyResource{
						ID:   "resource-1",
						Type: cdaenum.ResourceTypeDataAgentTpl,
						Name: "测试Agent模板",
					},
					Operation: &PolicyOperation{
						Allow: []*PolicyOperationItem{
							{
								ID:   cdapmsenum.AgentTplPublish,
								Name: "发布Agent模板",
							},
						},
						Deny: []*PolicyOperationItem{},
					},
				},
			},
			expected: 1, // 匹配模板发布操作
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res := &ListPolicyRes{
				TotalCount: len(tc.entries),
				Entries:    tc.entries,
			}

			err := res.FilterByOperation(tc.operation)
			assert.NoError(t, err)
			assert.Len(t, res.Entries, tc.expected)
			assert.Equal(t, tc.expected, res.TotalCount)
		})
	}
}
