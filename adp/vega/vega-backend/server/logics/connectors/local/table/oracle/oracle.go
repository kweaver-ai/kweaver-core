// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package oracle provides Oracle database connector implementation.
package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"vega-backend/interfaces"
	"vega-backend/logics/connectors"

	"github.com/mitchellh/mapstructure"
	_ "github.com/sijms/go-ora/v2"
)

type oracleConfig struct {
	Host        string         `mapstructure:"host"`
	Port        int            `mapstructure:"port"`
	ServiceName string         `mapstructure:"service_name"`
	Username    string         `mapstructure:"username"`
	Password    string         `mapstructure:"password"`
	Schemas     []string       `mapstructure:"schemas"`
	Options     map[string]any `mapstructure:"options"`
}

var (
	SYSTEM_SCHEMAS = []string{
		"SYS",
		"SYSTEM",
		"SYSMAN",
		"DBSNMP",
		"OUTLN",
		"MDSYS",
		"ORDSYS",
		"EXFSYS",
		"CTXSYS",
		"XDB",
		"WMSYS",
		"APEX_PUBLIC_USER",
		"FLOWS_FILES",
		"APEX_040000",
		"ORDDATA",
		"OLAPSYS",
	}
	SYSTEM_SCHEMAS_MAP = map[string]bool{
		"SYS":              true,
		"SYSTEM":           true,
		"SYSMAN":           true,
		"DBSNMP":           true,
		"OUTLN":            true,
		"MDSYS":            true,
		"ORDSYS":           true,
		"EXFSYS":           true,
		"CTXSYS":           true,
		"XDB":              true,
		"WMSYS":            true,
		"APEX_PUBLIC_USER": true,
		"FLOWS_FILES":      true,
		"APEX_040000":      true,
		"ORDDATA":          true,
		"OLAPSYS":          true,
	}
)

const (
	// SCHEMA_NAME_MAX_LENGTH Oracle schema name maximum length
	SCHEMA_NAME_MAX_LENGTH = 128
	// PORT_MIN valid port minimum value
	PORT_MIN = 1
	// PORT_MAX valid port maximum value
	PORT_MAX = 65535
)

// OracleConnector implements TableConnector for Oracle.
type OracleConnector struct {
	enabled   bool
	config    *oracleConfig
	connected bool
	db        *sql.DB
}

func (c *OracleConnector) ListDatabases(ctx context.Context) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

// NewOracleConnector creates Oracle connector builder
func NewOracleConnector() connectors.TableConnector {
	return &OracleConnector{}
}

// GetType returns the data source type.
func (c *OracleConnector) GetType() string {
	return interfaces.ConnectorTypeOracle
}

// GetName returns the connector name.
func (c *OracleConnector) GetName() string {
	return interfaces.ConnectorTypeOracle
}

// GetMode returns the connector mode.
func (c *OracleConnector) GetMode() string {
	return interfaces.ConnectorModeLocal
}

// GetCategory returns the connector category.
func (c *OracleConnector) GetCategory() string {
	return interfaces.ConnectorCategoryTable
}

// GetEnabled returns the enabled status.
func (c *OracleConnector) GetEnabled() bool {
	return c.enabled
}

