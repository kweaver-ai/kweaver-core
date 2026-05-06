// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package mariadb provides MariaDB database connector implementation.
package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"

	"vega-backend/interfaces"
)

// ListDatabases 列出实例下所有可访问的用户数据库（排除系统库）。
func (c *MariaDBConnector) ListDatabases(ctx context.Context) ([]string, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	rows, err := c.db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var databases []string
	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %w", err)
		}
		if !SYSTEM_DBS_MAP[db] {
			databases = append(databases, db)
		}
	}
	return databases, nil
}

// ListTables 返回数据库中的所有表。
// 如果 Config.Database 非空，只列出该数据库的表；
// 如果 Config.Database 为空（实例级连接），遍历所有用户数据库，返回的 TableMeta.Database 字段标记所属库。
func (c *MariaDBConnector) ListTables(ctx context.Context) ([]*interfaces.TableMeta, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	builder := sq.Select(
		"TABLE_SCHEMA",
		"TABLE_NAME",
		"TABLE_TYPE",
		"ENGINE",
		"TABLE_COLLATION",
		"TABLE_ROWS",
		"TABLE_COMMENT",
		"CREATE_TIME",
		"UPDATE_TIME",
		"DATA_LENGTH",
		"INDEX_LENGTH",
	).From("information_schema.TABLES")

	// Filter databases
	if len(c.config.Databases) > 0 {
		builder = builder.Where(sq.Eq{"TABLE_SCHEMA": c.config.Databases})
	} else {
		builder = builder.Where(sq.NotEq{"TABLE_SCHEMA": SYSTEM_DBS})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build list tables query: %w", err)
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tables []*interfaces.TableMeta
	for rows.Next() {
		var schema, name, tableType string
		var engine, collation, description sql.NullString
		var tableRows, dataLength, indexLength sql.NullInt64
		var createTime, updateTime sql.NullTime

		if err := rows.Scan(
			&schema,
			&name,
			&tableType,
			&engine,
			&collation,
			&tableRows,
			&description,
			&createTime,
			&updateTime,
			&dataLength,
			&indexLength,
		); err != nil {
			return nil, fmt.Errorf("failed to scan table info: %w", err)
		}

		tableType = strings.ToLower(tableType)
		if tableType != "view" {
			tableType = "table"
		}

		meta := &interfaces.TableMeta{
			Name:        name,
			TableType:   tableType,
			Description: description.String,
			Database:    schema,
		}

		// Populate Properties
		meta.Properties = make(map[string]any)
		meta.Properties["engine"] = engine.String
		meta.Properties["collation"] = collation.String
		meta.Properties["row_count"] = tableRows.Int64
		meta.Properties["data_length"] = dataLength.Int64
		meta.Properties["index_length"] = indexLength.Int64

		if createTime.Valid {
			meta.Properties["create_time"] = createTime.Time.UnixMilli()
		}
		if updateTime.Valid {
			meta.Properties["update_time"] = updateTime.Time.UnixMilli()
		}

		// Infer Charset from Collation
		if coll := collation.String; coll != "" {
			for i, ch := range coll {
				if ch == '_' {
					meta.Properties["charset"] = coll[:i]
					break
				}
			}
		}

		tables = append(tables, meta)
	}

	return tables, nil
}

// GetTableMeta returns metadata for a specific table.
// table 格式: "table_name" 或 "database.table_name"
func (c *MariaDBConnector) GetTableMeta(ctx context.Context, table *interfaces.TableMeta) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	// 1. 获取表基本信息（引擎、字符集、行数、注释）
	if err := c.fetchTableStatus(ctx, table); err != nil {
		return fmt.Errorf("failed to fetch table status: %w", err)
	}

	// 2. 获取字段信息
	if err := c.fetchColumns(ctx, table); err != nil {
		return fmt.Errorf("failed to fetch columns: %w", err)
	}

	// 3. 获取索引信息
	if err := c.fetchIndexes(ctx, table); err != nil {
		return fmt.Errorf("failed to fetch indexes: %w", err)
	}

	// 4. 获取外键信息
	if err := c.fetchForeignKeys(ctx, table); err != nil {
		return fmt.Errorf("failed to fetch foreign keys: %w", err)
	}

	return nil
}

