package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/utils"
)

type stubActionRecallService struct {
	called bool
}

func (s *stubActionRecallService) GetActionInfo(_ context.Context, _ *interfaces.KnActionRecallRequest) (*interfaces.KnActionRecallResponse, error) {
	s.called = true
	return &interfaces.KnActionRecallResponse{
		Headers:      map[string]string{"x-test": "ok"},
		DynamicTools: []interfaces.KnDynamicTool{},
	}, nil
}

func TestHandleGetActionInfo_IgnoresResponseFormatParam(t *testing.T) {
	convey.Convey("handleGetActionInfo should ignore response_format argument", t, func() {
		svc := &stubActionRecallService{}
		handler := handleGetActionInfo(svc)

		req := newCallToolRequest(map[string]any{
			"kn_id":                "kn-001",
			"at_id":                "at-001",
			"response_format":      "xml",
			"_instance_identities": []any{map[string]any{"id": "obj-001"}},
		})

		ctx := common.SetAccountAuthContextToCtx(context.Background(), &interfaces.AccountAuthContext{
			AccountID:   "acc-001",
			AccountType: interfaces.AccessorTypeUser,
		})

		result, err := handler(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.IsError, convey.ShouldBeFalse)
		convey.So(svc.called, convey.ShouldBeTrue)
		expectedText := utils.ObjectToJSON(&interfaces.KnActionRecallResponse{
			Headers:      map[string]string{"x-test": "ok"},
			DynamicTools: []interfaces.KnDynamicTool{},
		})
		raw, marshalErr := json.Marshal(result)
		convey.So(marshalErr, convey.ShouldBeNil)
		convey.So(string(raw), convey.ShouldContainSubstring, expectedText)
	})
}

var _ interfaces.IKnActionRecallService = (*stubActionRecallService)(nil)
