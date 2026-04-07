package service

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"go.uber.org/mock/gomock"
)

//nolint:unused
func getMockedDlm(ctrl *gomock.Controller) (dlmCmp *cmpmock.MockRedisDlmCmp) {
	dlmCmp = cmpmock.NewMockRedisDlmCmp(ctrl)
	mu := cmpmock.NewMockRedisDlmMutexCmp(ctrl)

	mu.EXPECT().Lock(gomock.Any()).Return(nil).Times(1)
	mu.EXPECT().Unlock().Return(nil).Times(1)

	dlmCmp.EXPECT().
		NewMutex(gomock.Any()).
		Return(mu).Times(1)

	return
}

// func getMockedSvcBase(ctrl *gomock.Controller) *SvcBase {
//	arTracer := arTracerMock.NewMockARTracer(ctrl)
//	arTracer.EXPECT().
//		SetInternalSpanName(gomock.Any()).
//		Times(1)
//
//	arTracer.EXPECT().
//		AddInternalTrace(gomock.Any()).Return(context.Background(), nil).
//		AnyTimes()
//
//	arTracer.EXPECT().
//		TelemetrySpanEnd(gomock.Any(), gomock.Any()).
//		AnyTimes()
//
//	return NewSvcBase(arTracer)
//}