// fetchTableStatus retrieves table status from information_schema.TABLES.
func (c *MariaDBConnector) fetchTableStatus(ctx context.Context, table *interfaces.TableMeta) error {
	query, args, err := sq.Select(
		"TABLE_TYPE",
		"AUTO_INCREMENT",
		"ENGINE",
		"TABLE_COLLATION",
		"TABLE_ROWS",
		"TABLE_COMMENT",
		"CREATE_TIME",
		"UPDATE_TIME",
		"DATA_LENGTH",
		"INDEX_LENGTH",
	).From("information_schema.TABLES").
		Where(sq.Eq{"TABLE_SCHEMA": table.Database}).
		Where(sq.Eq{"TABLE_NAME": table.Name}).
		ToSql()
	if err != nil {
		return err
	}

	var tableType, engine, collation, description sql.NullString
	var autoIncrement, tableRows, dataLength, indexLength sql.NullInt64
	var createTime, updateTime sql.NullTime

	row := c.db.QueryRowContext(ctx, query, args...)
	if err := row.Scan(
		&tableType,
		&autoIncrement,
		&engine,
		&collation,
		&tableRows,
		&description,
		&createTime,
		&updateTime,
		&dataLength,
		&indexLength,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	table.TableType = strings.ToLower(tableType.String)
	if table.TableType != "view" {
		table.TableType = "table"
	}

	// 初始化 Properties map
	if table.Properties == nil {
		table.Properties = make(map[string]any)
	}

	table.Properties["engine"] = engine.String
	table.Properties["collation"] = collation.String
	table.Properties["row_count"] = tableRows.Int64
	table.Properties["data_length"] = dataLength.Int64
	table.Properties["index_length"] = indexLength.Int64
	if autoIncrement.Valid {
		table.Properties["auto_increment"] = autoIncrement.Int64
	}
	table.Description = description.String

	if createTime.Valid {
		table.Properties["create_time"] = createTime.Time.UnixMilli()
	}
	if updateTime.Valid {
		table.Properties["update_time"] = updateTime.Time.UnixMilli()
	}

	// 从 Collation 推断 Charset
	if coll := collation.String; coll != "" {
		for i, ch := range coll {
			if ch == '_' {
				table.Properties["charset"] = coll[:i]
				break
			}
		}
	}
	return nil
}

// fetchColumns retrieves column metadata from information_schema.COLUMNS.
func (c *MariaDBConnector) fetchColumns(ctx context.Context, table *interfaces.TableMeta) error {
	query, args, err := sq.Select(
		"COLUMN_NAME",
		"DATA_TYPE",
		"COLUMN_TYPE",
		"IS_NULLABLE",
		"COLUMN_DEFAULT",
		"COLUMN_COMMENT",
		"CHARACTER_MAXIMUM_LENGTH",
		"NUMERIC_PRECISION",
		"NUMERIC_SCALE",
		"DATETIME_PRECISION",
		"CHARACTER_SET_NAME",
		"COLLATION_NAME",
		"ORDINAL_POSITION",
		"COLUMN_KEY",
	).From("information_schema.COLUMNS").
		Where(sq.Eq{"TABLE_SCHEMA": table.Database}).
		Where(sq.Eq{"TABLE_NAME": table.Name}).
		OrderBy("ORDINAL_POSITION").
		ToSql()
	if err != nil {
		return err
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	var columns []interfaces.TableColumnMeta
	var pkColumns []string

	for rows.Next() {
		var name, columnType, dataType, isNullable, columnKey sql.NullString
		var columnDefault, description, charset, collation sql.NullString
		var position, charMaxLen, numPrecision, numScale, datetimePrecision sql.NullInt64

		if err := rows.Scan(
			&name,
			&dataType,
			&columnType,
			&isNullable,
			&columnDefault,
			&description,
			&charMaxLen,
			&numPrecision,
			&numScale,
			&datetimePrecision,
			&charset,
			&collation,
			&position,
			&columnKey,
		); err != nil {
			return err
		}

		col := interfaces.TableColumnMeta{
			Name:              name.String,
			Type:              MapType(columnType.String), // 使用 COLUMN_TYPE 以正确识别 unsigned（如 "int unsigned"）
			OrigType:          columnType.String,
			Nullable:          isNullable.String == "YES",
			DefaultValue:      columnDefault.String,
			Description:       description.String,
			CharMaxLen:        int(charMaxLen.Int64),
			NumPrecision:      int(numPrecision.Int64),
			NumScale:          int(numScale.Int64),
			DatetimePrecision: int(datetimePrecision.Int64),
			Charset:           charset.String,
			Collation:         collation.String,
			OrdinalPosition:   int(position.Int64),
			ColumnKey:         columnKey.String,
		}
		columns = append(columns, col)

		// 检查是否为主键
		if columnKey.String == "PRI" {
			pkColumns = append(pkColumns, col.Name)
		}
	}

	table.Columns = columns
	table.PKs = pkColumns
	return nil
}

// fetchIndexes retrieves index metadata from information_schema.STATISTICS.
func (c *MariaDBConnector) fetchIndexes(ctx context.Context, table *interfaces.TableMeta) error {
	query, args, err := sq.Select(
		"INDEX_NAME",
		"COLUMN_NAME",
		"NON_UNIQUE",
		"SEQ_IN_INDEX",
	).From("information_schema.STATISTICS").
		Where(sq.Eq{"TABLE_SCHEMA": table.Database}).
		Where(sq.Eq{"TABLE_NAME": table.Name}).
		OrderBy("INDEX_NAME", "SEQ_IN_INDEX").
		ToSql()
	if err != nil {
		return err
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	indexMap := make(map[string]*interfaces.TableIndexMeta)

	for rows.Next() {
		var indexName, columnName sql.NullString
		var nonUnique, seqInIndex sql.NullInt64

		if err := rows.Scan(
			&indexName,
			&columnName,
			&nonUnique,
			&seqInIndex,
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
				Unique:  nonUnique.Int64 == 0,
				Primary: name == "PRIMARY",
			}
		}
	}

	var indices []interfaces.TableIndexMeta
	for _, idx := range indexMap {
		indices = append(indices, *idx)
	}
	table.Indices = indices
	return nil
}

// fetchForeignKeys retrieves foreign key metadata from information_schema.KEY_COLUMN_USAGE.
func (c *MariaDBConnector) fetchForeignKeys(ctx context.Context, table *interfaces.TableMeta) error {
	query, args, err := sq.Select(
		"CONSTRAINT_NAME",
		"COLUMN_NAME",
		"REFERENCED_TABLE_NAME",
		"REFERENCED_COLUMN_NAME",
	).From("information_schema.KEY_COLUMN_USAGE").
		Where(sq.Eq{"TABLE_SCHEMA": table.Database}).
		Where(sq.Eq{"TABLE_NAME": table.Name}).
		Where(sq.NotEq{"REFERENCED_TABLE_NAME": nil}).
		OrderBy("CONSTRAINT_NAME", "ORDINAL_POSITION").
		ToSql()
	if err != nil {
		return err
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

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

	// Note: Handling OnDelete/OnUpdate requires joining with REFERENTIAL_CONSTRAINTS, skipping for simplicity unless requested.

	var fks []interfaces.TableForeignKeyMeta
	for _, fk := range fkMap {
		fks = append(fks, *fk)
	}
	table.ForeignKeys = fks
	return nil
}

// GetMetadata returns the metadata for the catalog.
func (c *MariaDBConnector) GetMetadata(ctx context.Context) (map[string]any, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	// 2. Fetch critical global variables
	// 包含基础信息、字符集、时区、大小写敏感、SQL模式以及集群相关信息
	targetVars := []string{
		"version",
		"version_comment",
		"version_compile_os",
		"character_set_server",
		"collation_server",
		"time_zone",
		"system_time_zone",
		"lower_case_table_names",
		"sql_mode",
		// Cluster related
		"wsrep_on",                     // Galera Cluster
		"group_replication_group_name", // Group Replication / InnoDB Cluster
		"read_only",
		"super_read_only",
	}

	// Construct placeholders
	placeholders := make([]string, len(targetVars))
	args := make([]any, len(targetVars))
	for i, v := range targetVars {
		placeholders[i] = "?"
		args[i] = v
	}

	query := fmt.Sprintf("SHOW GLOBAL VARIABLES WHERE Variable_name IN (%s)", strings.Join(placeholders, ","))
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		// Just log error and return partial metadata if SHOW VARIABLES fails (unlikely)
		// But for now, we return error to be safe
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	metadata := make(map[string]any)
	for rows.Next() {
		var varName, varValue string
		if err := rows.Scan(&varName, &varValue); err == nil {
			metadata[varName] = varValue
		}
	}

	// 3. Infer Cluster Mode
	metadata["cluster_mode"] = "standalone" // Default
	if val, ok := metadata["wsrep_on"]; ok && strings.EqualFold(fmt.Sprint(val), "ON") {
		metadata["cluster_mode"] = "galera"
	} else if val, ok := metadata["group_replication_group_name"]; ok && fmt.Sprint(val) != "" {
		metadata["cluster_mode"] = "group_replication"
	}

	return metadata, nil
}