// SetEnabled sets the enabled status.
func (c *OracleConnector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// GetSensitiveFields returns the sensitive fields for Oracle connector.
func (c *OracleConnector) GetSensitiveFields() []string {
	return []string{"password"}
}

// GetFieldConfig returns the field configuration for Oracle connector.
func (c *OracleConnector) GetFieldConfig() map[string]interfaces.ConnectorFieldConfig {
	return map[string]interfaces.ConnectorFieldConfig{
		"host":         {Name: "主机地址", Type: "string", Description: "Oracle 服务器主机地址", Required: true, Encrypted: false},
		"port":         {Name: "端口号", Type: "integer", Description: "Oracle 服务器端口", Required: true, Encrypted: false},
		"service_name": {Name: "服务名", Type: "string", Description: "Oracle 服务名", Required: true, Encrypted: false},
		"username":     {Name: "用户名", Type: "string", Description: "数据库用户名", Required: true, Encrypted: false},
		"password":     {Name: "密码", Type: "string", Description: "数据库密码", Required: true, Encrypted: true},
		"schemas":      {Name: "模式列表", Type: "array", Description: "模式名称列表（可选，为空则连接实例级别）", Required: false, Encrypted: false},
		"options":      {Name: "连接参数", Type: "object", Description: "连接参数", Required: false, Encrypted: false},
	}
}

// New creates a new Oracle connector.
// Schemas is optional, if not specified, connects to instance level.
func (c *OracleConnector) New(cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	var oCfg oracleConfig
	if err := mapstructure.Decode(cfg, &oCfg); err != nil {
		return nil, fmt.Errorf("failed to decode oracle config: %w", err)
	}

	if oCfg.Host == "" || oCfg.Port == 0 || oCfg.Username == "" || oCfg.Password == "" || oCfg.ServiceName == "" {
		return nil, fmt.Errorf("oracle connector config is incomplete")
	}

	// Validate port range
	if oCfg.Port < PORT_MIN || oCfg.Port > PORT_MAX {
		return nil, fmt.Errorf("port %d is out of valid range (%d-%d)", oCfg.Port, PORT_MIN, PORT_MAX)
	}

	// Validate schema name length (Oracle schema name max 128 characters)
	for _, schema := range oCfg.Schemas {
		if len(schema) > SCHEMA_NAME_MAX_LENGTH {
			return nil, fmt.Errorf("schema name '%s' exceeds maximum length of %d characters", schema, SCHEMA_NAME_MAX_LENGTH)
		}
	}

	return &OracleConnector{
		config: &oCfg,
	}, nil
}

// Connect establishes connection to Oracle database.
// If Config.Schemas is empty, connects to instance level (without specifying schema).
func (c *OracleConnector) Connect(ctx context.Context) error {
	if c.connected {
		return nil
	}

	// Build connection string
	values := url.Values{}

	// Apply options
	for k, v := range c.config.Options {
		values.Set(k, fmt.Sprintf("%v", v))
	}
	//oracle://system:5208@0.0.0.0:1522/ORCL?
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		c.config.Username, c.config.Password, c.config.Host, c.config.Port, c.config.ServiceName)

	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}

	c.db = db
	c.connected = true

	return nil
}

// Close closes the database connection.
func (c *OracleConnector) Close(ctx context.Context) error {
	if c.db != nil {
		err := c.db.Close()
		c.connected = false
		c.db = nil
		return err
	}
	return nil
}

// Ping checks the database connection.
func (c *OracleConnector) Ping(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	return c.db.Ping()
}

// TestConnection tests the connection to Oracle database.
func (c *OracleConnector) TestConnection(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	// If schemas list is configured, verify these schemas exist
	if len(c.config.Schemas) > 0 {
		if err := c.validateSchemas(ctx); err != nil {
			return err
		}
	}

	return nil
}

// validateSchemas verifies that configured schemas exist
func (c *OracleConnector) validateSchemas(ctx context.Context) error {
	// Get all schemas list
	query := "SELECT USERNAME FROM ALL_USERS"
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()

	existingSchemas := make(map[string]bool)
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return fmt.Errorf("failed to scan schema name: %w", err)
		}
		existingSchemas[strings.ToUpper(schemaName)] = true
	}

	// Check if all configured schemas exist
	var notFoundSchemas []string
	for _, schema := range c.config.Schemas {
		if !existingSchemas[strings.ToUpper(schema)] {
			notFoundSchemas = append(notFoundSchemas, schema)
		}
	}

	if len(notFoundSchemas) > 0 {
		return fmt.Errorf("schemas not found: %v", notFoundSchemas)
	}

	return nil
}

// ListSchemas lists all accessible user schemas (excluding system schemas) under the instance.
func (c *OracleConnector) ListSchemas(ctx context.Context) ([]string, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	query := "SELECT USERNAME FROM ALL_USERS WHERE ORACLE_MAINTAINED = 'N' ORDER BY USERNAME"
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, fmt.Errorf("failed to scan schema name: %w", err)
		}
		if !SYSTEM_SCHEMAS_MAP[strings.ToUpper(schema)] {
			schemas = append(schemas, schema)
		}
	}
	return schemas, nil
}

