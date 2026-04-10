package agentsvc

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
)

// ==================== TerminateChat — deeper paths ====================

// stopchan存在但为nil的路径 (Line 52-56)
func TestTerminateChat_NilStopChan_NoInterruptedMsg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	// Store nil value in stopChanMap
	stopChanMap.Store("conv-nil-chan", nil)
	defer stopChanMap.Delete("conv-nil-chan")

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-nil-chan", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stopchan is nil")
}

// Not-owner path: stopchan nil but interruptedMsgID provided → update msg success
func TestTerminateChat_NilStopChan_WithInterruptedMsg_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		logger:              mockLogger,
		conversationMsgRepo: mockConvMsgRepo,
	}

	stopChanMap.Store("conv-nil-interrupted", nil)
	defer stopChanMap.Delete("conv-nil-interrupted")

	msgPO := &dapo.ConversationMsgPO{ID: "msg-nil"}
	mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "msg-nil").Return(msgPO, nil)
	mockConvMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-nil-interrupted", "", "msg-nil")
	assert.NoError(t, err)
}

// GetByID returns nil msgPO → no update needed, success
func TestTerminateChat_WithInterruptedMsgID_NilMsgPO(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		logger:              mockLogger,
		conversationMsgRepo: mockConvMsgRepo,
	}

	ch := make(chan struct{}, 1)
	stopChanMap.Store("conv-nil-po", ch)

	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "msg-nil-po").Return(nil, nil)

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-nil-po", "", "msg-nil-po")
	assert.NoError(t, err)
}

// ==================== addCiteDocMark — deeper paths ====================

func TestAddCiteDocMark_WithDocReference(t *testing.T) {
	svc := &agentSvc{}
	// answer中包含"参考信息1"模式，触发 markInDocIndex → HasRefrence=true
	answer := "这是第一个句子关于AI的介绍。第1个参考信息相关的内容描述。"
	cites := []*agentrespvo.CiteDoc{
		{Content: "Reference content 1", Slices: []*agentrespvo.V1Slice{{Content: "AI介绍"}}},
	}

	result := svc.addCiteDocMark(answer, cites)
	assert.NotEmpty(t, result)
}

func TestAddCiteDocMark_WithMultipleDocReferences(t *testing.T) {
	svc := &agentSvc{}
	// 包含"参考文档1和2"的模式
	answer := "根据参考文档1和2的内容，这个结论是正确的。后续还有更多的详细说明。"
	cites := []*agentrespvo.CiteDoc{
		{Content: "Doc 1 content", Slices: []*agentrespvo.V1Slice{{Content: "结论"}}},
		{Content: "Doc 2 content", Slices: []*agentrespvo.V1Slice{{Content: "说明"}}},
	}

	result := svc.addCiteDocMark(answer, cites)
	assert.NotEmpty(t, result)
}

func TestAddCiteDocMark_LongSentenceWithCites(t *testing.T) {
	svc := &agentSvc{}
	// 足够长的句子，使 splitSentences 合并生效
	answer := "短句。这是一个非常长的句子用于测试分句逻辑确保它能够正确处理和计算相似度。另一个长句子来确保覆盖到循环的不同分支和路径。"
	cites := []*agentrespvo.CiteDoc{
		{Content: "测试分句逻辑", Slices: []*agentrespvo.V1Slice{{Content: "分句"}, {Content: "相似度"}}},
		{Content: "循环分支", Slices: []*agentrespvo.V1Slice{{Content: "路径"}}},
	}

	result := svc.addCiteDocMark(answer, cites)
	assert.NotEmpty(t, result)
}

// ==================== markInDocIndex — deeper paths ====================

func TestMarkInDocIndex_WithPattern(t *testing.T) {
	text := "这是第1个参考信息的详细说明"
	has, docIndexs, newText := markInDocIndex(text, docRefPatternList)
	assert.True(t, has)
	assert.NotEmpty(t, docIndexs)
	assert.NotEqual(t, text, newText)
}

func TestMarkInDocIndex_MultiplePatterns(t *testing.T) {
	text := "参考文档1和2的内容说明"
	has, docIndexs, newText := markInDocIndex(text, docRefPatternList)
	assert.True(t, has)
	assert.Len(t, docIndexs, 2)
	assert.Contains(t, newText, "<i")
}

func TestMarkInDocIndex_NoPatternMatch(t *testing.T) {
	text := "这是一个没有任何引用的普通文本"
	has, docIndexs, newText := markInDocIndex(text, docRefPatternList)
	assert.False(t, has)
	assert.Empty(t, docIndexs)
	assert.Equal(t, text, newText)
}

func TestMarkInDocIndex_ParenPattern(t *testing.T) {
	text := "这是说明（参考信息1, 4）的内容"
	has, docIndexs, newText := markInDocIndex(text, docRefPatternList)
	assert.True(t, has)
	assert.NotEmpty(t, docIndexs)
	assert.Contains(t, newText, "<i")
}

func TestMarkInDocIndex_DocIDPattern(t *testing.T) {
	text := "详情请看（参考文档ID：第1个）"
	has, docIndexs, newText := markInDocIndex(text, docRefPatternList)
	assert.True(t, has)
	assert.NotEmpty(t, docIndexs)
	assert.Contains(t, newText, "<i")
}

// ==================== splitSentences ====================

func TestSplitSentences_ShortSentences(t *testing.T) {
	text := "短。很短。"
	result := splitSentences(text, 10)
	// short sentences should be combined
	assert.NotEmpty(t, result)
}

func TestSplitSentences_LongSentence(t *testing.T) {
	text := "这是一个非常长的中文句子用于测试分句逻辑。"
	result := splitSentences(text, 5)
	assert.NotEmpty(t, result)
}

func TestSplitSentences_Empty(t *testing.T) {
	result := splitSentences("", 10)
	assert.Empty(t, result)
}

// ==================== ResumeChat — error path ====================

func TestResumeChat_ConversationNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	// SessionMap does not have this conversation
	SessionMap.Delete("conv-not-found")

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.ResumeChat(context.Background(), "conv-not-found")
	assert.Error(t, err)
}
