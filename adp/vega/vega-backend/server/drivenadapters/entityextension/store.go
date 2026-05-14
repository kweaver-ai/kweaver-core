// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package entityextension 读写 t_entity_extension（Issue #382 方案 B）
package entityextension

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"

	"vega-backend/common"
)

const tableName = "t_entity_extension"

// 与 t_entity_extension.f_entity_kind 一致，区分 catalog / resource，避免同字符串 id 扩展行主键冲突
const (
	KindCatalog  = "catalog"
	KindResource = "resource"
)

var (
	storeOnce sync.Once
	st        *Store
)

// Store 实体级 extensions 行存储
type Store struct {
	db *sql.DB
}

// NewStore 单例，与 catalog/resource access 共用同一 DB 连接池
func NewStore(appSetting *common.AppSetting) *Store {
	storeOnce.Do(func() {
		st = &Store{db: libdb.NewDB(&appSetting.DBSetting)}
	})
	return st
}

// Replace 整包替换某实体下的全部 KV（空 map 表示删除全部行）
func (s *Store) Replace(ctx context.Context, kind string, entityID string, kv map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := deleteByEntityIDTx(ctx, tx, kind, entityID); err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	for k, v := range kv {
		q, args, err := sq.Insert(tableName).
			Columns("f_entity_kind", "f_entity_id", "f_key", "f_value", "f_create_time", "f_update_time").
			Values(kind, entityID, k, v, now, now).
			ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func deleteByEntityIDTx(ctx context.Context, tx *sql.Tx, kind string, entityID string) error {
	q, args, err := sq.Delete(tableName).Where(sq.Eq{
		"f_entity_kind": kind,
		"f_entity_id":   entityID,
	}).ToSql()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, q, args...)
	return err
}

// DeleteByEntityIDs 删除多个实体下的全部扩展行（用于批量删 catalog/resource）
func (s *Store) DeleteByEntityIDs(ctx context.Context, kind string, entityIDs []string) error {
	if len(entityIDs) == 0 {
		return nil
	}
	q, args, err := sq.Delete(tableName).Where(sq.Eq{
		"f_entity_kind": kind,
		"f_entity_id":   entityIDs,
	}).ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, q, args...)
	return err
}

// GetByEntityID 读取单实体 KV，无行时返回空 map（非 nil）
func (s *Store) GetByEntityID(ctx context.Context, kind string, entityID string) (map[string]string, error) {
	q, args, err := sq.Select("f_key", "f_value").From(tableName).
		Where(sq.Eq{"f_entity_kind": kind, "f_entity_id": entityID}).
		OrderBy("f_key").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

// GetByEntityIDs 批量读取，返回 entityID -> kv
func (s *Store) GetByEntityIDs(ctx context.Context, kind string, entityIDs []string) (map[string]map[string]string, error) {
	res := make(map[string]map[string]string)
	if len(entityIDs) == 0 {
		return res, nil
	}
	q, args, err := sq.Select("f_entity_id", "f_key", "f_value").From(tableName).
		Where(sq.Eq{"f_entity_kind": kind, "f_entity_id": entityIDs}).
		OrderBy("f_entity_id", "f_key").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var eid, k, v string
		if err := rows.Scan(&eid, &k, &v); err != nil {
			return nil, err
		}
		if res[eid] == nil {
			res[eid] = make(map[string]string)
		}
		res[eid][k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// ApplyJoinsForCatalog 为 FROM t_catalog 的查询追加 INNER JOIN 以实现多对 AND 扩展筛选
func ApplyJoinsForCatalog(builder sq.SelectBuilder, keys, values []string) sq.SelectBuilder {
	for i := range keys {
		alias := fmt.Sprintf("vex%d", i)
		join := fmt.Sprintf(
			"t_entity_extension %s ON %s.f_entity_kind = ? AND %s.f_entity_id = t_catalog.f_id AND %s.f_key = ? AND %s.f_value = ?",
			alias, alias, alias, alias, alias,
		)
		builder = builder.Join(join, KindCatalog, keys[i], values[i])
	}
	return builder
}

// ApplyJoinsForResource 为 FROM t_resource 的查询追加扩展筛选
func ApplyJoinsForResource(builder sq.SelectBuilder, keys, values []string) sq.SelectBuilder {
	for i := range keys {
		alias := fmt.Sprintf("vex%d", i)
		join := fmt.Sprintf(
			"t_entity_extension %s ON %s.f_entity_kind = ? AND %s.f_entity_id = t_resource.f_id AND %s.f_key = ? AND %s.f_value = ?",
			alias, alias, alias, alias, alias,
		)
		builder = builder.Join(join, KindResource, keys[i], values[i])
	}
	return builder
}

// FilterKeys 若 keysCSV 非空，仅保留列出的 key（用于 include_extension_keys）
func FilterKeys(in map[string]string, keysCSV string) map[string]string {
	if keysCSV == "" || len(in) == 0 {
		return in
	}
	parts := strings.Split(keysCSV, ",")
	allow := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			allow[p] = struct{}{}
		}
	}
	if len(allow) == 0 {
		return in
	}
	out := make(map[string]string)
	for k, v := range in {
		if _, ok := allow[k]; ok {
			out[k] = v
		}
	}
	return out
}