// ListTables returns all tables in the database.
// If Config.Schemas is not empty, only lists tables in those schemas;
// If Config.Schemas is empty (instance-level connection), iterates through all user schemas,
// and the TableMeta.Schema field marks the owning schema.
func (c *OracleConnector) ListTables(ctx context.Context) ([]*interfaces.TableMeta, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	baseQuery := "SELECT OWNER,OBJECT_NAME AS TABLE_NAME,OBJECT_TYPE AS TABLE_TYPE,LAST_DDL_TIME AS LAST_ANALYZED FROM all_objects WHERE 1=1 "
	// Filter schemas

	var query string
	//var args []interface{}

	if len(c.config.Schemas) > 0 {
		placeholders := make([]string, len(c.config.Schemas))
		//args = make([]interface{}, len(c.config.Schemas))
		for i, schema := range c.config.Schemas {
			// 使用字符串格式化添加单引号
			placeholders[i] = fmt.Sprintf("'%s'", strings.ToUpper(schema))
		}
		query = fmt.Sprintf("%s AND OWNER IN (%s)", baseQuery, strings.Join(placeholders, ", "))

		//// Convert to uppercase for Oracle
		//schemas := make([]string, len(c.config.Schemas))
		//for i, s := range c.config.Schemas {
		//	schemas[i] = strings.ToUpper(s)
		//}
		//builder = builder.Where(sq.Eq{"OWNER": schemas})
	} else {
		// Exclude system schemas
		//builder = builder.Where(sq.NotEq{"OWNER": SYSTEM_SCHEMAS})
		// 如果没有指定schemas，排除系统schemas
		query = baseQuery
		//args = []interface{}{}
	}
	//query, args, err := builder.ToSql()
	//if err != nil {
	//	return nil, fmt.Errorf("failed to build list tables query: %w", err)
	//}
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []*interfaces.TableMeta
	for rows.Next() {
		var schema, name, tableType string
		//var tableRows sql.NullInt64
		//var description sql.NullString
		var lastAnalyzed sql.NullTime

		if err := rows.Scan(
			&schema,
			&name,
			&tableType,
			&lastAnalyzed,
		); err != nil {
			return nil, fmt.Errorf("failed to scan table info: %w", err)
		}

		meta := &interfaces.TableMeta{
			Name:        name,
			TableType:   "table",
			Description: "",
			Database:    schema,
		}
		// Populate Properties
		meta.Properties = make(map[string]any)
		if lastAnalyzed.Valid {
			meta.Properties["last_analyzed"] = lastAnalyzed.Time.UnixMilli()
		}

		tables = append(tables, meta)
	}
	return tables, nil
}

// GetTableMeta returns metadata for a specific table.
// table format: "table_name" or "schema.table_name"
func (c *OracleConnector) GetTableMeta(ctx context.Context, table *interfaces.TableMeta) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	// 1. Get table basic info (row count, comment, etc.)
	if err := c.fetchTableStatus(ctx, table); err != nil {
		return fmt.Errorf("failed to fetch table status: %w", err)
	}

	// 2. Get column info
	if err := c.fetchColumns(ctx, table); err != nil {
		return fmt.Errorf("failed to fetch columns: %w", err)
	}

	// 3. Get index info
	if err := c.fetchIndexes(ctx, table); err != nil {
		return fmt.Errorf("failed to fetch indexes: %w", err)
	}

	// 4. Get foreign key info
	if err := c.fetchForeignKeys(ctx, table); err != nil {
		return fmt.Errorf("failed to fetch foreign keys: %w", err)
	}

	return nil
}

