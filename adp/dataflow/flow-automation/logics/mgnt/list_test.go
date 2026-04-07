package mgnt

import (
	"context"
	"errors"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"go.uber.org/mock/gomock"
)

func TestListDagWithFiltersBypassesFiltersInMockMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mod.NewMockStore(ctrl)
	prevStore := mod.GetStore()
	mod.SetStore(store)
	defer mod.SetStore(prevStore)

	patch := ApplyFunc(common.NewConfig, func() *common.Config {
		return &common.Config{Server: common.Server{AuthEnabled: "false"}}
	})
	defer patch.Reset()

	dags := []*entity.Dag{{BaseInfo: entity.BaseInfo{ID: "dag-1"}, Name: "mock-dag"}}

	store.EXPECT().
		ListDagCount(gomock.Any(), gomock.AssignableToTypeOf(&mod.ListDagInput{})).
		Return(int64(1), nil)
	store.EXPECT().
		ListDag(gomock.Any(), gomock.AssignableToTypeOf(&mod.ListDagInput{})).
		DoAndReturn(func(_ context.Context, input *mod.ListDagInput) ([]*entity.Dag, error) {
			if input.Type != common.DagTypeDataFlow {
				t.Fatalf("expected type %q, got %q", common.DagTypeDataFlow, input.Type)
			}
			return dags, nil
		})

	errFilter := func(c *listDagConfig) {
		c.filters = append(c.filters, func(context.Context, []string) ([]string, error) {
			return nil, errors.New("filter should be skipped in mock mode")
		})
	}

	res, total, err := ListDagWithFilters(context.Background(), QueryParams{Type: common.DagTypeDataFlow}, errFilter)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(res) != 1 || res[0].ID != "dag-1" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestListDagWithFiltersAppliesFiltersWhenAuthEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mod.NewMockStore(ctrl)
	prevStore := mod.GetStore()
	mod.SetStore(store)
	defer mod.SetStore(prevStore)

	patch := ApplyFunc(common.NewConfig, func() *common.Config {
		return &common.Config{Server: common.Server{AuthEnabled: "true"}}
	})
	defer patch.Reset()

	errFilter := func(c *listDagConfig) {
		c.filters = append(c.filters, func(context.Context, []string) ([]string, error) {
			return nil, errors.New("filter should run when auth is enabled")
		})
	}

	_, _, err := ListDagWithFilters(context.Background(), QueryParams{Type: common.DagTypeDataFlow}, errFilter)
	if err == nil {
		t.Fatal("expected filter error when auth is enabled")
	}
}
