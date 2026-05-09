// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mariadb

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"

	"vega-backend-tests/at/catalog/helpers"
	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)

// MariaDBPayloadBuilder MariaDB catalog payload构建器
type MariaDBPayloadBuilder struct {
	config     setup.MariaDBConfig
	testConfig *setup.TestConfig
}

// NewMariaDBPayloadBuilder 创建MariaDB payload构建器
func NewMariaDBPayloadBuilder(config setup.MariaDBConfig) *MariaDBPayloadBuilder {
	return &MariaDBPayloadBuilder{config: config}
}

// SetTestConfig 设置测试配置（包含加密器）
func (b *MariaDBPayloadBuilder) SetTestConfig(tc *setup.TestConfig) {
	b.testConfig = tc
}

// encryptPassword 加密密码
func (b *MariaDBPayloadBuilder) encryptPassword(password string) string {
	if b.testConfig != nil {
		return b.testConfig.EncryptString(password)
	}
	return password
}

// GetEncryptedPassword 返回加密后的正确密码
func (b *MariaDBPayloadBuilder) GetEncryptedPassword() string {
	return b.encryptPassword(b.config.Password)
}

// GetConnectorType 返回connector类型
func (b *MariaDBPayloadBuilder) GetConnectorType() string {
	return "mariadb"
}

// BuildCreatePayload 构建基本的MariaDB catalog创建payload
func (b *MariaDBPayloadBuilder) BuildCreatePayload() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("test-mariadb-catalog"),
		"connector_type": "mariadb",
		"connector_config": map[string]any{
			"host":      b.config.Host,
			"port":      b.config.Port,
			"databases": []string{b.config.Database},
			"username":  b.config.Username,
			"password":  b.encryptPassword(b.config.Password),
		},
	}
}

// BuildFullCreatePayload 构建完整的MariaDB catalog创建payload（包含所有可选字段）
func (b *MariaDBPayloadBuilder) BuildFullCreatePayload() map[string]any {
	payload := b.BuildCreatePayload()
	payload["description"] = "完整的测试catalog，包含所有可选字段"
	payload["tags"] = []string{"test", "mariadb", "at", "full-fields"}

	// 添加MariaDB options
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["options"] = map[string]any{
		"charset":   "utf8mb4",
		"parseTime": "true",
		"loc":       "Local",
	}

	return payload
}

// BuildCreatePayloadWithOptions 构建包含options的MariaDB catalog payload
func (b *MariaDBPayloadBuilder) BuildCreatePayloadWithOptions(options map[string]any) map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["options"] = options
	return payload
}

// BuildCreatePayloadWithWrongCredentials 构建错误凭证的MariaDB catalog payload
func (b *MariaDBPayloadBuilder) BuildCreatePayloadWithWrongCredentials() map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["password"] = b.encryptPassword("wrong_password_123")
	return payload
}

// BuildCreatePayloadWithInvalidConfig 构建无效配置的MariaDB catalog payload（不存在的数据库）
func (b *MariaDBPayloadBuilder) BuildCreatePayloadWithInvalidConfig() map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["databases"] = []string{"nonexistent_db_" + fmt.Sprintf("%d", time.Now().UnixNano())}
	return payload
}

// SupportsTestConnection MariaDB支持连接测试
func (b *MariaDBPayloadBuilder) SupportsTestConnection() bool {
	return true
}

// GetRequiredConfigFields 返回MariaDB connector_config必需的字段
// database 为可选字段，不指定时为实例级连接
func (b *MariaDBPayloadBuilder) GetRequiredConfigFields() []string {
	return []string{"host", "port", "username", "password"}
}

// BuildCreatePayloadWithoutDatabase 构建不含database的MariaDB catalog payload（实例级连接）
func (b *MariaDBPayloadBuilder) BuildCreatePayloadWithoutDatabase() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("test-mariadb-instance-catalog"),
		"connector_type": "mariadb",
		"connector_config": map[string]any{
			"host":     b.config.Host,
			"port":     b.config.Port,
			"username": b.config.Username,
			"password": b.encryptPassword(b.config.Password),
		},
	}
}