// fetchTableStatus retrieves table status from ALL_TABLES and ALL_TAB_COMMENTS.
func (c *OracleConnector) fetchTableStatus(ctx context.Context, table *interfaces.TableMeta) error {
	query := `
		SELECT 
			T.NUM_ROWS,
			C.COMMENTS,
			T.LAST_ANALYZED
		FROM ALL_TABLES T
		LEFT JOIN ALL_TAB_COMMENTS C ON T.OWNER = C.OWNER AND T.TABLE_NAME = C.TABLE_NAME
		WHERE T.OWNER = :1 AND T.TABLE_NAME = :2
	`

	var tableRows sql.NullInt64
	var description sql.NullString
	var lastAnalyzed sql.NullTime

	row := c.db.QueryRowContext(ctx, query, strings.ToUpper(table.Database), strings.ToUpper(table.Name))
	if err := row.Scan(
		&tableRows,
		&description,
		&lastAnalyzed,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	// Initialize Properties map
	if table.Properties == nil {
		table.Properties = make(map[string]any)
	}

	if table.TableType == "" {
		table.TableType = "table"
	}

	table.Properties["row_count"] = tableRows.Int64
	table.Description = description.String

	if lastAnalyzed.Valid {
		table.Properties["last_analyzed"] = lastAnalyzed.Time.UnixMilli()
	}

	return nil
}

// fetchColumns retrieves column metadata from ALL_TAB_COLUMNS.
func (c *OracleConnector) fetchColumns(ctx context.Context, table *interfaces.TableMeta) error {
	query := `
		SELECT 
			C.COLUMN_NAME,
			C.DATA_TYPE,
			C.CHAR_LENGTH,
			C.DATA_PRECISION,
			C.DATA_SCALE,
			C.NULLABLE,
			C.DATA_DEFAULT,
			CC.COMMENTS,
			C.CHARACTER_SET_NAME,
			C.COLUMN_ID
		FROM ALL_TAB_COLUMNS C
		LEFT JOIN ALL_COL_COMMENTS CC ON C.OWNER = CC.OWNER AND C.TABLE_NAME = CC.TABLE_NAME AND C.COLUMN_NAME = CC.COLUMN_NAME
		WHERE C.OWNER = :1 AND C.TABLE_NAME = :2
		ORDER BY COLUMN_ID
	`

	rows, err := c.db.QueryContext(ctx, query, strings.ToUpper(table.Database), strings.ToUpper(table.Name))
	if err != nil {
		return err
	}
	defer rows.Close()

	var columns []interfaces.TableColumnMeta

	for rows.Next() {
		var name, dataType, isNullable, columnDefault, description, charset sql.NullString
		var charMaxLen, numPrecision, numScale, position sql.NullInt64

		if err := rows.Scan(
			&name,
			&dataType,
			&charMaxLen,
			&numPrecision,
			&numScale,
			&isNullable,
			&columnDefault,
			&description,
			&charset,
			&position,
		); err != nil {
			return err
		}

		col := interfaces.TableColumnMeta{
			Name:            name.String,
			Type:            MapType(dataType.String),
			OrigType:        dataType.String,
			Nullable:        isNullable.String == "Y",
			DefaultValue:    columnDefault.String,
			Description:     description.String,
			CharMaxLen:      int(charMaxLen.Int64),
			NumPrecision:    int(numPrecision.Int64),
			NumScale:        int(numScale.Int64),
			Charset:         charset.String,
			OrdinalPosition: int(position.Int64),
		}
		columns = append(columns, col)
	}

	table.Columns = columns
	return nil
}

// fetchIndexes retrieves index metadata from ALL_INDEXES and ALL_IND_COLUMNS.
func (c *OracleConnector) fetchIndexes(ctx context.Context, table *interfaces.TableMeta) error {
	query := `
		SELECT 
			I.INDEX_NAME,
			IC.COLUMN_NAME,
			I.UNIQUENESS,
			IC.COLUMN_POSITION
		FROM ALL_INDEXES I
		JOIN ALL_IND_COLUMNS IC ON I.INDEX_NAME = IC.INDEX_NAME AND I.OWNER = IC.INDEX_OWNER
		WHERE I.TABLE_OWNER = :1 AND I.TABLE_NAME = :2
		ORDER BY I.INDEX_NAME, IC.COLUMN_POSITION
	`

	rows, err := c.db.QueryContext(ctx, query, strings.ToUpper(table.Database), strings.ToUpper(table.Name))
	if err != nil {
		return err
	}
	defer rows.Close()

	indexMap := make(map[string]*interfaces.TableIndexMeta)

	for rows.Next() {
		var indexName, columnName, uniqueness sql.NullString
		var position sql.NullInt64

		if err := rows.Scan(
			&indexName,
			&columnName,
			&uniqueness,
			&position,
		); err != nil {
			return err
		}

		name := indexName.String
		if idx, ok := indexMap[name]; ok {
			idx.Columns = append(idx.Columns, columnName.String)
		} else {
			indexMap[name] = &interfaces.TableIndexMeta{
				Name:    name,
				Columns: []string{columnName.String},
				Unique:  uniqueness.String == "UNIQUE",
				Primary: name == strings.ToUpper(table.Name)+"_PK", // Oracle primary key naming convention
			}
		}
	}

	var indexes []interfaces.TableIndexMeta
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}
	table.Indices = indexes
	return nil
}

