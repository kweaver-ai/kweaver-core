// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package factory provides connector factory for creating data source connectors.
package factory

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/common"
	connectorTypeAccess "vega-backend/drivenadapters/connector_type"
	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/remote"
)

var (
	factoryOnce sync.Once
	factory     *ConnectorFactory
)

// ConnectorFactory 创建和管理数据源连接器
// 支持 local 和 remote 两种模式:
// - local: 内置在 vega-backend 进程内运行的连接器
// - remote: 作为独立服务运行，通过 HTTP 调用的连接器
type ConnectorFactory struct {
	mu sync.RWMutex

	appSetting *common.AppSetting
	cta        interfaces.ConnectorTypeAccess // 数据库访问层

	connectors map[string]connectors.Connector // 内置 connector 构建器
}

// GetFactory returns the singleton connector factory instance.
func GetFactory() *ConnectorFactory {
	return factory
}

// Init 初始化 connector factory
func Init(appSetting *common.AppSetting) *ConnectorFactory {
	factoryOnce.Do(func() {
		factory = &ConnectorFactory{
			appSetting: appSetting,
			cta:        connectorTypeAccess.NewConnectorTypeAccess(appSetting),
			connectors: make(map[string]connectors.Connector, 0),
		}

		factory.InitLocalConnectors()
		factory.RegisterAllConnectors(appSetting)
	})
	return factory
}

// RegisterAllConnectors 注册所有 connector 构建器
func (cf *ConnectorFactory) RegisterAllConnectors(appSetting *common.AppSetting) {
	cts, _, err := cf.cta.List(context.Background(), interfaces.ConnectorTypesQueryParams{
		PaginationQueryParams: interfaces.PaginationQueryParams{
			Limit: -1,
		},
	})
	if err != nil {
		panic(fmt.Errorf("failed to get all connector types: %w", err))
	}

	ctx := context.Background()
	for _, ct := range cts {
		err = cf.RegisterConnector(ctx, ct.Type, ct)
		if err != nil {
			panic(fmt.Errorf("failed to register connector type %s:%s: %w", ct.Type, ct.Name, err))
		}
	}
}

// RegisterConnector 注册 connector 构建器
func (cf *ConnectorFactory) RegisterConnector(ctx context.Context, tp string, ct *interfaces.ConnectorType) error {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	connector, exist := cf.connectors[tp]
	if exist {
		if ct.Mode == interfaces.ConnectorModeLocal {
			// 验证 FieldConfig 一致性（代码是单一真相来源）
			codeFieldConfig := connector.GetFieldConfig()
			if !reflect.DeepEqual(codeFieldConfig, ct.FieldConfig) {
				logger.Fatalf("FieldConfig mismatch for connector type %s:\n  Code: %+v\n  DB:   %+v\nPlease update database migration to match code definition.",
					tp, codeFieldConfig, ct.FieldConfig)
			}
			connector.SetEnabled(ct.Enabled)
		} else {
			connector := remote.NewRemoteConnector(ct)
			cf.connectors[tp] = connector
		}
	} else {
		if ct.Mode == interfaces.ConnectorModeLocal {
			logger.Errorf("local connector %s:%s not implemented", tp, ct.Name)
			return fmt.Errorf("local connector %s:%s not implemented", tp, ct.Name)
		} else {
			connector := remote.NewRemoteConnector(ct)
			cf.connectors[tp] = connector
		}
	}
	return nil
}

// DeleteConnector 删除 connector 构建器
func (cf *ConnectorFactory) DeleteConnector(ctx context.Context, tp string) error {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	connector, exist := cf.connectors[tp]
	if exist {
		if connector.GetMode() == interfaces.ConnectorModeLocal {
			logger.Errorf("can not delete local connector %s:%s", tp, connector.GetName())
			return fmt.Errorf("can not delete local connector %s:%s", tp, connector.GetName())
		} else {
			delete(cf.connectors, tp)
		}
	} else {
		logger.Errorf("connector %s not implemented", tp)
		return fmt.Errorf("connector %s not implemented", tp)
	}
	return nil
}

func (cf *ConnectorFactory) SetConnectorEnabled(ctx context.Context, tp string, enabled bool) error {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	connector, exist := cf.connectors[tp]
	if exist {
		connector.SetEnabled(enabled)
	} else {
		logger.Errorf("connector %s not implemented", tp)
		return fmt.Errorf("connector %s not implemented", tp)
	}
	return nil
}

// CreateConnector 根据类型名称创建 connector 实例
func (cf *ConnectorFactory) CreateConnectorInstance(ctx context.Context, tp string, cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	if connector, ok := cf.connectors[tp]; ok {
		if !connector.GetEnabled() {
			return nil, fmt.Errorf("connector %s is disabled", tp)
		}

		cntor, err := connector.New(cfg)
		if err != nil {
			return nil, err
		}
		return cntor, nil
	}
	return nil, fmt.Errorf("connector %s not found", tp)
}

// GetSensitiveFields 根据 connector 类型返回敏感字段列表
func (cf *ConnectorFactory) GetSensitiveFields(tp string) []string {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	if connector, ok := cf.connectors[tp]; ok {
		return connector.GetSensitiveFields()
	}
	return nil
}
