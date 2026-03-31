package knsearch

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// mockLogger 模拟 Logger 接口
type mockLogger struct {
	logs []string
}

func (m *mockLogger) WithContext(ctx context.Context) interfaces.Logger {
	return m
}

func (m *mockLogger) Info(v ...interface{}) {
	m.logs = append(m.logs, fmt.Sprint(v...))
}

func (m *mockLogger) Debug(v ...interface{}) {
	m.logs = append(m.logs, fmt.Sprint(v...))
}

func (m *mockLogger) Warn(v ...interface{}) {
	m.logs = append(m.logs, fmt.Sprint(v...))
}

func (m *mockLogger) Error(v ...interface{}) {
	m.logs = append(m.logs, fmt.Sprint(v...))
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.logs = append(m.logs, fmt.Sprintf("[INFO] "+format, args...))
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {
	m.logs = append(m.logs, fmt.Sprintf("[DEBUG] "+format, args...))
}

func (m *mockLogger) Warnf(format string, args ...interface{}) {
	m.logs = append(m.logs, fmt.Sprintf("[WARN] "+format, args...))
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
	m.logs = append(m.logs, fmt.Sprintf("[ERROR] "+format, args...))
}

// mockBknBackend 模拟 BknBackendAccess 接口
type mockBknBackend struct {
	networkDetail      *interfaces.KnowledgeNetworkDetail
	networkError       error
	objectTypesResp    *interfaces.ObjectTypeConcepts
	objectTypesError   error
	relationTypesResp  *interfaces.RelationTypeConcepts
	relationTypesError error
}

func (m *mockBknBackend) GetKnowledgeNetworkDetail(ctx context.Context, knID string) (*interfaces.KnowledgeNetworkDetail, error) {
	return m.networkDetail, m.networkError
}

func (m *mockBknBackend) SearchObjectTypes(ctx context.Context, req *interfaces.QueryConceptsReq) (*interfaces.ObjectTypeConcepts, error) {
	return m.objectTypesResp, m.objectTypesError
}

func (m *mockBknBackend) SearchRelationTypes(ctx context.Context, req *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
	return m.relationTypesResp, m.relationTypesError
}

// 下面是接口中其他方法的空实现，满足接口定义
func (m *mockBknBackend) GetObjectTypeDetail(ctx context.Context, knID string, otIds []string, includeDetail bool) ([]*interfaces.ObjectType, error) {
	return nil, nil
}
func (m *mockBknBackend) GetRelationTypeDetail(ctx context.Context, knID string, rtIDs []string, includeDetail bool) ([]*interfaces.RelationType, error) {
	return nil, nil
}
func (m *mockBknBackend) SearchActionTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (actionTypes *interfaces.ActionTypeConcepts, err error) {
	return nil, nil
}
func (m *mockBknBackend) GetActionTypeDetail(ctx context.Context, knID string, atIDs []string, includeDetail bool) ([]*interfaces.ActionType, error) {
	return nil, nil
}
func (m *mockBknBackend) CreateFullBuildOntologyJob(ctx context.Context, knID string, req *interfaces.CreateFullBuildOntologyJobReq) (resp *interfaces.CreateJobResp, err error) {
	return nil, nil
}
func (m *mockBknBackend) ListOntologyJobs(ctx context.Context, knID string, req *interfaces.ListOntologyJobsReq) (resp *interfaces.ListOntologyJobsResp, err error) {
	return nil, nil
}

// mockOntologyQuery 模拟 DrivenOntologyQuery 接口
type mockOntologyQuery struct {
	instancesResp  *interfaces.QueryObjectInstancesResp
	instancesError error
	callCount      int
}

func (m *mockOntologyQuery) QueryObjectInstances(ctx context.Context, req *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
	m.callCount++
	return m.instancesResp, m.instancesError
}

func (m *mockOntologyQuery) QueryLogicProperties(ctx context.Context, req *interfaces.QueryLogicPropertiesReq) (*interfaces.QueryLogicPropertiesResp, error) {
	return nil, nil
}

func (m *mockOntologyQuery) QueryActions(ctx context.Context, req *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
	return nil, nil
}

func (m *mockOntologyQuery) QueryInstanceSubgraph(ctx context.Context, req *interfaces.QueryInstanceSubgraphReq) (resp *interfaces.QueryInstanceSubgraphResp, err error) {
	return nil, nil
}

// mockRerankClient 模拟 DrivenMFModelAPIClient 接口
type mockRerankClient struct {
	rerankResp  *interfaces.RerankResp
	rerankError error
}

func (m *mockRerankClient) Rerank(ctx context.Context, query string, documents []string) (*interfaces.RerankResp, error) {
	return m.rerankResp, m.rerankError
}

func (m *mockRerankClient) Chat(ctx context.Context, req *interfaces.LLMChatReq) (string, error) {
	return "", nil
}

// createMockNetworkDetail 创建测试用的知识网络详情
func createMockNetworkDetail(objectCount, relationCount, actionCount int) *interfaces.KnowledgeNetworkDetail {
	detail := &interfaces.KnowledgeNetworkDetail{
		ID:            "129",
		ObjectTypes:   make([]*interfaces.ObjectType, objectCount),
		RelationTypes: make([]*interfaces.RelationType, relationCount),
		ActionTypes:   make([]*interfaces.ActionType, actionCount),
	}

	// 生成对象类型（至少一个属性支持语义检索，供语义实例召回测试使用）
	for i := 0; i < objectCount; i++ {
		detail.ObjectTypes[i] = &interfaces.ObjectType{
			ID:      fmt.Sprintf("obj_%d", i),
			Name:    fmt.Sprintf("对象类型_%d", i),
			Comment: fmt.Sprintf("对象注释_%d", i),
			DataProperties: []*interfaces.DataProperty{
				{
					Name:                "prop1",
					DisplayName:         "属性1",
					Type:                "text",
					ConditionOperations: []interfaces.KnOperationType{interfaces.KnOperationTypeKnn, interfaces.KnOperationTypeMatch},
				},
				{Name: "prop2", DisplayName: "属性2"},
			},
		}
	}

	// 生成关系类型
	for i := 0; i < relationCount; i++ {
		detail.RelationTypes[i] = &interfaces.RelationType{
			ID:                 fmt.Sprintf("rel_%d", i),
			Name:               fmt.Sprintf("关系_%d", i),
			Comment:            fmt.Sprintf("关系注释_%d", i),
			SourceObjectTypeID: fmt.Sprintf("obj_%d", i%objectCount),
			TargetObjectTypeID: fmt.Sprintf("obj_%d", (i+1)%objectCount),
		}
	}

	// 生成操作类型
	for i := 0; i < actionCount; i++ {
		detail.ActionTypes[i] = &interfaces.ActionType{
			ID:           fmt.Sprintf("action_%d", i),
			Name:         fmt.Sprintf("操作_%d", i),
			Comment:      fmt.Sprintf("操作注释_%d", i),
			ObjectTypeID: fmt.Sprintf("obj_%d", i%objectCount),
		}
	}

	return detail
}

// createMockInstanceData 创建测试用的实例数据（预留供扩展测试使用）
//
//nolint:unused
func createMockInstanceData(count int) []interface{} {
	data := make([]interface{}, count)
	for i := 0; i < count; i++ {
		data[i] = map[string]interface{}{
			"unique_identities": map[string]interface{}{
				"id": fmt.Sprintf("inst_%d", i),
			},
			"instance_name": fmt.Sprintf("实例_%d", i),
			"field1":        fmt.Sprintf("值_%d", i),
			"_score":        0.9 - float64(i)*0.1,
		}
	}
	return data
}