// fetchForeignKeys retrieves foreign key metadata from ALL_CONSTRAINTS and ALL_CONS_COLUMNS.
func (c *OracleConnector) fetchForeignKeys(ctx context.Context, table *interfaces.TableMeta) error {
	query := `
		SELECT 
			C.CONSTRAINT_NAME,
			CC.COLUMN_NAME,
			R.TABLE_NAME,
			RCC.COLUMN_NAME
		FROM ALL_CONSTRAINTS C
		JOIN ALL_CONS_COLUMNS CC ON C.CONSTRAINT_NAME = CC.CONSTRAINT_NAME AND C.OWNER = CC.OWNER
		JOIN ALL_CONSTRAINTS R ON C.R_CONSTRAINT_NAME = R.CONSTRAINT_NAME AND C.R_OWNER = R.OWNER
		JOIN ALL_CONS_COLUMNS RCC ON R.CONSTRAINT_NAME = RCC.CONSTRAINT_NAME AND R.OWNER = RCC.OWNER
			AND CC.POSITION = RCC.POSITION
		WHERE C.OWNER = :1 AND C.TABLE_NAME = :2 AND C.CONSTRAINT_TYPE = 'R'
		ORDER BY C.CONSTRAINT_NAME, CC.POSITION
	`

	rows, err := c.db.QueryContext(ctx, query, strings.ToUpper(table.Database), strings.ToUpper(table.Name))
	if err != nil {
		return err
	}
	defer rows.Close()

	fkMap := make(map[string]*interfaces.TableForeignKeyMeta)

	for rows.Next() {
		var constraintName, columnName, refTableName, refColumnName sql.NullString

		if err := rows.Scan(
			&constraintName,
			&columnName,
			&refTableName,
			&refColumnName,
		); err != nil {
			return err
		}

		name := constraintName.String
		if fk, ok := fkMap[name]; ok {
			fk.Columns = append(fk.Columns, columnName.String)
			fk.RefColumns = append(fk.RefColumns, refColumnName.String)
		} else {
			fkMap[name] = &interfaces.TableForeignKeyMeta{
				Name:       name,
				Columns:    []string{columnName.String},
				RefTable:   refTableName.String,
				RefColumns: []string{refColumnName.String},
			}
		}
	}

	var fks []interfaces.TableForeignKeyMeta
	for _, fk := range fkMap {
		fks = append(fks, *fk)
	}
	table.ForeignKeys = fks
	return nil
}

func (c *OracleConnector) ExecuteQuery(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error) {
	return nil, nil
}

// GetMetadata returns the metadata for the catalog.
func (c *OracleConnector) GetMetadata(ctx context.Context) (map[string]any, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	// Fetch critical database information
	// Includes basic info, character set, timezone, version, etc.
	query := `
		SELECT 
			NAME,
			VALUE
		FROM V$PARAMETER
		WHERE NAME IN (
			'db_name',
			'instance_name',
			'db_unique_name',
			'open_cursors',
			'processes',
			'sessions',
			'nls_language',
			'nls_territory',
			'nls_characterset',
			'nls_nchar_characterset',
			'time_zone',
			'db_block_size'
		)
	`

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metadata := make(map[string]any)
	for rows.Next() {
		var paramName, paramValue string
		if err := rows.Scan(&paramName, &paramValue); err == nil {
			metadata[paramName] = paramValue
		}
	}

	// Get version info
	versionQuery := "SELECT BANNER FROM V$VERSION WHERE BANNER LIKE 'Oracle%'"
	versionRow := c.db.QueryRowContext(ctx, versionQuery)
	var version string
	if err := versionRow.Scan(&version); err == nil {
		metadata["version"] = version
	}

	// Get cluster mode (RAC or standalone)
	// Check if instance count > 1 for RAC
	racQuery := "SELECT COUNT(*) FROM V$ACTIVE_INSTANCES"
	racRow := c.db.QueryRowContext(ctx, racQuery)
	var instanceCount int
	if err := racRow.Scan(&instanceCount); err == nil {
		if instanceCount > 1 {
			metadata["cluster_mode"] = "rac"
		} else {
			metadata["cluster_mode"] = "standalone"
		}
	} else {
		metadata["cluster_mode"] = "standalone"
	}

	return metadata, nil
}
