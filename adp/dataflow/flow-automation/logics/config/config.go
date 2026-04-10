package config

import (
	"context"
	"strings"
	"sync"

	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
)

type ConfigHandler interface {
	ListConfigs(ctx context.Context, keys string) ([]rds.ConfModel, error)
	UpdateConfig(ctx context.Context, datas *ConfigReq) error
}

type ConfigImpl struct {
	ConfigModel rds.ConfDao
}

var oc sync.Once
var c ConfigHandler

func NewConfig() ConfigHandler {
	oc.Do(func() {
		c = &ConfigImpl{
			ConfigModel: rds.GetConfDao(),
		}
	})
	return c
}

type ConfigReq struct {
	ConfigItems []*ConfigItem `json:"configs"`
}

type ConfigItem struct {
	Keys   string `json:"key"`
	Values string `json:"value"`
}

// ListConfigs 获取配置列表
func (c *ConfigImpl) ListConfigs(ctx context.Context, keys string) ([]rds.ConfModel, error) {
	var err error
	var res []rds.ConfModel
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	opt := &rds.Options{}
	if keys != "" {
		opt.SearchOptions = append(opt.SearchOptions, &rds.SearchOption{
			Col:       "f_key",
			Val:       strings.Split(strings.TrimSpace(keys), ","),
			Condition: "IN",
		})
	}

	res, err = c.ConfigModel.ListConfigs(ctx, opt)
	if err != nil {
		log.Warnf("[ListConfigs] ListConfigs failed, detail: %s", err.Error())
		return res, err
	}

	if len(res) == 0 {
		res = make([]rds.ConfModel, 0)
	}

	return res, nil
}

// UpdateConfig 更新配置
func (c *ConfigImpl) UpdateConfig(ctx context.Context, datas *ConfigReq) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var configs []*rds.ConfModel
	for _, data := range datas.ConfigItems {
		configs = append(configs, &rds.ConfModel{
			Key:   &data.Keys,
			Value: &data.Values,
		})
	}

	err = c.ConfigModel.BatchUpdateConfig(ctx, configs)
	if err != nil {
		log.Warnf("[UpdateConfig] BatchUpdateConfig failed, detail: %s", err.Error())
		return err
	}

	return nil
}
