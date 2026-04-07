package mgnt

import (
	"context"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// HistoryDataResp 业务域历史数据响应结构体
type HistoryDataItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type HistoryDataResp struct {
	Items []*HistoryDataItem `json:"items"`
	Total int64              `json:"total"`
	Limit int64              `json:"limit"`
	Page  int64              `json:"page"`
}

// ListHistoryData 获取业务域历史数据
func (m *mgnt) ListHistoryData(ctx context.Context, page, limit int64) (HistoryDataResp, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	if limit > 1000 {
		limit = 1000
	}

	res := HistoryDataResp{
		Page:  page,
		Limit: limit,
	}

	filter := bson.M{}
	filter["$or"] = []bson.M{
		{"biz_domain_id": bson.M{"$exists": false}},
		{"biz_domain_id": nil},
		{"biz_domain_id": ""},
	}

	cnt, err := m.mongo.ListDagCountByFields(ctx, filter)
	if err != nil {
		log.Warnf("[logic.ListHistoryData] ListDagByFields err, detail: %s", err.Error())
		return res, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
	}

	res.Total = cnt

	if cnt == 0 {
		return res, nil
	}
	opt := options.FindOptions{}
	opt.Limit = &limit
	offset := limit * page
	opt.Skip = &offset
	opt.Sort = map[string]interface{}{"createdAt": 1}

	dags, err := m.mongo.ListDagByFields(ctx, filter, opt)
	if err != nil {
		log.Warnf("[logic.ListHistoryData] ListDagByFields err, detail: %s", err.Error())
		return res, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
	}

	for _, dag := range dags {
		res.Items = append(res.Items, &HistoryDataItem{
			ID:   dag.ID,
			Type: utils.IfNot(dag.Type == "", common.DagTypeDefault, dag.Type),
		})
	}

	return res, nil
}
