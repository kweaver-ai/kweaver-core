// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package postgresql 提供 PostgreSQL 表连接器：连接目标为单个 database；配置项 schemas 表示 schema 白名单。
package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
	"github.com/mitchellh/mapstructure"

	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
)

type postgresqlConfig struct {
	Host     string         `mapstructure:"host"`
	Port     int            `mapstructure:"port"`
	Username string         `mapstructure:"username"`
	Password string         `mapstructure:"password"`
	Database string         `mapstructure:"database"`
	Schemas  []string       `mapstructure:"schemas"`
	Options  map[string]any `mapstructure:"options"`
}

const (
	databaseNameMaxLength = 63 // PostgreSQL 标识符默认上限
	portMin               = 1
	portMax               = 65535
)

// PostgresqlConnector 实现 TableConnector（PostgreSQL）。
type PostgresqlConnector struct {
	enabled bool

	config *postgresqlConfig

	connected bool
	db        *sql.DB
}

// NewPostgresqlConnector 创建 PostgreSQL connector 构建器
func NewPostgresqlConnector() connectors.TableConnector {
	return &PostgresqlConnector{}
}

// GetType 返回数据源类型键（与 t_connector_type.f_type、factory 注册键一致）。
func (c *PostgresqlConnector) GetType() string {
	return "postgresql"
}

// GetName 返回连接器名称。
func (c *PostgresqlConnector) GetName() string {
	return "postgresql"
}

// GetMode 返回连接器模式。
func (c *PostgresqlConnector) GetMode() string {
	return interfaces.ConnectorModeLocal
}

// GetCategory 返回连接器分类。
func (c *PostgresqlConnector) GetCategory() string {
	return interfaces.ConnectorCategoryTable
}

// GetEnabled 是否启用。
func (c *PostgresqlConnector) GetEnabled() bool {
	return c.enabled
}

// SetEnabled 设置启用状态。
func (c *PostgresqlConnector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// GetSensitiveFields 敏感字段。
func (c *PostgresqlConnector) GetSensitiveFields() []string {
	return []string{"password"}
}

// GetFieldConfig 连接表单字段（须与迁移中 t_connector_type 的 JSON 完全一致）。
func (c *PostgresqlConnector) GetFieldConfig() map[string]interfaces.ConnectorFieldConfig {
	return map[string]interfaces.ConnectorFieldConfig{
		"host":     {Name: "主机地址", Type: "string", Description: "数据库服务器主机地址", Required: true, Encrypted: false},
		"port":     {Name: "端口号", Type: "integer", Description: "数据库服务器端口", Required: true, Encrypted: false},
		"username": {Name: "用户名", Type: "string", Description: "数据库用户名", Required: true, Encrypted: false},
		"password": {Name: "密码", Type: "string", Description: "数据库密码", Required: true, Encrypted: true},
		"database": {Name: "数据库名", Type: "string", Description: "PostgreSQL 连接目标 database", Required: true, Encrypted: false},
		"schemas":  {Name: "Schema 列表", Type: "array", Description: "可选；为空则扫描当前库下除系统 schema 外的用户 schema；非空则仅扫描列出的 schema", Required: false, Encrypted: false},
		"options":  {Name: "连接参数", Type: "object", Description: "连接参数（如 sslmode、connect_timeout 等）", Required: false, Encrypted: false},
	}
}

// New 根据配置创建连接器实例。
func (c *PostgresqlConnector) New(cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	var pCfg postgresqlConfig
	if err := mapstructure.Decode(cfg, &pCfg); err != nil {
		return nil, fmt.Errorf("failed to decode postgresql config: %w", err)
	}

	if pCfg.Host == "" || pCfg.Port == 0 || pCfg.Username == "" || pCfg.Password == "" || pCfg.Database == "" {
		return nil, fmt.Errorf("postgresql connector config is incomplete")
	}

	if pCfg.Port < portMin || pCfg.Port > portMax {
		return nil, fmt.Errorf("port %d is out of valid range (%d-%d)", pCfg.Port, portMin, portMax)
	}

	if len(pCfg.Database) > databaseNameMaxLength {
		return nil, fmt.Errorf("database name exceeds maximum length of %d characters", databaseNameMaxLength)
	}

	for _, s := range pCfg.Schemas {
		if len(s) > databaseNameMaxLength {
			return nil, fmt.Errorf("schema name '%s' exceeds maximum length of %d characters", s, databaseNameMaxLength)
		}
	}

	return &PostgresqlConnector{
		config: &pCfg,
	}, nil
}

func (c *PostgresqlConnector) buildConnString() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.config.Username, c.config.Password),
		Host:   fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Path:   "/" + strings.TrimPrefix(c.config.Database, "/"),
	}
	q := u.Query()
	if c.config.Options != nil {
		for k, v := range c.config.Options {
			q.Set(k, fmt.Sprintf("%v", v))
		}
	}
	if q.Get("sslmode") == "" {
		q.Set("sslmode", "disable")
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// Connect 建立连接。
func (c *PostgresqlConnector) Connect(ctx context.Context) error {
	if c.connected {
		return nil
	}

	db, err := sql.Open("postgres", c.buildConnString())
	if err != nil {
		return err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return err
	}

	c.db = db
	c.connected = true
	return nil
}

// Close 关闭连接。
func (c *PostgresqlConnector) Close(ctx context.Context) error {
	if c.db != nil {
		err := c.db.Close()
		c.connected = false
		c.db = nil
		return err
	}
	return nil
}

// Ping 检测连接。
func (c *PostgresqlConnector) Ping(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}
	return c.db.PingContext(ctx)
}

// TestConnection 测试连接；若配置了 schema 白名单则校验存在性。
func (c *PostgresqlConnector) TestConnection(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}
	if len(c.config.Schemas) > 0 {
		if err := c.validateSchemas(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *PostgresqlConnector) validateSchemas(ctx context.Context) error {
	for _, s := range c.config.Schemas {
		var exists bool
		err := c.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_namespace WHERE nspname = $1)`, s).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to validate schema %q: %w", s, err)
		}
		if !exists {
			return fmt.Errorf("schema not found: %s", s)
		}
	}
	return nil
}