// ========== MariaDB特定Payload生成函数 ==========

// BuildCreatePayloadWithInvalidPort 构建无效port的MariaDB payload
func (b *MariaDBPayloadBuilder) BuildCreatePayloadWithInvalidPort() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("invalid-port-catalog"),
		"connector_type": "mariadb",
		"connector_config": map[string]any{
			"host":      b.config.Host,
			"port":      "not_a_number",
			"databases": []string{b.config.Database},
			"username":  b.config.Username,
			"password":  b.encryptPassword(b.config.Password),
		},
	}
}

// BuildCreatePayloadWithNonExistentDB 构建不存在数据库的MariaDB payload
func (b *MariaDBPayloadBuilder) BuildCreatePayloadWithNonExistentDB() map[string]any {
	return b.BuildCreatePayloadWithInvalidConfig()
}

// GetConfig 返回MariaDB配置（供测试中直接使用）
func (b *MariaDBPayloadBuilder) GetConfig() setup.MariaDBConfig {
	return b.config
}

// ResetTestDatabase 重置测试数据库：清理at_db数据库并创建新的测试表
// tableSuffix: 时间戳参数，生成表名 at_tb_<timestamp>
// 返回新创建的表名列表和错误
func (b *MariaDBPayloadBuilder) ResetTestDatabase(tableSuffix string, tableSize int, recordSize int) ([]string, []any, error) {
	testDBName := "at_db"

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&loc=UTC&charset=utf8mb4&timeout=300s",
		b.config.Username,
		b.config.Password,
		b.config.Host,
		b.config.Port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to mysql: %w", err)
	}
	defer func() { _ = db.Close() }()

	// 清理并创建测试数据库
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to drop database: %w", err)
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", testDBName))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create database: %w", err)
	}

	// 插入10条测试数据
	records := []any{}
	for i := 1; i <= recordSize; i++ {
		record := []any{
			i, i + 100, i * 1000, i * 10000,

			i * 10, i * 20, i * 1, i * 2,

			i * 100, i * 200, float64(i) * 10.5, float64(i) * 1.5,

			float64(i) * 2.5, fmt.Sprintf("char_%d", i), fmt.Sprintf("varchar_%d", i), fmt.Sprintf("text_%d", i),

			fmt.Sprintf("mediumtext_%d", i), fmt.Sprintf("longtext_%d", i), fmt.Sprintf("2026-01-%02d", i), fmt.Sprintf("12:34:%02d", i),

			fmt.Sprintf("2026-01-%02d 12:34:56", i), fmt.Sprintf("2026-01-%02d 12:34:56.000000", i), fmt.Sprintf("2026-01-%02d 12:34:56", i), fmt.Sprintf("2026-01-%02d 12:34:56.000000", i),

			2026, fmt.Sprintf("binary_%d", i), fmt.Sprintf("varbinary_%d", i), fmt.Sprintf("blob_%d", i),

			fmt.Sprintf("longblob_%d", i), i, true, false,

			nil, fmt.Sprintf("not_null_%d", i), fmt.Sprintf("default_%d", i), fmt.Sprintf("comment_%d", i),

			"utf8mb4_unicode_ci",
		}
		records = append(records, record)
	}

	// 创建测试表，包含全部可能的字段类型
	testTableNames := []string{}
	for i := 0; i < tableSize; i++ {
		testTableName := fmt.Sprintf("at_tb_%s_%d", tableSuffix, i)
		testTableNames = append(testTableNames, testTableName)
		createTableSQL := fmt.Sprintf(`
		CREATE TABLE %s.%s (
			-- 数值类型
			c_int INT,
			c_int2 INT(11),
			c_bigint BIGINT,
			c_bigint2 BIGINT(20),
			c_smallint SMALLINT,
			c_smallint2 SMALLINT(6),
			c_tinyint TINYINT,
			c_tinyint2 TINYINT(4),
			c_mediumint MEDIUMINT,
			c_mediumint2 MEDIUMINT(8),
			c_decimal DECIMAL(10,2),
			c_float FLOAT,
			c_double DOUBLE,

			-- 字符串类型
			c_char CHAR(10),
			c_varchar VARCHAR(255) NULL,
			c_text TEXT,
			c_mediumtext MEDIUMTEXT,
			c_longtext LONGTEXT,

			-- 日期时间类型
			c_date DATE,
			c_time TIME,
			c_datetime DATETIME,
			c_datetime2 DATETIME(6),
			c_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			c_timestamp2 TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,
			c_year YEAR(4),

			-- 二进制类型
			c_binary BINARY(16),
			c_varbinary VARBINARY(255),
			c_blob BLOB,
			c_longblob LONGBLOB,

			-- 其他类型
			c_bit BIT(8),
			c_bool BOOL,
			c_boolean BOOLEAN,

			-- 其他
			c_null VARCHAR(20) NULL,
			c_not_null VARCHAR(20) NOT NULL,
			c_default VARCHAR(20) DEFAULT 'default_value',
			c_comment VARCHAR(20) COMMENT '这是注释',
			c_collate VARCHAR(20) COLLATE utf8mb4_unicode_ci,

			-- 索引和约束
			PRIMARY KEY (c_int),
			UNIQUE INDEX uk_c_int2 (c_int2),
			INDEX idx_c_bigint (c_bigint),
			INDEX idx_multi (c_smallint, c_smallint2)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`, testDBName, testTableName)

		_, err = db.Exec(createTableSQL)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create table: %w", err)
		}

		if recordSize == 0 {
			continue
		}

		insertBuilder := sq.Insert(fmt.Sprintf("%s.%s", testDBName, testTableName)).Columns(
			"c_int", "c_int2", "c_bigint", "c_bigint2",
			"c_smallint", "c_smallint2", "c_tinyint", "c_tinyint2",
			"c_mediumint", "c_mediumint2", "c_decimal", "c_float",
			"c_double", "c_char", "c_varchar", "c_text",
			"c_mediumtext", "c_longtext", "c_date", "c_time",
			"c_datetime", "c_datetime2", "c_timestamp", "c_timestamp2",
			"c_year", "c_binary", "c_varbinary", "c_blob",
			"c_longblob", "c_bit", "c_bool", "c_boolean",
			"c_null", "c_not_null", "c_default", "c_comment",
			"c_collate",
		)

		for _, record := range records {
			insertBuilder = insertBuilder.Values(record.([]any)...)
		}

		sqlStr, args, _ := insertBuilder.ToSql()
		_, err = db.Exec(sqlStr, args...)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to insert test data: %w", err)
		}
	}

	return testTableNames, records, nil
}

func (b *MariaDBPayloadBuilder) RunDiscoverTask(client *testutil.HTTPClient, catalogID string) error {
	discoverResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/discover", nil)
	if discoverResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to run discover task: %d", discoverResp.StatusCode)
	}

	taskID, ok := discoverResp.Body["id"].(string)
	if !ok || taskID == "" {
		return fmt.Errorf("failed to get task ID from discover response")
	}

	maxAttempts := 60
	for attempt := 0; attempt < maxAttempts; attempt++ {
		taskResp := client.GET("/api/vega-backend/v1/discover-tasks/" + taskID)
		if taskResp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get discover task status: %d", taskResp.StatusCode)
		}
		if status, ok := taskResp.Body["status"].(string); ok {
			if status == "completed" || status == "success" {
				break
			} else if status == "failed" || status == "error" {
				return fmt.Errorf("discover task failed: %s", status)
			}
		}
		time.Sleep(3 * time.Second)
	}
	return nil
}
