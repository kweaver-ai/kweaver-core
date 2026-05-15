// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func newTestBKNService(t *testing.T) (*bknService, *gomock.Controller, *bmock.MockKNService) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	kns := bmock.NewMockKNService(mockCtrl)
	svc := &bknService{
		appSetting: &common.AppSetting{},
		kns:        kns,
	}
	return svc, mockCtrl, kns
}

func Test_bknService_ExportToTar(t *testing.T) {
	Convey("Test bknService ExportToTar\n", t, func() {
		svc, mockCtrl, kns := newTestBKNService(t)
		defer mockCtrl.Finish()

		Convey("Success with empty KN (no sub-types)\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn1",
				KNName: "Test Network",
			}
			kns.EXPECT().GetKNByID(gomock.Any(), "kn1", interfaces.MAIN_BRANCH, interfaces.Mode_Export).Return(kn, nil)

			data, err := svc.ExportToTar(context.Background(), "kn1", interfaces.MAIN_BRANCH)

			So(err, ShouldBeNil)
			So(len(data), ShouldBeGreaterThan, 0)
		})

		Convey("Success with KN containing all sub-types\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn2",
				KNName: "Full Network",
				ObjectTypes: []*interfaces.ObjectType{
					{ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{OTID: "ot1", OTName: "OT1"}},
				},
				RelationTypes: []*interfaces.RelationType{
					{RelationTypeWithKeyField: interfaces.RelationTypeWithKeyField{RTID: "rt1", RTName: "RT1"}},
				},
				ActionTypes: []*interfaces.ActionType{
					{ActionTypeWithKeyField: interfaces.ActionTypeWithKeyField{ATID: "at1", ATName: "AT1"}},
				},
				ConceptGroups: []*interfaces.ConceptGroup{
					{CGID: "cg1", CGName: "CG1"},
				},
			}
			kns.EXPECT().GetKNByID(gomock.Any(), "kn2", interfaces.MAIN_BRANCH, interfaces.Mode_Export).Return(kn, nil)

			data, err := svc.ExportToTar(context.Background(), "kn2", interfaces.MAIN_BRANCH)

			So(err, ShouldBeNil)
			So(len(data), ShouldBeGreaterThan, 0)
		})

		Convey("Failed when GetKNByID returns error\n", func() {
			getErr := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_KnowledgeNetwork_InternalError,
				},
			}
			kns.EXPECT().GetKNByID(gomock.Any(), "kn-err", interfaces.MAIN_BRANCH, interfaces.Mode_Export).Return(nil, getErr)

			data, err := svc.ExportToTar(context.Background(), "kn-err", interfaces.MAIN_BRANCH)

			So(err, ShouldNotBeNil)
			So(errors.Is(err, getErr), ShouldBeTrue)
			So(data, ShouldBeNil)
		})

		Convey("Export tar preserves metrics readable by LoadNetworkFromTar\n", func() {
			kn := &interfaces.KN{
				KNID:   "kn-metrics",
				KNName: "KN With Metrics",
				Branch: interfaces.MAIN_BRANCH,
				Metrics: []*interfaces.MetricDefinition{
					{
						ID:         "pod_running_count",
						Name:       "Running Pods",
						MetricType: interfaces.MetricTypeAtomic,
						ScopeRef:   "pod",
						CalculationFormula: &interfaces.MetricCalculationFormula{
							Aggregation: interfaces.MetricAggregation{
								Property: "id",
								Aggr:     interfaces.MetricAggrCount,
							},
						},
					},
				},
			}
			kns.EXPECT().GetKNByID(gomock.Any(), "kn-metrics", interfaces.MAIN_BRANCH, interfaces.Mode_Export).Return(kn, nil)

			data, err := svc.ExportToTar(context.Background(), "kn-metrics", interfaces.MAIN_BRANCH)
			So(err, ShouldBeNil)
			loaded, err := bknsdk.LoadNetworkFromTar(bytes.NewReader(data))
			So(err, ShouldBeNil)
			So(len(loaded.Metrics), ShouldEqual, 1)
			So(loaded.Metrics[0].ID, ShouldEqual, "pod_running_count")
		})
	})
}
